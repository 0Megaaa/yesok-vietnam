package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/workflow"
)

type AdminUpdateOrderRequest struct {
	ActionName string `json:"action_name"`
	Remark     string `json:"remark"`
}

type SaveServiceRequest struct {
	ServiceCode string `json:"service_code"`
	ServiceName string `json:"service_name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	CoverImage  string `json:"cover_image"`
	Description string `json:"description"`
	BasePrice   int64  `json:"base_price"`
	Currency    string `json:"currency"`
	Unit        string `json:"unit"`
	SortOrder   int    `json:"sort_order"`
	Status      int    `json:"status"`
	IsHot       bool   `json:"is_hot"`
	FormSchema  string `json:"form_schema"`
}

type SaveSysUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
	Status   int    `json:"status"`
}

// AdminMe returns the current backend employee profile.
func AdminMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")
		uidVal, _ := uid.(uint)
		var user models.SysUser
		if uidVal > 0 && db.First(&user, uidVal).Error == nil {
			c.JSON(http.StatusOK, gin.H{"id": user.ID, "username": user.Username, "real_name": user.RealName, "role": user.Role, "status": user.Status, "is_admin": user.Role == models.RoleAdmin})
			return
		}
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{"id": uidVal, "username": "admin", "role": role, "is_admin": role == models.RoleAdmin})
	}
}

// DashboardStats aggregates dashboard core metrics.
func DashboardStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalUsers, totalOrders, pendingOrders, todayOrders int64
		db.Model(&models.AppUser{}).Count(&totalUsers)
		db.Model(&models.Order{}).Count(&totalOrders)
		db.Model(&models.Order{}).Where("macro_status IN ?", []string{"pending", "reviewing", "quoted"}).Count(&pendingOrders)
		todayStart := time.Now().Format("2006-01-02") + " 00:00:00"
		db.Model(&models.Order{}).Where("created_at >= ?", todayStart).Count(&todayOrders)
		var revenue struct{ Total int64 }
		db.Model(&models.PaymentRecord{}).Select("COALESCE(SUM(pay_amount), 0) as total").Where("status = ?", "success").Scan(&revenue)
		var latest []models.Order
		db.Order("created_at desc").Limit(5).Find(&latest)
		activities := make([]gin.H, 0, len(latest))
		for _, order := range latest {
			activities = append(activities, gin.H{"order_no": order.OrderNo, "service_name": order.ServiceName, "status": order.MacroStatus, "amount": order.TotalAmount, "created_at": order.CreatedAt})
		}
		c.JSON(http.StatusOK, gin.H{"total_users": totalUsers, "total_orders": totalOrders, "pending_orders": pendingOrders, "today_orders": todayOrders, "total_revenue": revenue.Total, "today_activities": activities})
	}
}

// AdminListOrders returns the order center list.
func AdminListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")
		query := db.Model(&models.Order{})
		if status != "" && status != "all" {
			query = query.Where("macro_status = ?", status)
		}
		var total int64
		query.Count(&total)
		var orders []models.Order
		if err := query.Order("created_at desc").Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
			return
		}
		list := make([]gin.H, 0, len(orders))
		for _, order := range orders {
			list = append(list, buildOrderPayload(db, order))
		}
		c.JSON(http.StatusOK, gin.H{"list": list, "orders": list, "total": total})
	}
}

// AdminGetOrder returns a single order with full detail.
func AdminGetOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusOK, buildOrderPayload(db, order))
	}
}

// AdminGetOrderActions returns all actionable workflow nodes for the current stage.
func AdminGetOrderActions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		var nodes []models.SysWorkflowNode
		if err := db.Where("service_id = ? AND stage_code = ?", order.ServiceID, order.CurrentStage).
			Order("sort_order asc, id asc").Find(&nodes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query actions"})
			return
		}
		actions := make([]gin.H, 0, len(nodes))
		for _, n := range nodes {
			if n.IsManual {
				actions = append(actions, gin.H{
					"id":               n.ID,
					"action_name":      n.ActionName,
					"next_stage_code":  n.NextStageCode,
					"stage_name":       n.StageName,
					"require_material": n.RequireMaterial,
					"notify_type":      n.NotifyType,
					"sort_order":       n.SortOrder,
				})
			}
		}
		c.JSON(http.StatusOK, gin.H{"actions": actions})
	}
}

// AdminUpdateOrder advances the order via OrderEngine.
func AdminUpdateOrder(db *gorm.DB, engine *workflow.OrderEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		var req AdminUpdateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if req.ActionName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "action_name is required"})
			return
		}
		if err := engine.AdvanceStage(uint(id), req.ActionName, "后台管家", req.Remark); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found after advance"})
			return
		}
		c.JSON(http.StatusOK, buildOrderPayload(db, order))
	}
}

// AdminListServices returns the backend service configuration list.
func AdminListServices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.SysService
		db.Order("sort_order asc, id asc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminSaveService creates a new service configuration.
func AdminSaveService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveServiceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		service := serviceFromRequest(req)
		if err := db.Create(&service).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create service"})
			return
		}
		c.JSON(http.StatusOK, service)
	}
}

// AdminUpdateService updates an existing service configuration.
func AdminUpdateService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req SaveServiceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		var service models.SysService
		if err := db.First(&service, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}
		updated := serviceFromRequest(req)
		updated.ID = service.ID
		if updated.Currency == "" {
			updated.Currency = service.Currency
		}
		if err := db.Model(&service).Select(
			"service_code", "service_name", "display_name", "icon", "cover_image",
			"description", "base_price", "currency", "unit",
			"sort_order", "status", "is_hot", "form_schema",
		).Updates(updated).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service"})
			return
		}
		db.First(&service, id)
		c.JSON(http.StatusOK, service)
	}
}

func serviceFromRequest(req SaveServiceRequest) models.SysService {
	if req.Currency == "" {
		req.Currency = models.DefaultCurrencyCode
	}
	formSchema := []byte{}
	if req.FormSchema != "" {
		formSchema = []byte(req.FormSchema)
	}
	return models.SysService{
		ServiceCode: req.ServiceCode, ServiceName: req.ServiceName, DisplayName: req.DisplayName,
		Icon: req.Icon, CoverImage: req.CoverImage, Description: req.Description,
		BasePrice: req.BasePrice, Currency: req.Currency, Unit: req.Unit,
		SortOrder: int64(req.SortOrder), Status: req.Status, IsHot: req.IsHot,
		FormSchema: formSchema,
	}
}

// AdminListPayments returns the financial ledger list.
func AdminListPayments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.PaymentRecord
		db.Order("created_at desc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminListAppUsers returns C-end customer profiles.
func AdminListAppUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.AppUser
		db.Order("created_at desc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminListSysUsers returns B-end staff account list.
func AdminListSysUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.SysUser
		db.Order("id asc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminCreateSysUser creates a B-end staff account.
func AdminCreateSysUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveSysUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if req.Password == "" {
			req.Password = "123456"
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if req.Role == "" {
			req.Role = models.RoleManager
		}
		if req.Status == 0 {
			req.Status = 1
		}
		user := models.SysUser{Username: req.Username, PasswordHash: string(hash), RealName: req.RealName, Role: req.Role, Status: req.Status}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sys user"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

// AdminUpdateSysUser updates a B-end staff account.
func AdminUpdateSysUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		updates := map[string]interface{}{}
		var req SaveSysUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if req.RealName != "" {
			updates["real_name"] = req.RealName
		}
		if req.Role != "" {
			updates["role"] = req.Role
		}
		updates["status"] = req.Status
		if req.Password != "" {
			hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			updates["password_hash"] = string(hash)
		}
		var user models.SysUser
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "sys user not found"})
			return
		}
		db.Model(&user).Updates(updates)
		db.First(&user, id)
		c.JSON(http.StatusOK, user)
	}
}

// AdminDeleteSysUser deletes a B-end staff account.
func AdminDeleteSysUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := db.Delete(&models.SysUser{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete sys user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// Legacy aliases.
func ListUsers(db *gorm.DB) gin.HandlerFunc      { return AdminListAppUsers(db) }
func UpdateUserRole(db *gorm.DB) gin.HandlerFunc { return AdminUpdateSysUser(db) }
func DeleteUser(db *gorm.DB) gin.HandlerFunc     { return AdminDeleteSysUser(db) }
