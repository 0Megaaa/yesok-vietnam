package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/workflow"
)

// loadUserFromContext extracts the user from context (set by middleware) or
// loads it from DB using the UID embedded in the JWT.
func loadUserFromContext(c *gin.Context, db *gorm.DB) (*models.User, error) {
	// Fast path: already set by middleware DB lookup.
	if u, exists := c.Get("user"); exists {
		if user, ok := u.(*models.User); ok {
			return user, nil
		}
	}

	// Load from DB using UID from JWT claims.
	uid, exists := c.Get("uid")
	if !exists {
		return nil, fmt.Errorf("uid not in context")
	}
	uidVal, ok := uid.(uint)
	if !ok || uidVal == 0 {
		return nil, fmt.Errorf("invalid uid")
	}

	var user models.User
	if err := db.First(&user, uidVal).Error; err != nil {
		return nil, err
	}
	c.Set("user", &user)
	return &user, nil
}

// ─── GET /api/state ───────────────────────────────────────────────────────────

// GetState returns the full application state.
//   - authenticated → user-specific or admin-aggregated data (Me, Users, Orders)
//   - anonymous     → public大盘: all orders with Me=nil, Users=nil (HTTP 200, never 401)
func GetState(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, _ := loadUserFromContext(c, db)
		var resp GetStateResponse

		if currentUser != nil && currentUser.Role == models.RoleAdmin {
			// admin sees everything
			var users []models.User
			if err := db.Order("id asc").Find(&users).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
				return
			}
			resp.Users = users

			var orders []models.Order
			if err := db.Order("created_at desc").Find(&orders).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
				return
			}
			resp.Orders = orders
			resp.Me = currentUser
		} else if currentUser != nil {
			// regular user sees only their own data
			resp.Me = currentUser

			var orders []models.Order
			if err := db.Where("user_id = ?", currentUser.ID).
				Order("created_at desc").
				Find(&orders).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
				return
			}
			resp.Orders = orders
		} else {
			// anonymous / 小程序冷启动: return public大盘, no 401
			var orders []models.Order
			if err := db.Order("created_at desc").Find(&orders).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
				return
			}
			resp.Orders = orders
			// Me=nil, Users=nil signals "not logged in" to the client
		}

		c.JSON(http.StatusOK, resp)
	}
}

// ─── PUT /api/state ───────────────────────────────────────────────────────────

// UpdateState handles order creation and status updates, persisting them to MySQL.
// It accepts a JSON body describing one or more order mutations:
//   - Create: { "action": "create", "order": { ...fields } }
//   - Update: { "action": "update", "id": 123, "fields": { ...patch } }
func UpdateState(db *gorm.DB, orderEngine *workflow.OrderEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, err := loadUserFromContext(c, db)
		if err != nil || currentUser == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var batch []StateMutation
		if err := c.ShouldBindJSON(&batch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		results := make([]MutationResult, 0, len(batch))
		for i, mut := range batch {
			var res MutationResult
			switch mut.Action {
			case "create":
				res = handleCreateOrder(db, currentUser, mut.Order)
			case "update":
				res = handleUpdateOrder(db, orderEngine, currentUser, mut.ID, mut.Fields)
			default:
				res = MutationResult{Success: false, Message: fmt.Sprintf("unknown action: %s", mut.Action)}
			}
			res.Index = i
			if !res.Success {
				res.Message = fmt.Sprintf("[%d] %s", i, res.Message)
			}
			results = append(results, res)
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
		})
	}
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func handleCreateOrder(db *gorm.DB, user *models.User, raw json.RawMessage) MutationResult {
	var order models.Order
	if raw != nil {
		if err := json.Unmarshal(raw, &order); err != nil {
			return MutationResult{Success: false, Message: "malformed order data: " + err.Error()}
		}
	}

	// Auto-generate order number if not provided.
	if order.OrderNo == "" {
		order.OrderNo = generateOrderNo()
	}

	// Enforce ownership.
	order.UserID = user.ID

	// Set defaults.
	if order.Currency == "" {
		order.Currency = "VND"
	}

	// 查询服务对应的首个流程节点作为初始 stage
	if order.CurrentStage == "" {
		if order.ServiceID > 0 {
			var firstNode models.SysWorkflowNode
			if err := db.Where("service_id = ?", order.ServiceID).Order("sort_order asc, id asc").First(&firstNode).Error; err == nil {
				order.CurrentStage = firstNode.StageCode
				order.MacroStatus = firstNode.MacroStatus
			}
		}
		// 兜底：仍未设置则使用默认值
		if order.CurrentStage == "" {
			order.CurrentStage = "start"
			if order.MacroStatus == "" {
				order.MacroStatus = string(models.OrderStatusPending)
			}
		}
	}

	if err := db.Create(&order).Error; err != nil {
		return MutationResult{Success: false, Message: "create failed: " + err.Error()}
	}

	return MutationResult{
		Success: true,
		Message: "created",
		OrderID: order.ID,
		OrderNo: order.OrderNo,
	}
}

func handleUpdateOrder(db *gorm.DB, orderEngine *workflow.OrderEngine, user *models.User, orderID uint, raw json.RawMessage) MutationResult {
	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return MutationResult{Success: false, Message: "order not found"}
		}
		return MutationResult{Success: false, Message: "query failed: " + err.Error()}
	}

	// Non-admin can only modify their own orders.
	if user.Role != models.RoleAdmin && order.UserID != user.ID {
		return MutationResult{Success: false, Message: "forbidden"}
	}

	if raw == nil {
		return MutationResult{Success: true, Message: "updated", OrderID: order.ID, OrderNo: order.OrderNo}
	}

	var patch map[string]interface{}
	if err := json.Unmarshal(raw, &patch); err != nil {
		return MutationResult{Success: false, Message: "malformed patch: " + err.Error()}
	}

	// 摘出只读字段和状态字段
	for _, k := range []string{"id", "user_id", "order_no", "created_at", "updated_at", "current_stage", "macro_status", "payment_status"} {
		delete(patch, k)
	}

	// 非状态字段直接更新
	if len(patch) > 0 {
		if err := db.Model(&order).Updates(patch).Error; err != nil {
			return MutationResult{Success: false, Message: "update fields failed: " + err.Error()}
		}
	}

	return MutationResult{
		Success: true,
		Message: "updated",
		OrderID: order.ID,
		OrderNo: order.OrderNo,
	}
}

func generateOrderNo() string {
	return "ORD" + strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// ─── Request/Response types ──────────────────────────────────────────────────

type GetStateResponse struct {
	Users  []models.User  `json:"users,omitempty"`
	Orders []models.Order `json:"orders"`
	Me     *models.User   `json:"me"`
}

// StateMutation describes a single atomic operation in a batch update request.
type StateMutation struct {
	Action string          `json:"action"` // "create" | "update"
	ID     uint            `json:"id,omitempty"`
	Order  json.RawMessage `json:"order,omitempty"`
	Fields json.RawMessage `json:"fields,omitempty"`
}

// MutationResult reports the outcome of a single mutation.
type MutationResult struct {
	Index   int    `json:"index"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	OrderID uint   `json:"order_id,omitempty"`
	OrderNo string `json:"order_no,omitempty"`
}
