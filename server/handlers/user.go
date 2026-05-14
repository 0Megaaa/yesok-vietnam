package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

// ─── GET /api/v1/client/user/me ───────────────────────────────────────────────

// GetMe returns the profile of the authenticated client user.
// UID is read from the JWT claims already injected by the auth middleware.
func GetMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")
		uidVal, ok := uid.(uint)
		if !ok || uidVal == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var user models.User
		if err := db.First(&user, uidVal).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
			"balance":    user.Balance,
			"language":   user.Language,
			"phone":      "",
			"avatar_url": user.AvatarURL,
		})
	}
}

// ─── GET /api/v1/admin/users ──────────────────────────────────────────────────

type UserListResponse struct {
	Users []models.User `json:"users"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

type PaginationQuery struct {
	Page  int `form:"page,default=1"`
	Limit int `form:"limit,default=20"`
}

func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var q PaginationQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			q.Page = 1
			q.Limit = 20
		}
		if q.Page < 1 {
			q.Page = 1
		}
		if q.Limit < 1 || q.Limit > 100 {
			q.Limit = 20
		}
		offset := (q.Page - 1) * q.Limit

		var total int64
		db.Model(&models.User{}).Count(&total)

		var users []models.User
		if err := db.Order("id asc").Offset(offset).Limit(q.Limit).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
			return
		}

		c.JSON(http.StatusOK, UserListResponse{
			Users: users,
			Total: total,
			Page:  q.Page,
			Limit: q.Limit,
		})
	}
}

// ─── PUT /api/v1/admin/users/:id/role ─────────────────────────────────────────

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func UpdateUserRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		var req UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		// Validate role value.
		validRoles := map[string]bool{
			models.RoleAdmin:  true,
			models.RoleUser:   true,
			models.RoleWorker: true,
		}
		if !validRoles[req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role value", "detail": "must be one of: admin, user, worker"})
			return
		}

		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		if err := db.Model(&user).Update("role", req.Role).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
			return
		}

		log.Printf("[UpdateUserRole] user_id=%d new_role=%s by admin", id, req.Role)
		c.JSON(http.StatusOK, gin.H{"message": "role updated", "role": req.Role})
	}
}

// ─── DELETE /api/v1/admin/users/:id ───────────────────────────────────────────

func DeleteUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		// Prevent self-deletion.
		uid, _ := c.Get("uid")
		if uid.(uint) == uint(id) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
			return
		}

		if err := db.Delete(&models.User{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
			return
		}

		log.Printf("[DeleteUser] user_id=%d deleted by admin", id)
		c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
	}
}

// ─── GET /api/v1/admin/auth/me ─────────────────────────────────────────────────

// AdminMe returns the admin's own profile (from JWT claims — no DB lookup needed).
func AdminMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")
		role, _ := c.Get("role")
		username, _ := c.Get("username") // may be empty for JWT-sourced admins

		// Admins (uid=0) are not in the users table; return a minimal profile.
		if uid.(uint) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"id":       0,
				"username": username,
				"role":     role,
				"is_admin": true,
			})
			return
		}

		var user models.User
		if err := db.First(&user, uid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
			"is_admin": user.Role == models.RoleAdmin,
		})
	}
}

// ─── GET /api/v1/admin/orders ─────────────────────────────────────────────────

type OrderListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
	Status string `form:"status"`
	UserID uint   `form:"user_id"`
}

type OrderListResponse struct {
	Orders []models.Order `json:"orders"`
	Total  int64          `json:"total"`
	Page   int            `json:"page"`
	Limit  int            `json:"limit"`
}

func AdminListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var q OrderListQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			q.Page, q.Limit = 1, 20
		}
		if q.Page < 1 {
			q.Page = 1
		}
		if q.Limit < 1 || q.Limit > 100 {
			q.Limit = 20
		}

		query := db.Model(&models.Order{})
		if q.Status != "" {
			query = query.Where("status = ?", q.Status)
		}
		if q.UserID > 0 {
			query = query.Where("user_id = ?", q.UserID)
		}

		var total int64
		query.Count(&total)

		offset := (q.Page - 1) * q.Limit
		var orders []models.Order
		if err := query.Order("created_at desc").Offset(offset).Limit(q.Limit).Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
			return
		}

		c.JSON(http.StatusOK, OrderListResponse{
			Orders: orders,
			Total:  total,
			Page:   q.Page,
			Limit:  q.Limit,
		})
	}
}

// ─── PUT /api/v1/admin/orders/:id ─────────────────────────────────────────────

type AdminUpdateOrderRequest struct {
	Status     string `json:"status"`
	WorkerTGID int64  `json:"worker_tg_id"`
	Note       string `json:"note"`
}

func AdminUpdateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}

		var req AdminUpdateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		updates := map[string]interface{}{}
		if req.Status != "" {
			validStatuses := map[string]bool{
				string(models.OrderStatusPending):   true,
				string(models.OrderStatusConfirmed): true,
				string(models.OrderStatusProgress):  true,
				string(models.OrderStatusCompleted): true,
				string(models.OrderStatusCancelled): true,
				string(models.OrderStatusFailed):    true,
				string(models.OrderStatusRefunded):  true,
			}
			if !validStatuses[req.Status] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
				return
			}
			updates["status"] = req.Status
		}
		if req.WorkerTGID > 0 {
			updates["worker_tg_id"] = req.WorkerTGID
		}
		if req.Note != "" {
			updates["note"] = req.Note
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		if err := db.Model(&order).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order"})
			return
		}

		log.Printf("[AdminUpdateOrder] order_id=%d updates=%v by admin", id, updates)
		c.JSON(http.StatusOK, gin.H{"message": "order updated"})
	}
}

// ─── GET /api/v1/admin/dashboard/stats ─────────────────────────────────────────

type DashboardStatsPayload struct {
	TotalUsers   int64   `json:"total_users"`
	TotalOrders  int64   `json:"total_orders"`
	TotalRevenue float64 `json:"total_revenue"`
}

func DashboardStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats DashboardStatsPayload

		db.Model(&models.User{}).Count(&stats.TotalUsers)
		db.Model(&models.Order{}).Count(&stats.TotalOrders)

		type result struct {
			Total float64
		}
		var r result
		db.Model(&models.Order{}).
			Select("COALESCE(SUM(amount), 0) as total").
			Scan(&r)
		stats.TotalRevenue = r.Total

		c.JSON(http.StatusOK, stats)
	}
}
