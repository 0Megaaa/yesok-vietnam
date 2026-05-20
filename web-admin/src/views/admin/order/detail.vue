<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const router = useRouter()
const route = useRoute()

const orderId = ref(route.params.id || 'YS20260515001')
const loading = ref(true)
const order = ref(null)

const statusMap = {
  pending: '等待受理',
  processing: '服务进行中',
  supplementing: '待客户补充',
  payment_pending: '待支付确认',
  completed: '已完成',
}

const formEntries = computed(() => {
  if (!order.value?.formData) return []
  return Object.entries(order.value.formData).map(([key, value]) => ({ key, value }))
})

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const loadOrderDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/v1/admin/orders/${orderId.value}`)
    order.value = res.data
  } catch (error) {
    showToast('订单详情加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/admin')
  }
}

const applyWorkflowAction = async (node) => {
  if (!order.value) return
  try {
    const res = await request.put(`/v1/admin/orders/${order.value.id}`, {
      targetStatus: node.targetStatus,
      statusText: statusMap[node.targetStatus] || '状态已更新',
    })
    order.value = res.data
    showToast(`${node.buttonName}已执行`, 'success')
  } catch (error) {
    showToast('流程推进失败', 'error')
  }
}

onMounted(() => {
  if (route.params.id) orderId.value = route.params.id
  loadOrderDetail()
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
      <div class="summary-card">
        <div class="summary-head">
          <div class="service-icon">{{ order.icon }}</div>
          <div class="summary-main">
            <span class="service-name">{{ order.serviceName }}</span>
            <span class="order-no">{{ order.orderNo }}</span>
          </div>
          <div class="status-pill" :class="order.currentStatus">{{ statusMap[order.currentStatus] || order.statusText }}</div>
        </div>
        <div class="summary-grid">
          <div>
            <span class="label">客户</span>
            <span class="value">{{ order.appUserName }}</span>
          </div>
          <div>
            <span class="label">联系电话</span>
            <span class="value">{{ order.appUserPhone }}</span>
          </div>
          <div>
            <span class="label">金额</span>
            <span class="value price">{{ order.totalAmountText }}</span>
          </div>
          <div>
            <span class="label">管家</span>
            <span class="value">{{ order.managerName }}</span>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="section-title">业务 JSON 详情</div>
        <div v-for="entry in formEntries" :key="entry.key" class="json-row">
          <span class="json-key">{{ entry.key }}</span>
          <span class="json-value">{{ entry.value }}</span>
        </div>
      </div>

      <div class="info-card">
        <div class="section-title">状态轨迹</div>
        <div v-for="timeline in order.timelines" :key="timeline.id" class="timeline-row">
          <div class="timeline-dot"></div>
          <div class="timeline-body">
            <span class="timeline-title">{{ timeline.label }}</span>
            <span class="timeline-desc">{{ timeline.remark }}</span>
            <span class="timeline-time">{{ timeline.operator }} · {{ timeline.createdAt }}</span>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="section-title">财务对账</div>
        <div v-if="!order.payments?.length" class="empty-line">当前订单暂无支付记录</div>
        <div v-for="payment in order.payments" :key="payment.id" class="payment-row">
          <div>
            <span class="payment-id">{{ payment.id }}</span>
            <span class="payment-method">{{ payment.payMethod }} · {{ payment.thirdTradeNo }}</span>
          </div>
          <div class="payment-right">
            <span class="payment-amount">{{ payment.payAmountText }}</span>
            <span class="payment-status">{{ payment.status }}</span>
          </div>
        </div>
      </div>

      <div class="info-card action-card">
        <div class="section-title">下一步动作</div>
        <div class="workflow-actions">
          <button
            v-for="node in order.actionNodes"
            :key="node.id"
            class="action-btn"
            :class="{ payment: node.triggerPayment, material: node.requiredMaterial }"
            @click="applyWorkflowAction(node)"
          >
            {{ node.buttonName }}
          </button>
          <button v-if="!order.actionNodes?.length" class="action-btn disabled" disabled>暂无下一步动作</button>
        </div>
      </div>
    </template>
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

.eyebrow, .hero-title, .hero-desc, .service-name, .order-no, .label, .value, .json-key, .json-value, .timeline-title, .timeline-desc, .timeline-time, .payment-id, .payment-method, .payment-amount, .payment-status, .empty-line, .section-title {
  display: block;
}

.eyebrow {
  margin-bottom: 8px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 1.4px;
}

.hero-title { font-size: 24px; font-weight: 900; }

.hero-desc {
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

.service-name { color: #12312c; font-size: 18px; font-weight: 900; }

.order-no { margin-top: 5px; color: #6b7c78; font-size: 11px; }

.status-pill {
  padding: 6px 10px;
  border-radius: 999px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  font-size: 10px;
  font-weight: 900;
}

.status-pill.payment_pending { color: #7a5a21; background: rgba(197, 160, 89, 0.2); }

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
  padding: 12px;
  border-radius: 24px;
  background: #f2f6f5;
}

.label { color: #87938f; font-size: 10px; }

.value { margin-top: 4px; color: #12312c; font-size: 13px; font-weight: 800; }
.value.price { color: #e97832; }

.section-title {
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

.timeline-title { color: #12312c; font-size: 13px; font-weight: 900; }

.timeline-desc { margin-top: 5px; color: #6b7c78; font-size: 12px; line-height: 1.6; }

.timeline-time { margin-top: 6px; color: #a0aaa7; font-size: 10px; }

.payment-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-radius: 22px;
  background: rgba(197, 160, 89, 0.09);
}

.payment-id { color: #12312c; font-size: 13px; font-weight: 900; }
.payment-method { margin-top: 5px; color: #6b7c78; font-size: 11px; }
.payment-right { text-align: right; }
.payment-amount { color: #e97832; font-size: 15px; font-weight: 900; }

.payment-status, .empty-line {
  margin-top: 4px;
  color: #7a5a21;
  font-size: 11px;
}

.workflow-actions { display: flex; flex-wrap: wrap; gap: 10px; }

.action-btn {
  height: 36px;
  margin: 0;
  padding: 0 16px;
  border: none;
  border-radius: 18px;
  color: #fff;
  background: #004d40;
  font-size: 12px;
  font-weight: 800;
  line-height: 36px;
  cursor: pointer;
}

.action-btn.payment { color: #12312c; background: #c5a059; }

.action-btn.material { color: #004d40; background: rgba(0, 77, 64, 0.1); }

.action-btn.disabled { color: #9aa3b5; background: #eef2f1; }
</style>
