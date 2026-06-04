export const ACTION_RENDERER = {
  BUTTON: 'button',
  DYNAMIC_FORM: 'dynamic_form',
  PAYMENT: 'payment',
  AUDIT: 'audit',
}

export function safeParseMetaJson(value) {
  if (!value) return null
  if (typeof value === 'object') return value
  if (typeof value !== 'string') return null

  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

export function hasUploadField(action = {}) {
  const fields = Array.isArray(action.form_fields) ? action.form_fields : []
  return fields.some((field) => field.type === 'image' || field.type === 'file')
}

export function isAuditAction(action = {}) {
  return action.is_audit_action === true ||
    action.action_name === 'audit_approve' ||
    action.action_name === 'audit_reject' ||
    action.action_name === 'audit_rejected'
}

export function getUiBehavior(action = {}) {
  const behavior = safeParseMetaJson(action.ui_behavior)

  if (behavior) {
    return behavior
  }

  if (isAuditAction(action)) {
    return {
      renderer: ACTION_RENDERER.AUDIT,
      variant: action.action_name === 'audit_reject' || action.action_name === 'audit_rejected'
        ? 'danger'
        : 'success',
      is_audit: true,
    }
  }

  if (action.action_type === 'form_input') {
    return {
      renderer: ACTION_RENDERER.DYNAMIC_FORM,
      variant: hasUploadField(action) ? 'upload' : 'primary',
      need_upload: hasUploadField(action),
    }
  }

  if (action.action_type === 'wx_pay') {
    return {
      renderer: ACTION_RENDERER.PAYMENT,
      variant: 'pay',
    }
  }

  return {
    renderer: ACTION_RENDERER.BUTTON,
    variant: 'primary',
  }
}

export function getActionRenderer(action = {}) {
  return getUiBehavior(action).renderer || ACTION_RENDERER.BUTTON
}

export function getActionVariant(action = {}) {
  return getUiBehavior(action).variant || 'primary'
}

export function groupWorkflowActions(actions = []) {
  const groups = {
    button: [],
    dynamicForm: [],
    payment: [],
    audit: [],
  }

  actions.forEach((action) => {
    const renderer = getActionRenderer(action)

    if (renderer === ACTION_RENDERER.AUDIT) {
      groups.audit.push(action)
    } else if (renderer === ACTION_RENDERER.DYNAMIC_FORM) {
      groups.dynamicForm.push(action)
    } else if (renderer === ACTION_RENDERER.PAYMENT) {
      groups.payment.push(action)
    } else {
      groups.button.push(action)
    }
  })

  return groups
}
