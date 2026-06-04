package handlers

import (
	"encoding/json"
	"strings"

	"yesok-vietnam/server/models"
)

func isEmptyJSONText(v models.JSONText) bool {
	raw := strings.TrimSpace(string(v))
	return raw == "" || raw == "null" || raw == "{}"
}

func makeDictMetaJSON(payload map[string]interface{}) models.JSONText {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return models.JSONText(data)
}

func generateDictMetaJSON(dictCode, dictLabel, dictValue, remark string) models.JSONText {
	dictCode = strings.TrimSpace(dictCode)
	dictLabel = strings.TrimSpace(dictLabel)
	dictValue = strings.TrimSpace(dictValue)
	remark = strings.TrimSpace(remark)

	label := dictLabel
	if label == "" {
		label = dictValue
	}

	base := map[string]interface{}{
		"label":        label,
		"desc":         defaultDictDesc(dictCode, label, dictValue, remark),
		"config_group": defaultDictGroup(dictCode),
		"scope":        defaultDictScope(dictCode),
	}

	switch dictCode {
	case "action_type":
		switch dictValue {
		case "button_click":
			base["renderer"] = "button"
			base["variant"] = "primary"
			base["confirm_required"] = true
			base["interaction"] = "点击按钮后直接推进工作流，不收集表单。"
		case "form_input":
			base["renderer"] = "dynamic_form"
			base["variant"] = "primary"
			base["confirm_required"] = false
			base["interaction"] = "点击后打开动态表单，提交 input_data 后推进工作流。"
		case "wx_pay":
			base["renderer"] = "payment"
			base["variant"] = "pay"
			base["confirm_required"] = true
			base["interaction"] = "点击后发起支付，支付成功后推进工作流。"
		default:
			base["renderer"] = "button"
			base["variant"] = "primary"
			base["interaction"] = "按动作类型配置执行。"
		}

	case "executor_role":
		base["role"] = dictValue
		switch dictValue {
		case "admin":
			base["visible_side"] = "web-admin"
			base["interaction"] = "仅后台订单详情展示并由管家/管理员执行。"
		case "client":
			base["visible_side"] = "web-client"
			base["interaction"] = "仅C端小程序订单详情展示并由客户执行。"
		case "both":
			base["visible_side"] = "both"
			base["interaction"] = "B端和C端都可以展示并执行。"
		default:
			base["visible_side"] = "unknown"
			base["interaction"] = "按执行角色配置展示。"
		}

	case "workflow_action":
		base["action_name"] = dictValue
		base["renderer"] = defaultWorkflowActionRenderer(dictValue)
		base["variant"] = defaultWorkflowActionVariant(dictValue)
		base["interaction"] = defaultWorkflowActionInteraction(dictValue, label)

	case "macro_status":
		base["status_code"] = dictValue
		base["variant"] = defaultStatusVariant(dictValue)
		base["interaction"] = "订单主状态展示，用于列表筛选和状态提示。"

	case "node_stage":
		base["stage_code"] = dictValue
		base["interaction"] = "工作流节点编码，用于订单当前节点和流转目标节点。"

	case "notify_type":
		base["notify_type"] = dictValue
		base["interaction"] = "动作执行后的通知方式配置。"

	case "service_category":
		base["service_code"] = dictValue
		base["interaction"] = "服务分类，用于服务列表、订单归类和展示。"

	case "article_category":
		base["article_category"] = dictValue
		base["interaction"] = "资讯分类，用于文章列表筛选和展示。"

	default:
		base["interaction"] = "通用字典项，用于后台配置和前端展示。"
	}

	return makeDictMetaJSON(base)
}

func defaultDictGroup(dictCode string) string {
	switch dictCode {
	case "action_type":
		return "动作类型"
	case "executor_role":
		return "执行角色"
	case "workflow_action":
		return "工作流动作"
	case "macro_status":
		return "订单主状态"
	case "node_stage":
		return "工作流节点"
	case "notify_type":
		return "通知类型"
	case "service_category":
		return "服务分类"
	case "article_category":
		return "资讯分类"
	default:
		return "通用字典"
	}
}

func defaultDictScope(dictCode string) string {
	switch dictCode {
	case "action_type", "executor_role", "workflow_action", "macro_status", "node_stage", "notify_type":
		return "workflow"
	case "service_category":
		return "service"
	case "article_category":
		return "article"
	default:
		return "common"
	}
}

func defaultDictDesc(dictCode, label, value, remark string) string {
	if remark != "" {
		return remark
	}

	switch dictCode {
	case "action_type":
		return "工作流动作类型：" + label
	case "executor_role":
		return "工作流执行角色：" + label
	case "workflow_action":
		return "工作流动作：" + label
	case "macro_status":
		return "订单主状态：" + label
	case "node_stage":
		return "工作流节点：" + label
	case "notify_type":
		return "通知方式：" + label
	case "service_category":
		return "服务分类：" + label
	case "article_category":
		return "资讯分类：" + label
	default:
		if label != "" {
			return "字典项：" + label
		}
		return "字典项：" + value
	}
}

func defaultWorkflowActionRenderer(actionName string) string {
	switch actionName {
	case "audit_approve", "audit_reject", "audit_rejected":
		return "audit"
	case "pay_order", "pay_final":
		return "payment"
	default:
		return "button"
	}
}

func defaultWorkflowActionVariant(actionName string) string {
	switch actionName {
	case "audit_approve":
		return "success"
	case "audit_reject", "audit_rejected", "process_failed", "external_rejected":
		return "danger"
	case "pay_order", "pay_final":
		return "pay"
	case "request_supplement", "external_supplement":
		return "warning"
	default:
		return "primary"
	}
}

func defaultWorkflowActionInteraction(actionName, label string) string {
	switch actionName {
	case "audit_approve":
		return "后台审核通过，订单进入配置的下一节点。"
	case "audit_reject", "audit_rejected":
		return "后台审核不通过，订单回退到配置的审核失败节点。"
	case "pay_order":
		return "C端用户支付首笔费用，全款或定金由报价数据决定。"
	case "pay_final":
		return "C端用户支付尾款。"
	case "process_failed":
		return "填写办理失败原因，C端展示红色办理失败状态。"
	case "complete_order":
		return "将订单推进到完成状态。"
	case "request_supplement", "external_supplement":
		return "要求用户补充资料。"
	default:
		return "点击后按工作流配置推进：" + label
	}
}

func defaultStatusVariant(status string) string {
	switch status {
	case "completed", "paid", "final_paid":
		return "success"
	case "failed", "process_failed", "cancelled", "rejected":
		return "danger"
	case "reviewing", "processing", "pending":
		return "warning"
	default:
		return "primary"
	}
}
