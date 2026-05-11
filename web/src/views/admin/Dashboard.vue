<script setup>
import { ref, onMounted } from 'vue'
import { getDashboardStats } from '@/api/admin/dashboard'

const stats = ref(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    stats.value = await getDashboardStats()
  } catch (e) {
    error.value = '加载失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dashboard">
    <h1>仪表盘</h1>
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">{{ stats?.total_users ?? 0 }}</div>
        <div class="stat-label">用户总数</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats?.total_orders ?? 0 }}</div>
        <div class="stat-label">订单总数</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats?.total_revenue?.toLocaleString() ?? 0 }}</div>
        <div class="stat-label">总收入 (VND)</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard h1 {
  margin-bottom: 1.5rem;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1.25rem;
}
.stat-card {
  background: var(--accent-bg);
  border: 1px solid var(--accent-border);
  border-radius: 10px;
  padding: 1.5rem;
}
.stat-value {
  font-size: 2rem;
  font-weight: 600;
  color: var(--accent);
  font-family: var(--mono);
}
.stat-label {
  font-size: 0.875rem;
  color: var(--text);
  margin-top: 0.375rem;
}
.loading, .error {
  color: var(--text);
  opacity: 0.6;
}
</style>
