package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

// OrderEngine 是订单流转核心引擎，负责在事务内完成状态推进、审核门控和时间线记录。
type OrderEngine struct {
	db *gorm.DB
}

// NewOrderEngine 创建并返回 OrderEngine 实例。
func NewOrderEngine(db *gorm.DB) *OrderEngine {
	return &OrderEngine{db: db}
}

// AdvanceStage 是订单流转的唯一入口方法。
// role 用于权限校验：只有节点的 executor_role 含请求角色时才允许执行。
// inputData 在 action_type=form_input 时由前端传入，包含 form_fields 定义的字段值。
//
// NeedAudit 门控逻辑：
//   - NeedAudit=false → 正常推进：current_stage → target_status，macro_status → node.MacroStatus
//   - NeedAudit=true  → 待审核状态：停在当前节点，macro_status 设为 pending_review，
//     写入 timeline entry（status=pending_audit），等待管理员确认
//
// 事务内完成：
//  1. 权限与节点定位（service_id + stage_code + action_name + role）
//  2. form_input 必填项校验（validateFormInput）
//  3. NeedAudit 门控分支
//  4. 更新 orders（仅在 NeedAudit=false 时推进）
//  5. 写入 order_timelines（含 audit payload）
func (e *OrderEngine) AdvanceStage(
	orderID uint,
	actionName string,
	operator string,
	operatorRole string,
	inputData map[string]interface{},
	remark string,
) error {
	if actionName == "" {
		return errors.New("actionName 不能为空")
	}

	// 禁止通过普通工作流入口执行审核动作，必须走专用审核接口
	if actionName == "audit_approve" || actionName == "audit_reject" {
		return errors.New("审核动作必须通过专用审核接口执行")
	}

	err := e.db.Transaction(func(tx *gorm.DB) error {
		// Step 1: 查出当前订单
		var order models.Order
		if err := tx.First(&order, orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("订单 %d 不存在", orderID)
			}
			return fmt.Errorf("查询订单失败: %w", err)
		}

		// Step 2: 定位节点（service_id + stage_code + action_name + role）
		var node models.SysWorkflowNode
		if err := tx.Where(
			"service_id = ? AND stage_code = ? AND action_name = ? AND (executor_role = ? OR executor_role = 'both')",
			order.ServiceID, order.CurrentStage, actionName, operatorRole,
		).First(&node).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("当前角色无权执行此操作或节点配置错误")
			}
			return fmt.Errorf("查询节点配置失败: %w", err)
		}

		// Step 3: form_input 必填项校验
		if node.ActionType == models.ActionTypeFormInput {
			if err := validateRequiredFields(node.FormFields, inputData); err != nil {
				return err
			}
			// send_quote 动作额外校验支付模式和金额
			if err := validatePaymentQuoteFields(actionName, inputData); err != nil {
				return err
			}
		}

		beforeStage := order.CurrentStage
		payloadBytes, _ := json.Marshal(inputData)

		// Step 4a: NeedAudit=true → 进入审核中节点，timeline 标记 pending，等待后台审核
		if node.NeedAudit {
			now := time.Now()
			timelineRemark := strings.TrimSpace(remark)
			if timelineRemark == "" {
				timelineRemark = defaultPendingAuditRemark(actionName, node.ButtonLabel, operatorRole)
			}

			payloadBytes, _ := json.Marshal(inputData)

			updates := map[string]interface{}{
				"current_stage": node.TargetStatus,
				"macro_status":  node.MacroStatus,
				"updated_at":    now,
			}

			// 合并 inputData 到 form_data
			if node.ActionType == models.ActionTypeFormInput && inputData != nil {
				extra := map[string]interface{}{
					"_last_action_name":   actionName,
					"_last_notify_type":   node.NotifyType,
					"_last_operator_role": operatorRole,
					"_last_submitted_at":  time.Now().Format(time.RFC3339),
				}
				updates["form_data"] = mergeJSONMap(order.FormData, inputData, extra)
			}

			if err := tx.Model(&order).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新订单状态失败: %w", err)
			}

			timeline := models.OrderTimeline{
				OrderID:      order.ID,
				BeforeStatus: beforeStage,
				AfterStatus:  node.TargetStatus,
				Operator:     operator,
				Remark:       timelineRemark,
				ActionName:   actionName,
				Payload:      payloadBytes,
				AuditStatus:  models.AuditStatusPending,
			}
			if err := tx.Create(&timeline).Error; err != nil {
				return fmt.Errorf("写入时间线记录失败: %w", err)
			}
			return nil
		}

		// Step 4b: NeedAudit=false → 正常推进
		// 根据支付状态决定是否跳过支付节点
		skippedPayNode := ""
		rawTarget := strings.TrimSpace(node.TargetStatus)
		targetStage := resolveTargetStatusByPayment(order, rawTarget)
		if targetStage != rawTarget {
			skippedPayNode = rawTarget
		}
		macroStatus := resolveMacroStatusByStage(tx, order.ServiceID, targetStage, node.MacroStatus)

		updates := map[string]interface{}{
			"current_stage": targetStage,
			"macro_status":  macroStatus,
		}

		// wx_pay：第一阶段模拟支付成功
		if node.ActionType == models.ActionTypeWxPay {
			// payment_status=paid 表示订单已支付过第一笔费用（全额或定金）
			updates["payment_status"] = "paid"

			// 提取支付金额：根据 actionName 和 form_data 决定
			payAmount := order.TotalAmount

			switch actionName {
			case "pay_order":
				// 客户首次支付：全额模式取 full_amount（回退 quote_amount/amount），定金模式取 deposit_amount
				paymentType := getPaymentType(order)
				if paymentType == "deposit" {
					if amt := getNumberFromMap(inputData, "deposit_amount"); amt > 0 {
						payAmount = amt
					}
				} else {
					if amt := getNumberFromMap(inputData, "full_amount"); amt > 0 {
						payAmount = amt
					} else if amt := getNumberFromMap(inputData, "quote_amount", "amount"); amt > 0 {
						payAmount = amt
					}
				}

			case "pay_final":
				// 客户支付尾款：取 final_amount
				if amt := getNumberFromMap(inputData, "final_amount"); amt > 0 {
					payAmount = amt
				}

			default:
				// 其他 wx_pay 动作兜底取 quote_amount/amount
				if amt := getNumberFromMap(inputData, "quote_amount", "amount"); amt > 0 {
					payAmount = amt
				}
			}

			// 创建支付记录
			tradeNo := fmt.Sprintf("MOCKWX%s%d", time.Now().Format("20060102150405"), order.ID)
			payment := models.PaymentRecord{
				OrderID:      order.ID,
				AppUserID:    order.AppUserID,
				PayerName:    order.ContactName,
				PayAmount:    payAmount,
				PayMethod:    "wx_pay",
				Status:       "success",
				ThirdTradeNo: tradeNo,
			}
			if err := tx.Create(&payment).Error; err != nil {
				return fmt.Errorf("创建支付记录失败: %w", err)
			}
		}

		// 从 inputData 中提取金额字段并写入 order.TotalAmount 和 order.Amount
		if inputData != nil {
			var finalAmount int64
			if amount, ok := inputData["amount"]; ok {
				if f, err := toFloat64(amount); err == nil && f > 0 {
					finalAmount = int64(f)
				}
			} else if amount, ok := inputData["quote_amount"]; ok {
				if f, err := toFloat64(amount); err == nil && f > 0 {
					finalAmount = int64(f)
				}
			}
			if finalAmount > 0 {
				updates["total_amount"] = finalAmount
				updates["amount"] = float64(finalAmount)
			}
		}

		// 使用 mergeJSONMap 合并 inputData 到 form_data，并添加上下文信息
		extra := map[string]interface{}{
			"_last_action_name":   actionName,
			"_last_notify_type":   node.NotifyType,
			"_last_operator_role": operatorRole,
			"_last_submitted_at":  time.Now().Format(time.RFC3339),
		}
		if node.ActionType == models.ActionTypeFormInput && inputData != nil {
			updates["form_data"] = mergeJSONMap(order.FormData, inputData, extra)
		}

		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// Step 5: 写入时间线记录（使用中文业务备注，不暴露 notify_type）
		timelineRemark := strings.TrimSpace(remark)
		if timelineRemark == "" {
			timelineRemark = defaultTimelineRemark(actionName, node.ButtonLabel, operatorRole)
		}
		// 跳过支付节点时，备注要体现已支付
		if skippedPayNode != "" && targetStage == "paid" {
			timelineRemark += "，订单已支付，自动跳过支付节点"
		} else if skippedPayNode != "" && targetStage == "deposit_paid" {
			timelineRemark += "，订单已付定金，自动跳过定金支付节点"
		}
		timeline := models.OrderTimeline{
			OrderID:      order.ID,
			BeforeStatus: beforeStage,
			AfterStatus:  targetStage,
			Operator:     operator,
			Remark:       timelineRemark,
			ActionName:   actionName,
			Payload:      payloadBytes,
			AuditStatus:  "", // 普通工作流动作不写 audit_status，避免前端误显示"审核通过"
		}
		if err := tx.Create(&timeline).Error; err != nil {
			return fmt.Errorf("写入时间线记录失败: %w", err)
		}

		return nil
	})

	return err
}

// resolveMacroStatusByStage 根据服务ID和阶段编码查找节点的宏状态。
func resolveMacroStatusByStage(tx *gorm.DB, serviceID uint, stageCode string, fallback string) string {
	var node models.SysWorkflowNode
	if err := tx.Where(
		"service_id = ? AND stage_code = ?",
		serviceID, stageCode,
	).Order("sort_order ASC, id ASC").First(&node).Error; err == nil {
		if strings.TrimSpace(node.MacroStatus) != "" {
			return strings.TrimSpace(node.MacroStatus)
		}
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return "supplement"
}

// resolveTargetStatusByPayment 根据订单支付状态和支付模式决定是否跳过支付节点。
// payment_status=paid 表示订单已支付过第一笔费用。
//   - 全额模式：paid → 进入 paid
//   - 定金模式：paid → 进入 paid（第一笔为定金），外部审批通过后由 wait_final_pay 处理尾款
func resolveTargetStatusByPayment(order models.Order, targetStatus string) string {
	normalized := strings.TrimSpace(targetStatus)
	if normalized == "" {
		return targetStatus
	}

	paymentStatus := strings.TrimSpace(order.PaymentStatus)
	paymentType := getPaymentType(order)

	switch normalized {
	case "wait_pay":
		// 第一笔费用支付节点。
		// 如果已支付过第一笔费用（全额或定金），直接进入 paid。
		if paymentStatus == "paid" {
			return "paid"
		}
		return "wait_pay"

	case "wait_final_pay":
		// 外部审批通过后的尾款分流。
		// 全额模式不需要尾款，外部审批通过后直接进入 final_paid。
		if paymentType != "deposit" {
			return "final_paid"
		}
		// 定金模式必须进入待支付尾款。
		return "wait_final_pay"
	}

	return normalized
}

// parseOrderFormData 将 orders.form_data JSON 字段解析为 map[string]interface{}。
func parseOrderFormData(order models.Order) map[string]interface{} {
	result := make(map[string]interface{})
	if len(order.FormData) == 0 {
		return result
	}
	_ = json.Unmarshal(order.FormData, &result)
	return result
}

// getStringFromMap 从 data 中按 keys 顺序查找第一个非空字符串值。
func getStringFromMap(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := data[key]; ok && v != nil {
			s := strings.TrimSpace(fmt.Sprintf("%v", v))
			if s != "" {
				return s
			}
		}
	}
	return ""
}

// getNumberFromMap 从 data 中按 keys 顺序查找第一个大于 0 的数值。
func getNumberFromMap(data map[string]interface{}, keys ...string) int64 {
	for _, key := range keys {
		if v, ok := data[key]; ok && v != nil {
			if f, err := toFloat64(v); err == nil && f > 0 {
				return int64(f)
			}
		}
	}
	return 0
}

// getPaymentMode 从订单 form_data 中读取 payment_mode，默认为 "full"。
func getPaymentMode(order models.Order) string {
	formData := parseOrderFormData(order)
	mode := getStringFromMap(formData, "payment_mode", "pay_mode")
	if mode == "" {
		return "full"
	}
	return mode
}

// getPaymentType 从订单 form_data 中读取 payment_type（优先级最高），默认为 "full"。
func getPaymentType(order models.Order) string {
	formData := parseOrderFormData(order)
	paymentType := getStringFromMap(formData, "payment_type", "payment_mode", "pay_mode")
	if paymentType == "" {
		return "full"
	}
	return paymentType
}

// toFloat64 将任意数值类型转为 float64。
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case json.Number:
		return val.Float64()
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("无法转换类型 %T", v)
	}
}

// ApproveAudit 管理员审核通过已提交的待审核节点。
// 将对应 timeline 条目状态改为 approved，并推进订单状态。
func (e *OrderEngine) ApproveAudit(orderID, timelineID uint, operator string, remark string) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		var timeline models.OrderTimeline
		if err := tx.First(&timeline, timelineID).Error; err != nil {
			return fmt.Errorf("时间线记录不存在: %w", err)
		}
		if timeline.OrderID != orderID {
			return errors.New("时间线记录与订单不匹配")
		}
		if timeline.AuditStatus != models.AuditStatusPending {
			return errors.New("该记录不在待审核状态")
		}

		// 定位节点以获取 target_status
		var node models.SysWorkflowNode
		if err := tx.Where(
			"service_id = (SELECT service_id FROM orders WHERE id = ?) AND stage_code = ? AND action_name = ?",
			orderID, timeline.BeforeStatus, timeline.ActionName,
		).First(&node).Error; err != nil {
			return fmt.Errorf("节点配置未找到: %w", err)
		}

		// 查出订单用于支付状态判断
		var order models.Order
		if err := tx.First(&order, orderID).Error; err != nil {
			return fmt.Errorf("订单不存在: %w", err)
		}

		// 更新 timeline 状态
		if err := tx.Model(&timeline).Updates(map[string]interface{}{
			"audit_status": models.AuditStatusApproved,
			"remark":       remark,
		}).Error; err != nil {
			return fmt.Errorf("更新时间线状态失败: %w", err)
		}

		// 根据支付状态决定是否跳过支付节点
		skippedPayNode := ""
		rawTarget := strings.TrimSpace(node.TargetStatus)
		targetStage := resolveTargetStatusByPayment(order, rawTarget)
		if targetStage != rawTarget {
			skippedPayNode = rawTarget
		}
		macroStatus := resolveMacroStatusByStage(tx, order.ServiceID, targetStage, node.MacroStatus)

		// 推进订单状态
		updates := map[string]interface{}{
			"current_stage": targetStage,
			"macro_status":  macroStatus,
		}
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 生成 timeline 备注（跳过支付节点时体现）
		tlRemark := strings.TrimSpace(remark)
		if skippedPayNode != "" && targetStage == "paid" {
			tlRemark += "，订单已支付，自动跳过支付节点"
		} else if skippedPayNode != "" && targetStage == "deposit_paid" {
			tlRemark += "，订单已付定金，自动跳过定金支付节点"
		}

		// 写入新的时间线
		newTimeline := models.OrderTimeline{
			OrderID:      orderID,
			BeforeStatus: timeline.BeforeStatus,
			AfterStatus:  targetStage,
			Operator:     operator,
			Remark:       tlRemark,
			ActionName:   "audit_approved",
			Payload:      timeline.Payload,
			AuditStatus:  models.AuditStatusApproved,
		}
		return tx.Create(&newTimeline).Error
	})
}

// GetAvailableActions 返回指定订单当前节点对给定角色可执行的动作列表。
func (e *OrderEngine) GetAvailableActions(orderID uint, role string) ([]models.SysWorkflowNode, error) {
	var order models.Order
	if err := e.db.First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("订单不存在: %w", err)
	}

	var nodes []models.SysWorkflowNode
	err := e.db.Where(
		"service_id = ? AND stage_code = ? AND (executor_role = ? OR executor_role = 'both')",
		order.ServiceID, order.CurrentStage, role,
	).Order("sort_order asc").Find(&nodes).Error

	return nodes, err
}

// defaultTimelineRemark 根据动作名称和按钮标签返回中文业务备注。
func defaultTimelineRemark(actionName, buttonLabel, operatorRole string) string {
	switch actionName {
	case "submit_request":
		return "客户提交资料"
	case "send_quote":
		return "后台已发送报价"
	case "pay_order":
		return "客户已完成支付"
	case "pay_final":
		return "客户已支付尾款"
	case "external_approved":
		return "外部审批通过"
	case "dispatch_driver":
		return "后台已安排司机"
	case "start_service":
		return "服务已开始"
	case "complete_order":
		return "订单已完成"
	case "upload_material":
		return "客户已上传资料"
	case "material_received":
		return "后台已确认资料收齐"
	case "approve":
		return "后台审核通过"
	case "process_failed":
		return "办理失败"
	default:
		if buttonLabel != "" {
			if operatorRole == "admin" {
				return "后台执行：" + buttonLabel
			}
			if operatorRole == "client" {
				return "客户执行：" + buttonLabel
			}
			return buttonLabel
		}
		return "流程状态已更新"
	}
}

// defaultPendingAuditRemark 根据动作名称返回待审核时间线备注。
func defaultPendingAuditRemark(actionName, buttonLabel, operatorRole string) string {
	switch actionName {
	case "upload_material":
		return "客户已提交资料，等待平台审核"
	case "supplement_material":
		return "客户已补充资料，等待平台审核"
	case "submit_request":
		return "客户提交需求，等待平台处理"
	default:
		if buttonLabel != "" {
			if operatorRole == "client" {
				return "客户已提交「" + buttonLabel + "」，等待平台审核"
			}
			return "已提交「" + buttonLabel + "」，等待平台审核"
		}
		return "资料已提交，等待平台审核"
	}
}

// isEmptyValue 检查值是否为空（nil、空字符串等）
func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []interface{}, map[string]interface{}:
		return val == nil || reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface())
	}
	return false
}

// mergeJSONMap 合并两个 JSON map，并添加额外的键值对
func mergeJSONMap(old []byte, input map[string]interface{}, extra map[string]interface{}) []byte {
	result := make(map[string]interface{})

	// 解析旧的 JSON
	if len(old) > 0 {
		json.Unmarshal(old, &result)
	}

	// 合并 input（兼容 nil）
	if input != nil {
		for k, v := range input {
			result[k] = v
		}
	}

	// 合并 extra（兼容 nil）
	if extra != nil {
		for k, v := range extra {
			result[k] = v
		}
	}

	out, _ := json.Marshal(result)
	return out
}

// validatePaymentQuoteFields 校验 send_quote 动作的支付类型和金额字段。
// 支持 payment_type: full（全款）/ deposit（定金+尾款）。
func validatePaymentQuoteFields(actionName string, inputData map[string]interface{}) error {
	if actionName != "send_quote" {
		return nil
	}

	paymentType := getStringFromMap(inputData, "payment_type", "payment_mode", "pay_mode")
	if paymentType == "" {
		return errors.New("请选择支付类型")
	}

	switch paymentType {
	case "full":
		// 全款模式：必须填写全款金额
		fullAmount := getNumberFromMap(inputData, "full_amount")
		if fullAmount <= 0 {
			return errors.New("选择全款支付时，请填写全款金额")
		}
		// 统一写入 quote_amount，方便后续支付金额读取
		inputData["quote_amount"] = fullAmount

	case "deposit":
		// 定金+尾款模式：必须填写定金和尾款金额
		depositAmount := getNumberFromMap(inputData, "deposit_amount")
		finalAmount := getNumberFromMap(inputData, "final_amount")

		if depositAmount <= 0 {
			return errors.New("选择定金尾款支付时，请填写定金金额")
		}

		if finalAmount <= 0 {
			return errors.New("选择定金尾款支付时，请填写尾款金额")
		}

		// 统一写入 quote_amount = 定金 + 尾款
		inputData["quote_amount"] = depositAmount + finalAmount

	default:
		return errors.New("支付类型不合法")
	}

	return nil
}

// matchFormCondition 判断 input[key] 是否满足 condition。
// condition 格式：{"key": "payment_type", "value": "full"}
// value 支持单值或数组：{"key": "x", "value": ["a","b"]}
func matchFormCondition(input map[string]interface{}, condition map[string]interface{}) bool {
	if condition == nil {
		return true
	}

	keyRaw, ok := condition["key"]
	if !ok {
		return true
	}

	key := strings.TrimSpace(fmt.Sprintf("%v", keyRaw))
	if key == "" {
		return true
	}

	expected, ok := condition["value"]
	if !ok {
		return true
	}

	actual, ok := input[key]
	if !ok {
		return false
	}

	// expected 为数组时，actual 在数组中即匹配
	if arr, isArr := expected.([]interface{}); isArr {
		actualStr := fmt.Sprintf("%v", actual)
		for _, e := range arr {
			if fmt.Sprintf("%v", e) == actualStr {
				return true
			}
		}
		return false
	}

	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// isFormFieldRequired 判断字段在当前 input 下是否为必填。
// 满足 required=true 或 required_when 条件之一即为必填。
func isFormFieldRequired(field models.FormFieldDef, input map[string]interface{}) bool {
	if field.Required {
		return true
	}
	if field.RequiredWhen != nil {
		return matchFormCondition(input, field.RequiredWhen)
	}
	return false
}

// validateRequiredFields 校验必填字段。
// 跳过 visible_when 条件不满足的字段；满足 required_when 条件才算必填。
func validateRequiredFields(fields models.FormFields, input map[string]interface{}) error {
	for _, f := range fields {
		// visible_when 条件不满足时跳过
		if f.VisibleWhen != nil && !matchFormCondition(input, f.VisibleWhen) {
			continue
		}
		if isFormFieldRequired(f, input) {
			val, ok := input[f.Key]
			if !ok || isEmptyFormValue(val) {
				return fmt.Errorf("缺少必填字段：%s", f.Label)
			}
		}
	}
	return nil
}

// isEmptyFormValue 判断表单值是否为空，支持 string / object / array。
func isEmptyFormValue(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case map[string]interface{}:
		if url, ok := val["url"].(string); ok && url != "" {
			return false
		}
		return true
	case []interface{}:
		if len(val) == 0 {
			return true
		}
		// 非空数组视为有值
		return false
	default:
		return false
	}
}
