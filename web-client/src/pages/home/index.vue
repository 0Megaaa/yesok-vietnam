<script setup>
import { computed, onMounted, ref } from 'vue'
import { get, ORIGIN_URL } from '@/api/request'
import AuthPopup from '@/components/AuthPopup.vue'
import { useClientStore } from '@/store/client'
import { useGlobalShare } from '@/composables/useGlobalShare'

const client = useClientStore()
const loading = ref(true)
const services = ref([])
const articles = ref([])
const configs = ref({})

useGlobalShare({ title: 'Yesok Vietnam｜越南本地生活管家', path: '/pages/home/index' })

// 容错计算属性，确保数据为 undefined 时不会报错
const hotServices = computed(() => {
  if (!services.value || !Array.isArray(services.value)) return []
  return services.value.filter((item) => item.is_hot).slice(0, 6)
})

const categories = computed(() => {
  if (!services.value || !Array.isArray(services.value)) return []
  return services.value.slice(0, 5).map((item) => ({
    id: item.id,
    icon: item.icon || '🌴',
    name: item.display_name || item.service_name
  }))
})

const bannerImage = computed(() => '/static/img.png')

const featuredArticles = computed(() => {
  if (!articles.value || !Array.isArray(articles.value)) return []
  return articles.value.slice(0, 3)
})

const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Yesok Client]', title)
}

const formatServicePrice = (item) => {
  const amount = Number(item?.base_price ?? item?.basePrice ?? item?.price ?? 0)
  if (!amount) return '面议'
  return `¥${amount.toLocaleString('zh-CN')}`
}

const toFullUrl = (url) => {
  if (!url) return '/static/img.png'
  if (/^https?:\/\//.test(url)) return url
  return `${ORIGIN_URL}${url.startsWith('/') ? '' : '/'}${url}`
}

const normalizeService = (item) => {
  return {
    ...item,
    display_name: item.display_name || item.service_name || item.name,
    price_text: formatServicePrice(item),
    cover_image: item.cover_image || bannerImage.value,
    service_code: item.service_code || item.code,
    icon: item.icon || '✨',
    unit: item.unit || '次'
  }
}

const normalizeArticle = (item) => ({
  ...item,
  cover_img: toFullUrl(item.cover_img || '/static/img.png'),
  category: item.category || 'guide',
  author: item.author || 'Yesok Vietnam',
  summary: item.summary || item.content || 'Yesok Vietnam 管家精选资讯',
})

// 独立并发加载，杜绝熔断
const loadHomeData = async () => {
  loading.value = true

  get('/v1/client/configs').then(res => {
    configs.value = res.data?.configs || res.data || {}
  }).catch(e => console.error('加载配置失败:', e))

  get('/v1/client/services').then(res => {
    services.value = (res.data?.list || res.data || []).map(normalizeService)
  }).catch(e => console.error('加载服务失败:', e))

  get('/v1/client/articles', { params: { limit: 3 } }).then(res => {
    articles.value = (res.data?.list || res.data || []).map(normalizeArticle)
  }).catch(e => console.error('加载资讯失败:', e))

  setTimeout(() => { loading.value = false }, 600)
}

const goPage = (page) => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.switchTab) uniApi.switchTab({ url: `/pages/${page}/index` })
}

const openServiceDetail = (service) => {
  const url = `/subpkg/service-detail/index?id=${encodeURIComponent(service.id)}&code=${encodeURIComponent(service.service_code || '')}`
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateTo) uniApi.navigateTo({ url })
}

const openNewsList = () => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.switchTab) uniApi.switchTab({ url: '/pages/news/index' })
}

const openArticleDetail = (article) => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateTo) {
    uniApi.navigateTo({
      url: `/subpkg/article-detail/index?id=${encodeURIComponent(article.id)}`,
    })
  }
}

onMounted(() => {
  loadHomeData()
})
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
      <view v-else-if="hotServices.length === 0" class="empty">暂无热门服务</view>
      <scroll-view v-else scroll-x class="hot-scroll">
        <view v-for="service in hotServices" :key="service.id" class="service-card" @click="openServiceDetail(service)">
          <image class="service-cover" :src="service.cover_image" mode="aspectFill" />
          <view class="service-body">
            <view class="service-title-row">
              <text class="service-icon">{{ service.icon || '✨' }}</text>
              <text class="service-name">{{ service.display_name }}</text>
            </view>
            <text class="service-desc">{{ service.description }}</text>
            <view class="service-bottom">
              <text class="service-price">
                {{ service.price_text }}
                <text class="service-unit">/{{ service.unit || '次' }}</text>
              </text>
              <view class="consult-btn" @click.stop="openServiceDetail(service)">去咨询</view>
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
      <view v-if="loading" class="empty">正在同步后台资讯...</view>
      <view v-else-if="featuredArticles.length === 0" class="empty">暂无相关资讯</view>
      <view
        v-for="article in featuredArticles"
        :key="article.id"
        class="news-card"
        @click="openArticleDetail(article)"
      >
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

    <AuthPopup />
  </view>
</template>

<style scoped>
.home-page { min-height: 100vh; padding-bottom: 92px; background: #f2f6f5; color: #12312c; }
.hero-bleed { position: relative; height: 340px; margin: 0; overflow: hidden; border-bottom-left-radius: 0; border-bottom-right-radius: 0; background: transparent; }
.hero-image { position: absolute; inset: 0; width: 100%; height: 100%; transform: scale(1.02); }
.hero-mask { position: absolute; inset: 0; background: linear-gradient(to bottom, rgba(15,61,62,.04) 0%, rgba(15,61,62,.08) 44%, #f2f6f5 100%); }
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
.hot-scroll { width: 100%; white-space: nowrap; overflow-x: auto; }
.service-card { display: inline-block; width: 265px; margin-right: 14px; border-radius: 28px; background: #fff; box-shadow: 0 16px 42px rgba(0,77,64,.08); vertical-align: top; white-space: normal; overflow: hidden; }
.service-cover { display: block; width: 100%; height: 128px; background: #dfeae6; }
.service-body { padding: 14px; }
.service-title-row { display: flex; align-items: center; gap: 8px; }
.service-icon { font-size: 20px; }
.service-name { color: #12312c; font-size: 16px; font-weight: 900; }
.service-desc { height: 38px; margin-top: 8px; color: #6b7c78; font-size: 12px; line-height: 1.6; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; white-space: normal; }
.service-bottom { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; }
.service-price { color: #e97832; font-size: 16px; font-weight: 900; }
.service-unit { color: #9aa7a3; font-size: 11px; font-weight: normal; margin-left: 2px; }
.consult-btn { height: 34px; padding: 0 16px; border-radius: 17px; color: #fff; background: #004d40; font-size: 12px; font-weight: 900; line-height: 34px; text-align: center; }
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
</style>
