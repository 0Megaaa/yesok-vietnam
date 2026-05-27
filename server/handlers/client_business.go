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
// 2.步骤 -> 校验服务、创建演示客户、查询服务首节点、序列化 form_data、生成订单号和初始时间线。
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

		// 查询服务对应的首个流程节点作为初始 stage
		var firstNode models.SysWorkflowNode
		if err := db.Where("service_id = ?", service.ID).Order("sort_order asc, id asc").First(&firstNode).Error; err != nil {
			// 没有配置节点时使用默认值
			firstNode.StageCode = "start"
			firstNode.MacroStatus = string(models.OrderStatusPending)
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
			CurrentStage:  firstNode.StageCode,
			MacroStatus:   firstNode.MacroStatus,
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
				BeforeStatus: "",
				AfterStatus:  order.CurrentStage,
				Operator:     "C端客户",
				Remark:       "客户提交订单，等待管家处理",
				ActionCode:   "客户下单",
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

// buildOrderPayload 统一转换订单聚合信息。
// 1.意图 -> 为后台和 C 端复用同一份订单详情结构。
// 2.步骤 -> 解析 form_data，加载时间线、支付记录和当前动作节点。
// 3.返回 -> 带 JSON 详情、动态按钮和财务流水的订单对象。
func buildOrderPayload(db *gorm.DB, order models.Order) gin.H {
	var timelines []models.OrderTimeline
	var payments []models.PaymentRecord
	var nodes []models.SysWorkflowNode
	db.Where("order_id = ?", order.ID).Order("created_at asc").Find(&timelines)
	db.Where("order_id = ?", order.ID).Order("created_at desc").Find(&payments)
	db.Where("service_id = ? AND stage_code = ?", order.ServiceID, order.MacroStatus).Order("sort_order asc, id asc").Find(&nodes)
	return gin.H{
		"id": order.ID, "order_no": order.OrderNo, "orderNo": order.OrderNo, "app_user_id": order.AppUserID,
		"service_id": order.ServiceID, "serviceId": order.ServiceID, "service_name": order.ServiceName, "serviceName": order.ServiceName,
		"contact_name": order.ContactName, "contact_phone": order.ContactPhone, "total_amount": order.TotalAmount,
		"amount": order.TotalAmount, "price": formatMoney(order.TotalAmount), "currency": order.Currency,
		"current_status": order.MacroStatus, "currentStatus": order.MacroStatus, "payment_status": order.PaymentStatus,
		"form_data": parseJSONString(string(order.FormData)), "formData": parseJSONString(string(order.FormData)), "remark": order.Remark,
		"created_at": order.CreatedAt, "updated_at": order.UpdatedAt, "timelines": timelines, "payments": payments, "actionNodes": nodes,
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
