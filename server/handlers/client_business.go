package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/workflow"
)

type CreateOrderRequest struct {
	ServiceID    uint                   `json:"service_id"`
	ServiceCode  string                 `json:"service_code"`
	AppUserID    uint                   `json:"app_user_id"`
	ContactName  string                 `json:"contact_name"`
	ContactPhone string                 `json:"contact_phone"`
	FormData     map[string]interface{} `json:"form_data"`
}

// ClientListServices 输出所有启用服务，驱动 C 端首页分类与热门卡片。
// 1.意图 -> 消灭前端服务分类、价格和图标硬编码。
// 2.步骤 -> 按 status=1 与 sort_order 查询 sys_services，并转换为前端友好字段。
// 3.返回 -> 服务数组，包含价格、图标、表单结构和热门标记。
func ClientListServices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var services []models.SysService
		if err := db.Where("status = ?", 1).Order("sort_order asc, id asc").Find(&services).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch services"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"list": buildServicePayloads(services), "total": len(services)})
	}
}

// ClientGetService 返回单个服务的完整信息（包含 form_schema）。
// 1.意图 -> 支撑服务详情页的动态表单渲染。
// 2.步骤 -> 按 id 精确查询 sys_services 并转换为前端友好字段。
// 3.返回 -> 单个服务对象，包含解析后的 form_schema。
func ClientGetService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		var service models.SysService
		var err error
		if idStr != "" {
			var id uint
			_, parseErr := fmt.Sscanf(idStr, "%d", &id)
			if parseErr == nil {
				err = db.Where("id = ? AND status = ?", id, 1).First(&service).Error
			} else {
				err = db.Where("service_code = ? AND status = ?", idStr, 1).First(&service).Error
			}
		}
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}
		payload := buildServicePayloads([]models.SysService{service})
		if len(payload) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}
		c.JSON(http.StatusOK, payload[0])
	}
}

// ClientGetServiceInitForm 返回指定服务的下单初始化表单。
// 数据来源：优先读取 stage_code='start' 且 action_name='submit_request' 的工作流节点，
// 取其 form_fields 作为下单表单配置。若无节点配置，则回退到 sys_services.form_schema。
func ClientGetServiceInitForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		var service models.SysService
		var err error
		if idStr != "" {
			var id uint
			if _, parseErr := fmt.Sscanf(idStr, "%d", &id); parseErr == nil {
				err = db.Where("id = ? AND status = ?", id, 1).First(&service).Error
			} else {
				err = db.Where("service_code = ? AND status = ?", idStr, 1).First(&service).Error
			}
		}
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}

		// 优先取 start 节点 + submit_request 的 form_fields
		var startNode models.SysWorkflowNode
		found := db.Where(
			"service_id = ? AND stage_code = 'start' AND action_name = 'submit_request' AND action_type = 'form_input' AND (executor_role = 'client' OR executor_role = 'both')",
			service.ID,
		).First(&startNode).Error == nil

		if found && len(startNode.FormFields) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"service_id":    service.ID,
				"service_name":  service.ServiceName,
				"action_name":   startNode.ActionName,
				"button_label":  startNode.ButtonLabel,
				"action_type":   startNode.ActionType,
				"form_fields":   startNode.FormFields,
				"target_status": startNode.TargetStatus,
				"macro_status":  startNode.MacroStatus,
				"notify_type":   startNode.NotifyType,
				"source":        "workflow_node",
			})
			return
		}

		// 回退：使用 sys_services.form_schema
		var fields []models.FormFieldDef
		if len(service.FormSchema) > 0 {
			// 1. 先尝试直接解析数组
			if err := json.Unmarshal(service.FormSchema, &fields); err != nil || fields == nil {
				// 2. 再尝试解析 { fields: [] }
				var schema struct {
					Fields []models.FormFieldDef `json:"fields"`
				}
				if err2 := json.Unmarshal(service.FormSchema, &schema); err2 == nil {
					fields = schema.Fields
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"service_id":   service.ID,
			"service_name": service.ServiceName,
			"form_fields":  fields,
			"source":       "form_schema",
		})
	}
}

// ClientGetConfigs 输出 C 端公开全局配置。
// 1.意图 -> 让 Banner、热线、主题色和全局文案由后台动态控制。
// 2.步骤 -> 仅读取 is_public=true 的 sys_configs，并按 key 组装为对象。
// 3.返回 -> configs 键值对象。
func ClientGetConfigs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var configs []models.SysConfig
		if err := db.Where("is_public = ?", true).Order("group_name asc, id asc").Find(&configs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch configs"})
			return
		}
		payload := gin.H{}
		for _, item := range configs {
			payload[item.ConfigKey] = parseConfigValue(item)
		}
		c.JSON(http.StatusOK, gin.H{"configs": payload})
	}
}

// ClientCreateOrder 接收 C 端订单并将动态表单写入 orders.form_data。
// 1.意图 -> 用统一订单表承接接机、签证、翻译等不同业务。
// 2.步骤 -> 校验服务、创建演示客户、查询 start+submit_request 节点、序列化 form_data、生成订单号和初始时间线。
// 3.返回 -> 新订单摘要，供 C 端立即展示订单状态。
func ClientCreateOrder(db *gorm.DB, engine *workflow.OrderEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		// 优先使用 JWT uid 创建订单，不信任前端传来的 app_user_id
		uidVal, ok := c.Get("uid")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		appUserID, ok := uidVal.(uint)
		if !ok || appUserID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		var appUser models.AppUser
		if err := db.First(&appUser, appUserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "app user not found, please login again"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		var service models.SysService
		query := db.Where("status = ?", 1)
		if req.ServiceID > 0 {
			query = query.Where("id = ?", req.ServiceID)
		} else if req.ServiceCode != "" {
			query = query.Where("service_code = ?", req.ServiceCode)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "service_id or service_code is required"})
			return
		}
		if err := query.First(&service).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}

		// 查询服务对应的 start + submit_request 节点作为初始状态
		var startNode models.SysWorkflowNode
		if err := db.Where(
			"service_id = ? AND stage_code = 'start' AND action_name = 'submit_request' AND action_type = 'form_input' AND (executor_role = 'client' OR executor_role = 'both')",
			service.ID,
		).First(&startNode).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前服务未配置下单工作流，请联系管理员"})
			return
		}

		// 校验 form_fields 必填字段
		for _, field := range startNode.FormFields {
			if field.Required {
				val, ok := req.FormData[field.Key]
				if !ok || val == nil || val == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("缺少必填字段：%s", field.Label)})
					return
				}
			}
		}

		// 优先使用 service_name，display_name 作为兜底
		serviceName := strings.TrimSpace(service.ServiceName)
		if serviceName == "" {
			serviceName = strings.TrimSpace(service.DisplayName)
		}
		formData := normalizeFormData(req.FormData, service, serviceName)
		order := models.Order{
			OrderNo:       fmt.Sprintf("YS%s%04d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000),
			AppUserID:     appUser.ID,
			ServiceID:     service.ID,
			ServiceName:   serviceName,
			ContactName:   req.ContactName,
			ContactPhone:  req.ContactPhone,
			TotalAmount:   service.BasePrice,
			Currency:      service.Currency,
			CurrentStage:  startNode.TargetStatus,
			MacroStatus:   startNode.MacroStatus,
			PaymentStatus: "unpaid",
			FormData:      marshalJSON(formData),
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&order).Error; err != nil {
				return err
			}
			return tx.Create(&models.OrderTimeline{
				OrderID:      order.ID,
				BeforeStatus: "start",
				AfterStatus:  order.CurrentStage,
				Operator:     "C端客户",
				Remark:       "客户提交资料",
				ActionName:   "submit_request",
			}).Error
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"order": buildOrderPayload(db, order), "message": "订单已提交，管家即将联系您"})
	}
}

// buildServicePayloads 统一转换服务配置。
// 1.意图 -> 屏蔽数据库字段命名与前端展示命名差异。
// 2.步骤 -> 遍历服务列表并解析 form_schema JSON。
// 3.返回 -> 可直接用于 C 端 v-for 的数组。
func buildServicePayloads(services []models.SysService) []gin.H {
	items := make([]gin.H, 0, len(services))
	for _, item := range services {
		name := item.DisplayName
		if name == "" {
			name = item.ServiceName
		}
		items = append(items, gin.H{
			"id": item.ID, "service_id": item.ID, "service_code": item.ServiceCode, "name": name,
			"service_name": item.ServiceName, "icon": item.Icon, "cover_image": item.CoverImage,
			"description": item.Description, "base_price": item.BasePrice, "price": formatMoney(item.BasePrice),
			"currency": item.Currency, "unit": item.Unit, "is_hot": item.IsHot, "sort_order": item.SortOrder,
			"form_schema": parseJSONString(string(item.FormSchema)),
		})
	}
	return items
}

// buildOrderPayload 统一转换订单聚合信息（兼容旧调用，默认为 client 角色）。
// 1.意图 -> 为后台和 C 端复用同一份订单详情结构。
// 2.步骤 -> 解析 form_data，加载时间线、支付记录和当前动作节点。
// 3.返回 -> 带 JSON 详情、动态按钮和财务流水的订单对象。
func buildOrderPayload(db *gorm.DB, order models.Order) gin.H {
	return buildOrderPayloadForRole(db, order, "client")
}

// buildOrderPayloadForRole 按角色返回订单详情，role 可以是 "client" 或 "admin"。
// client 只返回 client/both 角色的动作节点，admin 返回 admin/both 角色的动作节点。
func buildOrderPayloadForRole(db *gorm.DB, order models.Order, role string) gin.H {
	var timelines []models.OrderTimeline
	var payments []models.PaymentRecord
	db.Where("order_id = ?", order.ID).Order("created_at asc").Find(&timelines)
	db.Where("order_id = ?", order.ID).Order("created_at desc").Find(&payments)

	// 按角色查询动作节点
	var nodes []models.SysWorkflowNode
	db.Where(
		"service_id = ? AND stage_code = ? AND (executor_role = ? OR executor_role = 'both')",
		order.ServiceID, order.CurrentStage, role,
	).Order("sort_order asc, id asc").Find(&nodes)

	// 构建 form_items：业务字段中文展示
	formItems := buildFormItems(db, order)

	// 状态中文
	macroStatusText := dictLabel(db, "macro_status", order.MacroStatus)
	if macroStatusText == "" {
		macroStatusText = order.MacroStatus
	}
	currentStageText := workflowStageLabel(db, order.ServiceID, order.CurrentStage)
	paymentStatusText := paymentStatusLabel(order.PaymentStatus)

	// timeline 中文字段
	timelineItems := make([]gin.H, 0, len(timelines))
	for _, tl := range timelines {
		afterText := workflowStageLabel(db, order.ServiceID, tl.AfterStatus)
		beforeText := workflowStageLabel(db, order.ServiceID, tl.BeforeStatus)

		// is_audit_timeline：仅真正的审核类动作需要显示审核状态
		// upload_delivery_material 是普通交付动作，B端上传，C端查看，不属于审核动作
		isAuditTimeline := tl.AuditStatus == models.AuditStatusPending ||
			tl.ActionName == "audit_approve" ||
			tl.ActionName == "audit_reject" ||
			tl.ActionName == "upload_material" ||
			tl.ActionName == "supplement_material"

		timelineItems = append(timelineItems, gin.H{
			"id":                 tl.ID,
			"before_status":      tl.BeforeStatus,
			"before_status_text": beforeText,
			"after_status":       tl.AfterStatus,
			"after_status_text":  afterText,
			"remark":             tl.Remark,
			"action_name":        tl.ActionName,
			"payload":            tl.Payload,
			"created_at":         tl.CreatedAt,
			"audit_status":       tl.AuditStatus,
			"audit_status_text":  auditStatusText(tl.AuditStatus),
			"audit_remark":       tl.AuditRemark,
			"audit_operator":     tl.AuditOperator,
			"audited_at":         tl.AuditedAt,
			"is_audit_timeline":  isAuditTimeline,
		})
	}

	// actionNodes 中文字段（C 端过滤审核动作）
	actionNodes := make([]gin.H, 0, len(nodes))
	for _, n := range nodes {
		// C 端过滤审核动作
		if role == "client" && (n.ActionName == "audit_approve" || n.ActionName == "audit_reject" || n.ActionName == "audit_rejected") {
			continue
		}
		actionNameText := dictLabel(db, "workflow_action", n.ActionName)
		if actionNameText == "" {
			actionNameText = n.ButtonLabel
		}
		targetStatusText := workflowStageLabel(db, order.ServiceID, n.TargetStatus)
		macroText := dictLabel(db, "macro_status", n.MacroStatus)
		if macroText == "" {
			macroText = n.MacroStatus
		}
		notifyText := dictLabel(db, "notify_type", n.NotifyType)
		if notifyText == "" {
			notifyText = n.NotifyType
		}
		actionNodes = append(actionNodes, gin.H{
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

	// submitted_at 单独提取
	var submittedAt, serviceCode string
	rawFormData := parseJSONString(string(order.FormData))
	if m, ok := rawFormData.(map[string]interface{}); ok {
		if v, ok := m["submitted_at"].(string); ok {
			submittedAt = v
		}
		if v, ok := m["service_code"].(string); ok {
			serviceCode = v
		}
	}

	return gin.H{
		"id":                  order.ID,
		"order_no":            order.OrderNo,
		"app_user_id":         order.AppUserID,
		"service_id":          order.ServiceID,
		"service_code":        serviceCode,
		"service_name":        order.ServiceName,
		"contact_name":        order.ContactName,
		"contact_phone":       order.ContactPhone,
		"total_amount":        order.TotalAmount,
		"amount":              order.TotalAmount,
		"currency":            order.Currency,
		"macro_status":        order.MacroStatus,
		"macro_status_text":   macroStatusText,
		"current_stage":       order.CurrentStage,
		"current_stage_text":  currentStageText,
		"payment_status":      order.PaymentStatus,
		"payment_status_text": paymentStatusText,
		"form_data":           rawFormData,
		"form_items":          formItems,
		"submitted_at":        submittedAt,
		"remark":              order.Remark,
		"created_at":          order.CreatedAt,
		"updated_at":          order.UpdatedAt,
		"timelines":           timelineItems,
		"payments":            payments,
		"action_nodes":        actionNodes,
	}
}

// auditStatusText 将审核状态码转换为中文描述。
func auditStatusText(status string) string {
	switch status {
	case models.AuditStatusPending:
		return "待审核"
	case models.AuditStatusApproved:
		return "审核通过"
	case models.AuditStatusRejected:
		return "审核未通过"
	default:
		return ""
	}
}

// ensureOrderAppUser 创建或复用演示客户画像。
// 1.意图 -> 允许未登录 H5 演示环境也能真实写入 app_users 与 orders。
// 2.步骤 -> 优先使用传入 app_user_id，否则按手机号查找，不存在则创建。
// 3.返回 -> 可关联订单的 AppUser 实体。
func ensureOrderAppUser(db *gorm.DB, req CreateOrderRequest) models.AppUser {
	var user models.AppUser
	if req.AppUserID > 0 && db.First(&user, req.AppUserID).Error == nil {
		return user
	}
	phone := strings.TrimSpace(req.ContactPhone)
	if phone != "" && db.Where("phone = ?", phone).First(&user).Error == nil {
		return user
	}
	name := strings.TrimSpace(req.ContactName)
	if name == "" {
		name = "Yesok 贵宾客户"
	}
	user = models.AppUser{Phone: phone, Nickname: name, VipLevel: 1}
	db.Create(&user)
	return user
}

// normalizeFormData 补齐订单表单的业务上下文。
// 1.意图 -> 确保不同服务的 JSON 均带有 service_code 和提交时间。
// 2.步骤 -> 若表单为空则创建 map，并写入服务编码、服务名称、提交时间。
// 3.返回 -> 可序列化进入 orders.form_data 的 map。
func normalizeFormData(data map[string]interface{}, service models.SysService, serviceName string) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["service_code"] = service.ServiceCode
	data["service_name"] = serviceName
	data["submitted_at"] = time.Now().Format(time.RFC3339)
	return data
}

func parseConfigValue(item models.SysConfig) interface{} {
	if item.ValueType == "json" {
		return parseJSONString(item.ConfigValue)
	}
	return item.ConfigValue
}

func parseJSONString(raw string) interface{} {
	if strings.TrimSpace(raw) == "" {
		return gin.H{}
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return gin.H{}
	}
	return payload
}

func marshalJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

// formatMoney 将分单位金额转换为越南盾展示文案。
// 1.意图 -> 为前端提供可读的价格字符串，避免前端重复格式化金额。
// 2.步骤 -> 将分单位换算为主币单位，并拼接越南盾符号。
// 3.返回 -> 例如 650000 ₫ 的展示字符串。
func formatMoney(amount int64) string {
	return fmt.Sprintf("%d ₫", amount/100)
}

// CEndOrderActionRequest C 端执行订单动作的请求体。
type CEndOrderActionRequest struct {
	ActionName string                 `json:"action_name" binding:"required"`
	Remark     string                 `json:"remark"`
	InputData  map[string]interface{} `json:"input_data"`
}

// GetClientOrderActions 返回指定订单当前节点对 client 角色可执行的动作列表。
func GetClientOrderActions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUint(c.Param("id"))
		if err != nil || id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}

		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found")
			return
		}

		var nodes []models.SysWorkflowNode
		db.Where(
			"service_id = ? AND stage_code = ? AND (executor_role = ? OR executor_role = 'both')",
			order.ServiceID, order.CurrentStage, "client",
		).Order("sort_order asc").Find(&nodes)

		// 规范化返回字段，并过滤掉审核相关动作（audit_approve / audit_reject）
		// C 端绝不允许执行审核动作，这些动作仅供 B 端管理员使用
		actions := make([]gin.H, 0, len(nodes))
		for _, n := range nodes {
			if n.ActionName == "audit_approve" || n.ActionName == "audit_reject" || n.ActionName == "audit_rejected" {
				continue
			}
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
				"id":                 n.ID,
				"action_name":        n.ActionName,
				"action_name_text":   actionNameText,
				"button_label":       n.ButtonLabel,
				"action_type":        n.ActionType,
				"form_fields":        n.FormFields,
				"target_status":      n.TargetStatus,
				"target_status_text": targetStatusText,
				"macro_status":       n.MacroStatus,
				"macro_status_text":  macroText,
				"notify_type":        n.NotifyType,
				"notify_type_text":   notifyText,
				"need_audit":         n.NeedAudit,
				"sort_order":         n.SortOrder,
				"stage_code":         n.StageCode,
				"stage_name":         n.StageName,
			})
		}

		c.JSON(http.StatusOK, gin.H{"actions": actions})
	}
}

// ClientListOrders 返回当前 C 端用户的订单列表，支持 status 过滤。
func ClientListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidVal, ok := c.Get("uid")
		if !ok {
			httpError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
			return
		}
		appUserID, ok := uidVal.(uint)
		if !ok || appUserID == 0 {
			httpError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
			return
		}

		status := c.DefaultQuery("status", "all")

		query := db.Where("app_user_id = ?", appUserID)

		switch status {
		case "completed":
			query = query.Where("macro_status = ? OR current_stage = ?", "completed", "completed")
		case "active":
			query = query.Where("macro_status <> ? AND current_stage <> ?", "completed", "completed")
		case "all", "":
			// no extra filter
		default:
			query = query.Where("macro_status = ? OR current_stage = ?", status, status)
		}

		var orders []models.Order
		if err := query.Order("created_at desc").Find(&orders).Error; err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to fetch orders")
			return
		}

		list := make([]gin.H, 0, len(orders))
		for _, order := range orders {
			list = append(list, buildOrderPayloadForRole(db, order, "client"))
		}

		c.JSON(http.StatusOK, gin.H{
			"list":  list,
			"total": len(list),
		})
	}
}

// ClientGetOrder 返回 C 端用户可查看的订单详情。
func ClientGetOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUint(c.Param("id"))
		if err != nil || id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}

		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found")
			return
		}

		c.JSON(http.StatusOK, buildOrderPayloadForRole(db, order, "client"))
	}
}

// PostClientOrderAction C 端执行订单动作（推进状态）。
func PostClientOrderAction(db *gorm.DB, engine *workflow.OrderEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUint(c.Param("id"))
		if err != nil || id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}

		var req CEndOrderActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request: "+err.Error())
			return
		}

		// C 端主动拒绝审核动作
		if req.ActionName == "audit_approve" || req.ActionName == "audit_reject" || req.ActionName == "audit_rejected" {
			httpError(c, http.StatusForbidden, ErrCodeForbidden, "C端不允许执行审核动作")
			return
		}

		var order models.Order
		if err := db.First(&order, id).Error; err != nil {
			httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found")
			return
		}

		operatorID := fmt.Sprintf("client:%d", order.AppUserID)
		if err := engine.AdvanceStage(id, req.ActionName, operatorID, "client", req.InputData, req.Remark); err != nil {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
			return
		}

		db.First(&order, id)
		c.JSON(http.StatusOK, gin.H{
			"message":       "action executed",
			"order_no":      order.OrderNo,
			"macro_status":  order.MacroStatus,
			"current_stage": order.CurrentStage,
		})
	}
}

func parseUint(s string) (uint, error) {
	var v uint
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

// buildFormItems 将 form_data 转换为带中文 label 和 display_value 的列表。
// 1. 查询服务全部 workflow nodes，从所有节点的 form_fields 建立字段定义（按 sort_order 顺序）
// 2. 遍历 form_data，按节点定义顺序生成 form_items
// 3. 系统字段和 _last_* 字段不进入业务资料展示
func buildFormItems(db *gorm.DB, order models.Order) []gin.H {
	// 1. 查询当前服务全部节点（按 sort_order 排序）
	var allNodes []models.SysWorkflowNode
	db.Where("service_id = ?", order.ServiceID).
		Order("sort_order ASC, id ASC").
		Find(&allNodes)

	// 构建 fieldDefs 列表（保留顺序）和 fieldMap（用于快速查找）
	type fieldDefEntry struct {
		key   string
		field models.FormFieldDef
	}
	fieldDefs := make([]fieldDefEntry, 0)
	fieldMap := map[string]models.FormFieldDef{}
	for _, node := range allNodes {
		for _, f := range node.FormFields {
			if f.Key == "" {
				continue
			}
			if _, exists := fieldMap[f.Key]; !exists {
				fieldMap[f.Key] = f
				fieldDefs = append(fieldDefs, fieldDefEntry{key: f.Key, field: f})
			}
		}
	}

	// 2. 解析 form_data
	rawFormData := parseJSONString(string(order.FormData))
	data, ok := rawFormData.(map[string]interface{})
	if !ok {
		return nil
	}

	// 3. 系统字段不作为业务资料展示
	systemKeys := map[string]bool{
		"service_code":        true,
		"service_name":        true,
		"submitted_at":        true,
		"_last_action_name":   true,
		"_last_notify_type":   true,
		"_last_operator_role": true,
		"_last_submitted_at":  true,
	}

	// 4. 构建 form_items（按 fieldDefs 顺序，第一次出现的 key 优先）
	items := make([]gin.H, 0)
	seenKeys := map[string]bool{}
	for _, entry := range fieldDefs {
		if seenKeys[entry.key] {
			continue
		}
		seenKeys[entry.key] = true

		rawValue, exists := data[entry.key]
		if !exists || systemKeys[entry.key] || strings.HasPrefix(entry.key, "_") {
			continue
		}

		field := entry.field
		displayValue := ""

		// select 枚举值转中文
		if field.Type == "select" && len(field.Options) > 0 {
			val := fmt.Sprintf("%v", rawValue)
			for _, opt := range field.Options {
				if opt.Value == val {
					displayValue = opt.Label
					break
				}
			}
		}
		if displayValue == "" {
			displayValue = fmt.Sprintf("%v", rawValue)
		}

		items = append(items, gin.H{
			"key":           entry.key,
			"label":         field.Label,
			"value":         rawValue,
			"display_value": displayValue,
			"type":          field.Type,
		})
	}

	return items
}

// dictLabel 通过字典码和字典值查询中文标签。
func dictLabel(db *gorm.DB, dictCode, dictValue string) string {
	if dictCode == "" || dictValue == "" {
		return ""
	}
	var dictData models.SysDictData
	if err := db.Where("dict_code = ? AND dict_value = ? AND status = ?", dictCode, dictValue, 1).First(&dictData).Error; err != nil {
		return ""
	}
	return dictData.DictLabel
}

// workflowStageLabel 返回工作流节点的中文标签。
// 1.意图 -> 字典缺失时从服务工作流节点取 stage_name，仍缺失时用兜底映射。
// 2.步骤 -> 优先字典 → 查 SysWorkflowNode.stage_name → 兜底 map → 返回原始 code。
// 3.返回 -> 中文标签或原始 stage_code。
func workflowStageLabel(db *gorm.DB, serviceID uint, stageCode string) string {
	stageCode = strings.TrimSpace(stageCode)
	if stageCode == "" {
		return ""
	}

	// 1. 优先查字典
	if label := dictLabel(db, "node_stage", stageCode); label != "" {
		return label
	}

	// 2. 从当前服务工作流节点里取 stage_name
	var node models.SysWorkflowNode
	if err := db.Where(
		"service_id = ? AND stage_code = ?",
		serviceID, stageCode,
	).Order("sort_order ASC, id ASC").First(&node).Error; err == nil {
		if strings.TrimSpace(node.StageName) != "" {
			return strings.TrimSpace(node.StageName)
		}
	}

	// 3. 常用兜底映射
	fallback := map[string]string{
		"start":                    "开始",
		"wait_butler_contact":      "待管家沟通",
		"wait_quote":               "待报价",
		"wait_confirm":             "待用户确认",
		"wait_pay":                 "待支付",
		"wait_deposit_pay":         "待支付定金",
		"deposit_paid":             "定金已支付",
		"paid":                     "尾款已支付",
		"wait_final_pay":           "待支付尾款",
		"wait_upload_material":     "待上传资料",
		"material_review":          "平台资料初审",
		"wait_supplement":          "待补资料",
		"prepare_material":         "准备办理资料",
		"external_review":          "外部审批中",
		"approved":                 "审批通过",
		"rejected":                 "审批拒绝",
		"issued":                   "已出签/已交付",
		"processing":               "办理中",
		"service_following":        "服务跟进中",
		"delivering":               "交付中",
		"wait_delivery_upload":     "待上传交付资料",
		"delivery_review":          "交付资料审核",
		"delivery_confirmed":       "交付资料已审核",
		"aftersale_butler_contact": "待管家沟通",
		"aftersale_processing":     "管家介入中",
		"completed":                "已完成",
	}
	if label, ok := fallback[stageCode]; ok {
		return label
	}

	return stageCode
}

// paymentStatusLabel 返回支付状态的中文标签。
var paymentStatusMap = map[string]string{
	"unpaid":   "未支付",
	"paid":     "已支付",
	"refunded": "已退款",
}

func paymentStatusLabel(status string) string {
	if label, ok := paymentStatusMap[status]; ok {
		return label
	}
	return status
}
