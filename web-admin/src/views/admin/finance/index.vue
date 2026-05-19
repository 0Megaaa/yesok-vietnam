<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(true)
const payments = ref([])

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const normalizePayment = (payment) => {
  const payAmount = Number(payment.pay_amount ?? payment.payAmount ?? payment.amount ?? 0)
  return {
    ...payment,
    payment_no:
      payment.third_trade_no || payment.payment_no || payment.paymentNo || `PAY-${payment.id}`,
    amount: payAmount,
    amountText: `${Math.round(payAmount / 100)} ${payment.currency || 'VND'}`,
    pay_status: payment.status || payment.pay_status || payment.payStatus || 'pending',
  }
}

const loadPayments = async () => {
  loading.value = true
  try {
    console.log('[Finance] 发起 GET /v1/admin/payments')
    const res = await request.get('/v1/admin/payments')
    console.log('[Finance] ✅ 返回：', res.data)
    payments.value = (res.data.list || []).map(normalizePayment)
  } catch (error) {
    console.error('[Finance] ❌ 报错：', error)
    showToast(error?.message || '财务流水加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const statusLabel = (status) => {
  const map = { pending: '待确认', paid: '已确认', failed: '失败', refunded: '已退款' }
  return map[status] || status
}

const statusClass = (status) => {
  const map = { pending: 'status-pending', paid: 'status-paid', failed: 'status-failed', refunded: 'status-refunded' }
  return map[status] || ''
}

onMounted(() => {
  console.log('[Finance] 组件挂载，调用 loadPayments')
  loadPayments()
})
</script>

<template>
  <div class="finance-page">
    <div class="page-header">
      <span class="page-title">财务流水</span>
      <span class="page-sub">订单状态推进到收款节点后自动生成 payment_records</span>
      <el-button class="refresh-btn" type="default" :loading="loading" @click="loadPayments">
        刷新流水
      </el-button>
    </div>

    <div class="content-card">
      <div class="table-scroll">
        <div class="table-list">
          <div class="table-header">
            <span>流水号</span>
            <span>关联订单</span>
            <span>金额</span>
            <span>支付方式</span>
            <span>状态</span>
            <span>创建时间</span>
          </div>

          <div
            v-for="payment in payments"
            :key="payment.id"
            class="table-row"
          >
            <span class="payment-no">{{ payment.payment_no }}</span>
            <span class="order-ref">订单 #{{ payment.order_id }}</span>
            <span class="amount">{{ payment.amountText }}</span>
            <span class="method">{{ payment.pay_method || payment.payMethod || '-' }}</span>
            <span class="status-tag" :class="statusClass(payment.pay_status)">
              {{ statusLabel(payment.pay_status) }}
            </span>
            <span class="time muted">{{ payment.created_at }}</span>
          </div>

          <div v-if="!payments.length" class="empty">
            暂无财务流水，推进订单"确认收款"后自动生成。
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.finance-page {
  width: 100%;
  min-width: 800px;
}

.page-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 900;
  color: #12312c;
}

.page-sub {
  flex: 1;
  color: #6b7c78;
  font-size: 13px;
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

.content-card {
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
}

.table-scroll {
  width: 100%;
  overflow-x: auto;
}

.table-list {
  min-width: 800px;
}

.table-header {
  display: grid;
  grid-template-columns: 2fr 1fr 1.5fr 1fr 1fr 2fr;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 2px solid rgba(0, 77, 64, 0.1);
  color: #6b7c78;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.table-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1.5fr 1fr 1fr 2fr;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
  font-size: 13px;
  align-items: center;
}

.table-row:last-child {
  border-bottom: none;
}

.payment-no {
  font-weight: 700;
  color: #12312c;
  font-size: 12px;
}

.order-ref {
  color: #004d40;
  font-weight: 600;
}

.amount {
  font-weight: 900;
  color: #12312c;
}

.method {
  color: #6b7c78;
}

.time {
  color: #6b7c78;
  font-size: 12px;
}

.muted {
  color: #6b7c78;
}

.status-tag {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  text-align: center;
}

.status-pending {
  color: #6b7c78;
  background: rgba(0, 77, 64, 0.08);
}

.status-paid {
  color: #004d40;
  background: rgba(0, 77, 64, 0.12);
}

.status-failed {
  color: #b42318;
  background: rgba(180, 35, 24, 0.09);
}

.status-refunded {
  color: #c5a059;
  background: rgba(197, 160, 89, 0.12);
}

.empty {
  padding: 40px 20px;
  color: #6b7c78;
  text-align: center;
  font-size: 13px;
}
</style>
