<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'
import { get } from '@/api/request'
import AuthPopup from '@/components/AuthPopup.vue'
import { useGlobalShare } from '@/composables/useGlobalShare'

const client = useClientStore()
const serviceId = ref('')
const serviceData = ref(null)
const loading = ref(true)

const includes = computed(() => [
  '需求沟通与方案确认',
  '材料清单整理与中文说明',
  '本地资源匹配与预约协调',
  '办理进度实时反馈',
  '交付后注意事项提醒',
])

const steps = computed(() => [
  { no: '1', title: '填写需求', desc: '按服务类型填写专属表单，管家实时接收。' },
  { no: '2', title: '管家评估', desc: '专属管家确认可行性、材料清单与报价。' },
  { no: '3', title: '开始办理', desc: '进入动态工作流节点，按状态推进订单。' },
  { no: '4', title: '交付验收', desc: '完成服务后提交结果并同步后续注意事项。' },
])

useGlobalShare({ title: 'YesOK越南管家｜服务详情', path: '/pages/home/index' })

const unwrapResponse = (res) => {
  if (!res) return {}
  return res.data ?? res
}

const formatPrice = (amount) => {
  const n = Number(amount || 0)
  if (!n) return '面议'
  return `¥${n.toLocaleString('zh-CN')}`
}

const normalizeServiceDetail = (raw = {}) => {
  const serviceName = raw.service_name || raw.serviceName || raw.display_name || raw.name || '服务详情'
  const displayName = raw.display_name || raw.displayName || serviceName

  return {
    ...raw,
    id: raw.id || raw.service_id,
    service_id: raw.service_id || raw.id,
    service_code: raw.service_code || raw.serviceCode || raw.code || '',
    service_name: serviceName,
    display_name: displayName,
    title: serviceName,
    icon: raw.icon || '📋',
    cover_image: raw.cover_image || raw.coverImage || '',
    description: raw.description || raw.desc || '该服务由 YesOK 越南管家提供中文一站式协助。',
    base_price: Number(raw.base_price ?? raw.basePrice ?? 0),
    price_text: formatPrice(raw.base_price ?? raw.basePrice),
    unit: raw.unit || '次',
  }
}

onLoad((options) => {
  serviceId.value = options.id || ''
  loadService(serviceId.value)
})

const loadService = async (id) => {
  loading.value = true
  try {
    const svcRes = await get(`/v1/client/services/${id}`)
    serviceData.value = normalizeServiceDetail(unwrapResponse(svcRes))
  } catch {
    serviceData.value = {
      id: 'unknown',
      icon: '📋',
      bg: '#E3F0FF',
      name: '服务详情',
      price: '面议',
      unit: '',
      tags: ['专属管家'],
      description: '该服务详情正在准备中，请联系管家获取更多信息。',
    }
  } finally {
    loading.value = false
  }
}

const openOrderForm = () => {
  const serviceName = serviceData.value?.service_name || serviceData.value?.display_name || '该服务'
  if (!client.checkAuth(`预约「${serviceName}」`)) return

  const id = serviceData.value?.id || serviceData.value?.service_id || serviceId.value
  if (!id) {
    safeToast('服务信息异常，请稍后重试', 'none')
    return
  }

  uni.navigateTo({
    url: `/subpkg/dynamic-form/index?mode=service&service_id=${id}`,
  })
}

const goBack = () => {
  const pages = getCurrentPages()
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (pages.length > 1) {
    if (uniApi?.navigateBack) uniApi.navigateBack()
  } else {
    if (uniApi?.switchTab) uniApi.switchTab({ url: '/pages/home/index' })
  }
}

const contactManager = () => {
  const serviceName = serviceData.value?.service_name || serviceData.value?.display_name || '服务详情'
  if (!client.checkAuth(`咨询「${serviceName}」`)) return
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateTo) uniApi.navigateTo({ url: `/pages/chat/index?svc=${encodeURIComponent(serviceName)}` })
}

const safeToast = (title, icon = 'info') => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.showToast) {
    uniApi.showToast({ title, icon })
  } else {
    console.info(`[Toast] ${title}`)
  }
}
</script>

<template>
  <view class="detail-page">
    <!-- Hero -->
    <view class="hero" :style="{ background: serviceData?.bg || serviceData?.cover_image || '#E3F0FF' }">
      <view class="nav-back" @click="goBack">‹</view>
      <text class="hero-icon">{{ serviceData?.icon || '📋' }}</text>
    </view>

    <!-- 服务信息 -->
    <view class="title-card">
      <view class="title-row">
        <view class="service-icon-wrap" :style="{ background: serviceData?.bg || '#E3F0FF' }">
          <text class="service-icon">{{ serviceData?.icon || '📋' }}</text>
        </view>
        <view class="title-main">
          <text class="service-title">{{ serviceData?.service_name || serviceData?.display_name || '服务详情' }}</text>
          <view class="tag-row">
            <text v-for="tag in (serviceData?.tags || [])" :key="tag" class="tag">{{ tag }}</text>
          </view>
        </view>
      </view>
      <view class="price-row">
        <text class="price">{{ serviceData?.price_text || '面议' }}</text>
        <text class="unit">/{{ serviceData?.unit || '次' }}</text>
      </view>
      <text class="desc">{{ serviceData?.description || serviceData?.desc || '该服务由 YesOK 越南管家提供中文一站式协助。' }}</text>
    </view>

    <view class="section-card">
      <view class="section-head"><text class="section-title">服务包含</text></view>
      <view class="chips">
        <text v-for="item in includes" :key="item" class="chip">{{ item }}</text>
      </view>
    </view>

    <view class="section-card">
      <view class="section-head"><text class="section-title">服务流程</text></view>
      <view class="steps">
        <view v-for="step in steps" :key="step.no" class="step-item">
          <view class="step-no">{{ step.no }}</view>
          <view class="step-main">
            <text class="step-title">{{ step.title }}</text>
            <text class="step-desc">{{ step.desc }}</text>
          </view>
        </view>
      </view>
    </view>

    <AuthPopup />

    <view class="bottom-bar">
      <button class="ghost-btn" @click="contactManager">咨询管家</button>
      <button class="primary-btn" @click="openOrderForm">立即预约</button>
    </view>
  </view>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding-bottom: 32px;
  background: #f2f6f5;
}

.hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 180px;
  background: linear-gradient(135deg, #004d40, #07362f);
  border-radius: 0 0 32px 32px;
}

.nav-back {
  position: absolute;
  top: 16px;
  left: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.92);
  color: #102a55;
  font-size: 28px;
  line-height: 34px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.hero-icon {
  font-size: 72px;
}

.loading-card,
.error-card {
  margin: 14px 14px 0;
  padding: 40px;
  border-radius: 32px;
  background: #fff;
  text-align: center;
  color: #6b7c78;
}

.title-card,
.section-card {
  margin: 14px 14px 0;
  padding: 16px;
  border-radius: 32px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.05);
}

.title-card { margin-top: -16px; }

.title-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.service-icon-wrap {
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

.title-main { flex: 1; }

.service-title {
  display: block;
  color: #12312c;
  font-size: 18px;
  font-weight: 900;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.tag,
.chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(0, 77, 64, 0.08);
  color: #004d40;
  font-size: 12px;
}

.price-row {
  margin-top: 14px;
}

.price {
  color: #004d40;
  font-size: 24px;
  font-weight: 900;
}

.unit {
  color: #6b7c78;
  font-size: 12px;
}

.desc {
  display: block;
  margin-top: 12px;
  color: #4b5f5a;
  font-size: 14px;
  line-height: 1.8;
}

.section-head {
  margin-bottom: 12px;
}

.section-title {
  color: #12312c;
  font-size: 16px;
  font-weight: 900;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.steps {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: 20px;
  background: #f8fbfa;
}

.step-no {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(0, 77, 64, 0.12);
  color: #004d40;
  font-size: 14px;
  font-weight: 900;
}

.step-main {
  flex: 1;
}

.step-title {
  display: block;
  color: #12312c;
  font-size: 14px;
  font-weight: 900;
}

.step-desc {
  display: block;
  margin-top: 4px;
  color: #6b7c78;
  font-size: 12px;
  line-height: 1.6;
}

.bottom-bar {
  position: sticky;
  bottom: 0;
  z-index: 20;
  display: flex;
  gap: 12px;
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 -2px 16px rgba(0, 0, 0, 0.08);
}

.ghost-btn,
.primary-btn {
  flex: 1;
  height: 44px;
  margin: 0;
  border-radius: 22px;
  font-size: 15px;
  font-weight: 800;
  line-height: 44px;
}

.ghost-btn {
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
}

.primary-btn {
  color: #fff;
  background: linear-gradient(135deg, #004d40, #00695c);
}
</style>