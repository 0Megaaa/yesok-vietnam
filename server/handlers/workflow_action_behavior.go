package handlers

import (
	"github.com/gin-gonic/gin"
	"yesok-vietnam/server/models"
)

// hasUploadField 判断动态表单字段里是否包含图片/文件上传。
// 仅用于前端渲染建议，不参与真实工作流推进。
func hasUploadField(fields models.FormFields) bool {
	for _, f := range fields {
		if f.Type == "image" || f.Type == "file" {
			return true
		}
	}
	return false
}

// isAuditActionName 判断是否为审核语义动作。
// 仅用于前端渲染建议和按钮分组，不参与真实工作流推进。
func isAuditActionName(actionName string) bool {
	return actionName == "audit_approve" ||
		actionName == "audit_reject" ||
		actionName == "audit_rejected"
}

// inferWorkflowRenderer 根据工作流节点字段推导前端渲染器。
// 真实流程仍然只由 workflow engine 使用 action_name / target_status 推进。
func inferWorkflowRenderer(n models.SysWorkflowNode) string {
	if isAuditActionName(n.ActionName) {
		return "audit"
	}

	switch n.ActionType {
	case "form_input":
		return "dynamic_form"
	case "wx_pay":
		return "payment"
	case "button_click":
		return "button"
	default:
		return "button"
	}
}

// inferWorkflowVariant 根据动作语义推导前端按钮样式。
// 仅用于展示。
func inferWorkflowVariant(n models.SysWorkflowNode) string {
	if isAuditActionName(n.ActionName) {
		if n.ActionName == "audit_reject" || n.ActionName == "audit_rejected" {
			return "danger"
		}
		return "success"
	}

	if n.ActionType == "wx_pay" {
		return "pay"
	}

	if hasUploadField(n.FormFields) {
		return "upload"
	}

	switch n.ActionName {
	case "process_failed", "external_rejected":
		return "danger"
	case "request_supplement", "external_supplement":
		return "warning"
	case "complete_order":
		return "success"
	default:
		return "primary"
	}
}

// buildWorkflowUIBehavior 给前端返回通用渲染建议。
// 注意：ui_behavior 只做前端展示/交互提示，不参与真实工作流流转。
func buildWorkflowUIBehavior(n models.SysWorkflowNode) gin.H {
	renderer := inferWorkflowRenderer(n)

	return gin.H{
		"renderer":         renderer,
		"variant":          inferWorkflowVariant(n),
		"confirm_required": n.ActionType == "button_click" && renderer != "audit",
		"need_upload":      hasUploadField(n.FormFields),
		"is_audit":         isAuditActionName(n.ActionName),
	}
}
