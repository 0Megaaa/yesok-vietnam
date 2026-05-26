<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(true)
const stats = ref({})
const services = ref([])
const dictData = ref([])
const appUsers = ref([])
const orders = ref([])

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const normalizeOrder = (order) => ({
  ...order,
  current_status: order.current_status || order.currentStatus,
})

// stats 接口会返回订单总数/待受理/总收入
// 但当前端未见专门 stats 接口，改为并发读取必要数据
const loadStats = async () => {
  loading.value = true
  try {
    console.log('[Dashboard] 发起 API 请求')
    const [statsRes, ordersRes, servicesRes, dictDataRes, appUsersRes] = await Promise.all([
      request.get('/v1/admin/dashboard/stats'),
      request.get('/v1/admin/orders'),
      request.get('/v1/admin/services'),
      request.get('/v1/admin/dict-data'),
      request.get('/v1/admin/app-users'),
    ])
    console.log('[Dashboard] statsRes =', statsRes.data)
    stats.value = statsRes.data || {}
    orders.value = (ordersRes.data.list || ordersRes.data.orders || []).map(normalizeOrder)
    services.value = servicesRes.data.list || []
    dictData.value = dictDataRes.data.list || []
    appUsers.value = appUsersRes.data.list || []
  } catch (error) {
    console.error('[Dashboard] ❌ 报错：', error)
    showToast(error?.message || '数据加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const filters = computed(() => [
  { key: 'all', label: '全部', count: orders.value.length },
  { key: 'pending', label: '待受理', count: orders.value.filter((o) => o.current_status === 'pending').length },
  { key: 'quoted', label: '已报价', count: orders.value.filter((o) => o.current_status === 'quoted').length },
  { key: 'paid', label: '已收款', count: orders.value.filter((o) => o.current_status === 'paid').length },
  { key: 'in_progress', label: '履约中', count: orders.value.filter((o) => o.current_status === 'in_progress').length },
])

onMounted(() => {
  console.log('[Dashboard] 组件挂载，调用 loadStats')
  loadStats()
})

onUnmounted(() => {
  loading.value = false
})
</script>

<template>
  <div class="dashboard-page">
    <!-- 顶部 Banner -->
    <div class="hero-card">
      <div class="hero-text">
        <span class="eyebrow">REAL DATA BUSINESS LOOP</span>
        <span class="hero-title">越南奢华生活服务管家后台</span>
        <span class="hero-desc">后台配置服务、字典、资讯与流程，C 端动态展示，下单后由状态节点自动沉淀时间线与财务流水。</span>
      </div>
      <el-button class="refresh-btn" type="default" :loading="loading" @click="loadStats">
        刷新数据
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card green">
        <span>{{ stats.total_orders || stats.totalOrders || 0 }}</span>
        <span>总订单</span>
      </div>
      <div class="stat-card">
        <span>{{ stats.pending_orders || stats.pendingOrders || 0 }}</span>
        <span>待受理</span>
      </div>
      <div class="stat-card">
        <span>{{ services.length }}</span>
        <span>服务配置</span>
      </div>
      <div class="stat-card gold">
        <span>{{ stats.total_revenue_text || stats.total_revenue || 0 }}</span>
        <span>确认收入</span>
      </div>
    </div>

    <!-- 数据雷达 -->
    <div class="panel-grid">
      <div class="glass-panel wide">
        <span class="panel-title">今日履约雷达</span>
        <div v-for="filter in filters" :key="filter.key" class="radar-line">
          <span>{{ filter.label }}</span>
          <span class="strong">{{ filter.count }}</span>
        </div>
      </div>
      <div class="glass-panel">
        <span class="panel-title">服务上架</span>
        <span class="big-number">{{ services.filter((s) => s.status === 1).length }}</span>
        <span>项启用服务正在驱动 C 端</span>
      </div>
      <div class="glass-panel">
        <span class="panel-title">字典枚举</span>
        <span class="big-number">{{ dictData.length }}</span>
        <span>条可配置业务字典</span>
      </div>
      <div class="glass-panel">
        <span class="panel-title">客户矩阵</span>
        <span class="big-number">{{ appUsers.length }}</span>
        <span>位 C 端客户画像</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-page {
  width: 100%;
  min-width: 800px;
}

.hero-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 30px;
  border-radius: 36px;
  color: #fff;
  background: radial-gradient(
      circle at 85% 15%,
      rgba(245, 217, 143, 0.46),
      transparent 26%
    ),
    linear-gradient(135deg, #004d40, #0f3d3e);
  box-shadow: 0 28px 80px rgba(0, 77, 64, 0.18);
  margin-bottom: 18px;
}

.hero-text {
  display: flex;
  flex-direction: column;
}

.eyebrow {
  display: block;
  color: rgba(255, 255, 255, 0.72);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 1.8px;
}

.hero-title {
  display: block;
  margin-top: 8px;
  font-size: 28px;
  font-weight: 900;
}

.hero-desc {
  display: block;
  margin-top: 8px;
  max-width: 480px;
  color: rgba(255, 255, 255, 0.82);
  font-size: 13px;
  line-height: 1.7;
}

.refresh-btn {
  flex-shrink: 0;
  height: 42px;
  padding: 0 20px;
  color: #12312c;
  background: #f5d98f;
  border: 0;
  border-radius: 999px;
  font-weight: 900;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 18px;
}

.stat-card {
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
}

.stat-card > span:first-child {
  display: block;
  font-size: 28px;
  font-weight: 900;
  color: #12312c;
}

.stat-card > span:last-child {
  display: block;
  margin-top: 4px;
  color: #6b7c78;
  font-size: 13px;
}

.stat-card.green {
  color: #fff;
  background: #004d40;
}

.stat-card.gold {
  background: linear-gradient(135deg, #f5d98f, #fff);
}

.panel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.glass-panel {
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
}

.panel-title {
  display: block;
  margin-bottom: 12px;
  color: #12312c;
  font-size: 15px;
  font-weight: 700;
}

.radar-line {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
  color: #6b7c78;
  font-size: 13px;
}

.radar-line:last-child {
  border-bottom: none;
}

.strong {
  font-weight: 900;
  color: #12312c;
  font-size: 18px;
}

.big-number {
  display: block;
  font-size: 28px;
  font-weight: 900;
  color: #12312c;
}

.glass-panel > span:last-child {
  display: block;
  margin-top: 4px;
  color: #6b7c78;
  font-size: 13px;
}

.wide {
  grid-column: span 1;
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 520px) {
  .stats-grid,
  .panel-grid {
    grid-template-columns: 1fr;
  }
}
</style>
