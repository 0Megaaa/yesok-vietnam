<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderDetail, performOrderAction } from '@/api/admin/orders'
import { getServiceActions } from '@/api/admin/services'
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
    pending: '等待受理',
    reviewing: '资料审核中',
    quoted: '已报价',
    paid: '已收款',
    in_progress: '服务履约中',
    completed: '已完成',
    cancelled: '已取消',
    refunded: '已退款',
    failed: '异常',
  }
  return m[code] || code || '未知'
}

const formatMoney = (amount) => {
  if (!amount && amount !== 0) return '—'
  return `${(Number(amount) / 100).toLocaleString('vi-VN')} ₫`
}

const formatTime = (t) => {
  if (!t) return ''
  return new Date(t).toLocaleString('vi-VN', { dateStyle: 'short', timeStyle: 'short' })
}

const formEntries = computed(() => {
  if (!order.value?.form_data) return []
  const raw = order.value.form_data
  const data = typeof raw === 'string' ? JSON.parse(raw) : raw
  return Object.entries(data).map(([k, v]) => ({ key: k, value: typeof v === 'object' ? JSON.stringify(v) : v }))
})

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const loadOrderDetail = async () => {
  loading.value = true
  try {
    const res = await getOrderDetail(orderId.value)
    order.value = res
  } catch {
    showToast('订单详情加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const loadOrderActions = async () => {
  if (!order.value?.service_id || !order.value?.current_stage) return
  actionsLoading.value = true
  try {
    const res = await getServiceActions(order.value.service_id, order.value.current_stage, 'admin')
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

// 根据 action_type 分类
const buttonClickActions = computed(() => actions.value.filter(a => a.action_type === 'button_click'))
const formInputActions = computed(() => actions.value.filter(a => a.action_type === 'form_input'))
const wxPayActions = computed(() => actions.value.filter(a => a.action_type === 'wx_pay'))

// --- button_click 类型：直接执行备注确认 ---
const executeButtonClick = async (action) => {
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
    order.value = res
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
  // 调用 DynamicForm 的 validateAll()
  if (!dynamicFormRef.value?.validateAll()) return

  formInputLoading.value = true
  try {
    const res = await performOrderAction(order.value.id, {
      action_name: formInputAction.value.action_name,
      remark: '',
      input_data: formInputData.value,
    })
    order.value = res
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
    router.push('/admin/order')
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
          <div class="status-pill" :class="order.macro_status || order.currentStatus">
            {{ statusLabel(order.macro_status || order.currentStatus) }}
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
            <span class="value price">{{ formatMoney(order.total_amount || order.totalAmount) }}</span>
          </div>
          <div>
            <span class="label">当前节点</span>
            <span class="value">{{ order.current_stage || order.currentStatus || '—' }}</span>
          </div>
        </div>
      </div>

      <!-- 业务表单详情 -->
      <div class="info-card">
        <div class="section-title">业务表单</div>
        <div v-if="!formEntries.length" class="empty-line">暂无表单数据</div>
        <div v-for="entry in formEntries" :key="entry.key" class="json-row">
          <span class="json-key">{{ entry.key }}</span>
          <span class="json-value">{{ entry.value }}</span>
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
              {{ tl.after_status || tl.label || '—' }}
            </span>
            <span v-if="tl.remark" class="timeline-desc">{{ tl.remark }}</span>
            <span class="timeline-time">
              {{ tl.operator || '系统' }} · {{ formatTime(tl.created_at || tl.createdAt) }}
            </span>
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
            <span class="payment-status" :class="pay.status">{{ pay.status }}</span>
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
        <div v-else-if="!actions.length" class="empty-line">当前节点无操作动作</div>
        <div v-else class="workflow-actions">
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
</style>
