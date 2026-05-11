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
  <div class="adm-main">
    <!-- 数据看板 -->
    <div class="adm-pt">数据看板</div>
    <div class="adm-ps">实时业务概览</div>

    <!-- Stats Cards -->
    <div class="adm-stats">
      <div class="adm-stat s1">
        <div class="adm-sv">{{ stats?.total_users ?? '—' }}</div>
        <div class="adm-sl">用户总数</div>
      </div>
      <div class="adm-stat s2">
        <div class="adm-sv">{{ stats?.total_orders ?? '—' }}</div>
        <div class="adm-sl">订单总数</div>
      </div>
      <div class="adm-stat s3">
        <div class="adm-sv">{{ stats?.active_orders ?? '—' }}</div>
        <div class="adm-sl">进行中</div>
      </div>
      <div class="adm-stat s4">
        <div class="adm-sv">{{ stats?.total_revenue ? stats.total_revenue.toLocaleString() : '—' }}</div>
        <div class="adm-sl">总收入 (VND)</div>
      </div>
    </div>

    <!-- Two-column cards -->
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:14px;">
      <div class="adm-card">
        <div class="adm-card-t">基本信息</div>
        <div class="adm-fr">
          <label class="adm-fl">平台名称</label>
          <input class="adm-in" value="Yesok · 越南一站式管家服务">
        </div>
        <div class="adm-fr">
          <label class="adm-fl">客服微信</label>
          <input class="adm-in" value="Yesok_VN2024">
        </div>
        <div class="adm-fr">
          <label class="adm-fl">联系电话</label>
          <input class="adm-in" value="+84-xx-xxxx-xxxx">
        </div>
        <button class="adm-btn ab-gd" style="margin-top:4px;width:100%;justify-content:center;">保存设置</button>
      </div>
      <div class="adm-card">
        <div class="adm-card-t">最新订单</div>
        <div v-if="loading" style="font-size:12px;color:var(--tx3);">加载中…</div>
        <div v-else-if="error" style="font-size:12px;color:var(--rd);">{{ error }}</div>
        <div v-else style="font-size:12px;color:var(--tx2);">暂无最新订单数据</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Styles from style.css are inherited globally — adm-* classes are defined there */
</style>
