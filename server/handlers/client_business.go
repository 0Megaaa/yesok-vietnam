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
		if service.FormSchema != nil {
			json.Unmarshal(service.FormSchema, &fields)
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

		appUser := ensureOrderAppUser(db, req)
		formData := normalizeFormData(req.FormData, service)
		order := models.Order{
			OrderNo:       fmt.Sprintf("YS%s%04d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000),
			AppUserID:     appUser.ID,
			ServiceID:     service.ID,
			ServiceName:   service.DisplayName,
			ContactName:   req.ContactName,
			ContactPhone:  req.ContactPhone,
			TotalAmount:   service.BasePrice,
			Currency:      service.Currency,
			CurrentStage:  startNode.TargetStatus,
			MacroStatus:   startNode.MacroStatus,
			PaymentStatus: "unpaid",
			FormData:      marshalJSON(formData),
		}
		if order.ServiceName == "" {
			order.ServiceName = service.ServiceName
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

	actionNodes := make([]gin.H, 0, len(nodes))
	for _, n := range nodes {
		actionNodes = append(actionNodes, gin.H{
			"id":            n.ID,
			"action_name":   n.ActionName,
			"button_label":  n.ButtonLabel,
			"action_type":   n.ActionType,
			"form_fields":   n.FormFields,
			"target_status": n.TargetStatus,
			"macro_status":  n.MacroStatus,
			"notify_type":   n.NotifyType,
			"need_audit":    n.NeedAudit,
			"sort_order":    n.SortOrder,
			"stage_code":    n.StageCode,
			"stage_name":    n.StageName,
		})
	}

	return gin.H{
		"id": order.ID, "order_no": order.OrderNo, "orderNo": order.OrderNo, "app_user_id": order.AppUserID,
		"service_id": order.ServiceID, "serviceId": order.ServiceID,
		"service_name": order.ServiceName, "serviceName": order.ServiceName,
		"contact_name": order.ContactName, "contact_phone": order.ContactPhone,
		"total_amount": order.TotalAmount, "amount": order.TotalAmount, "price": formatMoney(order.TotalAmount),
		"currency":       order.Currency,
		"current_stage":  order.CurrentStage,
		"current_status": order.MacroStatus, "currentStatus": order.MacroStatus,
		"macro_status":   order.MacroStatus,
		"payment_status": order.PaymentStatus,
		"form_data":      parseJSONString(string(order.FormData)), "formData": parseJSONString(string(order.FormData)),
		"remark":     order.Remark,
		"created_at": order.CreatedAt, "updated_at": order.UpdatedAt,
		"timelines": timelines, "payments": payments, "actionNodes": actionNodes,
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
func normalizeFormData(data map[string]interface{}, service models.SysService) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["service_code"] = service.ServiceCode
	data["service_name"] = service.ServiceName
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

		// 规范化返回字段
		actions := make([]gin.H, 0, len(nodes))
		for _, n := range nodes {
			actions = append(actions, gin.H{
				"id":            n.ID,
				"action_name":   n.ActionName,
				"button_label":  n.ButtonLabel,
				"action_type":   n.ActionType,
				"form_fields":   n.FormFields,
				"target_status": n.TargetStatus,
				"macro_status":  n.MacroStatus,
				"notify_type":   n.NotifyType,
				"need_audit":    n.NeedAudit,
				"sort_order":    n.SortOrder,
				"stage_code":    n.StageCode,
				"stage_name":    n.StageName,
			})
		}

		c.JSON(http.StatusOK, gin.H{"actions": actions})
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
