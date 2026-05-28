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
		}

		beforeStage := order.CurrentStage
		payloadBytes, _ := json.Marshal(inputData)

		// Step 4a: NeedAudit=true → 停在当前节点，标记待审核
		if node.NeedAudit && operatorRole == "client" {
			// 合并 inputData 到 form_data，并添加上下文信息
			extra := map[string]interface{}{
				"_last_action_name":   actionName,
				"_last_notify_type":   node.NotifyType,
				"_last_operator_role": operatorRole,
				"_last_submitted_at":  time.Now().Format(time.RFC3339),
			}
			var updates map[string]interface{}
			if node.ActionType == models.ActionTypeFormInput && inputData != nil {
				updates = map[string]interface{}{
					"form_data": mergeJSONMap(order.FormData, inputData, extra),
				}
			} else {
				updates = map[string]interface{}{
					"form_data": mergeJSONMap(order.FormData, nil, extra),
				}
			}

			timeline := models.OrderTimeline{
				OrderID:      order.ID,
				BeforeStatus: beforeStage,
				AfterStatus:  beforeStage,
				Operator:     operator,
				Remark:       fmt.Sprintf("%s | 待后台审核", remark),
				ActionName:   actionName,
				Payload:      payloadBytes,
				AuditStatus:  models.AuditStatusPending,
			}
			if err := tx.Create(&timeline).Error; err != nil {
				return fmt.Errorf("写入时间线记录失败: %w", err)
			}
			if err := tx.Model(&order).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新订单状态失败: %w", err)
			}
			return nil
		}

		// Step 4b: NeedAudit=false → 正常推进
		updates := map[string]interface{}{
			"current_stage": node.TargetStatus,
			"macro_status":  node.MacroStatus,
		}

		// wx_pay：第一阶段模拟支付成功
		if node.ActionType == models.ActionTypeWxPay {
			updates["payment_status"] = "paid"

			// 提取支付金额
			payAmount := order.TotalAmount
			if inputData != nil {
				if amt, ok := inputData["amount"]; ok {
					if f, err := toFloat64(amt); err == nil && f > 0 {
						payAmount = int64(f)
					}
				} else if amt, ok := inputData["quote_amount"]; ok {
					if f, err := toFloat64(amt); err == nil && f > 0 {
						payAmount = int64(f)
					}
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

		// Step 5: 写入时间线记录
		timelineRemark := remark
		if node.NotifyType != "" {
			timelineRemark = fmt.Sprintf("%s [notify:%s]", remark, node.NotifyType)
		}
		timeline := models.OrderTimeline{
			OrderID:      order.ID,
			BeforeStatus: beforeStage,
			AfterStatus:  node.TargetStatus,
			Operator:     operator,
			Remark:       timelineRemark,
			ActionName:   actionName,
			Payload:      payloadBytes,
			AuditStatus:  models.AuditStatusApproved,
		}
		if err := tx.Create(&timeline).Error; err != nil {
			return fmt.Errorf("写入时间线记录失败: %w", err)
		}

		return nil
	})

	return err
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

		// 更新 timeline 状态
		if err := tx.Model(&timeline).Updates(map[string]interface{}{
			"audit_status": models.AuditStatusApproved,
			"remark":       remark,
		}).Error; err != nil {
			return fmt.Errorf("更新时间线状态失败: %w", err)
		}

		// 推进订单状态
		updates := map[string]interface{}{
			"current_stage": node.TargetStatus,
			"macro_status":  node.MacroStatus,
		}
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 写入新的时间线
		newTimeline := models.OrderTimeline{
			OrderID:      orderID,
			BeforeStatus: timeline.BeforeStatus,
			AfterStatus:  node.TargetStatus,
			Operator:     operator,
			Remark:       remark,
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

// validateRequiredFields 校验必填字段
func validateRequiredFields(fields models.FormFields, input map[string]interface{}) error {
	for _, f := range fields {
		if f.Required {
			val, ok := input[f.Key]
			if !ok || isEmptyValue(val) {
				return fmt.Errorf("缺少必填字段：%s", f.Label)
			}
		}
	}
	return nil
}
