<script setup>
import { computed, onMounted, ref } from 'vue'
import { get, ORIGIN_URL } from '@/api/request'
import { useGlobalShare } from '@/composables/useGlobalShare'

const loading = ref(true)
const activeCategory = ref('all')
const articles = ref([])
const categories = ref([
  { label: '全部', value: 'all' },
  { label: '落地指南', value: 'guide' },
  { label: '城市灵感', value: 'city' },
  { label: '服务公告', value: 'notice' },
])

useGlobalShare({ title: 'Yesok Vietnam｜越南资讯', path: '/pages/news/index' })

// 1.意图 -> 按当前分类筛选资讯列表，保持页面响应式渲染。
// 2.步骤 -> 当分类为 all 时返回全部资讯，否则匹配 article.category。
// 3.返回 -> 当前分类下可展示的资讯数组。
const visibleArticles = computed(() => {
  if (activeCategory.value === 'all') return articles.value
  return articles.value.filter((item) => item.category === activeCategory.value)
})

// showSafeToast 展示跨端提示。
// 1.意图 -> 数据加载失败时在 H5 与小程序都能反馈。
// 2.步骤 -> 优先使用 uni.showToast，缺失时降级 console.info。
// 3.返回 -> 无返回值。
const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Yesok News]', title)
}

const toFullUrl = (url) => {
  if (!url) return '/static/img.png'
  if (/^https?:\/\//.test(url)) return url

  // 小程序本地静态资源必须保持本地路径，不能拼 ORIGIN_URL
  if (url.startsWith('/static/') || url.startsWith('static/')) {
    return url.startsWith('/') ? url : `/${url}`
  }

  return `${ORIGIN_URL}${url.startsWith('/') ? '' : '/'}${url}`
}

// normalizeArticle 规范化资讯字段。
// 1.意图 -> 将接口返回的 sys_articles 字段转换为资讯卡片结构。
// 2.步骤 -> 补齐封面、作者、分类和摘要字段。
// 3.返回 -> 标准资讯对象。
const normalizeArticle = (item) => ({
  ...item,
  cover_img: toFullUrl(item.cover_img || '/static/img.png'),
  author: item.author || 'Yesok Vietnam',
  category: item.category || 'guide',
  summary: item.summary || item.content || 'Yesok Vietnam 管家精选资讯',
})

// loadArticles 拉取资讯列表。
// 1.意图 -> 让 C 端资讯 Tab 完全由后端 sys_articles 动态驱动。
// 2.步骤 -> 请求 /v1/articles 并写入响应式数组。
// 3.返回 -> Promise<void>。
const loadArticles = async () => {
  loading.value = true
  try {
    const res = await get('/v1/client/articles', { params: { limit: 20 } })
    articles.value = (res.data.list || []).map(normalizeArticle)
  } catch (error) {
    showSafeToast(error?.message || '资讯加载失败')
  } finally {
    loading.value = false
  }
}

// switchCategory 切换资讯分类。
// 1.意图 -> 在不刷新页面的情况下完成分类筛选。
// 2.步骤 -> 写入 activeCategory，visibleArticles 自动重新计算。
// 3.返回 -> 无返回值。
const switchCategory = (value) => {
  activeCategory.value = value
}

const openArticleDetail = (article) => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateTo) {
    uniApi.navigateTo({
      url: `/subpkg/article-detail/index?id=${encodeURIComponent(article.id)}`,
    })
  }
}

onMounted(loadArticles)
</script>

<template>
  <view class="news-page">
    <view class="top-hero">
      <image class="hero-bg" src="/static/img.png" mode="aspectFill" />
      <view class="hero-mask"></view>
<!--      <view class="hero-content">-->
<!--        <text class="kicker">YESOK JOURNAL</text>-->
<!--        <text class="title">越南商务与热带生活灵感</text>-->
<!--        <text class="subtitle">由后台资讯表实时驱动，每一条内容都可在 B 端配置。</text>-->
<!--      </view>-->
    </view>

    <scroll-view scroll-x class="category-scroll" enable-flex="true">
      <view class="category-row">
        <view v-for="item in categories" :key="item.value" class="category-btn" :class="{ active: activeCategory === item.value }" @click="switchCategory(item.value)">{{ item.label }}</view>
      </view>
    </scroll-view>

    <view v-if="loading" class="empty-card">正在读取后台资讯...</view>
    <view v-else-if="visibleArticles.length === 0" class="empty-card">暂无该分类资讯</view>
    <view v-else class="article-list">
      <view
      v-for="article in visibleArticles"
      :key="article.id"
      class="article-card"
      @click="openArticleDetail(article)"
    >
        <image class="article-cover" :src="article.cover_img" mode="aspectFill" />
        <view class="article-body">
          <view class="meta-row">
            <text class="tag">{{ article.category }}</text>
            <text class="views">{{ article.view_count || 0 }} 浏览</text>
          </view>
          <text class="article-title">{{ article.title }}</text>
          <text class="article-summary">{{ article.summary }}</text>
          <view class="author-row">
            <text class="author">{{ article.author }}</text>
            <text class="date">Yesok Vietnam</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.news-page { min-height: 100vh; padding-bottom: 96px; background: #f2f6f5; color: #12312c; }
.top-hero { position: relative; height: 280px; overflow: hidden; background: #12312c; }
.hero-bg { position: absolute; inset: 0; width: 100%; height: 100%; transform: scale(1.02); }
.hero-mask { position: absolute; inset: 0; background: linear-gradient(to bottom, rgba(0,77,64,.16), rgba(0,77,64,.34) 52%, #f2f6f5 100%); }
.hero-content { position: absolute; left: 22px; right: 22px; bottom: 42px; color: #fff; }
.kicker, .title, .subtitle, .article-title, .article-summary, .tag, .views, .author, .date { display: block; }
.kicker { color: #f5d98f; font-size: 11px; font-weight: 900; letter-spacing: 1.8px; }
.title { max-width: 330px; margin-top: 10px; font-size: 31px; font-weight: 900; line-height: 1.2; text-shadow: 0 14px 34px rgba(0,0,0,.28); }
.subtitle { max-width: 320px; margin-top: 12px; color: rgba(255,255,255,.88); font-size: 13px; line-height: 1.7; }
.category-scroll { position: relative; z-index: 2; width: 100%; margin-top: -24px; white-space: nowrap; }
.category-row { display: inline-flex; gap: 10px; padding: 0 16px 12px; }
.category-btn { height: 40px; margin: 0; padding: 0 18px; border: 0; border-radius: 999px; background: rgba(255,255,255,.78); color: #4c5d59; font-size: 13px; font-weight: 900; line-height: 40px; text-align: center; box-shadow: 0 12px 30px rgba(0,77,64,.08); }
.category-btn.active { color: #fff; background: #004d40; }
.empty-card { margin: 18px; padding: 28px; border-radius: 28px; background: rgba(255,255,255,.82); color: #6b7c78; text-align: center; box-shadow: 0 18px 52px rgba(0,77,64,.07); }
.article-list { padding: 4px 14px 18px; }
.article-card { overflow: hidden; margin-bottom: 16px; border-radius: 32px; background: rgba(255,255,255,.86); box-shadow: 0 18px 52px rgba(0,77,64,.07); }
.article-cover { display: block; width: 100%; height: 172px; background: #dfeae6; }
.article-body { padding: 16px; }
.meta-row, .author-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.tag { width: fit-content; padding: 4px 10px; border-radius: 999px; background: rgba(197,160,89,.16); color: #a37a29; font-size: 11px; font-weight: 900; }
.views, .date { color: #9aa7a3; font-size: 11px; }
.article-title { margin-top: 12px; color: #12312c; font-size: 20px; font-weight: 900; line-height: 1.35; }
.article-summary { margin-top: 8px; color: #6b7c78; font-size: 13px; line-height: 1.8; }
.author-row { margin-top: 14px; padding-top: 12px; border-top: 1px solid rgba(0,77,64,.08); }
.author { color: #004d40; font-size: 12px; font-weight: 900; }
@media (min-width: 768px) { .news-page { max-width: 560px; margin: 0 auto; box-shadow: 0 0 80px rgba(0,77,64,.08); } }
</style>
