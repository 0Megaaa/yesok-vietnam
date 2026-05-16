<script setup>
import { computed, onMounted, ref } from 'vue'
import request from '@/api/request'
import AuthPopup from '@/components/AuthPopup.vue'
import { useClientStore } from '@/store/client'

const client = useClientStore()
const services = ref([])
const loading = ref(true)
const keyword = ref('')

const filteredServices = computed(() => {
  const key = keyword.value.trim().toLowerCase()
  if (!key) return services.value
  return services.value.filter((item) => `${item.display_name}${item.service_name}${item.description}`.toLowerCase().includes(key))
})

// showSafeToast 展示跨端提示。
// 1.意图 -> 服务页操作结果兼容 H5 与小程序。
// 2.步骤 -> 优先 uni.showToast，缺失时使用 console.info。
// 3.返回 -> 无返回值。
const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) uni.showToast({ title, icon: 'none' })
  else console.info('[Yesok Services]', title)
}

// loadServices 加载真实服务配置。
// 1.意图 -> 消除服务页硬编码，让 B 端 sys_services 决定 C 端内容。
// 2.步骤 -> 请求 /v1/services 并规范化名称、价格和封面字段。
// 3.返回 -> Promise<void>。
const loadServices = async () => {
  loading.value = true
  try {
    const res = await request.get('/v1/services')
    services.value = (res.data.list || []).map((item) => ({ ...item, display_name: item.display_name || item.service_name, price_text: item.price || `${Math.round(item.base_price / 100)} ${item.currency}` }))
  } catch (error) {
    showSafeToast(error?.message || '服务加载失败')
  } finally {
    loading.value = false
  }
}

// consultService 发起服务咨询。
// 1.意图 -> 继续复用 AuthPopup 鉴权挡板，不破坏底层登录逻辑。
// 2.步骤 -> 校验登录，成功后跳转详情页。
// 3.返回 -> 无返回值。
const consultService = (service) => {
  if (!client.checkAuth(`咨询「${service.display_name}」`)) return
  uni.navigateTo({ url: `/pages/service-detail/index?id=${service.id}&code=${service.service_code}` })
}

onMounted(loadServices)
</script>

<template>
  <view class="services-page"><view class="top"><text class="title">服务</text><text class="sub">所有服务、价格与配置均来自后台 sys_services</text><view class="search"><input v-model="keyword" placeholder="搜索服务" /></view></view><view v-if="loading" class="empty">正在读取后台服务配置...</view><view v-else class="list"><view v-for="service in filteredServices" :key="service.id" class="card"><view class="icon">{{ service.icon || '🌴' }}</view><view class="main"><text class="name">{{ service.display_name }}</text><text class="desc">{{ service.description }}</text><view class="bottom"><text class="price">{{ service.price_text }}/{{ service.unit || '次' }}</text><view class="consult-button" @click="consultService(service)">去咨询</view></view></view></view></view><AuthPopup /></view>
</template>

<style scoped>
.services-page { min-height: 100vh; padding: 58px 14px 90px; background: #f2f6f5; color: #12312c; }
.title, .sub, .name, .desc { display: block; }
.title { font-size: 30px; font-weight: 900; }
.sub { margin-top: 8px; color: #6b7c78; font-size: 13px; }
.search { display: flex; align-items: center; height: 48px; margin-top: 18px; padding: 0 16px; border-radius: 24px; background: rgba(255,255,255,.72); box-shadow: 0 16px 42px rgba(0,77,64,.08); backdrop-filter: blur(15px); }
.search input { flex: 1; }
.list { display: grid; gap: 14px; margin-top: 18px; }
.card { display: flex; gap: 14px; padding: 16px; border-radius: 28px; background: #fff; box-shadow: 0 16px 46px rgba(0,77,64,.07); }
.icon { display: flex; align-items: center; justify-content: center; flex-shrink: 0; width: 54px; height: 54px; border-radius: 20px; background: linear-gradient(135deg,rgba(0,77,64,.1),rgba(245,217,143,.3)); font-size: 26px; }
.main { flex: 1; min-width: 0; }
.name { font-size: 17px; font-weight: 900; }
.desc { margin-top: 6px; color: #6b7c78; font-size: 12px; line-height: 1.6; }
.bottom { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; }
.price { color: #e97832; font-size: 16px; font-weight: 900; }
.consult-button { height: 34px; margin: 0; padding: 0 14px; border: 0; border-radius: 17px; color: #fff; background: #004d40; font-size: 12px; font-weight: 900; line-height: 34px; text-align: center; }
.empty { padding: 30px; color: #6b7c78; text-align: center; }
</style>
