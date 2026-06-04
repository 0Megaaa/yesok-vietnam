export function safeParseJson(value) {
  if (!value) return null
  if (typeof value === 'object') return value

  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

function defaultDictGroup(dictCode) {
  const map = {
    action_type: '动作类型',
    executor_role: '执行角色',
    workflow_action: '工作流动作',
    macro_status: '订单主状态',
    node_stage: '工作流节点',
    notify_type: '通知类型',
    service_category: '服务分类',
    article_category: '资讯分类',
  }
  return map[dictCode] || '通用字典'
}

function defaultDictScope(dictCode) {
  if (['action_type', 'executor_role', 'workflow_action', 'macro_status', 'node_stage', 'notify_type'].includes(dictCode)) {
    return 'workflow'
  }
  if (dictCode === 'service_category') return 'service'
  if (dictCode === 'article_category') return 'article'
  return 'common'
}

function defaultDesc(dictCode, label, value, remark) {
  if (remark) return remark

  const prefixMap = {
    action_type: '工作流动作类型',
    executor_role: '工作流执行角色',
    workflow_action: '工作流动作',
    macro_status: '订单主状态',
    node_stage: '工作流节点',
    notify_type: '通知方式',
    service_category: '服务分类',
    article_category: '资讯分类',
  }

  const prefix = prefixMap[dictCode] || '字典项'
  return `${prefix}：${label || value || ''}`
}

function workflowActionRenderer(actionName) {
  if (['audit_approve', 'audit_reject', 'audit_rejected'].includes(actionName)) return 'audit'
  if (['pay_order', 'pay_final'].includes(actionName)) return 'payment'
  return 'button'
}

function workflowActionVariant(actionName) {
  if (actionName === 'audit_approve') return 'success'
  if (['audit_reject', 'audit_rejected', 'process_failed', 'external_rejected'].includes(actionName)) return 'danger'
  if (['pay_order', 'pay_final'].includes(actionName)) return 'pay'
  if (['request_supplement', 'external_supplement'].includes(actionName)) return 'warning'
  return 'primary'
}

function workflowActionInteraction(actionName, label) {
  const map = {
    audit_approve: '后台审核通过，订单进入配置的下一节点。',
    audit_reject: '后台审核不通过，订单回退到配置的审核失败节点。',
    audit_rejected: '后台审核不通过，订单回退到配置的审核失败节点。',
    pay_order: 'C端用户支付首笔费用，全款或定金由报价数据决定。',
    pay_final: 'C端用户支付尾款。',
    process_failed: '填写办理失败原因，C端展示红色办理失败状态。',
    complete_order: '将订单推进到完成状态。',
    request_supplement: '要求用户补充资料。',
    external_supplement: '外部审批要求用户补充资料。',
  }
  return map[actionName] || `点击后按工作流配置推进：${label || actionName}`
}

function statusVariant(status) {
  if (['completed', 'paid', 'final_paid'].includes(status)) return 'success'
  if (['failed', 'process_failed', 'cancelled', 'rejected'].includes(status)) return 'danger'
  if (['reviewing', 'processing', 'pending'].includes(status)) return 'warning'
  return 'primary'
}

export function generateDictMetaJson(dictItem = {}) {
  const dictCode = dictItem.dict_code || ''
  const label = dictItem.dict_label || dictItem.label || dictItem.dict_value || ''
  const value = dictItem.dict_value || dictItem.value || ''
  const remark = dictItem.remark || ''

  const meta = {
    label,
    desc: defaultDesc(dictCode, label, value, remark),
    config_group: defaultDictGroup(dictCode),
    scope: defaultDictScope(dictCode),
  }

  if (dictCode === 'action_type') {
    if (value === 'button_click') {
      meta.renderer = 'button'
      meta.variant = 'primary'
      meta.confirm_required = true
      meta.interaction = '点击按钮后直接推进工作流，不收集表单。'
    } else if (value === 'form_input') {
      meta.renderer = 'dynamic_form'
      meta.variant = 'primary'
      meta.confirm_required = false
      meta.interaction = '点击后打开动态表单，提交 input_data 后推进工作流。'
    } else if (value === 'wx_pay') {
      meta.renderer = 'payment'
      meta.variant = 'pay'
      meta.confirm_required = true
      meta.interaction = '点击后发起支付，支付成功后推进工作流。'
    }
  }

  if (dictCode === 'executor_role') {
    meta.role = value
    if (value === 'admin') {
      meta.visible_side = 'web-admin'
      meta.interaction = '仅后台订单详情展示并由管家/管理员执行。'
    } else if (value === 'client') {
      meta.visible_side = 'web-client'
      meta.interaction = '仅C端小程序订单详情展示并由客户执行。'
    } else if (value === 'both') {
      meta.visible_side = 'both'
      meta.interaction = 'B端和C端都可以展示并执行。'
    }
  }

  if (dictCode === 'workflow_action') {
    meta.action_name = value
    meta.renderer = workflowActionRenderer(value)
    meta.variant = workflowActionVariant(value)
    meta.interaction = workflowActionInteraction(value, label)
  }

  if (dictCode === 'macro_status') {
    meta.status_code = value
    meta.variant = statusVariant(value)
    meta.interaction = '订单主状态展示，用于列表筛选和状态提示。'
  }

  if (dictCode === 'node_stage') {
    meta.stage_code = value
    meta.interaction = '工作流节点编码，用于订单当前节点和流转目标节点。'
  }

  if (dictCode === 'notify_type') {
    meta.notify_type = value
    meta.interaction = '动作执行后的通知方式配置。'
  }

  if (dictCode === 'service_category') {
    meta.service_code = value
    meta.interaction = '服务分类，用于服务列表、订单归类和展示。'
  }

  if (dictCode === 'article_category') {
    meta.article_category = value
    meta.interaction = '资讯分类，用于文章列表筛选和展示。'
  }

  return meta
}

export function stringifyMetaJson(meta) {
  return JSON.stringify(meta || {}, null, 2)
}

export function normalizeMetaJsonInput(value) {
  const parsed = safeParseJson(value)
  if (!parsed) return ''
  return stringifyMetaJson(parsed)
}

export function isValidMetaJson(value) {
  if (!value) return true
  return !!safeParseJson(value)
}
