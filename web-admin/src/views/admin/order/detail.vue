<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderDetail, getOrderActions, performOrderAction, auditOrder } from '@/api/admin/orders'
import { request, ORIGIN_URL } from '@/api/request'
import DynamicForm from '@/components/DynamicForm.vue'

const router = useRouter()
const route = useRoute()

const orderId = ref(route.params.id || '')
const loading = ref(true)
const actionsLoading = ref(false)
const order = ref(null)
const actions = ref([])

// DynamicForm 组件 ref，用于调用 validateAll()
const dynamicFormRef = ref(null)

// form_input 弹窗状态
const formInputVisible = ref(false)
const formInputAction = ref(null)
const formInputData = ref({})
const formInputLoading = ref(false)

const statusLabel = (code) => {
  const m = {
    start: '开始',
    pending: '等待受理',
    reviewing: '资料审核中',
    quoted: '已报价',
    paid: '已收款',
    wait_quote: '待报价',
    wait_pay: '待支付',
    dispatching: '派车中',
    in_progress: '服务履约中',
    processing: '服务中',
    completed: '已完成',
    cancelled: '已取消',
    refunded: '已退款',
    failed: '异常',
    wait_supplement: '待补资料',
    wait_upload_material: '待上传资料',
  }
  return m[code] || code || '未知'
}

const formatMoney = (amount) => {
  if (amount === null || amount === undefined || amount === '') return '—'
  const n = Number(amount || 0)
  return `¥${n.toLocaleString('zh-CN')}`
}

const formatTime = (t) => {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return t
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// normalizeOrder 标准化后端返回字段
const normalizeOrder = (raw = {}) => {
  const o = raw || {}
  return {
    ...o,
    order_no: o.order_no || o.orderNo || '',
    service_name: o.service_name || o.serviceName || '',
    macro_status: o.macro_status || o.current_status || o.currentStatus || '',
    macro_status_text: o.macro_status_text || '',
    current_stage: o.current_stage || '',
    current_stage_text: o.current_stage_text || '',
    payment_status: o.payment_status || 'unpaid',
    payment_status_text: o.payment_status_text || '',
    total_amount: o.total_amount || o.amount || 0,
    form_items: o.form_items || [],
    form_data: o.form_data || o.formData || {},
    timelines: o.timelines || [],
    payments: o.payments || [],
    action_nodes: o.action_nodes || o.actionNodes || [],
  }
}

// unwrapOrderFromResponse 统一从响应中提取 order 对象
const unwrapOrderFromResponse = (res) => {
  const body = res?.data ?? res ?? {}
  return body.order || body.data?.order || body.data || body
}

// 业务表单：只使用 form_items
const formEntries = computed(() => {
  if (Array.isArray(order.value?.form_items) && order.value.form_items.length) {
    return order.value.form_items.map((item) => ({
      key: item.key,
      label: item.label || item.key,
      value: item.display_value ?? item.value ?? '—',
    }))
  }
  return []
})

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
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

const isMaterialFile = (value) => {
  const url = getFileUrl(value)
  if (!url) return false
  return url.includes('/material/') || url.includes('/uploads/')
}

const isImageFile = (value) => {
  const url = getFileUrl(value)
  if (!isMaterialFile(url)) return false
  return /\.(jpg|jpeg|png)$/i.test(url)
}

const toFullFileUrl = (url) => {
  if (!url) return ''
  if (/^https?:\/\//.test(url)) return url
  const origin = String(ORIGIN_URL || '').replace(/\/+$/, '')
  const path = url.startsWith('/') ? url : `/${url}`
  return `${origin}${path}`
}

const getTimelineTime = (item) => new Date(item?.created_at || item?.updated_at || 0).getTime()

// pendingAuditTimeline 按时间排序取最新一条 pending timeline
// 防止多条 pending 时误取旧记录
const pendingAuditTimeline = computed(() => {
  return [...(order.value?.timelines || [])]
    .filter((tl) => tl.audit_status === 'pending')
    .sort((a, b) => {
      const bt = getTimelineTime(b)
      const at = getTimelineTime(a)
      if (bt !== at) return bt - at
      return Number(b.id || 0) - Number(a.id || 0)
    })[0] || null
})
const hasPendingAudit = computed(() => !!pendingAuditTimeline.value)

const auditLoading = ref(false)
const rejectAuditVisible = ref(false)
const rejectAuditRemark = ref('')

const approveAudit = async () => {
  if (!pendingAuditTimeline.value || !order.value?.id) return
  try {
    await ElMessageBox.confirm(
      '确认审核通过？通过后订单将进入下一流程。',
      '审核确认',
      { confirmButtonText: '审核通过', cancelButtonText: '取消', type: 'success' }
    )
    auditLoading.value = true
    const res = await auditOrder(order.value.id, {
      timeline_id: pendingAuditTimeline.value.id,
      result: 'approved',
      audit_remark: '资料审核通过',
    })
    const payload = unwrapOrderFromResponse(res)
    if (payload?.id) {
      order.value = normalizeOrder(payload)
    }
    await loadOrderDetail()
    await loadOrderActions()

    showToast('审核已通过', 'success')
  } catch (error) {
    if (error !== 'cancel') {
      showToast(error?.response?.data?.error || error?.message || '审核失败', 'error')
    }
  } finally {
    auditLoading.value = false
  }
}

const openRejectAudit = () => {
  rejectAuditRemark.value = ''
  rejectAuditVisible.value = true
}

const submitRejectAudit = async () => {
  if (!pendingAuditTimeline.value || !order.value?.id) return
  if (!rejectAuditRemark.value.trim()) {
    showToast('请填写审核失败原因', 'warning')
    return
  }
  auditLoading.value = true
  try {
    const res = await auditOrder(order.value.id, {
      timeline_id: pendingAuditTimeline.value.id,
      result: 'rejected',
      audit_remark: rejectAuditRemark.value.trim(),
    })
    const payload = unwrapOrderFromResponse(res)
    if (payload?.id) {
      order.value = normalizeOrder(payload)
    }

    await loadOrderDetail()
    await loadOrderActions()

    rejectAuditVisible.value = false
    rejectAuditRemark.value = ''

    showToast('已驳回，订单已回到补资料流程', 'success')
  } catch (error) {
    showToast(error?.response?.data?.error || error?.message || '审核失败', 'error')
  } finally {
    auditLoading.value = false
  }
}

const loadOrderDetail = async () => {
  loading.value = true
  try {
    const res = await getOrderDetail(orderId.value)
    order.value = normalizeOrder(res)
  } catch {
    showToast('订单详情加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const loadOrderActions = async () => {
  if (!order.value?.id) return
  actionsLoading.value = true
  try {
    const res = await getOrderActions(order.value.id)
    actions.value = res.actions || []
  } catch {
    actions.value = []
  } finally {
    actionsLoading.value = false
  }
}

const refresh = async () => {
  await loadOrderDetail()
  await loadOrderActions()
}

// 根据 action_type 分类（待审核状态下过滤掉 audit 类动作）
const isAuditAction = (action) => {
  if (!action) return false
  return action.is_audit_action === true ||
    ['audit_approve', 'audit_reject', 'audit_rejected'].includes(action.action_name)
}

const visibleWorkflowActions = computed(() => {
  // 待审核状态下，不展示任何普通工作流 action_nodes，只展示人工审核按钮
  if (hasPendingAudit.value) return []
  return (actions.value || []).filter((a) => !isAuditAction(a))
})

const buttonClickActions = computed(() => visibleWorkflowActions.value.filter(a => a.action_type === 'button_click'))
const formInputActions = computed(() => visibleWorkflowActions.value.filter(a => a.action_type === 'form_input'))
const wxPayActions = computed(() => visibleWorkflowActions.value.filter(a => a.action_type === 'wx_pay'))

// --- button_click 类型：直接执行备注确认 ---
const executeButtonClick = async (action) => {
  // 兜底拦截审核动作
  if (isAuditAction(action)) {
    showToast('审核动作必须通过审核按钮执行，请刷新页面后重试', 'warning')
    return
  }

  try {
    const { value: remark } = await ElMessageBox.prompt(
      `执行动作「${action.button_label}」，请输入备注（可选）：`,
      '确认操作',
      {
        confirmButtonText: '确认执行',
        cancelButtonText: '取消',
        inputPlaceholder: '备注信息（选填）',
      }
    )
    const res = await performOrderAction(order.value.id, {
      action_name: action.action_name,
      remark: remark || '',
    })
    const payload = unwrapOrderFromResponse(res)
    if (payload?.id) {
      order.value = normalizeOrder(payload)
    }
    await loadOrderActions()
    showToast(`「${action.button_label}」已执行`, 'success')
  } catch (err) {
    if (err !== 'cancel') {
      showToast('流程推进失败', 'error')
    }
  }
}

// --- form_input 类型：弹窗收集表单数据（DynamicForm 组件驱动）---
const openFormInput = (action) => {
  formInputAction.value = action
  formInputData.value = {}
  formInputVisible.value = true
}

const submitFormInput = async () => {
  // 兜底拦截审核动作
  if (isAuditAction(formInputAction.value)) {
    showToast('审核动作必须通过审核按钮执行，请刷新页面后重试', 'warning')
    return
  }

  // 调用 DynamicForm 的 validateAll()
  if (!dynamicFormRef.value?.validateAll()) return

  formInputLoading.value = true
  try {
    const res = await performOrderAction(order.value.id, {
      action_name: formInputAction.value.action_name,
      remark: '',
      input_data: formInputData.value,
    })
    const payload = unwrapOrderFromResponse(res)
    if (payload?.id) {
      order.value = normalizeOrder(payload)
    }
    await loadOrderActions()
    formInputVisible.value = false
    showToast(`「${formInputAction.value.button_label}」已执行`, 'success')
  } catch {
    showToast('流程推进失败', 'error')
  } finally {
    formInputLoading.value = false
  }
}

// --- wx_pay 类型：提示跳转到支付 ---
const triggerWxPay = (action) => {
  showToast(`「${action.button_label}」：请引导客户完成微信支付`, 'info')
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/orders')
  }
}

onMounted(async () => {
  if (route.params.id) orderId.value = route.params.id
  if (!orderId.value) {
    loading.value = false
    return
  }
  await loadOrderDetail()
  await loadOrderActions()
})
</script>

<template>
  <div class="detail-page">
    <div class="detail-hero">
      <div class="back-btn" @click="goBack">‹</div>
      <div>
        <span class="eyebrow">ORDER WORKFLOW</span>
        <span class="hero-title">订单履约详情</span>
        <span class="hero-desc">动态表单、状态轨迹、支付记录与下一步动作统一在此管理。</span>
      </div>
    </div>

    <div v-if="loading" class="info-card empty-card">正在加载订单详情...</div>

    <template v-else-if="order">
      <!-- 摘要卡片 -->
      <div class="summary-card">
        <div class="summary-head">
          <div class="service-icon">{{ order.icon || '📋' }}</div>
          <div class="summary-main">
            <span class="service-name">{{ order.service_name || order.serviceName }}</span>
            <span class="order-no">{{ order.order_no || order.orderNo }}</span>
          </div>
          <div class="status-pill" :class="order.macro_status">
            {{ order.macro_status_text || statusLabel(order.macro_status) }}
          </div>
        </div>
        <div class="summary-grid">
          <div>
            <span class="label">客户</span>
            <span class="value">{{ order.contact_name || order.appUserName || '—' }}</span>
          </div>
          <div>
            <span class="label">联系电话</span>
            <span class="value">{{ order.contact_phone || order.appUserPhone || '—' }}</span>
          </div>
          <div>
            <span class="label">金额</span>
            <span class="value price">{{ formatMoney(order.total_amount || order.amount) }}</span>
          </div>
          <div>
            <span class="label">当前节点</span>
            <span class="value">{{ order.current_stage_text || statusLabel(order.current_stage) || '—' }}</span>
          </div>
        </div>
      </div>

      <!-- 业务表单详情 -->
      <div class="info-card">
        <div class="section-title">业务表单</div>
        <div v-if="!formEntries.length" class="empty-line">暂无表单数据</div>
        <div v-for="entry in formEntries" :key="entry.key" class="json-row">
          <span class="json-key">{{ entry.label }}</span>
          <span v-if="isImageFile(entry.value)" class="json-value image-cell">
            <el-image
              :src="toFullFileUrl(getFileUrl(entry.value))"
              :preview-src-list="[toFullFileUrl(getFileUrl(entry.value))]"
              fit="cover"
              class="material-thumb"
            />
          </span>
          <span v-else-if="isMaterialFile(entry.value)" class="json-value">
            <el-link :href="toFullFileUrl(getFileUrl(entry.value))" target="_blank" type="primary">
              查看文件
            </el-link>
          </span>
          <span v-else class="json-value">{{ entry.value }}</span>
        </div>
      </div>

      <!-- 状态轨迹 -->
      <div class="info-card">
        <div class="section-title">状态轨迹</div>
        <div v-if="!order.timelines?.length" class="empty-line">暂无轨迹记录</div>
        <div v-for="tl in order.timelines" :key="tl.id" class="timeline-row">
          <div class="timeline-dot"></div>
          <div class="timeline-body">
            <span class="timeline-title">
              {{ tl.after_status_text || statusLabel(tl.after_status) || '—' }}
            </span>
            <span v-if="tl.remark" class="timeline-desc">{{ tl.remark }}</span>
            <span v-if="tl.audit_status === 'pending'" class="audit-tag pending">待平台审核</span>
            <span v-if="tl.audit_status === 'approved'" class="audit-tag approved">审核通过</span>
            <span v-if="tl.audit_status === 'rejected'" class="audit-tag rejected">审核未通过</span>
            <span v-if="tl.audit_remark" class="audit-remark">审核备注：{{ tl.audit_remark }}</span>
            <span v-if="tl.audited_at" class="audit-time">审核时间：{{ formatTime(tl.audited_at) }}</span>
            <span class="timeline-time">{{ formatTime(tl.created_at || tl.createdAt) }}</span>
          </div>
        </div>
      </div>

      <!-- 财务对账 -->
      <div class="info-card">
        <div class="section-title">财务对账</div>
        <div v-if="!order.payments?.length" class="empty-line">当前订单暂无支付记录</div>
        <div v-for="pay in order.payments" :key="pay.id" class="payment-row">
          <div>
            <span class="payment-id">#{{ pay.id }}</span>
            <span class="payment-method">{{ pay.pay_method || pay.payMethod }} · {{ pay.third_trade_no || pay.thirdTradeNo || '—' }}</span>
          </div>
          <div class="payment-right">
            <span class="payment-amount">{{ formatMoney(pay.pay_amount || pay.payAmount) }}</span>
            <span class="payment-status" :class="pay.status">
              {{ pay.status === 'success' ? '支付成功' : pay.status === 'failed' ? '支付失败' : pay.status || '—' }}
            </span>
          </div>
        </div>
      </div>

      <!-- 下一步动作 -->
      <div class="info-card action-card">
        <div class="section-title">下一步动作</div>
        <div v-if="actionsLoading" class="actions-skeleton">
          <div class="skeleton-btn"></div>
          <div class="skeleton-btn"></div>
        </div>
        <div v-else-if="!hasPendingAudit && !visibleWorkflowActions.length" class="empty-line">当前节点无操作动作</div>
        <div v-else class="workflow-actions">
          <!-- 待审核：只展示人工审核按钮 -->
          <template v-if="hasPendingAudit">
            <button
              class="action-btn action-audit-pass"
              :disabled="auditLoading"
              @click="approveAudit"
            >
              审核通过
            </button>
            <button
              class="action-btn action-audit-reject"
              :disabled="auditLoading"
              @click="openRejectAudit"
            >
              审核不通过
            </button>
          </template>

          <!-- 普通工作流动作（待审核时已被 visibleWorkflowActions 过滤为空，此分支不会渲染） -->
          <template v-else>
            <!-- button_click -->
            <button
              v-for="action in buttonClickActions"
              :key="action.id"
              class="action-btn"
              :title="`${action.button_label}（${action.stage_name}）`"
              @click="executeButtonClick(action)"
            >
              {{ action.button_label }}
            </button>
            <!-- form_input -->
            <button
              v-for="action in formInputActions"
              :key="action.id"
              class="action-btn action-material"
              :title="`${action.button_label}（需填写表单）`"
              @click="openFormInput(action)"
            >
              {{ action.button_label }}
            </button>
            <!-- wx_pay -->
            <button
              v-for="action in wxPayActions"
              :key="action.id"
              class="action-btn action-payment"
              :title="`${action.button_label}（微信支付）`"
              @click="triggerWxPay(action)"
            >
              {{ action.button_label }}
            </button>
          </template>
        </div>
      </div>
    </template>

    <div v-else class="info-card empty-card">订单不存在或加载失败</div>

    <!-- form_input 动态表单弹窗（DynamicForm 引擎驱动）-->
    <el-dialog
      v-model="formInputVisible"
      :title="`填写：${formInputAction?.button_label || ''}`"
      width="580px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <DynamicForm
        ref="dynamicFormRef"
        v-model="formInputData"
        :fields="formInputAction?.form_fields || []"
      />
      <template #footer>
        <el-button @click="formInputVisible = false">取消</el-button>
        <el-button type="primary" :loading="formInputLoading" @click="submitFormInput">确认执行</el-button>
      </template>
    </el-dialog>

    <!-- 审核不通过弹窗 -->
    <el-dialog v-model="rejectAuditVisible" title="审核不通过" width="520px" :close-on-click-modal="false">
      <el-form label-position="top">
        <el-form-item label="审核失败原因" required>
          <el-input
            v-model="rejectAuditRemark"
            type="textarea"
            :rows="4"
            placeholder="请填写审核失败原因，例如：护照首页照片模糊，请重新上传"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectAuditVisible = false">取消</el-button>
        <el-button type="danger" :loading="auditLoading" @click="submitRejectAudit">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding-bottom: 32px;
  background: #f2f6f5;
}

.detail-hero {
  display: flex;
  gap: 16px;
  align-items: center;
  padding: 54px 22px 30px;
  color: #fff;
  background: radial-gradient(circle at 88% 18%, rgba(197, 160, 89, 0.42), transparent 32%), linear-gradient(135deg, #004d40, #07362f);
  border-radius: 0 0 32px 32px;
  box-shadow: 0 22px 54px rgba(0, 77, 64, 0.2);
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  color: #004d40;
  background: rgba(255, 255, 255, 0.9);
  font-size: 30px;
  line-height: 34px;
  cursor: pointer;
}

.eyebrow {
  display: block;
  margin-bottom: 8px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 1.4px;
}

.hero-title {
  display: block;
  font-size: 24px;
  font-weight: 900;
}

.hero-desc {
  display: block;
  max-width: 280px;
  margin-top: 8px;
  color: rgba(255, 255, 255, 0.82);
  font-size: 12px;
  line-height: 1.7;
}

.summary-card, .info-card {
  margin: 14px 14px 0;
  padding: 16px;
  border-radius: 32px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.05);
}

.summary-card { margin-top: -16px; }

.empty-card { color: #6b7c78; font-size: 13px; text-align: center; }

.summary-head { display: flex; gap: 12px; align-items: center; }

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

.service-name { display: block; color: #12312c; font-size: 18px; font-weight: 900; }

.order-no { display: block; margin-top: 5px; color: #6b7c78; font-size: 11px; }

.status-pill {
  padding: 6px 10px;
  border-radius: 999px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  font-size: 10px;
  font-weight: 900;
}

.status-pill.pending    { color: #7a5a21; background: rgba(197, 160, 89, 0.2); }
.status-pill.reviewing  { color: #1a4a7a; background: rgba(70, 130, 180, 0.15); }
.status-pill.quoted     { color: #1a6a2a; background: rgba(60, 160, 80, 0.15); }
.status-pill.paid       { color: #5a3a10; background: rgba(197, 160, 89, 0.25); }
.status-pill.in_progress { color: #2a3a8a; background: rgba(90, 100, 200, 0.15); }
.status-pill.completed  { color: #004d40; background: rgba(0, 77, 64, 0.1); }

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
  padding: 12px;
  border-radius: 24px;
  background: #f2f6f5;
}

.label { display: block; color: #87938f; font-size: 10px; }
.value { display: block; margin-top: 4px; color: #12312c; font-size: 13px; font-weight: 800; }
.value.price { color: #e97832; }

.section-title {
  display: block;
  margin-bottom: 12px;
  color: #12312c;
  font-size: 16px;
  font-weight: 900;
}

.json-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.07);
}

.json-key { flex-shrink: 0; color: #87938f; font-size: 12px; font-weight: 800; }

.json-value {
  color: #12312c;
  font-size: 13px;
  line-height: 1.55;
  text-align: right;
}

.timeline-row {
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

.timeline-body { flex: 1; }

.timeline-title { display: block; color: #12312c; font-size: 13px; font-weight: 900; }
.timeline-desc { display: block; margin-top: 5px; color: #6b7c78; font-size: 12px; line-height: 1.6; }
.timeline-time { display: block; margin-top: 6px; color: #a0aaa7; font-size: 10px; }

.payment-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-radius: 22px;
  background: rgba(197, 160, 89, 0.09);
  margin-top: 8px;
}

.payment-id { display: block; color: #12312c; font-size: 13px; font-weight: 900; }
.payment-method { display: block; margin-top: 5px; color: #6b7c78; font-size: 11px; }
.payment-right { text-align: right; }
.payment-amount { color: #e97832; font-size: 15px; font-weight: 900; display: block; }
.payment-status { display: block; margin-top: 4px; color: #7a5a21; font-size: 11px; }
.payment-status.success { color: #2a7a3a; }

.workflow-actions { display: flex; flex-wrap: wrap; gap: 10px; }

/* 操作按钮样式 */
.action-btn {
  height: 36px;
  padding: 0 16px;
  border: none;
  border-radius: 18px;
  color: #fff;
  background: #004d40;
  font-size: 12px;
  font-weight: 800;
  line-height: 36px;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.15s;
}

.action-btn:hover { opacity: 0.85; transform: translateY(-1px); }
.action-btn:active { transform: translateY(0); }

.action-btn.action-payment { color: #12312c; background: #c5a059; }
.action-btn.action-material { color: #004d40; background: rgba(0, 77, 64, 0.1); border: 1.5px solid rgba(0, 77, 64, 0.2); }
.action-btn.action-audit-pass { background: #0b6b55; color: #fff; border-color: #0b6b55; }
.action-btn.action-audit-pass:hover { background: #075f4b; border-color: #075f4b; }
.action-btn.action-audit-reject { background: #fff5f5; color: #d64545; border-color: #ffb8b8; }
.action-btn.action-audit-reject:hover { background: #ffecec; color: #c92f2f; border-color: #ff9b9b; }

/* 骨架屏占位 */
.actions-skeleton { display: flex; gap: 10px; }
.skeleton-btn {
  width: 80px;
  height: 36px;
  border-radius: 18px;
  background: linear-gradient(90deg, #e8eeed 25%, #f5f7f6 50%, #e8eeed 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.empty-line {
  color: #9aa3b5;
  font-size: 12px;
  padding: 8px 0;
}

.field-error {
  color: #e53e3e;
  font-size: 11px;
  margin-top: 4px;
}

.audit-panel {
  margin: 12px 16px;
  padding: 16px;
  border-radius: 16px;
  background: #fff7e6;
  border: 1px solid #ffd591;
}
.audit-title { font-size: 16px; font-weight: 900; color: #7a3e00; }
.audit-desc { margin-top: 6px; color: #8a5a16; font-size: 13px; }
.audit-actions { margin-top: 12px; display: flex; gap: 10px; }
.audit-tag { display: inline-flex; margin-top: 6px; padding: 3px 8px; border-radius: 999px; font-size: 12px; font-weight: 800; }
.audit-tag.pending { background: #fff7e6; color: #ad6800; }
.audit-tag.approved { background: #f6ffed; color: #389e0d; }
.audit-tag.rejected { background: #fff1f0; color: #cf1322; }
.audit-remark, .audit-time { display: block; margin-top: 4px; color: #5f6f6a; font-size: 12px; }

.material-thumb { width: 88px; height: 88px; border-radius: 8px; border: 1px solid #edf2f0; display: block; }
.image-cell { display: inline-flex; }
</style>
