<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const router = useRouter()
const loading = ref(true)
const activeFilter = ref('all')
const orders = ref([])

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
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

const normalizeOrder = (order) => ({
  ...order,
  order_no: order.order_no || order.orderNo,
  service_name: order.service_name || order.serviceName,
  current_status: order.macro_status || order.current_status || order.currentStatus,
  current_stage: order.current_stage || '',
  total_amount: order.total_amount || order.amount || 0,
  totalAmountText:
    order.price || `${Math.round((order.total_amount || 0) / 100)} ${order.currency || 'VND'}`,
  form_data: order.form_data || order.formData || {},
  payment_status: order.payment_status || order.pay_status || 'pending',
  macro_status: order.macro_status || order.current_status || order.currentStatus,
  // actionNodes 字段映射：后端返回 button_label → 按钮文案, target_status → 目标状态
  actionNodes: (order.actionNodes || []).map((node) => ({
    id: node.id,
    action_name: node.action_name || '',
    button_label: node.button_label || '',
    target_status: node.target_status || node.targetStatus || '',
    need_audit: node.need_audit || false,
  })),
})

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

const applyWorkflowAction = async (order, node) => {
  try {
    const res = await request.put(`/v1/admin/orders/${order.id}`, {
      action_name: node.action_name,
      remark: `${node.action_name}：由后台管家执行`,
    })
    orders.value = orders.value.map((item) =>
      item.id === order.id ? normalizeOrder(res.data.order || res.data) : item
    )
    showToast(`${node.action_name}已执行`, 'success')
  } catch (error) {
    showToast(error?.message || '流程推进失败', 'error')
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
            <span class="status">{{ statusMap[order.current_status] || order.current_status }}</span>
            <span v-if="order.current_stage && order.current_stage !== order.current_status" class="stage-tag">{{ order.current_stage }}</span>
            <span class="payment-tag" :class="order.payment_status">{{ paymentStatusMap[order.payment_status] || order.payment_status }}</span>
          </div>
        </div>

        <div class="order-meta">
          <span>金额：{{ order.totalAmountText }}</span>
          <span>支付：{{ order.payment_status }}</span>
          <span>时间：{{ order.created_at }}</span>
        </div>

        <div class="json-box">
          <span v-for="(value, key) in order.form_data" :key="key">
            {{ key }}：{{ value }}
          </span>
        </div>

        <div class="actions" @click.stop>
          <el-button
            v-for="node in order.actionNodes"
            :key="node.id"
            class="action-btn"
            :class="{ payment: node.require_material }"
            type="default"
            @click="applyWorkflowAction(order, node)"
          >
            {{ node.action_name }}
          </el-button>
          <span v-if="!order.actionNodes.length" class="muted">暂无下一步动作</span>
          <el-button class="detail-btn" type="default" @click.stop="goToDetail(order.id)">
            查看详情
          </el-button>
        </div>
      </div>

      <div v-if="!filteredOrders.length" class="empty">
        暂无符合条件的订单
      </div>
    </div>
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
</style>
