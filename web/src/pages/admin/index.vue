<script setup>
import { computed, onMounted, ref } from 'vue'
import request from '@/api/request'

const loading = ref(true)
const stats = ref({})
const orders = ref([])
const activeFilter = ref('all')

const statusMap = {
  pending: '等待受理',
  processing: '服务进行中',
  supplementing: '待客户补充',
  payment_pending: '待支付确认',
  completed: '已完成',
}

const filters = computed(() => [
  { key: 'all', label: '全部订单', count: orders.value.length },
  { key: 'pending', label: '待受理', count: orders.value.filter((item) => item.currentStatus === 'pending').length },
  { key: 'processing', label: '办理中', count: orders.value.filter((item) => item.currentStatus === 'processing').length },
  { key: 'payment_pending', label: '收款确认', count: orders.value.filter((item) => item.currentStatus === 'payment_pending').length },
])

const filteredOrders = computed(() => {
  if (activeFilter.value === 'all') return orders.value
  return orders.value.filter((item) => item.currentStatus === activeFilter.value)
})

// showSafeToast 安全展示后台操作结果。
// 意图：让 H5、微信小程序和普通浏览器预览都能获得明确反馈。
// 实现步骤：
// 1. 优先调用 UniApp showToast。
// 2. 非 UniApp 环境降级到控制台输出。
// 3. 不阻断主业务流程。
// 返回：无返回值。
const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Yesok Admin]', title)
}

// loadDashboard 加载 B 端看板数据。
// 意图：从 Mock Admin API 获取统计数据和订单列表，验证后台闭环。
// 实现步骤：
// 1. 并发请求订单统计与订单列表。
// 2. 将动态工作流按钮随订单一起写入页面状态。
// 3. 结束加载态并保证失败时有中文提示。
// 返回：Promise<void>。
const loadDashboard = async () => {
  loading.value = true
  try {
    const [statsRes, ordersRes] = await Promise.all([
      request.get('/v1/admin/dashboard/stats'),
      request.get('/v1/admin/orders'),
    ])
    stats.value = statsRes.data
    orders.value = ordersRes.data.list || []
  } catch (error) {
    showSafeToast('后台数据加载失败')
  } finally {
    loading.value = false
  }
}

// openDetail 打开订单详情页面。
// 意图：让运营人员可以查看订单 form_data JSON 业务详情与状态轨迹。
// 实现步骤：
// 1. 读取当前订单号。
// 2. 优先使用 UniApp navigateTo。
// 3. 普通浏览器预览降级为查询参数路由，避免静态服务深链 404。
// 返回：无返回值。
const openDetail = (order) => {
  const url = `/pages/admin/order-detail?id=${encodeURIComponent(order.id)}`
  const staticPreviewUrl = `/?admin=detail&id=${encodeURIComponent(order.id)}`
  if (typeof uni !== 'undefined' && uni?.navigateTo) {
    uni.navigateTo({ url })
    return
  }
  window.location.href = staticPreviewUrl
}

// applyWorkflowAction 执行动态流程按钮动作。
// 意图：后台按钮完全由 sys_workflow_nodes 配置驱动，而非写死在页面中。
// 实现步骤：
// 1. 读取按钮节点的 targetStatus 作为订单目标状态。
// 2. 调用 Mock PUT 接口模拟后端更新。
// 3. 在列表中局部替换订单状态与下一批可用按钮。
// 返回：Promise<void>。
const applyWorkflowAction = async (order, node) => {
  try {
    const res = await request.put(`/v1/admin/orders/${order.id}`, {
      targetStatus: node.targetStatus,
      statusText: statusMap[node.targetStatus] || '状态已更新',
    })
    orders.value = orders.value.map((item) => (item.id === order.id ? res.data : item))
    showSafeToast(`${node.buttonName}已执行`)
  } catch (error) {
    showSafeToast('流程推进失败，请稍后重试')
  }
}

onMounted(loadDashboard)
</script>

<template>
  <view class="admin-page">
    <view class="admin-hero">
      <view>
        <text class="eyebrow">YESOK COMMAND CENTER</text>
        <text class="hero-title">越南管家订单看板</text>
        <text class="hero-desc">以动态流程节点驱动后台动作，统一承接公司注册、签证、接机与高端通道订单。</text>
      </view>
      <view class="hero-avatar">M</view>
    </view>

    <view class="stats-grid">
      <view class="stat-card primary">
        <text class="stat-value">{{ stats.totalOrders || 0 }}</text>
        <text class="stat-label">总订单</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ stats.pendingOrders || 0 }}</text>
        <text class="stat-label">待受理</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ stats.processingOrders || 0 }}</text>
        <text class="stat-label">办理中</text>
      </view>
      <view class="stat-card champagne">
        <text class="stat-value">{{ stats.paymentPendingOrders || 0 }}</text>
        <text class="stat-label">收款确认</text>
      </view>
    </view>

    <scroll-view class="filter-scroll" scroll-x>
      <view class="filter-inner">
        <view
          v-for="filter in filters"
          :key="filter.key"
          class="filter-chip"
          :class="{ active: activeFilter === filter.key }"
          @click="activeFilter = filter.key"
        >
          <text>{{ filter.label }}</text>
          <text class="filter-count">{{ filter.count }}</text>
        </view>
      </view>
    </scroll-view>

    <view class="order-list">
      <view v-if="loading" class="empty-card">正在加载后台看板...</view>
      <view v-else-if="!filteredOrders.length" class="empty-card">当前筛选条件下暂无订单</view>
      <template v-else>
      <view v-for="order in filteredOrders" :key="order.id" class="order-card">
        <view class="order-head" @click="openDetail(order)">
          <view class="service-icon">{{ order.icon }}</view>
          <view class="order-main">
            <text class="service-name">{{ order.serviceName }}</text>
            <text class="order-no">{{ order.orderNo }} · {{ order.appUserName }}</text>
          </view>
          <view class="status-pill" :class="order.currentStatus">{{ statusMap[order.currentStatus] || order.statusText }}</view>
        </view>

        <view class="order-meta">
          <view>
            <text class="meta-label">专属管家</text>
            <text class="meta-value">{{ order.managerName }}</text>
          </view>
          <view>
            <text class="meta-label">订单金额</text>
            <text class="meta-value price">{{ order.totalAmountText }}</text>
          </view>
          <view>
            <text class="meta-label">更新时间</text>
            <text class="meta-value">{{ order.updatedAt }}</text>
          </view>
        </view>

        <view class="form-preview" @click="openDetail(order)">
          <text v-for="(value, key) in order.formData" :key="key" class="form-line">{{ key }}：{{ value }}</text>
        </view>

        <view class="workflow-actions">
          <button
            v-for="node in order.actionNodes"
            :key="node.id"
            class="action-btn"
            :class="{ payment: node.triggerPayment, material: node.requiredMaterial }"
            @click.stop="applyWorkflowAction(order, node)"
          >
            {{ node.buttonName }}
          </button>
          <button v-if="!order.actionNodes?.length" class="action-btn disabled" disabled>暂无下一步动作</button>
        </view>
      </view>
      </template>
    </view>
  </view>
</template>

<style scoped>
/* 意图：搭建 B 端后台的热带现代主义底色与安全留白。 */
/* 步骤：使用薄荷灰青背景、顶部深绿色英雄区和底部滚动空间。 */
/* 返回：可在 H5 与小程序中稳定展示的后台页面容器。 */
.admin-page {
  min-height: 100vh;
  padding-bottom: 32px;
  background: #f2f6f5;
}

.admin-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 54px 22px 28px;
  color: #fff;
  background: radial-gradient(circle at 85% 18%, rgba(197, 160, 89, 0.42), transparent 32%), linear-gradient(135deg, #004d40, #07362f);
  border-radius: 0 0 32px 32px;
  box-shadow: 0 22px 54px rgba(0, 77, 64, 0.2);
}

.eyebrow,
.hero-title,
.hero-desc,
.stat-value,
.stat-label,
.service-name,
.order-no,
.meta-label,
.meta-value,
.form-line {
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
  line-height: 1.25;
}

.hero-desc {
  max-width: 260px;
  margin-top: 10px;
  color: rgba(255, 255, 255, 0.82);
  font-size: 12px;
  line-height: 1.7;
}

.hero-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border: 1px solid rgba(255, 255, 255, 0.38);
  border-radius: 50%;
  color: #7a5a21;
  background: rgba(255, 255, 255, 0.86);
  font-weight: 900;
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.14);
}

/* 意图：让关键经营指标形成高端管理驾驶舱视觉。 */
/* 步骤：四宫格布局，每张卡片使用大圆角与深绿色轻投影。 */
/* 返回：后台人员进入页面即可快速理解运营状态。 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin: -18px 14px 14px;
}

.stat-card {
  min-height: 84px;
  padding: 16px;
  border-radius: 28px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.06);
}

.stat-card.primary {
  color: #fff;
  background: linear-gradient(135deg, #004d40, #00695c);
}

.stat-card.champagne {
  background: linear-gradient(135deg, rgba(197, 160, 89, 0.18), #fff);
}

.stat-value {
  color: inherit;
  font-size: 26px;
  font-weight: 900;
}

.stat-label {
  margin-top: 6px;
  color: rgba(18, 49, 44, 0.58);
  font-size: 12px;
  font-weight: 700;
}

.stat-card.primary .stat-label {
  color: rgba(255, 255, 255, 0.78);
}

.filter-scroll {
  width: 100%;
  white-space: nowrap;
}

.filter-inner {
  display: inline-flex;
  gap: 10px;
  padding: 2px 14px 14px;
}

.filter-chip {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  padding: 9px 14px;
  border-radius: 999px;
  color: #6b7c78;
  background: rgba(255, 255, 255, 0.82);
  font-size: 12px;
  font-weight: 800;
}

.filter-chip.active {
  color: #fff;
  background: #004d40;
  box-shadow: 0 12px 28px rgba(0, 77, 64, 0.18);
}

.filter-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  background: rgba(197, 160, 89, 0.2);
  font-size: 10px;
}

.order-list {
  padding: 0 14px;
}

.empty-card,
.order-card {
  margin-bottom: 14px;
  border-radius: 32px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.05);
}

.empty-card {
  padding: 24px;
  color: #6b7c78;
  font-size: 13px;
  text-align: center;
}

.order-card {
  padding: 16px;
}

.order-head {
  display: flex;
  gap: 12px;
  align-items: center;
}

.service-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 50px;
  height: 50px;
  border-radius: 18px;
  background: rgba(0, 77, 64, 0.08);
  font-size: 24px;
}

.order-main {
  flex: 1;
  min-width: 0;
}

.service-name {
  color: #12312c;
  font-size: 16px;
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

.order-meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
  padding: 12px;
  border-radius: 22px;
  background: #f2f6f5;
}

.meta-label {
  color: #87938f;
  font-size: 10px;
}

.meta-value {
  margin-top: 4px;
  color: #12312c;
  font-size: 12px;
  font-weight: 800;
}

.meta-value.price {
  color: #e97832;
}

.form-preview {
  margin-top: 12px;
  padding: 12px;
  border-left: 3px solid #c5a059;
  border-radius: 18px;
  background: rgba(197, 160, 89, 0.08);
}

.form-line {
  color: #6b7c78;
  font-size: 11px;
  line-height: 1.65;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 意图：根据动态流程节点渲染后台动作按钮。 */
/* 步骤：普通动作使用西贡绿，支付动作用香槟金，材料动作使用浅绿色强调。 */
/* 返回：运营人员无需理解代码即可按状态推进订单。 */
.workflow-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 14px;
}

.action-btn {
  height: 34px;
  margin: 0;
  padding: 0 14px;
  border: none;
  border-radius: 17px;
  color: #fff;
  background: #004d40;
  font-size: 12px;
  font-weight: 800;
  line-height: 34px;
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
