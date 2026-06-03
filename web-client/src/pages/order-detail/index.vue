<script setup>
import { computed, onMounted, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'
import { get, post, ORIGIN_URL } from '@/api/request'

const client = useClientStore()
const orderId = ref('')
const order = ref(null)
const actions = ref([])
const loading = ref(true)
const actionsLoading = ref(false)
const submitting = ref(false)



// isLocalTempFile 判断是否是本地临时文件路径（不上传后端则不走此路径）
const isLocalTempFile = (url) => {
  if (!url || typeof url !== 'string') return false
  return (
    url.startsWith('wxfile://') ||
    url.startsWith('http://tmp/') ||
    url.startsWith('file://') ||
    url.includes('/tmp/') ||
    url.includes('/temp/')
  )
}

// toFullFileUrl：静态资源用 ORIGIN_URL 拼接，本地临时文件直接返回
const toFullFileUrl = (url) => {
  if (!url) return ''
  if (isLocalTempFile(url)) return url
  if (/^https?:\/\//.test(url)) return url
  const origin = String(ORIGIN_URL || '').replace(/\/+$/, '')
  const path = url.startsWith('/') ? url : `/${url}`
  return `${origin}${path}`
}

// previewUploadedFile 预览图片（支持本地临时和后端 URL）
const previewUploadedFile = (url, urls = []) => {
  const fullUrl = toFullFileUrl(url)
  if (!fullUrl) return
  const fullUrls = urls.length ? urls.map(toFullFileUrl) : [fullUrl]
  uni.previewImage({ urls: fullUrls, current: fullUrl })
}

// getFileUrl 从字符串/对象/数组中提取文件 URL
const getFileUrl = (value) => {
  if (!value) return ''
  if (typeof value === 'string') return value
  if (Array.isArray(value)) {
    const first = value[0]
    if (typeof first === 'string') return first
    return first?.url || first?.path || ''
  }
  if (typeof value === 'object') {
    return value.url || value.path || ''
  }
  return ''
}

// isMaterialFile 判断是否为平台素材文件
const isMaterialFile = (value) => {
  const url = getFileUrl(value)
  if (!url) return false
  return url.includes('/material/') || url.includes('/uploads/') || isLocalTempFile(url)
}

// isImageFile 判断是否为图片文件
const isImageFile = (value) => {
  const url = getFileUrl(value)
  if (!url) return false
  if (isLocalTempFile(url)) return true
  return /\.(jpg|jpeg|png)$/i.test(url)
}

// isImageEntry 判断 entry 是否为图片字段
const isImageEntry = (entry) => {
  if (!entry) return false
  if (entry.type === 'image' || entry.type === 'file') {
    return isImageFile(entry.raw_value || entry.value)
  }
  return isImageFile(entry.raw_value || entry.value)
}

// getFileUrls 提取多个文件 URL（支持未来数组场景）
const getFileUrls = (value) => {
  if (!value) return []
  if (typeof value === 'string') return [value]
  if (Array.isArray(value)) {
    return value
      .map((item) => {
        if (typeof item === 'string') return item
        return item?.url || item?.path || ''
      })
      .filter(Boolean)
  }
  if (typeof value === 'object') {
    const url = value.url || value.path || ''
    return url ? [url] : []
  }
  return []
}

// getOptionLabel 返回 select 字段选中项的中文 label
const getOptionLabel = (field, value) => {
  const option = (field.options || []).find((item) => item.value === value)
  return option?.label || value || ''
}

// unwrapResponse 统一解包 request.js 返回的 { data: ... } 结构
const unwrapResponse = (res) => {
  if (!res) return {}
  return res.data ?? res
}

const statusMap = {
  start: '开始',
  pending: '等待受理',
  reviewing: '资料审核中',
  quoted: '已报价',
  paid: '尾款已支付',
  unpaid: '未支付',
  deposit_paid: '定金已支付',
  wait_quote: '待报价',
  wait_confirm: '待用户确认',
  wait_pay: '待支付',
  wait_deposit_pay: '待支付定金',
  wait_final_pay: '待支付尾款',
  wait_butler_contact: '待管家沟通',
  wait_upload_material: '待上传资料',
  material_review: '平台资料初审',
  wait_supplement: '待补资料',
  prepare_material: '准备办理资料',
  external_review: '外部审批中',
  approved: '审批通过',
  rejected: '审批拒绝',
  issued: '已出签/已交付',
  dispatching: '派车中',
  in_progress: '服务中',
  processing: '服务中',
  service_following: '服务跟进中',
  delivering: '交付中',
  wait_delivery_upload: '待上传交付资料',
  delivery_review: '交付资料审核',
  delivery_confirmed: '交付资料已审核',
  aftersale_butler_contact: '待管家沟通',
  aftersale_processing: '管家介入中',
  completed: '已完成',
  cancelled: '已取消',
  refunded: '已退款',
}

const statusLabel = (code) => statusMap[code] || code || '未知'

// normalizeTimeline 标准化 timeline 字段，确保中文状态兜底
const normalizeTimeline = (tl = {}) => {
  const afterStatus = tl.after_status || tl.afterStatus || ''
  const beforeStatus = tl.before_status || tl.beforeStatus || ''

  return {
    ...tl,
    before_status: beforeStatus,
    after_status: afterStatus,
    before_status_text: tl.before_status_text || tl.beforeStatusText || statusLabel(beforeStatus),
    after_status_text: tl.after_status_text || tl.afterStatusText || statusLabel(afterStatus),
    audit_status: tl.audit_status || tl.auditStatus || '',
    audit_status_text: tl.audit_status_text || tl.auditStatusText || '',
    audit_remark: tl.audit_remark || tl.auditRemark || '',
    action_name: tl.action_name || tl.actionName || '',
    created_at: tl.created_at || tl.createdAt || '',
    updated_at: tl.updated_at || tl.updatedAt || '',
  }
}

// normalizeOrder 标准化后端返回字段，保留完整 action_nodes 信息
const normalizeOrder = (raw = {}) => {
  const timelines = raw.timelines || raw.timeline_items || raw.timelineItems || []
  return {
    ...raw,
    action_nodes: raw.action_nodes || raw.actionNodes || [],
    form_items: raw.form_items || [],
    form_data: raw.form_data || raw.formData || {},
    timelines: timelines.map(normalizeTimeline),
    current_stage: raw.current_stage || raw.currentStage || '',
    current_stage_text: raw.current_stage_text || raw.currentStageText || statusLabel(raw.current_stage || raw.currentStage || ''),
    macro_status: raw.macro_status || raw.current_status || raw.currentStatus || '',
    macro_status_text: raw.macro_status_text || raw.macroStatusText || statusLabel(raw.macro_status || raw.current_status || raw.currentStatus || ''),
    total_amount: raw.total_amount ?? raw.amount ?? 0,
    amount: raw.amount ?? raw.total_amount ?? 0,
    payment_status: raw.payment_status || 'unpaid',
    payment_status_text: raw.payment_status_text || '',
    payment_info: raw.payment_info || {},
  }
}

const formatMoney = (amount) => {
  if (amount === null || amount === undefined || amount === '') return '—'
  const n = Number(amount || 0)
  return `¥${n.toLocaleString('zh-CN')}`
}

const formatTime = (t) => {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN', { dateStyle: 'short', timeStyle: 'short' })
}

const paymentInfo = computed(() => order.value?.payment_info || {})

// paymentSummaryRows 订单摘要中的支付行
const paymentSummaryRows = computed(() => {
  const p = paymentInfo.value
  if (!p || !p.payment_type) {
    return [{ label: '订单金额', value: formatMoney(order.value?.total_amount || 0) }]
  }
  if (p.payment_type === 'deposit') {
    return [
      { label: '支付类型', value: p.payment_type_text || '定金尾款支付' },
      { label: '定金金额', value: formatMoney(p.deposit_amount || 0) },
      { label: '尾款金额', value: formatMoney(p.final_amount || 0) },
      { label: '已付金额', value: formatMoney(p.paid_amount || 0) },
    ]
  }
  return [
    { label: '支付类型', value: p.payment_type_text || '全款支付' },
    { label: '全款金额', value: formatMoney(p.full_amount || p.quote_amount || 0) },
    { label: '已付金额', value: formatMoney(p.paid_amount || 0) },
  ]
})

// getActionPayAmount 从 action.pay_amount 优先取金额，兜底从 payment_info 算
const getActionPayAmount = (action) => {
  if (action?.pay_amount || action?.payAmount) {
    return Number(action.pay_amount || action.payAmount || 0)
  }
  const p = paymentInfo.value
  if (action?.action_name === 'pay_final') {
    return Number(p.final_amount || 0)
  }
  if (action?.action_name === 'pay_order') {
    if (p.payment_type === 'deposit') {
      return Number(p.deposit_amount || 0)
    }
    return Number(p.full_amount || p.quote_amount || order.value?.total_amount || 0)
  }
  return Number(order.value?.total_amount || 0)
}


// 业务资料：保留 type/raw_value/display_value 以支持图片识别
const formEntries = computed(() => {
  if (Array.isArray(order.value?.form_items) && order.value.form_items.length) {
    return order.value.form_items.map((item) => ({
      key: item.key,
      label: item.label || item.key,
      type: item.type || 'text',
      value: item.display_value ?? item.value ?? '',
      display_value: item.display_value ?? item.value ?? '',
      raw_value: item.value ?? item.display_value ?? '',
    }))
  }
  return []
})

// deliveryEntries 展示交付资料（C端不显示上传按钮，由B端上传交付资料后在此查看）
const deliveryEntries = computed(() => {
  const fd = order.value?.form_data || {}
  const entries = []
  if (fd.business_license_image) {
    entries.push({ key: 'business_license_image', label: '营业执照', value: fd.business_license_image, type: 'image' })
  }
  if (fd.tax_registration_image) {
    entries.push({ key: 'tax_registration_image', label: '税务登记', value: fd.tax_registration_image, type: 'image' })
  }
  if (fd.bank_account_assist_image) {
    entries.push({ key: 'bank_account_assist_image', label: '开户协助', value: fd.bank_account_assist_image, type: 'image' })
  }
  if (fd.delivery_remark) {
    entries.push({ key: 'delivery_remark', label: '交付说明', value: fd.delivery_remark, type: 'text' })
  }
  return entries
})

const timelineEntries = computed(() => {
  const list = order.value?.timelines || []
  return [...list].map(normalizeTimeline).reverse()
})

// getTimelineTime 提取 timeline 时间戳用于排序
const getTimelineTime = (item) => {
  return new Date(item?.created_at || item?.updated_at || 0).getTime()
}

// ─── 失败类判断工具（全局通用，不写死 service_id）─────────────────────────────
const FAILURE_ACTIONS = [
  'audit_reject',
  'external_rejected',
  'process_failed',
  'request_supplement',
  'external_supplement',
]

const FAILURE_STATUSES = [
  'rejected',
  'process_failed',
  'wait_supplement',
  'aftersale_butler_contact',
]

const FAILURE_KEYWORDS = [
  '失败',
  '不通过',
  '拒绝',
  '驳回',
  '异常',
  '补资料',
  '需补',
  '补充资料',
]

// isFailureTimeline 判断某条 timeline 是否为失败类
const isFailureTimeline = (item) => {
  if (!item) return false
  if (item.audit_status === 'rejected') return true
  if (FAILURE_ACTIONS.includes(item.action_name)) return true
  if (FAILURE_STATUSES.includes(item.after_status)) return true
  const text = [
    item.remark,
    item.audit_remark,
    item.after_status_text,
    item.before_status_text,
  ].filter(Boolean).join(' ')
  return FAILURE_KEYWORDS.some((keyword) => text.includes(keyword))
}

// failureTitleOf 根据动作名称返回失败/异常类型的展示标题
const failureTitleOf = (item) => {
  const action = item?.action_name || ''
  const afterStatus = item?.after_status || ''
  if (action === 'audit_reject' || item?.audit_status === 'rejected') return '审核未通过'
  if (action === 'external_rejected' || afterStatus === 'rejected') return '审批拒绝'
  if (action === 'process_failed' || afterStatus === 'process_failed') return '办理失败'
  if (action === 'request_supplement' || action === 'external_supplement' || afterStatus === 'wait_supplement') return '需补资料'
  if (afterStatus === 'aftersale_butler_contact') return '管家介入'
  return '异常'
}

// extractFailureReason 统一提取失败原因
const extractFailureReason = (item) => {
  if (!item) return ''
  if (item.audit_remark) return item.audit_remark
  if (item.payload) {
    try {
      const payload = typeof item.payload === 'string' ? JSON.parse(item.payload) : item.payload
      if (payload.failed_reason) return payload.failed_reason
      if (payload.external_reject_reason) return payload.external_reject_reason
      if (payload.external_supplement_reason) return payload.external_supplement_reason
      if (payload.supplement_reason) return payload.supplement_reason
      if (payload.audit_remark) return payload.audit_remark
      if (payload.remark) return payload.remark
      if (payload.reason) return payload.reason
    } catch (e) {}
  }
  return String(item.remark || '')
    .replace(/^审核未通过[:：]\s*/, '')
    .replace(/^审核不通过[:：]\s*/, '')
    .replace(/^办理失败[:：]\s*/, '')
    .replace(/^审批拒绝[:：]\s*/, '')
    .replace(/^审批需补资料[:：]\s*/, '')
    .replace(/^要求补资料[:：]\s*/, '')
    .trim()
}

// isAuditTimeline 判断是否为真正的审核类动作
const isAuditTimeline = (tl) => {
  if (!tl) return false
  if (tl.is_audit_timeline === true) return true
  return (
    tl.audit_status === 'pending' ||
    tl.action_name === 'audit_approve' ||
    tl.action_name === 'audit_reject' ||
    tl.action_name === 'upload_material' ||
    tl.action_name === 'supplement_material'
    // upload_delivery_material 是普通交付动作，B端上传，C端查看，不属于审核动作
  )
}

// shouldShowAuditApproved 仅在真正的审核类动作且审核通过时显示绿色"审核通过"
const shouldShowAuditApproved = (tl) => {
  if (!tl) return false
  if (isFailureTimeline(tl)) return false
  if (!isAuditTimeline(tl)) return false
  return tl.audit_status === 'approved'
}

// currentFailureTimeline 仅展示当前仍处于失败状态的最新一条 timeline
const currentFailureTimeline = computed(() => {
  const currentStage = order.value?.current_stage || ''
  const macroStatus = order.value?.macro_status || ''
  const list = [...(order.value?.timelines || [])]
    .filter(isFailureTimeline)
    .sort((a, b) => {
      const bt = getTimelineTime(b)
      const at = getTimelineTime(a)
      if (bt !== at) return bt - at
      return Number(b.id || 0) - Number(a.id || 0)
    })
  const latest = list[0] || null
  if (!latest) return null
  // 当前处于这些状态时展示失败提示
  if (
    currentStage === 'wait_supplement' ||
    currentStage === 'aftersale_butler_contact' ||
    currentStage === 'rejected' ||
    macroStatus === 'aftersale' ||
    macroStatus === 'rejected'
  ) {
    return latest
  }
  // 审核拒绝状态也展示
  if (latest.audit_status === 'rejected') return latest
  return null
})

const currentFailureTitle = computed(() => {
  const item = currentFailureTimeline.value
  return item ? failureTitleOf(item) : ''
})

const currentFailureReason = computed(() => {
  const item = currentFailureTimeline.value
  return item ? extractFailureReason(item) : ''
})

// extractAuditRemarkFromPayload 从 payload 中提取审核备注
const extractAuditRemarkFromPayload = (payload) => {
  if (!payload) return ''
  try {
    const data = typeof payload === 'string' ? JSON.parse(payload) : payload
    return data?.audit_remark || data?.remark || data?.reason || ''
  } catch (e) {
    return ''
  }
}

// auditTimelines 收集所有审核相关 timeline 并按时间倒序
const auditTimelines = computed(() => {
  const list = order.value?.timelines || []
  return list
    .filter((item) => ['pending', 'approved', 'rejected'].includes(item.audit_status))
    .sort((a, b) => {
      const bt = getTimelineTime(b)
      const at = getTimelineTime(a)
      if (bt !== at) return bt - at
      return Number(b.id || 0) - Number(a.id || 0)
    })
})

// latestAuditTimeline 最近一次审核 timeline
const latestAuditTimeline = computed(() => auditTimelines.value[0] || null)

// isProcessFailedTimeline 判断某条 timeline 是否为办理失败
const isProcessFailedTimeline = (item) => item?.action_name === 'process_failed'

// timelineRemarkText 生成 timeline 展示文本：办理失败时显示 failed_reason
const timelineRemarkText = (item) => {
  if (!item) return ''
  if (isProcessFailedTimeline(item)) {
    if (item.payload) {
      try {
        const payload = typeof item.payload === 'string' ? JSON.parse(item.payload) : item.payload
        if (payload?.failed_reason) return payload.failed_reason
      } catch (e) {}
    }
    const formData = order.value?.form_data || order.value?.formData || {}
    if (formData.failed_reason) return formData.failed_reason
    return item.remark || '办理失败'
  }
  return item.remark || ''
}

// timelineAuditRejectReason 提取时间线中审核失败的原因
const timelineAuditRejectReason = (item) => {
  if (!item || item.audit_status !== 'rejected') return ''
  return (
    item.audit_remark ||
    extractAuditRemarkFromPayload(item.payload) ||
    String(item.remark || '')
      .replace(/^审核未通过[:：]\s*/, '')
      .replace(/^资料审核未通过[:：]\s*/, '')
      .trim()
  )
}

const buttonClickActions = computed(() => actions.value.filter(a => a.action_type === 'button_click'))
const formInputActions = computed(() => actions.value.filter(a => a.action_type === 'form_input'))
const wxPayActions = computed(() => actions.value.filter(a => a.action_type === 'wx_pay'))

const safeToast = (title, icon = 'info') => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.showToast) {
    uniApi.showToast({ title, icon })
  } else {
    console.info(`[Toast] ${title}`)
  }
}

const loadOrderDetail = async () => {
  loading.value = true
  try {
    const res = await get(`/v1/client/orders/${orderId.value}`)
    const payload = unwrapResponse(res)
    const normalized = normalizeOrder(payload.order || payload)
    order.value = normalized

    // 临时调试日志：确认接口返回和本地兜底生效
    console.log('[order-detail] raw payload:', payload)
    console.log('[order-detail] normalized timelines:', normalized.timelines)

    // 后端订单详情已经返回 action_nodes，先用它兜底渲染按钮
    if (Array.isArray(normalized.action_nodes)) {
      actions.value = normalized.action_nodes
    }
  } catch (error) {
    console.error('[order-detail] loadOrderDetail failed:', error)
    safeToast('订单详情加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const loadOrderActions = async () => {
  if (!orderId.value) return
  actionsLoading.value = true
  try {
    const res = await get(`/v1/client/orders/${orderId.value}/actions`)
    const payload = unwrapResponse(res)

    actions.value =
      payload.actions ||
      payload.action_nodes ||
      payload.actionNodes ||
      []
  } catch (error) {
    console.warn('[order-detail] loadOrderActions failed:', error)
    // 如果单独 actions 接口失败，不要直接清空；
    // 订单详情里的 action_nodes 仍可作为兜底。
    actions.value = order.value?.action_nodes || []
  } finally {
    actionsLoading.value = false
  }
}

const executeButtonClick = async (action) => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.showModal) {
    uniApi.showModal({
      title: `执行「${action.button_label}」`,
      content: '确认执行此操作？',
      success: async (res) => {
        if (res.confirm) {
          await performAction(action, {})
        }
      },
    })
  } else {
    await performAction(action, {})
  }
}

const openFormInput = (action) => {
  if (!orderId.value || !action?.action_name) {
    safeToast('操作信息异常，请稍后重试', 'none')
    return
  }

  uni.navigateTo({
    url: `/pages/dynamic-form/index?mode=order&order_id=${orderId.value}&action_name=${encodeURIComponent(action.action_name)}`,
  })
}

const closeForm = () => {
  formVisible.value = false
  currentAction.value = null
  formData.value = {}
  localFileMap.value = {}
  uploadingMap.value = {}
  resetKeyboardSheet()
}

const validateField = (field) => {
  if (!isFieldVisible(field)) return true
  if (!isFieldRequired(field)) return true

  const val = getFormValue(field)

  if (field.type === 'image' || field.type === 'file') {
    if (!val || typeof val !== 'string' || val.trim() === '') {
      safeToast(`请上传${field.label}`, 'none')
      return false
    }
    return true
  }

  const str = (val || '').toString().trim()
  if (!str) {
    safeToast(`请填写 ${field.label}`, 'none')
    return false
  }

  return true
}

const submitForm = async () => {
  if (submitting.value) return

  const fields = normalizeActionFields(currentAction.value)

  for (const field of fields) {
    if (!validateField(field)) return
  }

  submitting.value = true

  try {
    uni.showLoading({ title: '提交中...' })

    const finalInputData = await uploadPendingLocalFiles(fields)

    const ok = await performAction(currentAction.value, finalInputData)

    if (ok) {
      closeForm()
    }
  } catch (error) {
    uni.showToast({ title: error?.message || '提交失败', icon: 'none' })
  } finally {
    uni.hideLoading()
    submitting.value = false
  }
}

const executeWxPay = async (action) => {
  if (submitting.value) return

  const uniApi = typeof uni !== 'undefined' ? uni : null
  const payAmount = getActionPayAmount(action)

  const confirmPay = async () => {
    await simulateWxPay(action, payAmount)
  }

  if (uniApi?.showModal) {
    uniApi.showModal({
      title: action.button_label || '立即支付',
      content: `确认支付 ${formatMoney(payAmount)}？`,
      confirmText: '确认支付',
      cancelText: '取消',
      success: async (res) => {
        if (res.confirm) {
          await confirmPay()
        }
      },
    })
  } else {
    await confirmPay()
  }
}

// TODO: 正式接入微信支付时改造这里：
// 1. 调后端创建支付单，获取 prepay 参数
// 2. 调 uni.requestPayment(wxPayParams)
// 3. 支付成功后调用 /client/orders/:id/action 或专门的支付确认接口
// 4. 支付失败/取消时不推进工作流
const simulateWxPay = async (action, payAmount) => {
  const uniApi = typeof uni !== 'undefined' ? uni : null

  submitting.value = true
  if (uniApi?.showLoading) uniApi.showLoading({ title: '支付处理中...' })

  try {
    await post(`/v1/client/orders/${orderId.value}/action`, {
      action_name: action.action_name,
      remark: '客户已完成支付',
      input_data: {},
    })

    if (uniApi?.hideLoading) uniApi.hideLoading()
    safeToast('支付成功', 'success')

    await loadOrderDetail()
    await loadOrderActions()
  } catch (error) {
    if (uniApi?.hideLoading) uniApi.hideLoading()
    console.error('[order-detail] simulateWxPay failed:', error)
    safeToast(error?.error || error?.message || '支付失败，请稍后重试', 'none')
  } finally {
    submitting.value = false
  }
}

const performAction = async (action, inputData) => {
  try {
    await post(`/v1/client/orders/${orderId.value}/action`, {
      action_name: action.action_name,
      remark: '',
      input_data: inputData,
    })
    safeToast('操作成功', 'success')
    await loadOrderDetail()
    await loadOrderActions()
    return true
  } catch (err) {
    safeToast(err?.message || '操作失败', 'error')
    return false
  }
}

const goBack = () => {
  const pages = getCurrentPages()
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (pages.length > 1) {
    if (uniApi?.navigateBack) uniApi.navigateBack()
  } else {
    if (uniApi?.switchTab) uniApi.switchTab({ url: '/pages/profile/index' })
  }
}

onLoad((options) => {
  orderId.value = options.id || ''
})

onMounted(async () => {
  if (orderId.value) {
    await Promise.all([loadOrderDetail(), loadOrderActions()])
  } else {
    loading.value = false
  }
})
</script>

<template>
  <view class="detail-page">
    <!-- Hero -->
    <view class="hero">
      <view class="nav-back" @click="goBack">‹</view>
      <text class="hero-icon">{{ order?.service_name?.charAt(0) || '📋' }}</text>
    </view>

    <!-- Loading -->
    <view v-if="loading" class="loading-card">
      <text>正在加载订单详情...</text>
    </view>

    <template v-else-if="order">
      <!-- 订单摘要 -->
      <view class="summary-card">
        <view class="summary-head">
          <view class="service-icon">{{ order.service_name?.charAt(0) || '📋' }}</view>
          <view class="summary-main">
            <text class="service-name">{{ order.service_name || '服务订单' }}</text>
            <text class="order-no">{{ order.order_no || order.orderNo }}</text>
          </view>
          <view class="status-pill" :class="order.macro_status || order.current_status">
            {{ order.macro_status_text || statusLabel(order.macro_status || order.current_status) }}
          </view>
        </view>
        <view class="summary-grid">
          <view class="grid-item">
            <text class="label">联系人</text>
            <text class="value">{{ order.contact_name || '—' }}</text>
          </view>
          <view class="grid-item">
            <text class="label">联系电话</text>
            <text class="value">{{ order.contact_phone || '—' }}</text>
          </view>
          <view
            v-for="row in paymentSummaryRows"
            :key="row.label"
            class="grid-item"
          >
            <text class="label">{{ row.label }}</text>
            <text class="value price">{{ row.value }}</text>
          </view>
        </view>
      </view>

      <!-- 统一失败提示（展示当前仍处于失败状态的最新 timeline） -->
      <view v-if="currentFailureTimeline" class="audit-reject-card">
        <text class="audit-reject-title">{{ currentFailureTitle }}</text>
        <text class="audit-reject-desc">
          {{ currentFailureReason || '订单出现异常，管家将介入处理' }}
        </text>
      </view>

      <!-- 业务资料 -->
      <view class="section-card">
        <view class="section-title"><view class="section-bar"></view>业务资料</view>
        <view v-if="!formEntries.length" class="empty-text">暂无资料</view>
        <view v-for="entry in formEntries" :key="entry.key" class="form-row" :class="{ 'form-row-image': isImageEntry(entry) }">
          <text class="form-key">{{ entry.label }}</text>
          <view v-if="isImageEntry(entry)" class="form-image-list">
            <image
              v-for="url in getFileUrls(entry.raw_value || entry.value)"
              :key="url"
              class="form-image"
              :src="toFullFileUrl(url)"
              mode="aspectFill"
              @click.stop="previewUploadedFile(url, getFileUrls(entry.raw_value || entry.value))"
            />
          </view>
          <text v-else class="form-value">{{ entry.display_value || entry.value || '—' }}</text>
        </view>
      </view>

      <!-- 交付资料（B端上传，C端仅查看） -->
      <view v-if="deliveryEntries.length" class="section-card">
        <view class="section-title"><view class="section-bar"></view>交付资料</view>
        <view v-for="entry in deliveryEntries" :key="entry.key" class="form-row" :class="{ 'form-row-image': entry.type === 'image' }">
          <text class="form-key">{{ entry.label }}</text>
          <view v-if="entry.type === 'image'" class="form-image-list">
            <image
              v-for="url in getFileUrls(entry.value)"
              :key="url"
              class="form-image"
              :src="toFullFileUrl(url)"
              mode="aspectFill"
              @click.stop="previewUploadedFile(url, getFileUrls(entry.value))"
            />
          </view>
          <text v-else class="form-value">{{ entry.value || '—' }}</text>
        </view>
      </view>

      <!-- 状态轨迹 -->
      <view class="section-card">
        <view class="section-title"><view class="section-bar"></view>订单进度</view>
        <view v-if="!timelineEntries.length" class="empty-text">暂无进度记录</view>
        <view
          v-for="tl in timelineEntries"
          :key="tl.id"
          class="timeline-item"
          :class="{ 'timeline-item-failed': isFailureTimeline(tl) }"
        >
          <view class="timeline-dot"></view>
          <view class="timeline-body">
            <text class="timeline-title">{{ tl.after_status_text || statusLabel(tl.after_status) || '—' }}</text>
            <text v-if="tl.audit_status === 'pending'" class="audit-status pending">平台审核中</text>
            <text v-if="shouldShowAuditApproved(tl)" class="audit-status approved">审核通过</text>
            <text v-if="tl.audit_status === 'rejected'" class="audit-status rejected">审核未通过</text>
            <text
              v-if="isFailureTimeline(tl) && tl.audit_status !== 'rejected'"
              class="timeline-failed-tag"
            >
              {{ failureTitleOf(tl) }}
            </text>
            <text
              v-if="isFailureTimeline(tl) && extractFailureReason(tl)"
              class="timeline-failed-reason"
            >原因：{{ extractFailureReason(tl) }}</text>
            <text v-if="tl.remark && !isProcessFailedTimeline(tl) && !isFailureTimeline(tl)" class="timeline-desc">{{ tl.remark }}</text>
            <text class="timeline-time">{{ formatTime(tl.created_at) }}</text>
          </view>
        </view>
      </view>

      <!-- 下一步动作 -->
      <view class="section-card action-card">
        <view class="section-title"><view class="section-bar"></view>下一步操作</view>
        <view v-if="actionsLoading" class="loading-text">加载中...</view>
        <view v-else-if="!actions.length" class="empty-text">当前没有可执行的操作</view>
        <view v-else class="action-list">
          <!-- button_click -->
          <button
            v-for="action in buttonClickActions"
            :key="action.id"
            class="action-btn"
            @click="executeButtonClick(action)"
          >
            {{ action.button_label }}
          </button>
          <!-- form_input -->
          <button
            v-for="action in formInputActions"
            :key="action.id"
            class="action-btn action-material"
            @click="openFormInput(action)"
          >
            {{ action.button_label }}
          </button>
          <!-- wx_pay -->
          <button
            v-for="action in wxPayActions"
            :key="action.id"
            class="action-btn action-pay"
            :disabled="submitting"
            @click="executeWxPay(action)"
          >
            {{ submitting ? '处理中...' : `${action.button_label || '立即支付'}${action.pay_amount_text ? ' ' + action.pay_amount_text : ''}` }}
          </button>
        </view>
      </view>
    </template>

    <view v-else class="error-card">
      <text>订单不存在或加载失败</text>
    </view>

    <!-- 表单输入弹窗 -->
    <view v-if="formVisible" class="form-overlay">
      <view class="form-mask" @tap="closeForm" @touchmove.stop.prevent></view>

      <view
        class="form-sheet"
        :style="formSheetStyle"
        @tap.stop
        @click.stop
      >
        <view class="form-header">
          <text class="form-title">填写 {{ currentAction?.button_label }}</text>
          <view class="form-close" @click="closeForm">✕</view>
        </view>
        <scroll-view
          scroll-y
          class="form-body"
          :enhanced="true"
          :enable-flex="true"
          :scroll-with-animation="true"
          :scroll-into-view="activeFieldAnchor"
          :show-scrollbar="false"
          @tap.stop
          @click.stop
        >
          <view class="form-scroll-content" :style="formScrollContentStyle">
            <view
              v-for="field in normalizeActionFields(currentAction)"
              v-show="isFieldVisible(field)"
              :id="fieldDomId(field)"
              :key="getFieldKey(field)"
              class="field-wrap"
            >
              <view class="field-label">
                <text>{{ field.label }}</text>
                <text v-if="isFieldRequired(field)" class="required">*</text>
              </view>
              <input
                v-if="field.type === 'text' || field.type === 'phone' || field.type === 'number'"
                :value="getFormValue(field)"
                :type="field.type === 'phone' ? 'number' : field.type === 'number' ? 'digit' : 'text'"
                class="field-input"
                :placeholder="`请输入${field.label}`"
                :adjust-position="false"
                cursor-spacing="20"
                @input="(e) => setFormValue(field, e.detail.value)"
                @focus="handleFieldFocus(field)"
                @keyboardheightchange="handleKeyboardHeightChange"
                @blur="handleFieldBlur(field)"
              />
              <input
                v-else-if="field.type === 'date'"
                :value="getFormValue(field)"
                class="field-input"
                type="text"
                placeholder="格式：2025-01-15"
                :adjust-position="false"
                cursor-spacing="20"
                @input="(e) => setFormValue(field, e.detail.value)"
                @focus="handleFieldFocus(field)"
                @keyboardheightchange="handleKeyboardHeightChange"
                @blur="handleFieldBlur(field)"
              />
              <input
                v-else-if="field.type === 'datetime'"
                :value="getFormValue(field)"
                class="field-input"
                type="text"
                placeholder="格式：2025-01-15 14:30"
                :adjust-position="false"
                cursor-spacing="20"
                @input="(e) => setFormValue(field, e.detail.value)"
                @focus="handleFieldFocus(field)"
                @keyboardheightchange="handleKeyboardHeightChange"
                @blur="handleFieldBlur(field)"
              />
              <textarea
                v-else-if="field.type === 'textarea'"
                :value="getFormValue(field)"
                class="field-textarea"
                :placeholder="`请输入${field.label}`"
                :auto-height="false"
                :adjust-position="false"
                cursor-spacing="20"
                @input="(e) => setFormValue(field, e.detail.value)"
                @focus="handleFieldFocus(field)"
                @keyboardheightchange="handleKeyboardHeightChange"
                @blur="handleFieldBlur(field)"
              />
              <picker
                v-else-if="field.type === 'select'"
                mode="selector"
                :value="0"
                :range="field.options || []"
                :range-key="'label'"
                @change="(e) => { setFormValue(field, field.options[e.detail.value]?.value ?? field.options[e.detail.value] ?? '') }"
              >
                <view class="field-picker">
                  <text>{{ getOptionLabel(field, getFormValue(field)) || `请选择${field.label}` }}</text>
                  <text class="arrow">›</text>
                </view>
              </picker>
              <view v-else-if="field.type === 'image' || field.type === 'file'" class="upload-field">
                <view
                  class="upload-box"
                  :class="{ uploading: uploadingMap[getFieldKey(field)] }"
                  @click.stop="chooseAndUploadFile(field)"
                >
                  <image
                    v-if="getFormValue(field)"
                    :src="toFullFileUrl(getFormValue(field))"
                    class="upload-preview"
                    mode="aspectFill"
                    @click.stop="previewUploadedFile(getFormValue(field))"
                  />
                  <view v-else class="upload-placeholder">
                    <text class="upload-plus">+</text>
                    <text class="upload-text">上传图片</text>
                  </view>
                  <view v-if="uploadingMap[getFieldKey(field)]" class="upload-mask">
                    <text class="upload-loading">上传中...</text>
                  </view>
                </view>
                <text v-if="getFormValue(field)" class="upload-change" @click.stop="chooseAndUploadFile(field)">
                  点击更换
                </text>
              </view>
              <view v-else class="field-input">
                <text class="placeholder-text">暂不支持此字段类型：{{ field.type }}</text>
              </view>
            </view>
          </view>
        </scroll-view>
        <view v-show="!isKeyboardVisible" class="form-footer" @tap.stop @click.stop>
          <button class="submit-btn" :disabled="submitting" @click="submitForm">
            {{ submitting ? '提交中...' : '确认提交' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding-bottom: 32px;
  background: #f2f6f5;
}

.hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 180px;
  background: linear-gradient(135deg, #004d40, #07362f);
  border-radius: 0 0 32px 32px;
}

.nav-back {
  position: absolute;
  top: 16px;
  left: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.92);
  color: #102a55;
  font-size: 28px;
  line-height: 34px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.hero-icon {
  font-size: 72px;
}

.loading-card,
.error-card {
  margin: 14px 14px 0;
  padding: 40px;
  border-radius: 32px;
  background: #fff;
  text-align: center;
  color: #6b7c78;
}

.summary-card,
.section-card {
  margin: 14px 14px 0;
  padding: 16px;
  border-radius: 32px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.05);
}

.summary-card { margin-top: -16px; }

.summary-head {
  display: flex;
  gap: 12px;
  align-items: center;
}

.service-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 52px;
  height: 52px;
  border-radius: 18px;
  background: rgba(0, 77, 64, 0.08);
  font-size: 24px;
}

.summary-main { flex: 1; }

.service-name {
  display: block;
  color: #12312c;
  font-size: 18px;
  font-weight: 900;
}

.order-no {
  display: block;
  margin-top: 5px;
  color: #6b7c78;
  font-size: 11px;
}

.status-pill {
  padding: 6px 10px;
  border-radius: 999px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  font-size: 10px;
  font-weight: 900;
}

.status-pill.pending { color: #7a5a21; background: rgba(197, 160, 89, 0.2); }
.status-pill.reviewing { color: #1a4a7a; background: rgba(70, 130, 180, 0.15); }
.status-pill.quoted { color: #1a6a2a; background: rgba(60, 160, 80, 0.15); }
.status-pill.paid { color: #5a3a10; background: rgba(197, 160, 89, 0.25); }
.status-pill.completed { color: #004d40; background: rgba(0, 77, 64, 0.1); }

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
  padding: 12px;
  border-radius: 24px;
  background: #f2f6f5;
}

.grid-item { display: flex; flex-direction: column; }

.label {
  display: block;
  color: #87938f;
  font-size: 10px;
}

.value {
  display: block;
  margin-top: 4px;
  color: #12312c;
  font-size: 13px;
  font-weight: 800;
}

.value.price { color: #e97832; }

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  color: #12312c;
  font-size: 16px;
  font-weight: 900;
}

.section-bar {
  width: 4px;
  height: 16px;
  border-radius: 4px;
  background: #004d40;
}

.empty-text {
  color: #9aa3b5;
  font-size: 12px;
  padding: 8px 0;
}

.loading-text {
  color: #6b7c78;
  font-size: 13px;
  padding: 8px 0;
  text-align: center;
}

.form-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.07);
}

.form-key {
  flex-shrink: 0;
  color: #87938f;
  font-size: 12px;
  font-weight: 800;
}

.form-value {
  color: #12312c;
  font-size: 13px;
  line-height: 1.55;
  text-align: right;
}

.form-row-image {
  align-items: flex-start;
}

.form-image-list {
  display: flex;
  justify-content: flex-end;
  gap: 12rpx;
  flex-wrap: wrap;
  max-width: 380rpx;
}

.form-image {
  width: 144rpx;
  height: 144rpx;
  border-radius: 16rpx;
  background: #f3f7f5;
  border: 1rpx solid #e2eeea;
}

.timeline-item {
  position: relative;
  display: flex;
  gap: 12px;
  padding-bottom: 18px;
}

.timeline-dot {
  flex-shrink: 0;
  width: 11px;
  height: 11px;
  margin-top: 4px;
  border: 3px solid rgba(197, 160, 89, 0.36);
  border-radius: 50%;
  background: #004d40;
}

/* 失败类时间线：红色圆点 + 红色标题 */
.timeline-item-failed .timeline-dot {
  background: #e53935;
  border-color: rgba(229, 57, 53, 0.3);
}

.timeline-item-failed .timeline-title {
  color: #d93025;
}

.timeline-body { flex: 1; }

.timeline-title {
  display: block;
  color: #12312c;
  font-size: 13px;
  font-weight: 900;
}

.timeline-desc {
  display: block;
  margin-top: 5px;
  color: #6b7c78;
  font-size: 12px;
  line-height: 1.6;
}

.timeline-time {
  display: block;
  margin-top: 6px;
  color: #a0aaa7;
  font-size: 10px;
}

.action-card { margin-bottom: 20px; }

.action-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.action-btn {
  height: 40px;
  padding: 0 20px;
  border: none;
  border-radius: 20px;
  color: #fff;
  background: #004d40;
  font-size: 14px;
  font-weight: 800;
  line-height: 40px;
}

.action-btn.action-material {
  color: #004d40;
  background: rgba(0, 77, 64, 0.1);
  border: 1.5px solid rgba(0, 77, 64, 0.2);
}

.action-btn.action-pay {
  color: #12312c;
  background: #c5a059;
}

/* 表单弹窗 */
.form-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: flex-end;
  background: rgba(0, 0, 0, 0.36);
}

.form-sheet {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-height: 88vh;
  border-radius: 28px 28px 0 0;
  background: #fff;
  box-shadow: 0 -18px 60px rgba(0, 0, 0, 0.18);
}

.form-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 20px 20px 16px;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
}

.form-title {
  color: #102a55;
  font-size: 17px;
  font-weight: 900;
}

.form-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #f2f6f5;
  color: #6b7c78;
  font-size: 14px;
}

.form-body {
  flex: 1;
  padding: 16px 18px;
  max-height: 60vh;
}

.field-wrap {
  margin-bottom: 18px;
}

.field-label {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 8px;
  color: #102a55;
  font-size: 13px;
  font-weight: 800;
}

.required { color: #e53e3e; }

.field-input {
  box-sizing: border-box;
  width: 100%;
  height: 44px;
  padding: 0 14px;
  border: 1.5px solid rgba(0, 77, 64, 0.15);
  border-radius: 14px;
  background: #f8fbfa;
  color: #102a55;
  font-size: 14px;
}

.field-textarea {
  box-sizing: border-box;
  width: 100%;
  min-height: 80px;
  padding: 10px 14px;
  border: 1.5px solid rgba(0, 77, 64, 0.15);
  border-radius: 14px;
  background: #f8fbfa;
  color: #102a55;
  font-size: 14px;
}

.field-picker {
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  width: 100%;
  height: 44px;
  padding: 0 14px;
  border: 1.5px solid rgba(0, 77, 64, 0.15);
  border-radius: 14px;
  background: #f8fbfa;
  color: #9aa3b5;
  font-size: 14px;
}

.image-field {
  display: flex;
  flex-direction: column;
  height: auto;
  padding: 8px 0;
}

.field-url-input {
  width: 100%;
  height: 36px;
  padding: 0 10px;
  border: 1px solid rgba(0, 77, 64, 0.2);
  border-radius: 8px;
  background: #f8fbfa;
  font-size: 12px;
}

.placeholder-text {
  color: #c0c4cc;
  font-size: 12px;
}

.arrow {
  color: #9aa3b5;
  font-size: 18px;
  font-weight: 700;
}

.form-footer {
  flex-shrink: 0;
  padding: 12px 18px calc(12px + env(safe-area-inset-bottom));
  border-top: 1px solid rgba(0, 77, 64, 0.08);
}

.submit-btn {
  width: 100%;
  height: 46px;
  border: none;
  border-radius: 23px;
  background: linear-gradient(135deg, #004d40, #00695c);
  color: #fff;
  font-size: 15px;
  font-weight: 900;
  line-height: 46px;
}

.audit-reject-card {
  margin: 24rpx;
  padding: 28rpx;
  border-radius: 24rpx;
  background: #fff5f5;
  border: 1rpx solid #ffb4b4;
}
.audit-reject-title {
  display: block;
  color: #d93025;
  font-size: 30rpx;
  font-weight: 700;
  margin-bottom: 12rpx;
}
.audit-reject-desc {
  display: block;
  color: #9f1c1c;
  font-size: 26rpx;
  line-height: 1.6;
  word-break: break-all;
}

.audit-status {
  display: inline-flex;
  margin-top: 6px;
  padding: 4px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 800;
}
.audit-status.pending { background: #fff7e6; color: #ad6800; }
.audit-status.approved { background: #f6ffed; color: #389e0d; }
.audit-status.rejected {
  background: #fff0f0;
  color: #d93025;
  border: 1rpx solid #ffb4b4;
}
.audit-reason,
.timeline-failed-reason {
  display: block;
  margin-top: 8rpx;
  color: #d93025;
  font-size: 24rpx;
  line-height: 1.5;
  font-weight: 600;
}

.upload-field { margin-top: 8rpx; }

.upload-box {
  position: relative;
  width: 180rpx;
  height: 180rpx;
  border-radius: 20rpx;
  border: 2rpx dashed #d8e8e3;
  background: #f8fbfa;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.upload-preview { width: 100%; height: 100%; }
.upload-placeholder { display: flex; flex-direction: column; align-items: center; justify-content: center; }
.upload-plus { font-size: 48rpx; line-height: 1; color: #0b6b55; font-weight: 800; }
.upload-text { margin-top: 8rpx; color: #7d918c; font-size: 24rpx; }
.upload-change { display: block; margin-top: 8rpx; color: #0b6b55; font-size: 24rpx; }
.upload-mask {
  position: absolute; inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
}
.upload-loading { color: #fff; font-size: 24rpx; }

/* 办理失败红色提示 */
.process-failed-alert {
  margin: 24rpx 32rpx;
  padding: 28rpx 30rpx;
  border-radius: 24rpx;
  background: #fff1f1;
  border: 2rpx solid #ffb6b6;
}
.process-failed-title {
  font-size: 34rpx;
  font-weight: 800;
  color: #d93025;
  margin-bottom: 14rpx;
}
.process-failed-reason {
  font-size: 28rpx;
  line-height: 1.6;
  color: #9f2f2f;
  font-weight: 600;
}

/* 时间线办理失败标签 */
.timeline-failed-tag {
  display: inline-flex;
  align-items: center;
  margin-top: 10rpx;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  background: #ffe8e8;
  color: #d93025;
  font-size: 24rpx;
  font-weight: 700;
}
.timeline-failed-reason {
  margin-top: 10rpx;
  color: #9f2f2f;
  font-size: 26rpx;
  line-height: 1.5;
  font-weight: 600;
}
</style>
