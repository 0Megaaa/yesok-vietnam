package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"yesok-vietnam/models"
)

// ─── GET /api/state ───────────────────────────────────────────────────────────

// GetState returns the full application state for the authenticated user.
//   - admin  → all users + all orders
//   - user   → only their own orders + basic profile
func GetState(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		currentUser := u.(*models.User)

		var resp GetStateResponse

		if currentUser.Role == models.RoleAdmin {
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
		} else {
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
		}

		c.JSON(http.StatusOK, resp)
	}
}

// ─── PUT /api/state ───────────────────────────────────────────────────────────

// UpdateState handles order creation and status updates, persisting them to MySQL.
// It accepts a JSON body describing one or more order mutations:
//   - Create: { "action": "create", "order": { ...fields } }
//   - Update: { "action": "update", "id": 123, "fields": { ...patch } }
func UpdateState(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		currentUser := u.(*models.User)

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
				res = handleUpdateOrder(db, currentUser, mut.ID, mut.Fields)
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
	if order.Status == "" {
		order.Status = models.OrderStatusPending
	}
	if order.Currency == "" {
		order.Currency = "VND"
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

func handleUpdateOrder(db *gorm.DB, user *models.User, orderID uint, raw json.RawMessage) MutationResult {
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

	// Parse optional field patch.
	if raw != nil {
		var patch map[string]interface{}
		if err := json.Unmarshal(raw, &patch); err != nil {
			return MutationResult{Success: false, Message: "malformed patch: " + err.Error()}
		}
		// Remove read-only fields from patch.
		delete(patch, "id")
		delete(patch, "user_id")
		delete(patch, "order_no")
		delete(patch, "created_at")
		delete(patch, "updated_at")

		if len(patch) > 0 {
			if err := db.Model(&order).Updates(patch).Error; err != nil {
				return MutationResult{Success: false, Message: "update failed: " + err.Error()}
			}
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
