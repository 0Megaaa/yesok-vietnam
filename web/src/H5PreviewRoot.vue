<template>
  <view class="h5-preview-root">
    <AdminDetailPage v-if="currentRoute === 'admin-detail'" />
    <AdminIndexPage v-else-if="currentRoute === 'admin-index'" />
    <HomePage v-else />
  </view>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import HomePage from './pages/home/index.vue'
import AdminIndexPage from './pages/admin/index.vue'
import AdminDetailPage from './pages/admin/order-detail.vue'
import { useClientStore } from './store/client'

const client = useClientStore()

// currentRoute 计算 H5 预览当前页面。
// 意图：在无完整 UniApp 路由运行时，也能通过 URL path 直接展示 B/C 端验收页面。
// 实现步骤：
// 1. 读取 window.location.pathname。
// 2. 兼容 UniApp path、静态服务 query 和 hash 三类访问方式。
// 3. 匹配后台订单详情、后台看板与默认首页。
// 返回：admin-detail、admin-index 或 home。
const currentRoute = computed(() => {
  if (typeof window === 'undefined') return 'home'
  const path = window.location.pathname
  const routeText = `${path}${window.location.search}${window.location.hash}`
  if (routeText.includes('/pages/admin/order-detail') || routeText.includes('admin=detail')) return 'admin-detail'
  if (routeText.includes('/pages/admin/index') || routeText.includes('admin=index')) return 'admin-index'
  return 'home'
})

// H5PreviewRoot 仅用于 H5 浏览器预览兜底挂载。
// 实现步骤：
// 1. 在普通 Vite/H5 入口未注入 UniApp 页面容器时，根据 URL 切换预览页面。
// 2. 首页内部已经挂载 AuthPopup，因此根组件不再重复渲染登录弹窗。
// 3. 进入页面后主动初始化 Mock 聚合状态，保证热门服务、攻略精选和订单数据可见。
onMounted(async () => {
  await client.initMockState()
})
</script>

<style scoped>
.h5-preview-root {
  min-height: 100vh;
  background: #f2f6f5;
}
</style>
