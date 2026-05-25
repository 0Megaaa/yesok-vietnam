<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'
import { get, post } from '@/api/request'
import AuthPopup from '@/components/AuthPopup.vue'
import { useGlobalShare } from '@/composables/useGlobalShare'

const client = useClientStore()
const serviceId = ref('company')
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
  { no: '1', title: '提交需求', desc: '说明服务目标、时间计划与预算范围。' },
  { no: '2', title: '管家评估', desc: '专属管家确认可行性、材料清单与报价。' },
  { no: '3', title: '开始办理', desc: '进入动态工作流节点，按状态推进订单。' },
  { no: '4', title: '交付验收', desc: '完成服务后提交结果并同步后续注意事项。' },
])

useGlobalShare({ title: 'YesOK越南管家｜服务详情', path: '/pages/home/index' })

onLoad((options) => {
  serviceId.value = options.id || 'company'
  loadService(serviceId.value)
})

// loadService 加载服务详情。
// 实现步骤：
// 1. 读取路由中的服务 ID。
// 2. 通过统一 request 访问 Mock 服务详情路由。
// 3. 若 Mock 数据缺失则展示兜底服务，防止页面空白。
const loadService = async (id) => {
  loading.value = true
  try {
    const res = await get(`/v1/client/services/${id}`)
    serviceData.value = res.data
  } catch (error) {
    serviceData.value = {
      id: 'unknown',
      icon: '📋',
      color: '#1565C0',
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

// goBack 返回上一页。
// 实现步骤：
// 1. 优先调用 navigateBack。
// 2. 若没有历史栈则回到首页，避免 H5 直达详情页无法返回。
const goBack = () => {
  const pages = getCurrentPages()
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (pages.length > 1) {
    if (uniApi?.navigateBack) uniApi.navigateBack()
  } else {
    if (uniApi?.switchTab) uniApi.switchTab({ url: '/pages/home/index' })
  }
}

// contactManager 发起服务咨询。
// 实现步骤：
// 1. 先执行登录拦截，未登录时打开底部登录弹窗。
// 2. 已登录后进入聊天页并带上服务名称。
const contactManager = () => {
  const serviceName = serviceData.value?.name || '服务详情'
  if (!client.checkAuth(`咨询「${serviceName}」`)) return
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateTo) uniApi.navigateTo({ url: `/pages/chat/index?svc=${encodeURIComponent(serviceName)}` })
}

// createConsultOrder 创建演示咨询订单。
// 实现步骤：
// 1. 登录拦截，保证订单与用户身份绑定。
// 2. 调用 Mock 下单接口生成订单。
// 3. 写入 store 并提示用户后续可在首页查看订单动态。
const createConsultOrder = async () => {
  const serviceName = serviceData.value?.name || '服务详情'
  if (!client.checkAuth(`预约「${serviceName}」`)) return
  const res = await post('/v1/client/orders', { serviceId: serviceData.value?.id || serviceId.value })
  client.addOrder(res.data)
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.showToast) uniApi.showToast({ title: '已生成咨询订单', icon: 'success' })
}
</script>

<template>
  <view class="detail-page">
    <view class="hero" :style="{ background: serviceData?.bg || '#E3F0FF' }">
      <view class="nav-back" @click="goBack">‹</view>
      <text class="hero-icon">{{ serviceData?.icon || '📋' }}</text>
    </view>

    <view class="title-card">
      <view class="title-row">
        <view class="service-icon" :style="{ background: serviceData?.bg || '#E3F0FF' }">{{ serviceData?.icon || '📋' }}</view>
        <view class="title-main">
          <text class="service-title">{{ serviceData?.name || '服务详情' }}</text>
          <view class="tag-row">
            <text v-for="tag in serviceData?.tags || []" :key="tag" class="tag">{{ tag }}</text>
          </view>
        </view>
      </view>
      <view class="price-row">
        <text class="price">{{ serviceData?.price || '面议' }}</text>
        <text class="unit">{{ serviceData?.unit || '' }}</text>
      </view>
      <text class="desc">{{ serviceData?.description || serviceData?.desc || '该服务由 YesOK 越南管家提供中文一站式协助。' }}</text>
    </view>

    <view class="section-card">
      <view class="section-title"><text class="section-bar"></text>包含服务</view>
      <view v-for="item in includes" :key="item" class="include-item">
        <text class="check">✓</text>
        <text>{{ item }}</text>
      </view>
    </view>

    <view class="section-card">
      <view class="section-title"><text class="section-bar"></text>办理流程</view>
      <view v-for="step in steps" :key="step.no" class="step-item">
        <view class="step-no">{{ step.no }}</view>
        <view class="step-content">
          <text class="step-title">{{ step.title }}</text>
          <text class="step-desc">{{ step.desc }}</text>
        </view>
      </view>
    </view>

    <view class="bottom-bar">
      <button class="ghost-btn" @click="contactManager">去咨询</button>
      <button class="primary-btn" @click="createConsultOrder">立即预约</button>
    </view>

    <AuthPopup />
  </view>
</template>

<style scoped>
/* 意图：服务详情页同步 Yesok 2.0 薄荷灰青底色，保证从首页跳转后的品牌一致性。 */
/* 步骤：保留原有结构，只替换背景色并预留底部安全区。 */
/* 返回：视觉连续且不影响业务逻辑的详情页容器。 */
.detail-page {
  min-height: 100vh;
  padding-bottom: 92px;
  background: #f2f6f5;
}

.hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 220px;
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
  font-size: 88px;
}

.title-card,
.section-card {
  margin: 10px 12px 0;
  padding: 16px;
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.05);
}

.title-row {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.service-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 50px;
  height: 50px;
  border-radius: 14px;
  font-size: 24px;
}

.title-main {
  flex: 1;
}

.service-title,
.desc,
.step-title,
.step-desc {
  display: block;
}

.service-title {
  margin-bottom: 8px;
  color: #102a55;
  font-size: 19px;
  font-weight: 800;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag {
  padding: 3px 8px;
  border-radius: 12px;
  background: #f0f6ff;
  color: #0d47a1;
  font-size: 10px;
  font-weight: 700;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
  margin: 14px 0 10px;
}

.price {
  color: #e97832;
  font-size: 23px;
  font-weight: 900;
}

.unit {
  color: #9aa3b5;
  font-size: 13px;
}

.desc {
  color: #4a5568;
  font-size: 13px;
  line-height: 1.8;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  color: #102a55;
  font-size: 16px;
  font-weight: 800;
}

.section-bar {
  width: 4px;
  height: 16px;
  border-radius: 4px;
  background: #004d40;
}

.include-item {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f6ff;
  color: #4a5568;
  font-size: 13px;
}

.check {
  color: #2e7d32;
  font-weight: 900;
}

.step-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f6ff;
}

.step-no {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #004d40;
  color: #fff;
  font-size: 13px;
  font-weight: 800;
}

.step-content {
  flex: 1;
}

.step-title {
  color: #102a55;
  font-size: 14px;
  font-weight: 800;
}

.step-desc {
  margin-top: 4px;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.6;
}

.bottom-bar {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
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
