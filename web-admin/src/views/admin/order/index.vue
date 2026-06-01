<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request, { ORIGIN_URL } from '@/api/request'
import DynamicForm from '@/components/DynamicForm.vue'

const router = useRouter()
const loading = ref(true)
const activeFilter = ref('all')
const orders = ref([])

// form_input 弹窗状态
const dynamicFormRef = ref(null)
const formInputVisible = ref(false)
const formInputOrder = ref(null)
const formInputAction = ref(null)
const formInputData = ref({})
const formInputLoading = ref(false)

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

// materialImageItems 返回前3张图片项，避免列表卡片过重
const materialImageItems = (order) => {
  return (order.form_items || [])
    .filter((item) => isImageFile(item.display_value ?? item.value))
    .slice(0, 3)
}

// textFormItems 返回非图片的文本项，最多8条
const textFormItems = (order) => {
  return (order.form_items || [])
    .filter((item) => !isMaterialFile(item.display_value ?? item.value))
    .slice(0, 8)
}

const statusMap = {
  pending: '待受理',
  reviewing: '资料审核中',
  quoted: '已报价',
  paid: '已收款',
  in_progress: '履约中',
  processing: '服务中',
  completed: '已完成',
  cancelled: '已取消',
  refunded: '已退款',
}

const paymentStatusMap = {
  unpaid: '未支付',
  paid: '已支付',
  refunded: '已退款',
}

// formatMoney 格式化金额为人民币
const formatMoney = (amount) => {
  const n = Number(amount || 0)
  return `¥${n.toLocaleString('zh-CN')}`
}

const normalizeOrder = (order) => ({
  ...order,
  order_no: order.order_no || order.orderNo,
  service_name: order.service_name || order.serviceName,
  current_status: order.macro_status || order.current_status || order.currentStatus,
  current_stage: order.current_stage || '',
  current_stage_text: order.current_stage_text || '',
  macro_status: order.macro_status || order.current_status || order.currentStatus,
  macro_status_text: order.macro_status_text || '',
  payment_status: order.payment_status || order.pay_status || 'unpaid',
  payment_status_text: order.payment_status_text || '',
  total_amount: order.total_amount || order.amount || 0,
  totalAmountText: formatMoney(order.total_amount || order.amount || 0),
  form_items: order.form_items || [],
  form_data: order.form_data || order.formData || {},
  timelines: order.timelines || [],
  // action_nodes 完整字段
  action_nodes: (order.action_nodes || order.actionNodes || []).map((node) => ({
    id: node.id,
    action_name: node.action_name || '',
    action_name_text: node.action_name_text || '',
    button_label: node.button_label || node.action_name_text || node.action_name || '',
    action_type: node.action_type || 'button_click',
    form_fields: node.form_fields || [],
    target_status: node.target_status || node.targetStatus || '',
    target_status_text: node.target_status_text || '',
    macro_status: node.macro_status || '',
    macro_status_text: node.macro_status_text || '',
    notify_type: node.notify_type || '',
    notify_type_text: node.notify_type_text || '',
    need_audit: node.need_audit || false,
    audit_reject_status: node.audit_reject_status || '',
  })),
})

const hasPendingAudit = (order) => {
  return (order.timelines || []).some((tl) => tl.audit_status === 'pending')
}

const latestRejectedAudit = (order) => {
  return [...(order.timelines || [])].reverse().find((tl) => tl.audit_status === 'rejected')
}

const pendingAuditTimelineOf = (order) => {
  return (order.timelines || []).find((tl) => tl.audit_status === 'pending')
}

// 待审核状态下过滤掉 audit 类动作
const isAuditAction = (action) => {
  if (!action) return false
  return action.is_audit_action === true ||
    ['audit_approve', 'audit_reject', 'audit_rejected'].includes(action.action_name)
}

// 列表可见动作：待审核时返回空数组，隐藏普通 action_nodes
const visibleActionNodes = (order) => {
  if (!order) return []
  if (hasPendingAudit(order)) return []
  return (order.action_nodes || []).filter((a) => !isAuditAction(a))
}

const approveAuditFromList = async (order) => {
  const pending = pendingAuditTimelineOf(order)
  if (!pending || !order?.id) return

  try {
    await ElMessageBox.confirm(
      `确认审核通过订单「${order.order_no || order.orderNo}」？`,
      '审核确认',
      { confirmButtonText: '审核通过', cancelButtonText: '取消', type: 'success' }
    )

    const res = await request.post(`/v1/admin/orders/${order.id}/audit`, {
      timeline_id: pending.id,
      result: 'approved',
      audit_remark: '资料审核通过',
    })

    const next = normalizeOrder(res.data?.order || res.data)
    orders.value = orders.value.map((item) => (item.id === order.id ? next : item))

    showToast('审核已通过', 'success')
  } catch (error) {
    if (error !== 'cancel') {
      showToast(error?.response?.data?.error || error?.message || '审核失败', 'error')
    }
  }
}

const rejectAuditFromList = async (order) => {
  const pending = pendingAuditTimelineOf(order)
  if (!pending || !order?.id) return

  try {
    const { value } = await ElMessageBox.prompt(
      `请填写订单「${order.order_no || order.orderNo}」审核不通过原因`,
      '审核不通过',
      {
        confirmButtonText: '确认驳回',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputPlaceholder: '例如：护照首页照片模糊，请重新上传',
        inputValidator: (val) => {
          if (!val || !val.trim()) return '请填写审核不通过原因'
          return true
        },
      }
    )

    const res = await request.post(`/v1/admin/orders/${order.id}/audit`, {
      timeline_id: pending.id,
      result: 'rejected',
      audit_remark: value.trim(),
    })

    const next = normalizeOrder(res.data?.order || res.data)
    orders.value = orders.value.map((item) => (item.id === order.id ? next : item))

    showToast('已驳回，订单已回到补资料流程', 'success')
  } catch (error) {
    if (error !== 'cancel') {
      showToast(error?.response?.data?.error || error?.message || '审核失败', 'error')
    }
  }
}

const loadOrders = async () => {
  loading.value = true
  try {
    console.log('[Orders] 发起 GET /v1/admin/orders')
    const res = await request.get('/v1/admin/orders')
    console.log('[Orders] ✅ 返回：', res.data)
    orders.value = (res.data.list || res.data.orders || []).map(normalizeOrder)
  } catch (error) {
    console.error('[Orders] ❌ 报错：', error)
    showToast(error?.message || '订单加载失败', 'error')
  } finally {
    loading.value = false
  }
}

// 根据 action_type 分流处理工作流动作
const handleWorkflowAction = async (order, node) => {
  if (node.action_type === 'form_input') {
    openFormInput(order, node)
    return
  }

  if (node.action_type === 'wx_pay') {
    showToast('支付动作请客户在 C 端完成，后台不直接执行支付', 'info')
    return
  }

  await executeButtonClick(order, node)
}

// 直接执行 button_click 动作
const executeButtonClick = async (order, node) => {
  try {
    const { value: remark } = await ElMessageBox.prompt(
      `确认执行「${node.button_label || node.action_name}」？可填写备注：`,
      '确认操作',
      {
        confirmButtonText: '确认执行',
        cancelButtonText: '取消',
        inputPlaceholder: '备注信息（选填）',
      }
    )

    const res = await request.post(`/v1/admin/orders/${order.id}/action`, {
      action_name: node.action_name,
      remark: remark || '',
      input_data: {},
    })

    const next = normalizeOrder(res.data?.order || res.data)
    orders.value = orders.value.map((item) => (item.id === order.id ? next : item))
    showToast(`「${node.button_label || node.action_name}」已执行`, 'success')
  } catch (error) {
    if (error !== 'cancel') {
      showToast(error?.response?.data?.error || error?.message || '流程推进失败', 'error')
    }
  }
}

// 打开 form_input 弹窗
const openFormInput = (order, node) => {
  formInputOrder.value = order
  formInputAction.value = node
  formInputData.value = {}
  formInputVisible.value = true
}

// 提交 form_input
const submitFormInput = async () => {
  if (!formInputOrder.value || !formInputAction.value) return

  if (dynamicFormRef.value?.validateAll && !dynamicFormRef.value.validateAll()) {
    return
  }

  formInputLoading.value = true
  try {
    const res = await request.post(`/v1/admin/orders/${formInputOrder.value.id}/action`, {
      action_name: formInputAction.value.action_name,
      remark: '',
      input_data: formInputData.value,
    })

    const next = normalizeOrder(res.data?.order || res.data)
    orders.value = orders.value.map((item) =>
      item.id === formInputOrder.value.id ? next : item
    )

    formInputVisible.value = false
    formInputOrder.value = null
    formInputAction.value = null
    formInputData.value = {}

    showToast(`「${next.service_name || '订单'}」流程已推进`, 'success')
  } catch (error) {
    showToast(error?.response?.data?.error || error?.message || '流程推进失败', 'error')
  } finally {
    formInputLoading.value = false
  }
}

const goToDetail = (orderId) => {
  router.push(`/admin/order/${orderId}`)
}

const filters = computed(() => [
  { key: 'all', label: '全部', count: orders.value.length },
  {
    key: 'pending',
    label: '待受理',
    count: orders.value.filter((o) => o.current_status === 'pending').length,
  },
  {
    key: 'reviewing',
    label: '资料审核',
    count: orders.value.filter((o) => o.current_status === 'reviewing').length,
  },
  {
    key: 'quoted',
    label: '已报价',
    count: orders.value.filter((o) => o.current_status === 'quoted').length,
  },
  {
    key: 'paid',
    label: '已收款',
    count: orders.value.filter((o) => o.current_status === 'paid').length,
  },
  {
    key: 'processing',
    label: '服务中',
    count: orders.value.filter((o) => o.current_status === 'processing' || o.current_status === 'in_progress').length,
  },
  {
    key: 'completed',
    label: '已完成',
    count: orders.value.filter((o) => o.current_status === 'completed').length,
  },
])

const filteredOrders = computed(() =>
  activeFilter.value === 'all'
    ? orders.value
    : orders.value.filter((o) => o.current_status === activeFilter.value)
)

onMounted(() => {
  console.log('[Orders] 组件挂载，调用 loadOrders')
  loadOrders()
})

onUnmounted(() => {
  loading.value = false
})
</script>

<template>
  <div class="orders-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <span class="page-title">订单中心</span>
      <el-button class="refresh-btn" type="default" :loading="loading" @click="loadOrders">
        刷新订单
      </el-button>
    </div>

    <!-- 状态筛选 -->
    <div class="chip-row">
      <el-button
        v-for="filter in filters"
        :key="filter.key"
        class="chip"
        :class="{ active: activeFilter === filter.key }"
        type="default"
        @click="activeFilter = filter.key"
      >
        {{ filter.label }} {{ filter.count }}
      </el-button>
    </div>

    <!-- 订单列表 -->
    <div v-if="loading" class="empty">正在加载真实订单...</div>
    <div v-else class="order-list">
      <div
        v-for="order in filteredOrders"
        :key="order.id"
        class="order-card"
        @click="goToDetail(order.id)"
      >
        <div class="order-top">
          <div class="order-info">
            <span class="order-title">{{ order.service_name }}</span>
            <span class="muted">
              {{ order.order_no }} · {{ order.contact_name }}
              {{ order.contact_phone }}
            </span>
          </div>
          <div class="status-badges">
            <span v-if="hasPendingAudit(order)" class="audit-badge pending">待审核</span>
            <span v-else-if="latestRejectedAudit(order)" class="audit-badge rejected">审核未通过</span>
            <span class="status">{{ order.macro_status_text || statusMap[order.current_status] || order.current_status }}</span>
            <span v-if="order.current_stage_text && order.current_stage_text !== order.current_status" class="stage-tag">{{ order.current_stage_text }}</span>
            <span class="payment-tag" :class="order.payment_status">{{ order.payment_status_text || paymentStatusMap[order.payment_status] || order.payment_status }}</span>
          </div>
        </div>

        <div class="order-meta">
          <span>金额：{{ order.totalAmountText }}</span>
          <span>时间：{{ order.created_at }}</span>
        </div>

        <div class="json-box">
          <template v-if="order.form_items && order.form_items.length">
            <span v-for="item in textFormItems(order)" :key="item.key">
              {{ item.label }}：{{ item.display_value ?? item.value ?? '—' }}
            </span>

            <div v-if="materialImageItems(order).length" class="material-preview-row" @click.stop>
              <el-image
                v-for="item in materialImageItems(order)"
                :key="item.key"
                :src="toFullFileUrl(getFileUrl(item.display_value ?? item.value))"
                :preview-src-list="materialImageItems(order).map(img => toFullFileUrl(getFileUrl(img.display_value ?? img.value)))"
                fit="cover"
                class="material-list-thumb"
              />
            </div>
          </template>
          <span v-else class="muted">暂无业务资料</span>
        </div>

        <div class="actions" @click.stop>
          <template v-if="hasPendingAudit(order)">
            <el-button
              class="action-btn audit-pass"
              type="success"
              @click.stop="approveAuditFromList(order)"
            >
              审核通过
            </el-button>
            <el-button
              class="action-btn audit-reject"
              type="danger"
              plain
              @click.stop="rejectAuditFromList(order)"
            >
              审核不通过
            </el-button>
          </template>

          <template v-else>
            <el-button
              v-for="node in visibleActionNodes(order)"
              :key="node.id"
              class="action-btn"
              :class="{ payment: node.need_audit }"
              type="default"
              @click.stop="handleWorkflowAction(order, node)"
            >
              {{ node.button_label || node.action_name_text || node.action_name }}
            </el-button>
            <span v-if="!visibleActionNodes(order).length" class="muted">暂无下一步动作</span>
          </template>

          <el-button class="detail-btn" type="default" @click.stop="goToDetail(order.id)">
            查看详情
          </el-button>
        </div>
      </div>

      <div v-if="!filteredOrders.length" class="empty">
        暂无符合条件的订单
      </div>
    </div>

    <!-- form_input 弹窗 -->
    <el-dialog
      v-model="formInputVisible"
      :title="`填写：${formInputAction?.button_label || formInputAction?.action_name || ''}`"
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
        <el-button type="primary" :loading="formInputLoading" @click="submitFormInput">
          确认执行
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.orders-page {
  width: 100%;
  min-width: 800px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 900;
  color: #12312c;
}

.refresh-btn {
  height: 38px;
  padding: 0 18px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 18px;
}

.chip {
  height: 34px;
  padding: 0 14px;
  color: #6b7c78;
  background: #eef5f2;
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.chip.active {
  color: #fff;
  background: #004d40;
}

.order-list {
  display: grid;
  gap: 14px;
}

.order-card {
  min-width: 1200px;
  padding: 18px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 12px 40px rgba(0, 77, 64, 0.08);
  cursor: pointer;
  transition: box-shadow 0.2s;
  border: 1px solid rgba(255, 255, 255, 0.72);
}

.order-card:hover {
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.15);
}

.order-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.order-info {
  display: flex;
  flex-direction: column;
}

.order-title {
  font-size: 17px;
  font-weight: 900;
  color: #12312c;
}

.muted {
  display: block;
  margin-top: 4px;
  color: #6b7c78;
  font-size: 12px;
}

.status {
  flex-shrink: 0;
  padding: 7px 12px;
  border-radius: 999px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  font-size: 12px;
  font-weight: 900;
}

.status-badges {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.stage-tag {
  padding: 4px 10px;
  border-radius: 999px;
  color: #7a5a21;
  background: rgba(197, 160, 89, 0.15);
  font-size: 10px;
  font-weight: 700;
}

.payment-tag {
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 700;
}

.payment-tag.paid {
  color: #2a7a3a;
  background: rgba(60, 160, 80, 0.15);
}

.payment-tag.unpaid {
  color: #7a2a2a;
  background: rgba(200, 80, 80, 0.15);
}

.audit-badge {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 800;
}
.audit-badge.pending { background: #fff7e6; color: #ad6800; }
.audit-badge.rejected { background: #fff1f0; color: #cf1322; }

.order-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  margin: 12px 0;
  color: #6b7c78;
  font-size: 12px;
}

.json-box {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  padding: 12px;
  border-radius: 18px;
  background: #f2f6f5;
  color: #4c5d59;
  font-size: 12px;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
}

.action-btn {
  height: 34px;
  padding: 0 16px;
  color: #fff;
  background: #004d40;
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.action-btn.payment {
  color: #12312c;
  background: #f5d98f;
}

.action-btn.audit-pass {
  background: #0b6b55;
  border-color: #0b6b55;
  color: #fff;
}

.action-btn.audit-reject {
  background: #fff5f5;
  border-color: #ffb8b8;
  color: #d64545;
}

.detail-btn {
  height: 34px;
  padding: 0 16px;
  color: #6b7c78;
  background: #eef5f2;
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.empty {
  padding: 40px 20px;
  color: #6b7c78;
  text-align: center;
  font-size: 14px;
}

.material-preview-row {
  display: flex;
  gap: 8px;
  margin-top: 10px;
  flex-wrap: wrap;
}
.material-list-thumb {
  width: 64px;
  height: 64px;
  border-radius: 10px;
  border: 1px solid #e5efeb;
  background: #f6faf8;
  overflow: hidden;
}
</style>
