<template>
  <view class="h5-preview-root">
    <HomePage />
    <AuthLoginSheet />
  </view>
</template>

<script setup>
import { onMounted } from 'vue'
import HomePage from './pages/home/index.vue'
import AuthLoginSheet from './components/AuthLoginSheet.vue'
import { useClientStore } from './store/client'

const client = useClientStore()

// H5PreviewRoot 仅用于 H5 浏览器预览兜底挂载。
// 实现步骤：
// 1. 在普通 Vite/H5 入口未注入 UniApp 页面容器时，直接挂载首页组件。
// 2. 保留全局登录弹窗组件，确保“去咨询”等登录拦截交互可展示。
// 3. 进入页面后主动初始化 Mock 聚合状态，保证热门服务、攻略精选和订单数据可见。
onMounted(async () => {
  await client.initMockState()
})
</script>

<style scoped>
.h5-preview-root {
  min-height: 100vh;
  background: #f5f8ff;
}
</style>
