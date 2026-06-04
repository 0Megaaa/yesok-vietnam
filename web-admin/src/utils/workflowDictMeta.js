export const actionTypeMeta = {
  button_click: {
    label: '普通按钮',
    desc: '点击后直接推进工作流，不需要填写表单。',
    renderer: 'button',
  },
  form_input: {
    label: '填写表单',
    desc: '点击后填写动态表单，提交 input_data 后推进流程。',
    renderer: 'dynamic_form',
  },
  wx_pay: {
    label: '微信支付',
    desc: 'C端用户支付动作，支付成功后推进流程。',
    renderer: 'payment',
  },
}

export const executorRoleMeta = {
  admin: {
    label: '后台管家',
    desc: '只在B端后台订单详情展示和执行。',
  },
  client: {
    label: 'C端用户',
    desc: '只在小程序订单详情展示和执行。',
  },
  both: {
    label: '双方可执行',
    desc: 'B端和C端都可以展示并执行。',
  },
}

export const workflowActionMeta = {
  audit_approve: {
    label: '审核通过',
    desc: '平台内部审核通过，订单进入下一工作流节点。',
    renderer: 'audit',
  },
  audit_reject: {
    label: '审核不通过',
    desc: '平台内部审核失败，订单回退到配置的审核失败节点。',
    renderer: 'audit',
  },
  send_quote: {
    label: '发送报价',
    desc: '后台填写报价信息，C端用户确认后进入支付或下一节点。',
    renderer: 'dynamic_form',
  },
  pay_order: {
    label: '支付首笔费用',
    desc: 'C端支付全款或定金。',
    renderer: 'payment',
  },
  pay_final: {
    label: '支付尾款',
    desc: 'C端支付尾款。',
    renderer: 'payment',
  },
  process_failed: {
    label: '办理失败',
    desc: '填写办理失败原因，C端展示红色失败状态。',
    renderer: 'dynamic_form',
  },
  complete_order: {
    label: '完成订单',
    desc: '将订单推进到完成状态。',
    renderer: 'button',
  },
}

export function safeParseMetaJson(value) {
  if (!value) return null

  if (typeof value === 'object') {
    return value
  }

  if (typeof value !== 'string') {
    return null
  }

  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

export function getLocalDictMeta(dictCode, value) {
  if (!value) return null

  if (dictCode === 'action_type') {
    return actionTypeMeta[value] || null
  }

  if (dictCode === 'executor_role') {
    return executorRoleMeta[value] || null
  }

  if (dictCode === 'workflow_action') {
    return workflowActionMeta[value] || null
  }

  return null
}

export function getDictMeta(dictCode, item) {
  const dbMeta = safeParseMetaJson(item?.meta_json)

  if (dbMeta) {
    return dbMeta
  }

  const value = item?.dict_value || item?.value
  return getLocalDictMeta(dictCode, value)
}

export function getDictHelp(dictCode, item) {
  const meta = getDictMeta(dictCode, item)

  return (
    meta?.desc ||
    meta?.interaction ||
    item?.remark ||
    item?.dict_label ||
    item?.label ||
    ''
  )
}

export function getDictRenderer(dictCode, item) {
  const meta = getDictMeta(dictCode, item)
  return meta?.renderer || ''
}

export function getDictVariant(dictCode, item) {
  const meta = getDictMeta(dictCode, item)
  return meta?.variant || 'primary'
}
