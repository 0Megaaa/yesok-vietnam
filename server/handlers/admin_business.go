package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

type AdminUpdateOrderRequest struct {
	TargetStatus      string `json:"targetStatus"`
	TargetStatusSnake string `json:"target_status"`
	CurrentStatus     string `json:"current_status"`
	Remark            string `json:"remark"`
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

// AdminMe 返回当前后台员工资料。
// 1.意图 -> 让后台登录态来自 sys_users，而不是硬编码账号。
// 2.步骤 -> 从 JWT 读取 uid，查询 sys_users 并屏蔽密码哈希。
// 3.返回 -> 管家后台当前用户信息。
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

// DashboardStats 汇总数据看板核心指标。
// 1.意图 -> 为后台首页提供真实订单、客户、收入和待办数据。
// 2.步骤 -> 分别统计 app_users、orders、payment_records 与今日订单。
// 3.返回 -> 四个统计卡片与今日动态数组。
func DashboardStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalUsers, totalOrders, pendingOrders, todayOrders int64
		db.Model(&models.AppUser{}).Count(&totalUsers)
		db.Model(&models.Order{}).Count(&totalOrders)
		db.Model(&models.Order{}).Where("current_status IN ?", []string{"pending", "reviewing", "quoted"}).Count(&pendingOrders)
		todayStart := time.Now().Format("2006-01-02") + " 00:00:00"
		db.Model(&models.Order{}).Where("created_at >= ?", todayStart).Count(&todayOrders)
		var revenue struct{ Total int64 }
		db.Model(&models.PaymentRecord{}).Select("COALESCE(SUM(pay_amount), 0) as total").Where("status = ?", "success").Scan(&revenue)
		var latest []models.Order
		db.Order("created_at desc").Limit(5).Find(&latest)
		activities := make([]gin.H, 0, len(latest))
		for _, order := range latest {
			activities = append(activities, gin.H{"order_no": order.OrderNo, "service_name": order.ServiceName, "status": order.CurrentStatus, "amount": order.TotalAmount, "created_at": order.CreatedAt})
		}
		c.JSON(http.StatusOK, gin.H{"total_users": totalUsers, "total_orders": totalOrders, "pending_orders": pendingOrders, "today_orders": todayOrders, "total_revenue": revenue.Total, "today_activities": activities})
	}
}

// AdminListOrders 返回订单中心列表。
// 1.意图 -> 将订单、动态 JSON、当前流程按钮合并给后台矩阵页面。
// 2.步骤 -> 按状态筛选 orders，并逐条挂载 sys_workflow_nodes、轨迹和财务流水。
// 3.返回 -> list/total 格式订单聚合列表。
func AdminListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")
		query := db.Model(&models.Order{})
		if status != "" && status != "all" {
			query = query.Where("current_status = ?", status)
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

// AdminGetOrder 返回单个订单完整详情。
// 1.意图 -> 支撑点击行查看完整 JSON、时间线和财务流水。
// 2.步骤 -> 根据 ID 查询订单并复用订单聚合转换函数。
// 3.返回 -> 单个订单聚合对象。
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

// AdminUpdateOrder 根据流程节点更新订单状态，并按配置生成财务流水。
// 1.意图 -> 跑通“后台改状态 -> 时间线 -> 财务流水”的履约闭环。
// 2.步骤 -> 校验目标状态、写入 orders、追加 order_timelines，并在 trigger_payment 时创建 payment_records。
// 3.返回 -> 更新后的订单聚合对象。
func AdminUpdateOrder(db *gorm.DB) gin.HandlerFunc {
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
		target := strings.TrimSpace(req.TargetStatus)
		if target == "" {
			target = strings.TrimSpace(req.TargetStatusSnake)
		}
		if target == "" {
			target = strings.TrimSpace(req.CurrentStatus)
		}
		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "targetStatus is required"})
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		before := order.CurrentStatus
		var node models.SysWorkflowNode
		db.Where("service_id = ? AND current_status = ? AND target_status = ?", order.ServiceID, before, target).First(&node)
		err = db.Transaction(func(tx *gorm.DB) error {
			updates := map[string]interface{}{"current_status": target}
			if target == "paid" {
				updates["payment_status"] = "paid"
			}
			if req.Remark != "" {
				updates["remark"] = req.Remark
			}
			if err := tx.Model(&order).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.OrderTimeline{OrderID: order.ID, BeforeStatus: before, AfterStatus: target, Operator: "后台管家", Remark: defaultRemark(req.Remark, node.ButtonName)}).Error; err != nil {
				return err
			}
			if node.TriggerPayment || target == "paid" || target == "payment_pending" {
				payStatus := "pending"
				if target == "paid" {
					payStatus = "success"
				}
				return tx.Create(&models.PaymentRecord{OrderID: order.ID, AppUserID: order.AppUserID, PayerName: order.ContactName, PayAmount: order.TotalAmount, PayMethod: "bank_transfer", Status: payStatus, ThirdTradeNo: fmt.Sprintf("YS-PAY-%d", time.Now().UnixNano())}).Error
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order"})
			return
		}
		db.First(&order, id)
		c.JSON(http.StatusOK, buildOrderPayload(db, order))
	}
}

func defaultRemark(remark, button string) string {
	if remark != "" {
		return remark
	}
	if button != "" {
		return "执行动作：" + button
	}
	return "后台管家更新订单状态"
}

// AdminListServices 返回后台服务配置列表。
// 1.意图 -> 支撑后台直接维护价格、图标和上下架状态。
// 2.步骤 -> 查询 sys_services 全量数据并按排序输出。
// 3.返回 -> 服务配置数组。
func AdminListServices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.SysService
		db.Order("sort_order asc, id asc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminSaveService 新增服务配置。
// 1.意图 -> 允许后台扩展新的 C 端服务入口。
// 2.步骤 -> 绑定请求并写入 sys_services。
// 3.返回 -> 新服务记录。
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

// AdminUpdateService 更新服务配置。
// 1.意图 -> 让 B 端价格和图标修改实时影响 C 端接口。
// 2.步骤 -> 根据 ID 保存字段到 sys_services。
// 3.返回 -> 更新后的服务记录。
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
		// Currency 若前端没传，保持数据库原值
		if updated.Currency == "" {
			updated.Currency = service.Currency
		}
		// 用 Select 精确指定可更新字段
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
	return models.SysService{ServiceCode: req.ServiceCode, ServiceName: req.ServiceName, DisplayName: req.DisplayName, Icon: req.Icon, CoverImage: req.CoverImage, Description: req.Description, BasePrice: req.BasePrice, Currency: req.Currency, Unit: req.Unit, SortOrder: req.SortOrder, Status: req.Status, IsHot: req.IsHot, FormSchema: req.FormSchema}
}

// AdminListPayments 返回财务流水列表。
// 1.意图 -> 展示付款人、金额、支付方式和关联订单。
// 2.步骤 -> 查询 payment_records 并按创建时间倒序输出。
// 3.返回 -> 财务流水数组。
func AdminListPayments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.PaymentRecord
		db.Order("created_at desc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminListAppUsers 返回 C 端客户画像。
// 1.意图 -> 支撑用户矩阵中的客户画像模块。
// 2.步骤 -> 查询 app_users 并统计基础信息。
// 3.返回 -> C 端客户数组。
func AdminListAppUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.AppUser
		db.Order("created_at desc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminListSysUsers 返回 B 端员工账号列表。
// 1.意图 -> 支撑员工账号管理。
// 2.步骤 -> 查询 sys_users 并屏蔽密码字段。
// 3.返回 -> 员工账号数组。
func AdminListSysUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.SysUser
		db.Order("id asc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminCreateSysUser 创建 B 端员工账号。
// 1.意图 -> 允许管理员新增管家后台账号。
// 2.步骤 -> 对密码 bcrypt 加密并写入 sys_users。
// 3.返回 -> 新员工账号。
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

// AdminUpdateSysUser 更新 B 端员工账号。
// 1.意图 -> 允许管理员修改员工姓名、角色、状态和密码。
// 2.步骤 -> 根据请求构造 updates，密码存在时重新哈希。
// 3.返回 -> 更新后的员工账号。
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

// AdminDeleteSysUser 删除 B 端员工账号。
// 1.意图 -> 清理无效员工账号。
// 2.步骤 -> 按 ID 删除 sys_users 记录，并避免删除最后一个管理员由业务侧控制。
// 3.返回 -> 删除成功标识。
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

// 旧接口别名，保持现有路由编译兼容。
func ListUsers(db *gorm.DB) gin.HandlerFunc      { return AdminListAppUsers(db) }
func UpdateUserRole(db *gorm.DB) gin.HandlerFunc { return AdminUpdateSysUser(db) }
func DeleteUser(db *gorm.DB) gin.HandlerFunc     { return AdminDeleteSysUser(db) }
