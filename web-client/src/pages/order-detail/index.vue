<script setup>
import { computed, onMounted, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'
import { get, post } from '@/api/request'

const client = useClientStore()
const orderId = ref('')
const order = ref(null)
const actions = ref([])
const loading = ref(true)
const actionsLoading = ref(false)
const formData = ref({})
const formVisible = ref(false)
const currentAction = ref(null)
const submitting = ref(false)

const statusMap = {
  pending: '等待受理',
  reviewing: '资料审核中',
  quoted: '已报价',
  paid: '已收款',
  wait_pay: '待支付',
  processing: '服务中',
  completed: '已完成',
  cancelled: '已取消',
  refunded: '已退款',
}

const statusLabel = (code) => statusMap[code] || code || '未知'

const formatMoney = (amount) => {
  if (amount === null || amount === undefined || amount === '') return '—'
  const n = Number(amount || 0)
  return `¥${n.toLocaleString('zh-CN')}`
}

const formatTime = (t) => {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN', { dateStyle: 'short', timeStyle: 'short' })
}

const formEntries = computed(() => {
  // 优先使用后端 form_items
  if (Array.isArray(order.value?.form_items) && order.value.form_items.length) {
    return order.value.form_items.map((item) => ({
      key: item.key,
      label: item.label || item.key,
      value: item.display_value ?? item.value ?? '—',
    }))
  }

  // fallback：旧数据
  if (!order.value?.form_data) return []
  const raw = order.value.form_data
  const data = typeof raw === 'string' ? JSON.parse(raw) : raw
  return Object.entries(data)
    .filter(([k]) => !k.startsWith('_') && !['service_code', 'service_name', 'submitted_at'].includes(k))
    .map(([k, v]) => ({
      key: k,
      label: k,
      value: typeof v === 'object' ? JSON.stringify(v) : v,
    }))
})

const timelineEntries = computed(() => {
  if (!order.value?.timelines) return []
  return [...order.value.timelines].reverse()
})

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
    // 后端直接返回订单对象，兼容 res.order 或 res.data
    order.value = res.order || res.data?.order || res.data || res
  } catch {
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
    actions.value = res.actions || []
  } catch {
    actions.value = []
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
  currentAction.value = action
  formData.value = {}
  formVisible.value = true
}

const closeForm = () => {
  formVisible.value = false
  currentAction.value = null
  formData.value = {}
}

const validateField = (field) => {
  if (field.required) {
    const val = (formData.value[field.key] || '').toString().trim()
    if (!val) {
      safeToast(`请填写 ${field.label}`, 'none')
      return false
    }
  }
  return true
}

const submitForm = async () => {
  const fields = currentAction.value?.form_fields || []
  for (const field of fields) {
    if (!validateField(field)) return
  }
  await performAction(currentAction.value, formData.value)
  closeForm()
}

const executeWxPay = async (action) => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  const payAmount = order.value?.total_amount || order.value?.amount || 0

  // 构造 input_data 包含金额
  const inputData = { amount: payAmount }

  if (uniApi?.showModal) {
    uniApi.showModal({
      title: `执行「${action.button_label}」`,
      content: `确认支付 ${formatMoney(payAmount)}？`,
      success: async (res) => {
        if (res.confirm) {
          await performAction(action, inputData)
        }
      },
    })
  } else {
    await performAction(action, inputData)
  }
}

const performAction = async (action, inputData) => {
  submitting.value = true
  try {
    await post(`/v1/client/orders/${orderId.value}/action`, {
      action_name: action.action_name,
      remark: '',
      input_data: inputData,
    })
    safeToast('操作成功', 'success')
    await loadOrderDetail()
    await loadOrderActions()
  } catch (err) {
    safeToast(err?.message || '操作失败', 'error')
  } finally {
    submitting.value = false
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
          <view class="grid-item">
            <text class="label">订单金额</text>
            <text class="value price">{{ formatMoney(order.total_amount || order.totalAmount) }}</text>
          </view>
          <view class="grid-item">
            <text class="label">支付状态</text>
            <text class="value">{{ order.payment_status_text || (order.payment_status === 'paid' ? '已支付' : '未支付') }}</text>
          </view>
        </view>
      </view>

      <!-- 业务表单 -->
      <view class="section-card">
        <view class="section-title"><view class="section-bar"></view>业务资料</view>
        <view v-if="!formEntries.length" class="empty-text">暂无资料</view>
        <view v-for="entry in formEntries" :key="entry.key" class="form-row">
          <text class="form-key">{{ entry.label }}</text>
          <text class="form-value">{{ entry.value }}</text>
        </view>
      </view>

      <!-- 状态轨迹 -->
      <view class="section-card">
        <view class="section-title"><view class="section-bar"></view>订单进度</view>
        <view v-if="!timelineEntries.length" class="empty-text">暂无进度记录</view>
        <view v-for="tl in timelineEntries" :key="tl.id" class="timeline-item">
          <view class="timeline-dot"></view>
          <view class="timeline-body">
            <text class="timeline-title">{{ tl.after_status_text || statusLabel(tl.after_status) || '—' }}</text>
            <text v-if="tl.remark" class="timeline-desc">{{ tl.remark }}</text>
            <text class="timeline-time">{{ tl.operator || '系统' }} · {{ formatTime(tl.created_at) }}</text>
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
            @click="executeWxPay(action)"
          >
            {{ action.button_label }}
          </button>
        </view>
      </view>
    </template>

    <view v-else class="error-card">
      <text>订单不存在或加载失败</text>
    </view>

    <!-- 表单输入弹窗 -->
    <view v-if="formVisible" class="form-overlay" @click.self="closeForm">
      <view class="form-sheet">
        <view class="form-header">
          <text class="form-title">填写 {{ currentAction?.button_label }}</text>
          <view class="form-close" @click="closeForm">✕</view>
        </view>
        <scroll-view scroll-y class="form-body">
          <view v-for="field in currentAction?.form_fields || []" :key="field.key" class="field-wrap">
            <view class="field-label">
              <text>{{ field.label }}</text>
              <text v-if="field.required" class="required">*</text>
            </view>
            <input
              v-if="field.type === 'text' || field.type === 'phone' || field.type === 'number'"
              v-model="formData[field.key]"
              :type="field.type === 'phone' ? 'number' : field.type === 'number' ? 'digit' : 'text'"
              class="field-input"
              :placeholder="`请输入${field.label}`"
            />
            <input
              v-else-if="field.type === 'date'"
              v-model="formData[field.key]"
              class="field-input"
              type="text"
              placeholder="格式：2025-01-15"
            />
            <input
              v-else-if="field.type === 'datetime'"
              v-model="formData[field.key]"
              class="field-input"
              type="text"
              placeholder="格式：2025-01-15 14:30"
            />
            <textarea
              v-else-if="field.type === 'textarea'"
              v-model="formData[field.key]"
              class="field-textarea"
              :placeholder="`请输入${field.label}`"
            />
            <picker
              v-else-if="field.type === 'select'"
              mode="selector"
              :value="0"
              :range="field.options || []"
              :range-key="'label'"
              @change="(e) => { formData[field.key] = field.options[e.detail.value]?.value ?? field.options[e.detail.value] ?? '' }"
            >
              <view class="field-picker">
                <text>{{ formData[field.key] || `请选择${field.label}` }}</text>
                <text class="arrow">›</text>
              </view>
            </picker>
            <view v-else-if="field.type === 'image' || field.type === 'file'" class="field-input image-field">
              <input
                v-model="formData[field.key]"
                class="field-url-input"
                placeholder="请输入文件URL，或联系管家上传"
              />
            </view>
            <view v-else class="field-input">
              <text class="placeholder-text">暂不支持此字段类型：{{ field.type }}</text>
            </view>
          </view>
        </scroll-view>
        <view class="form-footer">
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
</style>
