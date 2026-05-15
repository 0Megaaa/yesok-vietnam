<script setup>
import { computed, ref } from 'vue'
import { useClientStore } from '@/store/client'
import { PLATFORM_GUARANTEES } from '@/api/mockData'
import AuthPopup from '@/components/AuthPopup.vue'
import { useGlobalShare } from '@/composables/useGlobalShare'

const client = useClientStore()
const activeNewsTab = ref('全部')

useGlobalShare({
  title: 'YesOK越南管家｜越南一站式中文服务平台',
  path: '/pages/home/index',
})

const services = computed(() => client.services || [])
const orders = computed(() => client.orders || [])
const newsCategories = computed(() => client.newsCategories?.length ? client.newsCategories : ['全部'])
const filteredNews = computed(() => {
  if (activeNewsTab.value === '全部') return client.news || []
  return (client.news || []).filter((item) => item.tag === activeNewsTab.value)
})

const STATUS_CONFIG = {
  pending: { icon: '📋', label: '等待审核', badge: 'stb-bl', progress: 10 },
  requirement_submitted: { icon: '📝', label: '需求已提交', badge: 'stb-bl', progress: 20 },
  processing: { icon: '⚙️', label: '服务进行中', badge: 'stb-gd', progress: 45 },
  supplementing: { icon: '⚠️', label: '需要补充材料', badge: 'stb-rd', progress: 65 },
  completed: { icon: '✅', label: '服务已完成', badge: 'stb-gr', progress: 100 },
}

// getStatus 获取订单状态展示配置。
// 实现步骤：
// 1. 根据订单状态码查找状态字典。
// 2. 未命中时使用 pending 配置兜底。
// 3. 页面统一消费 label、badge 和 progress，避免魔术数字散落在模板中。
const getStatus = (statusKey) => STATUS_CONFIG[statusKey] || STATUS_CONFIG.pending

// getOrderClass 获取订单卡片样式类。
// 实现步骤：
// 1. 需要补充材料时返回警示样式。
// 2. 已完成时返回完成态样式。
// 3. 其他状态统一使用进行中样式。
const getOrderClass = (statusKey) => {
  if (statusKey === 'supplementing') return 'oc-w'
  if (statusKey === 'completed') return 'oc-d'
  return 'oc-a'
}

// goPage 切换底部 Tab 页面。
// 实现步骤：
// 1. 接收页面短名称。
// 2. 拼接 UniApp tabBar 页面路径。
// 3. 调用 switchTab 保持小程序与 H5 行为一致。
const goPage = (page) => {
  uni.switchTab({ url: `/pages/${page}/index` })
}

// openChat 打开咨询入口。
// 实现步骤：
// 1. 先通过 store 执行统一登录校验。
// 2. 未登录时弹出底部登录弹窗并停止跳转。
// 3. 已登录时进入聊天页或后续真实客服页。
const openChat = (serviceName) => {
  if (!client.checkAuth(`咨询「${serviceName}」`)) return
  uni.navigateTo({ url: `/pages/chat/index?svc=${encodeURIComponent(serviceName)}` })
}

// openServiceDetail 打开服务详情页。
// 实现步骤：
// 1. 使用服务 ID 作为详情页参数。
// 2. 详情页继续使用 Mock API 读取服务信息。
const openServiceDetail = (serviceId) => {
  uni.navigateTo({ url: `/pages/service-detail/index?id=${encodeURIComponent(serviceId)}` })
}

// selectNewsTab 切换攻略精选分类。
// 实现步骤：
// 1. 更新当前分类。
// 2. filteredNews 自动重新计算。
const selectNewsTab = (category) => {
  activeNewsTab.value = category
}
</script>

<template>
  <view class="home-page">
    <view class="hero">
      <view class="hero-gradient"></view>
      <view class="hero-topbar">
        <view class="logo">Yesok <text class="logo-accent">Vietnam</text></view>
        <view class="country-badge">🇻🇳</view>
      </view>
      <view class="hero-copy">
        <text class="hero-title">越南一站式管家服务</text>
        <text class="hero-subtitle">签证、公司、租房、接机、翻译，全程中文响应。</text>
      </view>
    </view>

    <view class="search-wrap">
      <view class="search-bar">
        <text class="search-icon">🔍</text>
        <input class="search-input" placeholder="搜索越南管家服务..." />
      </view>
    </view>

    <view class="category-card">
      <view class="category-item" @click="goPage('services')">
        <view class="category-icon">✈️</view>
        <text class="category-name">签证旅游</text>
      </view>
      <view class="category-item" @click="goPage('services')">
        <view class="category-icon">🏢</view>
        <text class="category-name">房产投资</text>
      </view>
      <view class="category-item" @click="goPage('services')">
        <view class="category-icon">📋</view>
        <text class="category-name">公司注册</text>
      </view>
      <view class="category-item" @click="goPage('services')">
        <view class="category-icon">🚗</view>
        <text class="category-name">接送出行</text>
      </view>
      <view class="category-item" @click="goPage('services')">
        <view class="category-icon">🗣️</view>
        <text class="category-name">翻译服务</text>
      </view>
    </view>

    <view class="section-card">
      <view class="section-head">
        <text class="section-title">热门服务</text>
        <text class="section-more" @click="goPage('services')">全部服务 ></text>
      </view>
      <scroll-view class="hot-scroll" scroll-x>
        <view v-for="service in services" :key="service.id" class="service-card" @click="openServiceDetail(service.id)">
          <view class="service-main">
            <view class="service-icon" :style="{ background: service.bg }">{{ service.icon }}</view>
            <view class="service-info">
              <text class="service-name" :style="{ color: service.color }">{{ service.name }}</text>
              <text class="service-sub">{{ service.sub }}</text>
              <view class="tag-row">
                <text v-for="tag in service.tags" :key="tag" class="service-tag">{{ tag }}</text>
              </view>
              <view class="service-bottom">
                <text class="service-price">{{ service.price }}<text class="service-unit">/{{ service.unit }}</text></text>
                <button class="consult-btn" :style="{ background: service.color }" @click.stop="openChat(service.name)">去咨询</button>
              </view>
            </view>
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="section-head guide-head">
      <text class="section-title">攻略精选</text>
      <text class="section-more">了解真实越南 ></text>
    </view>
    <scroll-view scroll-x class="news-tab-scroll">
      <view class="news-tab-inner">
        <text
          v-for="category in newsCategories"
          :key="category"
          :class="['news-tab', activeNewsTab === category ? 'active' : '']"
          @click="selectNewsTab(category)"
        >
          {{ category }}
        </text>
      </view>
    </scroll-view>
    <view class="news-list">
      <view v-for="item in filteredNews" :key="item.id" class="news-item">
        <view class="news-icon">📰</view>
        <view class="news-content">
          <text class="news-title"><text v-if="item.top" class="pin-text">📌 </text>{{ item.title }}</text>
          <text class="news-meta">{{ item.tag }} · {{ item.source }}</text>
        </view>
        <text class="news-date">{{ item.date }}</text>
      </view>
    </view>

    <view class="guarantee-card">
      <text class="guarantee-title">平台保障</text>
      <view class="guarantee-grid">
        <view v-for="item in PLATFORM_GUARANTEES" :key="item.title" class="guarantee-item">
          <view class="guarantee-icon">{{ item.icon }}</view>
          <text class="guarantee-name">{{ item.title }}</text>
          <text class="guarantee-desc">{{ item.desc }}</text>
        </view>
      </view>
    </view>

    <view class="order-section">
      <view class="section-head">
        <text class="section-title">📋 订单动态</text>
      </view>
      <view class="ord-list">
        <view v-if="orders.length === 0" class="empty">
          <view class="empty-ic">📭</view>
          <text class="empty-ti">暂无订单</text>
        </view>
        <view v-for="order in orders" v-else :key="order.id" :class="['ord-card', getOrderClass(order.status || order.sk)]">
          <view class="oc-top">
            <view>
              <text class="oc-svc">{{ order.icon || '📋' }} {{ order.serviceName || order.svc }}</text>
              <text class="oc-no">订单号 {{ order.orderNo || order.id }}</text>
            </view>
            <text :class="['stb', getStatus(order.status || order.sk).badge]">
              {{ getStatus(order.status || order.sk).label }}
            </text>
          </view>
          <view class="order-progress-wrap">
            <view class="oc-pbar">
              <view class="oc-pfill" :style="{ width: getStatus(order.status || order.sk).progress + '%' }"></view>
            </view>
            <view class="oc-plbl">
              <text>{{ getStatus(order.status || order.sk).label }}</text>
              <text>{{ getStatus(order.status || order.sk).progress }}% 完成</text>
            </view>
          </view>
          <view class="oc-bot">
            <view class="oc-mgr">
              <view class="oc-mav">{{ (order.managerName || order.mg || '管')[0] }}</view>
              <text>{{ order.managerName || order.mg || '专属管家' }}</text>
            </view>
            <text class="order-price">{{ order.price || order.pr || '—' }}</text>
          </view>
        </view>
      </view>
    </view>

    <AuthPopup />
    <view class="safe-bottom"></view>
  </view>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  overflow-x: hidden;
  background: #f2f6f5;
}

/* 意图：打造 340px 沉浸式融边 Banner，彻底消除微信顶部白条并呈现热带奢华第一视觉。 */
/* 步骤：使用 100vw 负边距铺满视口，以 /static/img.png 为背景，并在底部叠加灰青色渐变遮罩。 */
/* 返回：一个顶部无白边、底部自然融入薄荷灰青背景的品牌英雄区。 */
.hero {
  position: relative;
  width: 100vw;
  height: 340px;
  margin-top: -24px;
  margin-left: calc((100% - 100vw) / 2);
  padding-top: 64px;
  overflow: hidden;
  background-image: url('/static/img.png');
  background-size: cover;
  background-position: center top;
}

.hero::before {
  position: absolute;
  inset: 0;
  content: '';
  background: linear-gradient(180deg, rgba(0, 77, 64, 0.08) 0%, rgba(0, 77, 64, 0.28) 58%, rgba(242, 246, 245, 0.96) 100%);
}

.hero-gradient {
  position: absolute;
  right: 0;
  bottom: -1px;
  left: 0;
  height: 128px;
  background: linear-gradient(180deg, rgba(242, 246, 245, 0) 0%, #f2f6f5 88%);
}

.hero-topbar {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 22px;
}

.logo {
  color: #fff;
  font-size: 22px;
  font-weight: 800;
}

.logo-accent {
  color: #c5a059;
}

.country-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid rgba(197, 160, 89, 0.45);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.86);
  box-shadow: 0 12px 30px rgba(0, 77, 64, 0.16);
  backdrop-filter: blur(12px);
}

.hero-copy {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  padding: 58px 24px 0;
}

.hero-title {
  color: #fff;
  font-size: 26px;
  font-weight: 900;
  line-height: 1.25;
}

.hero-subtitle {
  max-width: 286px;
  margin-top: 10px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 13px;
  line-height: 1.65;
  text-shadow: 0 2px 12px rgba(0, 77, 64, 0.28);
}

/* 意图：实现悬浮于 Banner 底部的毛玻璃搜索胶囊。 */
/* 步骤：通过负 margin 上浮，使用高透白底、15px 背景模糊和 30px 圆角形成玻璃质感。 */
/* 返回：一个 H5 与小程序都能稳定展示的高端搜索入口。 */
.search-wrap {
  position: relative;
  z-index: 5;
  padding: 0 22px;
  margin-top: -42px;
}

.search-bar {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 14px 18px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 18px 45px rgba(0, 77, 64, 0.14);
  backdrop-filter: blur(15px);
}

.search-input {
  flex: 1;
  color: #1a2340;
  font-size: 14px;
}

.category-card,
.section-card,
.news-list,
.guarantee-card,
.order-section {
  margin: 14px 12px 0;
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.04);
}

.category-card {
  display: flex;
  justify-content: space-around;
  padding: 14px 4px;
}

.category-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  width: 62px;
}

.category-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: rgba(0, 77, 64, 0.08);
  color: #004d40;
  font-size: 22px;
}

.category-name {
  color: #1a2340;
  font-size: 11px;
  font-weight: 600;
}

/* 意图：为业务模块注入卡片呼吸感，同时不触碰攻略精选与平台保障保护模块。 */
/* 步骤：仅覆盖分类卡和热门服务容器，使用 32px 圆角与极浅深绿色彩色投影。 */
/* 返回：植物掩映光影感的核心业务卡片容器。 */
.category-card,
.section-card {
  border-radius: 32px;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.05);
}

.section-card {
  padding: 16px 0 6px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px 12px;
}

.section-title {
  color: #1a2340;
  font-size: 17px;
  font-weight: 800;
}

.section-more {
  color: #9aa3b5;
  font-size: 12px;
}

.hot-scroll {
  width: 100%;
  white-space: nowrap;
}

/* 意图：让热门服务卡片具备热带奢华管家的呼吸感与高价值感。 */
/* 步骤：扩大圆角到 32px，叠加深绿色极浅投影，并设置固定最小高度稳定横滑排版。 */
/* 返回：纯白、柔和、可横向滑动的高端业务服务卡片。 */
.service-card {
  display: inline-block;
  width: 286px;
  min-height: 168px;
  margin: 8px 0 16px 14px;
  padding: 18px;
  border: 1px solid rgba(0, 77, 64, 0.04);
  border-radius: 32px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(0, 77, 64, 0.05);
  vertical-align: top;
}

.service-main {
  display: flex;
  gap: 12px;
}

.service-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 52px;
  height: 52px;
  border-radius: 18px;
  font-size: 24px;
}

.service-info {
  flex: 1;
  min-width: 0;
}

.service-name,
.service-sub,
.service-price,
.news-title,
.news-meta,
.guarantee-title,
.guarantee-name,
.guarantee-desc,
.oc-svc,
.oc-no {
  display: block;
}

.service-name {
  margin-bottom: 4px;
  font-size: 15px;
  font-weight: 800;
}

.service-sub {
  margin-bottom: 8px;
  color: var(--tx2);
  font-size: 12px;
}

.tag-row {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;
}

.service-tag {
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--gy);
  color: var(--tx3);
  font-size: 10px;
}

.service-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 10px;
  border-top: 1px dashed #eee;
}

.service-price {
  color: #e97832;
  font-size: 17px;
  font-weight: 900;
}

.service-unit {
  color: var(--tx3);
  font-size: 10px;
  font-weight: 400;
}

.consult-btn {
  height: 30px;
  margin: 0;
  padding: 0 18px;
  border: none;
  border-radius: 15px;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  line-height: 30px;
  background: #004d40 !important;
}

.guide-head {
  padding-top: 24px;
}

.news-tab-scroll {
  width: 100%;
  white-space: nowrap;
}

.news-tab-inner {
  display: inline-flex;
  gap: 8px;
  padding: 0 16px 12px;
}

.news-tab {
  padding: 6px 14px;
  border-radius: 16px;
  background: var(--gy);
  color: var(--tx2);
  font-size: 12px;
  font-weight: 700;
}

.news-tab.active {
  color: #fff;
  background: var(--bl);
}

.news-list {
  overflow: hidden;
}

.news-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--gy);
}

.news-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: var(--gdl);
  font-size: 18px;
}

.news-content {
  flex: 1;
  min-width: 0;
}

.news-title {
  color: var(--tx);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
}

.pin-text {
  color: var(--rd);
}

.news-meta {
  margin-top: 4px;
  color: var(--tx3);
  font-size: 11px;
}

.news-date {
  align-self: flex-start;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--gy2);
  color: var(--tx3);
  font-size: 10px;
  font-weight: 700;
}

.guarantee-card {
  padding: 20px 12px;
}

.guarantee-title {
  margin-bottom: 16px;
  color: var(--tx);
  font-size: 16px;
  font-weight: 800;
  text-align: center;
}

.guarantee-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.guarantee-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.guarantee-icon {
  margin-bottom: 8px;
  font-size: 24px;
}

.guarantee-name {
  color: var(--tx);
  font-size: 12px;
  font-weight: 800;
}

.guarantee-desc {
  margin-top: 4px;
  color: var(--tx3);
  font-size: 10px;
  line-height: 1.35;
}

.order-section {
  padding-top: 14px;
}

.order-progress-wrap {
  margin: 9px 0;
}

.order-price {
  color: var(--tx3);
  font-size: 11px;
}

.safe-bottom {
  height: 20px;
}
</style>
