<script setup>
import { computed, onMounted, ref } from 'vue'
import request from '@/api/request'
import AuthPopup from '@/components/AuthPopup.vue'
import { useClientStore } from '@/store/client'
import { useGlobalShare } from '@/composables/useGlobalShare'

const client = useClientStore()
const loading = ref(true)
const services = ref([])
const articles = ref([])
const configs = ref({})
const selectedService = ref(null)
const orderForm = ref({ contact_name: '', contact_phone: '', hotel_address: '', remark: '' })

useGlobalShare({ title: 'Yesok Vietnam｜越南本地生活管家', path: '/pages/home/index' })

// 1.意图 -> 从服务接口中提取首页热门服务、分类入口和 Banner 图片。
// 2.步骤 -> 使用 computed 对响应式服务列表与配置对象进行派生，避免页面硬编码服务名称和价格。
// 3.返回 -> 可被模板直接消费的热门服务、分类和首页文案。
const hotServices = computed(() => services.value.filter((item) => item.is_hot).slice(0, 6))
const categories = computed(() => services.value.slice(0, 5).map((item) => ({ id: item.id, icon: item.icon || '🌴', name: item.display_name || item.service_name })))
const bannerImage = computed(() => '/static/img.png')
const featuredArticles = computed(() => articles.value.slice(0, 3))

// showSafeToast 展示跨端提示。
// 1.意图 -> 下单、咨询、加载失败时在 H5 与小程序都能反馈。
// 2.步骤 -> 优先使用 uni.showToast，缺失时降级 console.info。
// 3.返回 -> 无返回值。
const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Yesok Client]', title)
}

// normalizeService 规范化 C 端服务字段。
// 1.意图 -> 将 Go 接口返回的 sys_services 字段转换为首页卡片可直接消费的数据。
// 2.步骤 -> 补齐展示名、价格、单位、封面图和服务编码。
// 3.返回 -> 标准服务对象。
const normalizeService = (item) => ({
  ...item,
  display_name: item.display_name || item.service_name || item.name,
  price_text: item.price || `${Math.round((item.base_price || 0) / 100)} ${item.currency || 'VND'}`,
  cover_image: item.cover_image || bannerImage.value,
  service_code: item.service_code || item.code,
})

// normalizeArticle 规范化 C 端资讯字段。
// 1.意图 -> 将 sys_articles 接口字段转换为首页资讯卡片所需结构。
// 2.步骤 -> 补齐封面、分类、作者和摘要，保证空字段不影响渲染。
// 3.返回 -> 标准资讯对象。
const normalizeArticle = (item) => ({
  ...item,
  cover_img: item.cover_img || bannerImage.value,
  category: item.category || 'guide',
  author: item.author || 'Yesok Vietnam',
  summary: item.summary || item.content || 'Yesok Vietnam 管家精选资讯',
})

// loadHomeData 加载首页真实数据。
// 1.意图 -> 让首页 Banner 文案、服务分类、价格、热门卡片和资讯全部由后端动态驱动。
// 2.步骤 -> 并发请求 /v1/configs、/v1/services 与 /v1/articles，再写入页面响应式状态。
// 3.返回 -> Promise<void>。
const loadHomeData = async () => {
  loading.value = true
  try {
    const [configRes, serviceRes, articleRes] = await Promise.all([
      request.get('/v1/configs'),
      request.get('/v1/services'),
      request.get('/v1/articles', { params: { limit: 3 } }),
    ])
    configs.value = configRes.data.configs || configRes.data || {}
    services.value = (serviceRes.data.list || []).map(normalizeService)
    articles.value = (articleRes.data.list || []).map(normalizeArticle)
  } catch (error) {
    showSafeToast(error?.message || '首页数据加载失败')
  } finally {
    loading.value = false
  }
}

// goPage 切换原生 Tab 页面。
// 1.意图 -> 使用 pages.json 标准 Tabbar 唤醒首页、资讯、服务、我的。
// 2.步骤 -> 调用 uni.switchTab 并传入目标页面。
// 3.返回 -> 无返回值。
const goPage = (page) => {
  uni.switchTab({ url: `/pages/${page}/index` })
}

// openServiceDetail 打开服务详情。
// 1.意图 -> 将服务 ID 与服务编码带给详情页，保持真实服务链路。
// 2.步骤 -> 优先使用 uni.navigateTo，缺失时降级 H5 地址跳转。
// 3.返回 -> 无返回值。
const openServiceDetail = (service) => {
  const url = `/pages/service-detail/index?id=${encodeURIComponent(service.id)}&code=${encodeURIComponent(service.service_code)}`
  if (typeof uni !== 'undefined' && uni?.navigateTo) uni.navigateTo({ url })
}

// openOrderSheet 打开轻量下单面板。
// 1.意图 -> 在首页直接演示 C 端提交订单到 orders 表。
// 2.步骤 -> 通过 AuthPopup 兼容方法做登录挡板，通过后设置 selectedService。
// 3.返回 -> 无返回值。
const openOrderSheet = (service) => {
  if (!client.checkAuth(`咨询「${service.display_name}」`)) return
  selectedService.value = service
}

// openNewsList 跳转资讯 Tab。
// 1.意图 -> 从首页资讯模块进入四栏 Tabbar 的资讯频道。
// 2.步骤 -> 使用 uni.switchTab 跳转到 /pages/news/index。
// 3.返回 -> 无返回值。
const openNewsList = () => {
  uni.switchTab({ url: '/pages/news/index' })
}

// submitOrder 提交真实订单。
// 1.意图 -> 将 C 端需求写入 orders.form_data，并触发后台订单中心可见。
// 2.步骤 -> 组装服务编码、联系人和动态表单 JSON 后 POST /v1/orders。
// 3.返回 -> Promise<void>。
const submitOrder = async () => {
  if (!selectedService.value) return
  try {
    const payload = {
      service_code: selectedService.value.service_code,
      contact_name: orderForm.value.contact_name || '微信客户',
      contact_phone: orderForm.value.contact_phone || '+84000000000',
      form_data: {
        hotel_address: orderForm.value.hotel_address,
        remark: orderForm.value.remark,
        source: 'C端首页服务咨询入口',
      },
    }
    const res = await request.post('/v1/orders', payload)
    client.addOrder?.(res.data.order)
    selectedService.value = null
    orderForm.value = { contact_name: '', contact_phone: '', hotel_address: '', remark: '' }
    showSafeToast('订单已提交，管家即将联系您')
  } catch (error) {
    showSafeToast(error?.message || '下单失败')
  }
}

onMounted(loadHomeData)
</script>

<template>
  <view class="home-page">
    <view class="hero-bleed">
      <image class="hero-image" src="/static/img.png" mode="aspectFill" />
      <view class="hero-mask"></view>
    </view>

    <view class="search-capsule">
      <text class="search-icon">⌕</text>
      <input class="search-input" placeholder="搜索接机、签证、商务陪同..." />
    </view>

    <view class="category-card">
      <view v-for="item in categories" :key="item.id" class="category-item" @click="goPage('services')">
        <view class="category-icon">{{ item.icon }}</view>
        <text>{{ item.name }}</text>
      </view>
    </view>

    <view class="section-card">
      <view class="section-head">
        <text class="section-title">热门服务</text>
        <text class="section-more" @click="goPage('services')">全部服务 &gt;</text>
      </view>
      <view v-if="loading" class="empty">正在从后台读取服务价格...</view>
      <scroll-view v-else scroll-x class="hot-scroll" enable-flex="true">
        <view v-for="service in hotServices" :key="service.id" class="service-card" @click="openServiceDetail(service)">
          <image class="service-cover" :src="service.cover_image" mode="aspectFill" />
          <view class="service-body">
            <view class="service-title-row">
              <text class="service-icon">{{ service.icon || '🌴' }}</text>
              <text class="service-name">{{ service.display_name }}</text>
            </view>
            <text class="service-desc">{{ service.description }}</text>
            <view class="service-bottom">
              <text class="service-price">{{ service.price_text }}<text class="service-unit">/{{ service.unit || '次' }}</text></text>
              <view class="consult-btn" @click.stop="openOrderSheet(service)">去咨询</view>
            </view>
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="section-card news-section">
      <view class="section-head">
        <text class="section-title">越南灵感</text>
        <text class="section-more" @click="openNewsList">进入资讯 &gt;</text>
      </view>
      <view v-if="featuredArticles.length === 0" class="empty">正在同步后台资讯...</view>
      <view v-for="article in featuredArticles" :key="article.id" class="news-card" @click="openNewsList">
        <image class="news-cover" :src="article.cover_img" mode="aspectFill" />
        <view class="news-body">
          <text class="news-tag">{{ article.category }}</text>
          <text class="news-title">{{ article.title }}</text>
          <text class="news-summary">{{ article.summary }}</text>
        </view>
      </view>
    </view>

    <view class="why-choose-us" style="padding: 20px; background: #fff; border-radius: 16px; margin: 15px; box-shadow: 0 4px 24px rgba(0,77,64,0.04);">
      <view style="font-size: 18px; font-weight: bold; color: #004D40; margin-bottom: 20px; text-align: center; letter-spacing: 1px;">为什么选择 YesOk</view>
      <view style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
         <view style="text-align: center; padding: 10px;"><view style="font-size: 28px; margin-bottom: 8px;">🤝</view><view style="font-size: 14px; font-weight: bold; color: #333; margin-bottom: 4px;">华人团队</view><view style="font-size: 11px; color: #888; line-height: 1.4;">无缝沟通 懂你所需</view></view>
         <view style="text-align: center; padding: 10px;"><view style="font-size: 28px; margin-bottom: 8px;">🛡️</view><view style="font-size: 14px; font-weight: bold; color: #333; margin-bottom: 4px;">资金合规</view><view style="font-size: 11px; color: #888; line-height: 1.4;">平台担保 安全无忧</view></view>
         <view style="text-align: center; padding: 10px;"><view style="font-size: 28px; margin-bottom: 8px;">💰</view><view style="font-size: 14px; font-weight: bold; color: #333; margin-bottom: 4px;">透明报价</view><view style="font-size: 11px; color: #888; line-height: 1.4;">明码标价 拒绝隐形</view></view>
         <view style="text-align: center; padding: 10px;"><view style="font-size: 28px; margin-bottom: 8px;">⚡</view><view style="font-size: 14px; font-weight: bold; color: #333; margin-bottom: 4px;">极速响应</view><view style="font-size: 11px; color: #888; line-height: 1.4;">专属管家 1对1服务</view></view>
      </view>
    </view>

    <view v-if="selectedService" class="order-modal">
      <view class="order-sheet">
        <text class="panel-title">咨询 {{ selectedService.display_name }}</text>
        <input v-model="orderForm.contact_name" class="form-input" placeholder="联系人" />
        <input v-model="orderForm.contact_phone" class="form-input" placeholder="联系电话" />
        <input v-model="orderForm.hotel_address" class="form-input" placeholder="酒店/目的地" />
        <input v-model="orderForm.remark" class="form-input" placeholder="补充需求" />
        <view class="modal-actions">
          <view class="cancel-btn" @click="selectedService = null">取消</view>
          <view class="submit-btn" @click="submitOrder">提交订单</view>
        </view>
      </view>
    </view>

    <AuthPopup />
  </view>
</template>

<style scoped>
.home-page { min-height: 100vh; padding-bottom: 92px; background: #f2f6f5; color: #12312c; }
.hero-bleed { position: relative; height: 340px; margin: 0; overflow: hidden; border-bottom-left-radius: 0; border-bottom-right-radius: 0; background: transparent; }
.hero-image { position: absolute; inset: 0; width: 100%; height: 100%; transform: scale(1.02); }
.hero-mask { position: absolute; inset: 0; background: linear-gradient(to bottom, rgba(15,61,62,.04) 0%, rgba(15,61,62,.08) 44%, #f2f6f5 100%); }
.hero-topbar { position: relative; z-index: 1; display: flex; align-items: center; justify-content: space-between; padding: 62px 22px 0; color: #fff; }
.brand { font-size: 20px; font-weight: 900; letter-spacing: .4px; text-shadow: 0 8px 26px rgba(0,0,0,.28); }
.locale { padding: 7px 12px; border-radius: 999px; background: rgba(255,255,255,.22); backdrop-filter: blur(12px); font-size: 11px; font-weight: 900; }
.hero-copy { position: absolute; left: 22px; right: 22px; bottom: 54px; z-index: 1; color: #fff; }
.hero-title, .hero-subtitle, .section-title, .section-more, .service-name, .service-desc, .panel-kicker, .panel-title, .panel-desc, .news-tag, .news-title, .news-summary { display: block; }
.hero-title { max-width: 320px; font-size: 34px; font-weight: 900; line-height: 1.18; text-shadow: 0 14px 38px rgba(0,0,0,.32); }
.hero-subtitle { max-width: 310px; margin-top: 12px; color: rgba(255,255,255,.88); font-size: 14px; line-height: 1.7; }
.search-capsule { position: relative; z-index: 3; display: flex; align-items: center; gap: 10px; height: 52px; margin: -38px 18px 14px; padding: 0 18px; border: 1px solid rgba(255,255,255,.65); border-radius: 26px; background: rgba(255,255,255,.65); box-shadow: 0 18px 48px rgba(0,77,64,.16); backdrop-filter: blur(15px); }
.search-icon { color: #004d40; font-size: 20px; font-weight: 900; }
.search-input { flex: 1; color: #12312c; font-size: 14px; }
.category-card { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 8px; margin: 0 14px 16px; padding: 16px 10px; border-radius: 30px; background: rgba(255,255,255,.78); box-shadow: 0 18px 48px rgba(0,77,64,.08); backdrop-filter: blur(12px); }
.category-item { display: flex; flex-direction: column; align-items: center; gap: 7px; color: #4c5d59; font-size: 11px; font-weight: 800; text-align: center; }
.category-icon { display: flex; align-items: center; justify-content: center; width: 42px; height: 42px; border-radius: 18px; background: linear-gradient(135deg, rgba(0,77,64,.1), rgba(245,217,143,.28)); font-size: 22px; }
.section-card, .luxury-panel { margin: 14px; padding: 18px; border-radius: 32px; background: rgba(255,255,255,.82); box-shadow: 0 18px 52px rgba(0,77,64,.07); }
.section-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
.section-title { color: #12312c; font-size: 20px; font-weight: 900; }
.section-more { color: #c5a059; font-size: 12px; font-weight: 900; }
.hot-scroll { width: 100%; white-space: nowrap; }
.service-card { display: inline-block; width: 265px; margin-right: 14px; overflow: hidden; border-radius: 28px; background: #fff; box-shadow: 0 16px 42px rgba(0,77,64,.08); vertical-align: top; }
.service-cover { display: block; width: 100%; height: 128px; background: #dfeae6; }
.service-body { padding: 14px; }
.service-title-row { display: flex; align-items: center; gap: 8px; }
.service-icon { font-size: 20px; }
.service-name { color: #12312c; font-size: 16px; font-weight: 900; }
.service-desc { height: 38px; margin-top: 8px; color: #6b7c78; font-size: 12px; line-height: 1.6; white-space: normal; overflow: hidden; }
.service-bottom { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; }
.service-price { color: #e97832; font-size: 16px; font-weight: 900; }
.service-unit { color: #9aa7a3; font-size: 11px; }
.consult-btn { height: 34px; margin: 0; padding: 0 14px; border: 0; border-radius: 17px; color: #fff; background: #004d40; font-size: 12px; font-weight: 900; line-height: 34px; text-align: center; }
.news-section { padding-bottom: 8px; }
.news-card { display: flex; gap: 12px; margin-bottom: 12px; padding: 10px; border-radius: 24px; background: rgba(242,246,245,.78); }
.news-cover { flex: 0 0 92px; width: 92px; height: 92px; border-radius: 20px; background: #dfeae6; }
.news-body { flex: 1; min-width: 0; }
.news-tag { width: fit-content; padding: 3px 8px; border-radius: 999px; background: rgba(197,160,89,.16); color: #a37a29; font-size: 10px; font-weight: 900; }
.news-title { margin-top: 7px; color: #12312c; font-size: 15px; font-weight: 900; line-height: 1.35; }
.news-summary { height: 38px; margin-top: 5px; color: #6b7c78; font-size: 12px; line-height: 1.6; overflow: hidden; }
.panel-kicker { color: #c5a059; font-size: 10px; font-weight: 900; letter-spacing: 1.6px; }
.panel-title { margin-top: 8px; color: #12312c; font-size: 20px; font-weight: 900; line-height: 1.35; }
.panel-desc { margin-top: 8px; color: #6b7c78; font-size: 13px; line-height: 1.8; }
.empty { padding: 22px; color: #6b7c78; text-align: center; }
.order-modal { position: fixed; inset: 0; z-index: 30; display: flex; align-items: flex-end; background: rgba(0,0,0,.32); }
.order-sheet { width: 100%; padding: 22px; border-radius: 30px 30px 0 0; background: #fff; box-shadow: 0 -18px 60px rgba(0,0,0,.16); }
.form-input { box-sizing: border-box; width: 100%; height: 46px; margin-top: 10px; padding: 0 14px; border: 1px solid rgba(0,77,64,.1); border-radius: 16px; background: #f8fbfa; }
.modal-actions { display: flex; gap: 12px; margin-top: 14px; }
.cancel-btn, .submit-btn { flex: 1; height: 44px; border: 0; border-radius: 22px; font-weight: 900; line-height: 44px; text-align: center; }
.cancel-btn { color: #6b7c78; background: #eef5f2; }
.submit-btn { color: #fff; background: #004d40; }
@media (min-width: 768px) { .home-page { max-width: 560px; margin: 0 auto; box-shadow: 0 0 80px rgba(0,77,64,.08); } }
</style>
