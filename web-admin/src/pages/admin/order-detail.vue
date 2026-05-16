<script setup>
import { computed, onMounted, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import request from '@/api/request'

const orderId = ref('YS20260515001')
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

// showSafeToast 安全展示详情页反馈。
// 意图：兼容 UniApp 与普通 H5 预览环境。
// 实现步骤：
// 1. 优先使用 UniApp showToast。
// 2. 若运行环境不存在 uni，则降级输出控制台。
// 3. 保持页面加载失败时仍能稳定展示提示。
// 返回：无返回值。
const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Yesok Admin Detail]', title)
}

// loadOrderDetail 加载订单详情。
// 意图：读取主订单、动态表单、流程轨迹和支付记录，形成后台完整履约视图。
// 实现步骤：
// 1. 使用订单号请求 Mock Admin 详情接口。
// 2. 将后端返回的 formData 解析为可渲染条目。
// 3. 结束加载态，失败时展示中文轻提示。
// 返回：Promise<void>。
const loadOrderDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/v1/admin/orders/${orderId.value}`)
    order.value = res.data
  } catch (error) {
    showSafeToast('订单详情加载失败')
  } finally {
    loading.value = false
  }
}

// goBack 返回后台订单看板。
// 意图：保证 UniApp 和普通浏览器预览都能从详情页回到 B 端看板。
// 实现步骤：
// 1. 优先调用 navigateBack。
// 2. 无历史栈或非 UniApp 环境时跳转到管理端首页路径。
// 3. 避免 H5 直达详情页无法返回。
// 返回：无返回值。
const goBack = () => {
  if (typeof uni !== 'undefined' && uni?.navigateBack) {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
    if (pages.length > 1) {
      uni.navigateBack()
      return
    }
  }
  window.location.href = '/pages/admin/index'
}

// applyWorkflowAction 在详情页推进订单流程。
// 意图：与订单看板保持同一套动态按钮逻辑，验证状态驱动闭环。
// 实现步骤：
// 1. 读取节点配置中的目标状态。
// 2. 调用 Mock PUT 接口模拟后端更新。
// 3. 用返回值刷新详情页动作区与订单状态。
// 返回：Promise<void>。
const applyWorkflowAction = async (node) => {
  if (!order.value) return
  try {
    const res = await request.put(`/v1/admin/orders/${order.value.id}`, {
      targetStatus: node.targetStatus,
      statusText: statusMap[node.targetStatus] || '状态已更新',
    })
    order.value = res.data
    showSafeToast(`${node.buttonName}已执行`)
  } catch (error) {
    showSafeToast('流程推进失败')
  }
}

onLoad((options) => {
  if (options?.id) orderId.value = options.id
  loadOrderDetail()
})

onMounted(() => {
  if (typeof window !== 'undefined') {
    const params = new URLSearchParams(window.location.search)
    if (params.get('id')) orderId.value = params.get('id')
  }
  loadOrderDetail()
})
</script>

<template>
  <view class="detail-page">
    <view class="detail-hero">
      <view class="back-btn" @click="goBack">‹</view>
      <view>
        <text class="eyebrow">ORDER WORKFLOW</text>
        <text class="hero-title">订单履约详情</text>
        <text class="hero-desc">动态表单、状态轨迹、支付记录与下一步动作统一在此管理。</text>
      </view>
    </view>

    <view v-if="loading" class="info-card empty-card">正在加载订单详情...</view>
    <template v-else-if="order">
      <view class="summary-card">
        <view class="summary-head">
          <view class="service-icon">{{ order.icon }}</view>
          <view class="summary-main">
            <text class="service-name">{{ order.serviceName }}</text>
            <text class="order-no">{{ order.orderNo }}</text>
          </view>
          <view class="status-pill" :class="order.currentStatus">{{ statusMap[order.currentStatus] || order.statusText }}</view>
        </view>
        <view class="summary-grid">
          <view>
            <text class="label">客户</text>
            <text class="value">{{ order.appUserName }}</text>
          </view>
          <view>
            <text class="label">联系电话</text>
            <text class="value">{{ order.appUserPhone }}</text>
          </view>
          <view>
            <text class="label">金额</text>
            <text class="value price">{{ order.totalAmountText }}</text>
          </view>
          <view>
            <text class="label">管家</text>
            <text class="value">{{ order.managerName }}</text>
          </view>
        </view>
      </view>

      <view class="info-card">
        <view class="section-title">业务 JSON 详情</view>
        <view v-for="entry in formEntries" :key="entry.key" class="json-row">
          <text class="json-key">{{ entry.key }}</text>
          <text class="json-value">{{ entry.value }}</text>
        </view>
      </view>

      <view class="info-card">
        <view class="section-title">状态轨迹</view>
        <view v-for="timeline in order.timelines" :key="timeline.id" class="timeline-row">
          <view class="timeline-dot"></view>
          <view class="timeline-body">
            <text class="timeline-title">{{ timeline.label }}</text>
            <text class="timeline-desc">{{ timeline.remark }}</text>
            <text class="timeline-time">{{ timeline.operator }} · {{ timeline.createdAt }}</text>
          </view>
        </view>
      </view>

      <view class="info-card">
        <view class="section-title">财务对账</view>
        <view v-if="!order.payments?.length" class="empty-line">当前订单暂无支付记录</view>
        <view v-for="payment in order.payments" :key="payment.id" class="payment-row">
          <view>
            <text class="payment-id">{{ payment.id }}</text>
            <text class="payment-method">{{ payment.payMethod }} · {{ payment.thirdTradeNo }}</text>
          </view>
          <view class="payment-right">
            <text class="payment-amount">{{ payment.payAmountText }}</text>
            <text class="payment-status">{{ payment.status }}</text>
          </view>
        </view>
      </view>

      <view class="info-card action-card">
        <view class="section-title">下一步动作</view>
        <view class="workflow-actions">
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
        </view>
      </view>
    </template>
  </view>
</template>

<style scoped>
/* 意图：构建后台详情页的薄荷灰青背景和纵向信息流。 */
/* 步骤：页面底色使用 Yesok 2.0 灰青色，并通过卡片承载订单、JSON、轨迹和支付记录。 */
/* 返回：结构清晰、适合运营人员快速核查的订单详情视图。 */
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
}

.eyebrow,
.hero-title,
.hero-desc,
.service-name,
.order-no,
.label,
.value,
.json-key,
.json-value,
.timeline-title,
.timeline-desc,
.timeline-time,
.payment-id,
.payment-method,
.payment-amount,
.payment-status,
.empty-line {
  display: block;
}

.eyebrow {
  margin-bottom: 8px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 1.4px;
}

.hero-title {
  font-size: 24px;
  font-weight: 900;
}

.hero-desc {
  max-width: 280px;
  margin-top: 8px;
  color: rgba(255, 255, 255, 0.82);
  font-size: 12px;
  line-height: 1.7;
}

.summary-card,
.info-card {
  margin: 14px 14px 0;
  padding: 16px;
  border-radius: 32px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.05);
}

.summary-card {
  margin-top: -16px;
}

.empty-card {
  color: #6b7c78;
  font-size: 13px;
  text-align: center;
}

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

.summary-main {
  flex: 1;
}

.service-name {
  color: #12312c;
  font-size: 18px;
  font-weight: 900;
}

.order-no {
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

.status-pill.payment_pending {
  color: #7a5a21;
  background: rgba(197, 160, 89, 0.2);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
  padding: 12px;
  border-radius: 24px;
  background: #f2f6f5;
}

.label {
  color: #87938f;
  font-size: 10px;
}

.value {
  margin-top: 4px;
  color: #12312c;
  font-size: 13px;
  font-weight: 800;
}

.value.price {
  color: #e97832;
}

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

.json-key {
  flex-shrink: 0;
  color: #87938f;
  font-size: 12px;
  font-weight: 800;
}

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

.timeline-body {
  flex: 1;
}

.timeline-title {
  color: #12312c;
  font-size: 13px;
  font-weight: 900;
}

.timeline-desc {
  margin-top: 5px;
  color: #6b7c78;
  font-size: 12px;
  line-height: 1.6;
}

.timeline-time {
  margin-top: 6px;
  color: #a0aaa7;
  font-size: 10px;
}

.payment-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-radius: 22px;
  background: rgba(197, 160, 89, 0.09);
}

.payment-id {
  color: #12312c;
  font-size: 13px;
  font-weight: 900;
}

.payment-method {
  margin-top: 5px;
  color: #6b7c78;
  font-size: 11px;
}

.payment-right {
  text-align: right;
}

.payment-amount {
  color: #e97832;
  font-size: 15px;
  font-weight: 900;
}

.payment-status,
.empty-line {
  margin-top: 4px;
  color: #7a5a21;
  font-size: 11px;
}

.workflow-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

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
}

.action-btn.payment {
  color: #12312c;
  background: #c5a059;
}

.action-btn.material {
  color: #004d40;
  background: rgba(0, 77, 64, 0.1);
}

.action-btn.disabled {
  color: #9aa3b5;
  background: #eef2f1;
}
</style>
