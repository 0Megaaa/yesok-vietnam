<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { get, ORIGIN_URL } from '@/api/request'
import { useGlobalShare } from '@/composables/useGlobalShare'

const articleId = ref('')
const loading = ref(true)
const article = ref(null)

const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Article Detail]', title)
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

const coverUrl = computed(() => toFullUrl(article.value?.cover_img || '/static/img.png'))

const contentHtml = computed(() => {
  return article.value?.content || '<p>暂无正文内容</p>'
})

const formatTime = (t) => {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

const loadArticle = async () => {
  if (!articleId.value) {
    loading.value = false
    showSafeToast('资讯不存在')
    return
  }

  loading.value = true
  try {
    const res = await get(`/v1/client/articles/${articleId.value}`)
    const payload = res.data || res
    article.value = payload.article || payload
  } catch (error) {
    console.error('[article-detail] load failed:', error)
    showSafeToast(error?.message || '资讯加载失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateBack) {
    uniApi.navigateBack()
  }
}

useGlobalShare({
  title: computed(() => article.value?.title || 'Yesok Vietnam 资讯'),
  path: computed(() => `/subpkg/article-detail/index?id=${articleId.value}`),
})

onLoad((query = {}) => {
  articleId.value = query.id || ''
  loadArticle()
})
</script>

<template>
  <view class="article-page">
    <view class="top-bar">
      <view class="back-btn" @click="goBack">‹</view>
      <text class="top-title">资讯详情</text>
    </view>

    <view v-if="loading" class="empty-card">正在加载资讯...</view>

    <view v-else-if="!article" class="empty-card">资讯不存在或已下架</view>

    <view v-else class="detail-card">
      <image class="cover" :src="coverUrl" mode="aspectFill" />

      <view class="body">
        <view class="meta-row">
          <text class="tag">{{ article.category || 'guide' }}</text>
          <text class="views">{{ article.view_count || 0 }} 浏览</text>
        </view>

        <text class="title">{{ article.title }}</text>

        <view class="sub-row">
          <text>{{ article.author || 'Yesok Vietnam' }}</text>
          <text>{{ formatTime(article.created_at) }}</text>
        </view>

        <text v-if="article.summary" class="summary">{{ article.summary }}</text>

        <rich-text class="rich-content" :nodes="contentHtml"></rich-text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.article-page {
  min-height: 100vh;
  padding: 72px 14px 28px;
  background: #f2f6f5;
  color: #12312c;
}

.top-bar {
  position: fixed;
  z-index: 10;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(242, 246, 245, 0.92);
  backdrop-filter: blur(12px);
}

.back-btn {
  position: absolute;
  left: 16px;
  top: 14px;
  width: 36px;
  height: 36px;
  border-radius: 18px;
  background: rgba(255,255,255,.86);
  color: #12312c;
  font-size: 30px;
  line-height: 32px;
  text-align: center;
  box-shadow: 0 10px 26px rgba(0,77,64,.1);
}

.top-title {
  color: #12312c;
  font-size: 16px;
  font-weight: 900;
}

.detail-card {
  overflow: hidden;
  border-radius: 32px;
  background: rgba(255,255,255,.9);
  box-shadow: 0 18px 52px rgba(0,77,64,.07);
}

.cover {
  display: block;
  width: 100%;
  height: 210px;
  background: #dfeae6;
}

.body {
  padding: 18px;
}

.meta-row,
.sub-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.tag {
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(197,160,89,.16);
  color: #a37a29;
  font-size: 11px;
  font-weight: 900;
}

.views,
.sub-row {
  color: #9aa7a3;
  font-size: 12px;
}

.title {
  display: block;
  margin-top: 14px;
  color: #12312c;
  font-size: 24px;
  font-weight: 900;
  line-height: 1.35;
}

.sub-row {
  margin-top: 12px;
}

.summary {
  display: block;
  margin-top: 16px;
  padding: 14px;
  border-radius: 18px;
  background: rgba(0,77,64,.06);
  color: #4c5d59;
  font-size: 14px;
  line-height: 1.8;
}

.rich-content {
  display: block;
  margin-top: 18px;
  color: #12312c;
  font-size: 15px;
  line-height: 1.9;
}

.empty-card {
  margin-top: 40px;
  padding: 28px;
  border-radius: 28px;
  background: rgba(255,255,255,.82);
  color: #6b7c78;
  text-align: center;
  box-shadow: 0 18px 52px rgba(0,77,64,.07);
}

@media (min-width: 768px) {
  .article-page {
    max-width: 560px;
    margin: 0 auto;
    box-shadow: 0 0 80px rgba(0,77,64,.08);
  }

  .top-bar {
    max-width: 560px;
    margin: 0 auto;
  }
}
</style>
