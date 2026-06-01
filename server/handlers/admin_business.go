package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/workflow"
)

type AdminUpdateOrderRequest struct {
	ActionName string                 `json:"action_name"`
	Remark     string                 `json:"remark"`
	InputData  map[string]interface{} `json:"input_data"`
}

// ServiceInfoRequest 是服务基础信息的请求体。
type ServiceInfoRequest struct {
	ID          uint   `json:"id"`
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

// SaveServiceRequest 合并服务基础信息与工作流节点配置。
type SaveServiceRequest struct {
	ServiceInfo   ServiceInfoRequest       `json:"service_info"`
	WorkflowNodes []models.SysWorkflowNode `json:"workflow_nodes"`
}

type SaveSysUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
	Status   int    `json:"status"`
}

// Error codes for unified frontend error display
const (
	ErrCodeInvalidRequest    = "INVALID_REQUEST"    // 参数校验失败
	ErrCodeNotFound          = "NOT_FOUND"          // 资源不存在
	ErrCodeTransactionFailed = "TRANSACTION_FAILED" // 事务执行失败
	ErrCodeInternalError     = "INTERNAL_ERROR"     // 内部错误
	ErrCodeUnauthorized      = "UNAUTHORIZED"       // 未授权
)

// userError 包装用户可见的业务错误。
// 前端可通过 err.code 判断错误类型，err.message 显示给用户。
type userError struct {
	code    string
	message string
}

func (e *userError) Error() string { return e.message }

func newUserError(code, message string) *userError { return &userError{code: code, message: message} }

// httpError 构造可直接写入 JSON 响应的 gin.H。
// 前端统一通过 response.data?.error 或 response.message 获取错误信息。
func httpError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": message, "code": code})
}

// safeHTTPError 是带 recover 的安全包装，防止 panic 泄露。
func safeHTTPError(c *gin.Context, status int, code, message string) {
	defer func() {
		if r := recover(); r != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "服务器内部异常，请稍后重试")
		}
	}()
	httpError(c, status, code, message)
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
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to fetch orders")
			return
		}
		list := make([]gin.H, 0, len(orders))
		for _, order := range orders {
			list = append(list, buildOrderPayloadForRole(db, order, "admin"))
		}
		c.JSON(http.StatusOK, gin.H{"list": list, "total": total})
	}
}

// AdminGetOrder returns a single order with full detail.
func AdminGetOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found")
			return
		}
		c.JSON(http.StatusOK, buildOrderPayloadForRole(db, order, "admin"))
	}
}

// AdminGetOrderActions returns all actionable workflow nodes for the current stage (admin role).
func AdminGetOrderActions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found")
			return
		}
		var nodes []models.SysWorkflowNode
		if err := db.Where(
			"service_id = ? AND stage_code = ? AND (executor_role = ? OR executor_role = 'both')",
			order.ServiceID, order.CurrentStage, "admin",
		).Order("sort_order asc").Find(&nodes).Error; err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to query actions")
			return
		}

		// 规范化返回字段（添加中文字段）
		actions := make([]gin.H, 0, len(nodes))
		for _, n := range nodes {
			actionNameText := dictLabel(db, "workflow_action", n.ActionName)
			if actionNameText == "" {
				actionNameText = n.ButtonLabel
			}
			targetStatusText := dictLabel(db, "node_stage", n.TargetStatus)
			if targetStatusText == "" {
				targetStatusText = n.TargetStatus
			}
			macroText := dictLabel(db, "macro_status", n.MacroStatus)
			if macroText == "" {
				macroText = n.MacroStatus
			}
			notifyText := dictLabel(db, "notify_type", n.NotifyType)
			if notifyText == "" {
				notifyText = n.NotifyType
			}
			actions = append(actions, gin.H{
				"id":                  n.ID,
				"action_name":         n.ActionName,
				"action_name_text":    actionNameText,
				"button_label":        n.ButtonLabel,
				"action_type":         n.ActionType,
				"form_fields":         n.FormFields,
				"target_status":       n.TargetStatus,
				"target_status_text":  targetStatusText,
				"macro_status":        n.MacroStatus,
				"macro_status_text":   macroText,
				"notify_type":         n.NotifyType,
				"notify_type_text":    notifyText,
				"need_audit":          n.NeedAudit,
				"audit_reject_status": n.AuditRejectStatus,
				"sort_order":          n.SortOrder,
				"stage_code":          n.StageCode,
				"stage_name":          n.StageName,
			})
		}

		c.JSON(http.StatusOK, gin.H{"actions": actions})
	}
}

// AdminUpdateOrder advances the order via OrderEngine.
func AdminUpdateOrder(db *gorm.DB, engine *workflow.OrderEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}
		var req AdminUpdateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
			return
		}
		if req.ActionName == "" {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "action_name is required")
			return
		}
		if err := engine.AdvanceStage(uint(id), req.ActionName, "后台管家", "admin", req.InputData, req.Remark); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeTransactionFailed, "流程推进失败: "+err.Error())
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found after advance")
			return
		}
		c.JSON(http.StatusOK, buildOrderPayloadForRole(db, order, "admin"))
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

// AdminPostOrderAction 是工作流动作的专用 POST 入口。
// 参数与 AdvanceStage 签名完全对齐：action_name / input_data / remark。
func AdminPostOrderAction(db *gorm.DB, engine *workflow.OrderEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}
		var req AdminUpdateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request: "+err.Error())
			return
		}
		if req.ActionName == "" {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "action_name is required")
			return
		}
		if err := engine.AdvanceStage(uint(id), req.ActionName, "后台管家", "admin", req.InputData, req.Remark); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeTransactionFailed, "流程推进失败: "+err.Error())
			return
		}
		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found after advance")
			return
		}
		c.JSON(http.StatusOK, buildOrderPayloadForRole(db, order, "admin"))
	}
}

// AdminGetServiceDetail 返回服务基础信息与工作流节点列表的复合对象。
// 前端单次调用即可获取完整配置，无需拆分请求。
func AdminGetServiceDetail(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid service id")
			return
		}
		var service models.SysService
		if err := db.First(&service, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "service not found")
			return
		}
		var nodes []models.SysWorkflowNode
		db.Where("service_id = ?", id).Order("sort_order asc, id asc").Find(&nodes)
		c.JSON(http.StatusOK, gin.H{
			"service_info":   service,
			"workflow_nodes": nodes,
		})
	}
}

// AdminGetServiceWorkflow 返回指定服务的所有工作流节点，供编辑时回显。
// 前端在打开编辑弹窗时会并行请求 GET /admin/services/:id/workflow。
func AdminGetServiceWorkflow(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid service id")
			return
		}
		var service models.SysService
		if err := db.First(&service, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "service not found")
			return
		}
		var nodes []models.SysWorkflowNode
		if err := db.Where("service_id = ?", id).Order("sort_order asc, id asc").Find(&nodes).Error; err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to query workflow nodes")
			return
		}
		c.JSON(http.StatusOK, gin.H{"workflow_nodes": nodes})
	}
}

// AdminGetServiceActions 返回指定服务在给定 stage_code 和角色下可执行的动作列表。
// role 参数从 query string 传入，默认为 "admin"。
// 用于订单详情页：通过 order.service_id + order.current_stage 动态渲染操作按钮。
func AdminGetServiceActions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid service id")
			return
		}
		stage := c.Query("stage")
		if stage == "" {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "stage query param is required")
			return
		}
		role := c.DefaultQuery("role", "admin")

		var nodes []models.SysWorkflowNode
		if err := db.Where(
			"service_id = ? AND stage_code = ? AND (executor_role = ? OR executor_role = 'both')",
			id, stage, role,
		).Order("sort_order asc").Find(&nodes).Error; err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to query actions")
			return
		}
		c.JSON(http.StatusOK, gin.H{"actions": nodes})
	}
}

// AdminSaveService 在事务中同时保存服务基础信息与工作流节点配置。
// 前端 payload：{service_info: {...}, workflow_nodes: [...]}
// 事务保证：主表与从表同时成功或同时回滚。
func AdminSaveService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveServiceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request: "+err.Error())
			return
		}
		if req.ServiceInfo.ServiceCode == "" {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "service_code is required")
			return
		}

		// 后端工作流节点校验
		if err := validateWorkflowNodes(req.WorkflowNodes); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
			return
		}

		var service models.SysService
		if req.ServiceInfo.ID > 0 {
			if err := db.First(&service, req.ServiceInfo.ID).Error; err != nil {
				httpError(c, http.StatusNotFound, ErrCodeNotFound, "service not found")
				return
			}
		}

		svc := buildServiceFromRequest(req.ServiceInfo)
		if req.ServiceInfo.ID > 0 {
			svc.ID = req.ServiceInfo.ID
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&svc).Error; err != nil {
				return fmt.Errorf("save service failed: %w", err)
			}
			if err := tx.Where("service_id = ?", svc.ID).Delete(&models.SysWorkflowNode{}).Error; err != nil {
				return fmt.Errorf("clear workflow nodes failed: %w", err)
			}
			if len(req.WorkflowNodes) > 0 {
				nodes := make([]models.SysWorkflowNode, len(req.WorkflowNodes))
				for i := range req.WorkflowNodes {
					nodes[i] = req.WorkflowNodes[i]
					nodes[i].ID = 0
					nodes[i].ServiceID = svc.ID
				}
				if err := tx.Create(&nodes).Error; err != nil {
					return fmt.Errorf("create workflow nodes failed: %w", err)
				}
			}
			return nil
		}); err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeTransactionFailed, "事务执行失败："+err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{"service": svc, "message": "service saved with workflow nodes"})
	}
}

// AdminUpdateService 在事务中同时更新服务基础信息与工作流节点配置。
func AdminUpdateService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid service id")
			return
		}
		var req SaveServiceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request: "+err.Error())
			return
		}

		// 后端工作流节点校验
		if err := validateWorkflowNodes(req.WorkflowNodes); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
			return
		}

		var service models.SysService
		if err := db.First(&service, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "service not found")
			return
		}

		svc := buildServiceFromRequest(req.ServiceInfo)
		svc.ID = uint(id)

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&svc).Error; err != nil {
				return fmt.Errorf("update service failed: %w", err)
			}
			if err := tx.Where("service_id = ?", svc.ID).Delete(&models.SysWorkflowNode{}).Error; err != nil {
				return fmt.Errorf("clear workflow nodes failed: %w", err)
			}
			if len(req.WorkflowNodes) > 0 {
				nodes := make([]models.SysWorkflowNode, len(req.WorkflowNodes))
				for i := range req.WorkflowNodes {
					nodes[i] = req.WorkflowNodes[i]
					nodes[i].ID = 0
					nodes[i].ServiceID = svc.ID
				}
				if err := tx.Create(&nodes).Error; err != nil {
					return fmt.Errorf("create workflow nodes failed: %w", err)
				}
			}
			return nil
		}); err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeTransactionFailed, "事务执行失败："+err.Error())
			return
		}

		db.First(&service, id)
		c.JSON(http.StatusOK, gin.H{"service": service, "message": "service updated with workflow nodes"})
	}
}

// buildServiceFromRequest 从请求体构建 SysService 模型。
func buildServiceFromRequest(req ServiceInfoRequest) models.SysService {
	if req.Currency == "" {
		req.Currency = models.DefaultCurrencyCode
	}
	formSchema := []byte("{}")
	if req.FormSchema != "" && json.Valid([]byte(req.FormSchema)) {
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

// validateWorkflowNodes 校验工作流节点的必填字段和合法性。
func validateWorkflowNodes(nodes []models.SysWorkflowNode) error {
	// 合法的字段类型
	validFieldTypes := map[string]bool{
		"text": true, "textarea": true, "number": true,
		"date": true, "datetime": true, "select": true,
		"image": true, "file": true, "phone": true,
	}
	// 合法的执行角色
	validExecutorRoles := map[string]bool{"admin": true, "client": true, "both": true}
	// 合法的动作类型
	validActionTypes := map[string]bool{"button_click": true, "form_input": true, "wx_pay": true}

	// 同一服务内不允许 stage_code + executor_role + action_name 重复
	seen := make(map[string]bool)

	for i := range nodes {
		n := &nodes[i] // 使用指针以便修改原切片

		// 基础必填校验
		if strings.TrimSpace(n.StageCode) == "" {
			return fmt.Errorf("第 %d 个节点：stage_code 不能为空", i+1)
		}
		if strings.TrimSpace(n.StageName) == "" {
			return fmt.Errorf("第 %d 个节点：stage_name 不能为空", i+1)
		}
		if strings.TrimSpace(n.ActionName) == "" {
			return fmt.Errorf("第 %d 个节点：action_name 不能为空", i+1)
		}
		if strings.TrimSpace(n.ButtonLabel) == "" {
			return fmt.Errorf("第 %d 个节点：button_label 不能为空", i+1)
		}

		// executor_role 校验
		if !validExecutorRoles[n.ExecutorRole] {
			return fmt.Errorf("第 %d 个节点：executor_role 必须是 admin/client/both", i+1)
		}

		// action_type 校验
		if !validActionTypes[n.ActionType] {
			return fmt.Errorf("第 %d 个节点：action_type 必须是 button_click/form_input/wx_pay", i+1)
		}

		// target_status 校验
		if strings.TrimSpace(n.TargetStatus) == "" {
			return fmt.Errorf("第 %d 个节点：target_status 不能为空", i+1)
		}

		// macro_status 校验
		if strings.TrimSpace(n.MacroStatus) == "" {
			return fmt.Errorf("第 %d 个节点：macro_status 不能为空", i+1)
		}

		// notify_type 为空时自动设为 none
		if strings.TrimSpace(n.NotifyType) == "" {
			n.NotifyType = "none"
		}

		// form_fields 字段类型校验
		for j := range n.FormFields {
			f := &n.FormFields[j]
			if !validFieldTypes[f.Type] {
				return fmt.Errorf("第 %d 个节点第 %d 个字段：type '%s' 不合法", i+1, j+1, f.Type)
			}
		}

		// 重复校验
		uniqueKey := fmt.Sprintf("%s_%s_%s", n.StageCode, n.ExecutorRole, n.ActionName)
		if seen[uniqueKey] {
			return fmt.Errorf("工作流动作重复：stage_code=%s + executor_role=%s + action_name=%s 不允许重复", n.StageCode, n.ExecutorRole, n.ActionName)
		}
		seen[uniqueKey] = true
	}

	return nil
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
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request: "+err.Error())
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
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to create sys user: "+err.Error())
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
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request: "+err.Error())
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
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "sys user not found")
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
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to delete sys user: "+err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// AdminAuditOrderRequest 审核订单的请求结构。
type AdminAuditOrderRequest struct {
	TimelineID  uint   `json:"timeline_id" binding:"required"`
	Result      string `json:"result" binding:"required"` // approved/rejected
	AuditRemark string `json:"audit_remark"`
}

// AdminAuditOrder 处理后台审核订单（审核通过 / 审核失败）。
func AdminAuditOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || orderID == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}

		var req AdminAuditOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request: "+err.Error())
			return
		}

		if req.Result != models.AuditStatusApproved && req.Result != models.AuditStatusRejected {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "result must be approved or rejected")
			return
		}

		var updatedOrder models.Order
		err = db.Transaction(func(tx *gorm.DB) error {
			// Step 1: 查出订单
			var order models.Order
			if err := tx.First(&order, orderID).Error; err != nil {
				return fmt.Errorf("订单不存在: %w", err)
			}

			// Step 2: 查出待审核时间线
			var timeline models.OrderTimeline
			if err := tx.First(&timeline, req.TimelineID).Error; err != nil {
				return fmt.Errorf("时间线记录不存在: %w", err)
			}
			if timeline.OrderID != order.ID {
				return fmt.Errorf("时间线记录与订单不匹配")
			}
			if timeline.AuditStatus != models.AuditStatusPending {
				return fmt.Errorf("该审核记录已处理，请勿重复审核")
			}

			// Step 3: 查找原工作流节点（用于获取 audit_reject_status）
			var originNode models.SysWorkflowNode
			found := tx.Where(
				"service_id = ? AND action_name = ? AND target_status = ?",
				order.ServiceID, timeline.ActionName, timeline.AfterStatus,
			).First(&originNode).Error == nil

			if !found {
				tx.Where(
					"service_id = ? AND action_name = ?",
					order.ServiceID, timeline.ActionName,
				).First(&originNode)
			}

			now := time.Now()

			if req.Result == models.AuditStatusApproved {
				// ======================================================
				// 审核通过：查找 audit_approve 节点推进到下一状态
				// ======================================================
				var approveNode models.SysWorkflowNode
				if err := tx.Where(
					"service_id = ? AND stage_code = ? AND action_name = ? AND executor_role IN ?",
					order.ServiceID, order.CurrentStage, "audit_approve",
					[]string{"admin", "both"},
				).Order("sort_order ASC, id ASC").First(&approveNode).Error; err != nil {
					return fmt.Errorf("缺少审核通过工作流节点 audit_approve")
				}

				// 更新原 timeline
				auditRemark := strings.TrimSpace(req.AuditRemark)
				if auditRemark == "" {
					auditRemark = "资料审核通过"
				}
				tx.Model(&timeline).Updates(map[string]interface{}{
					"audit_status":   models.AuditStatusApproved,
					"audit_remark":   auditRemark,
					"audit_operator": "后台审核",
					"audited_at":     now,
					"updated_at":     now,
				})

				// 更新订单状态
				tx.Model(&order).Updates(map[string]interface{}{
					"current_stage": approveNode.TargetStatus,
					"macro_status":  approveNode.MacroStatus,
					"updated_at":    now,
				})

				// 写入审核通过 timeline
				tx.Create(&models.OrderTimeline{
					OrderID:       order.ID,
					BeforeStatus:  order.CurrentStage,
					AfterStatus:   approveNode.TargetStatus,
					Operator:      "后台审核",
					Remark:        "审核通过：" + auditRemark,
					ActionName:    "audit_approve",
					AuditStatus:   models.AuditStatusApproved,
					AuditRemark:   auditRemark,
					AuditOperator: "后台审核",
					AuditedAt:     &now,
					CreatedAt:     &now,
					UpdatedAt:     &now,
				})

				tx.First(&order, order.ID)
				updatedOrder = order
				return nil
			}

			// ======================================================
			// 审核失败：回退到配置的回退节点
			// ======================================================
			auditRemark := strings.TrimSpace(req.AuditRemark)
			if auditRemark == "" {
				return fmt.Errorf("审核失败原因不能为空")
			}

			// 更新原 timeline
			tx.Model(&timeline).Updates(map[string]interface{}{
				"audit_status":   models.AuditStatusRejected,
				"audit_remark":   auditRemark,
				"audit_operator": "后台审核",
				"audited_at":     now,
				"updated_at":     now,
			})

			// 计算回退节点
			rejectStatus := strings.TrimSpace(originNode.AuditRejectStatus)
			if rejectStatus == "" {
				rejectStatus = "wait_supplement"
			}

			// 更新订单状态（macro_status 固定为 reviewing，表示待补资料状态）
			tx.Model(&order).Updates(map[string]interface{}{
				"current_stage": rejectStatus,
				"macro_status":  "reviewing",
				"updated_at":    now,
			})

			// 写入审核失败 timeline
			tx.Create(&models.OrderTimeline{
				OrderID:       order.ID,
				BeforeStatus:  order.CurrentStage,
				AfterStatus:   rejectStatus,
				Operator:      "后台审核",
				Remark:        "审核未通过：" + auditRemark,
				ActionName:    "audit_reject",
				AuditStatus:   models.AuditStatusRejected,
				AuditRemark:   auditRemark,
				AuditOperator: "后台审核",
				AuditedAt:     &now,
				CreatedAt:     &now,
				UpdatedAt:     &now,
			})

			tx.First(&order, order.ID)
			updatedOrder = order
			return nil
		})

		if err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"order": buildOrderPayloadForRole(db, updatedOrder, "admin"),
		})
	}
}

// Legacy aliases.
func ListUsers(db *gorm.DB) gin.HandlerFunc      { return AdminListAppUsers(db) }
func UpdateUserRole(db *gorm.DB) gin.HandlerFunc { return AdminUpdateSysUser(db) }
func DeleteUser(db *gorm.DB) gin.HandlerFunc     { return AdminDeleteSysUser(db) }
