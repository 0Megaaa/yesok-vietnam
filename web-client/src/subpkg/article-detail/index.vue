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

const normalizeRichTextHtml = (html = '') => {
  const origin = String(ORIGIN_URL || '').replace(/\/+$/, '')
  let text = String(html || '').trim()

  if (!text) return '<p>暂无正文内容</p>'

  text = text.replace(/src=(["'])(?!https?:\/\/)([^"']+)\1/g, (match, quote, rawSrc) => {
    const src = String(rawSrc || '').trim()
    if (!src) return match

    if (src.startsWith('/static/') || src.startsWith('static/')) {
      const localPath = src.startsWith('/') ? src : `/${src}`
      return `src=${quote}${localPath}${quote}`
    }

    if (
      src.startsWith('/uploads/') ||
      src.startsWith('uploads/') ||
      src.startsWith('/material/') ||
      src.startsWith('material/')
    ) {
      const path = src.startsWith('/') ? src : `/${src}`
      return `src=${quote}${origin}${path}${quote}`
    }

    if (src.startsWith('/')) {
      return `src=${quote}${origin}${src}${quote}`
    }

    return `src=${quote}${src}${quote}`
  })

  text = text.replace(/<img\b([^>]*)>/gi, (match, attrs = '') => {
    const cleanedAttrs = attrs
      .replace(/\sstyle=(["']).*?\1/gi, '')
      .replace(/\swidth=(["']).*?\1/gi, '')
      .replace(/\sheight=(["']).*?\1/gi, '')

    return `<img${cleanedAttrs} style="display:block;width:100%;max-width:100%;height:auto;margin:24rpx auto;border-radius:20rpx;" />`
  })

  return text
}

const contentHtml = computed(() => normalizeRichTextHtml(article.value?.content || ''))

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
    <view v-if="loading" class="empty-state">
      <view class="floating-back" @click="goBack">‹</view>
      <view class="empty-card">正在加载资讯...</view>
    </view>

    <view v-else-if="!article" class="empty-state">
      <view class="floating-back" @click="goBack">‹</view>
      <view class="empty-card">资讯不存在或已下架</view>
    </view>

    <view v-else class="article-detail">
      <view class="cover-wrap">
        <image class="cover" :src="coverUrl" mode="aspectFill" />
        <view class="cover-mask"></view>
        <view class="back-btn" @click="goBack">‹</view>
      </view>

      <view class="body">
        <view class="meta-row">
          <text class="tag">{{ article.category || 'guide' }}</text>
          <text class="views">{{ article.view_count || 0 }} 浏览</text>
        </view>

        <text class="title">{{ article.title }}</text>

        <text v-if="article.summary" class="summary">{{ article.summary }}</text>

        <rich-text class="rich-content" :nodes="contentHtml"></rich-text>

        <view class="footer-row">
          <text>{{ article.author || 'Yesok Vietnam' }}</text>
          <text>{{ formatTime(article.created_at) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.article-page {
  min-height: 100vh;
  padding: 0;
  background: #f2f6f5;
  color: #12312c;
}

.article-detail {
  min-height: 100vh;
  background: #fff;
}

.cover-wrap {
  position: relative;
  width: 100%;
  height: 520rpx;
  overflow: hidden;
  background: #dfeae6;
}

.cover {
  display: block;
  width: 100%;
  height: 100%;
}

.cover-mask {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 180rpx;
  background: linear-gradient(to bottom, rgba(0,0,0,0), rgba(0,0,0,.22));
}

.back-btn,
.floating-back {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: rgba(255,255,255,.92);
  color: #12312c;
  font-size: 54rpx;
  line-height: 72rpx;
  box-shadow: 0 10rpx 28rpx rgba(0,0,0,.14);
}

.back-btn {
  position: absolute;
  z-index: 2;
  left: 28rpx;
  top: calc(env(safe-area-inset-top) + 24rpx);
}

.floating-back {
  position: fixed;
  z-index: 2;
  left: 28rpx;
  top: calc(env(safe-area-inset-top) + 24rpx);
}

.body {
  position: relative;
  z-index: 1;
  margin-top: -36rpx;
  padding: 36rpx 32rpx 48rpx;
  border-radius: 36rpx 36rpx 0 0;
  background: #fff;
}

.meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
}

.tag {
  padding: 10rpx 24rpx;
  border-radius: 999rpx;
  background: rgba(197,160,89,.16);
  color: #a37a29;
  font-size: 24rpx;
  font-weight: 900;
}

.views {
  color: #9aa7a3;
  font-size: 24rpx;
  font-weight: 700;
}

.title {
  display: block;
  margin-top: 32rpx;
  color: #12312c;
  font-size: 46rpx;
  font-weight: 900;
  line-height: 1.32;
}

.summary {
  display: block;
  margin-top: 28rpx;
  padding: 28rpx;
  border-radius: 24rpx;
  background: rgba(0,77,64,.06);
  color: #4c5d59;
  font-size: 29rpx;
  line-height: 1.85;
}

.rich-content {
  display: block;
  margin-top: 34rpx;
  color: #12312c;
  font-size: 31rpx;
  line-height: 2;
}

.rich-content :deep(img) {
  display: block;
  max-width: 100%;
  width: 100%;
  height: auto;
  margin: 24rpx auto;
  border-radius: 20rpx;
}

.rich-content :deep(p) {
  margin: 20rpx 0;
}

.footer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  margin-top: 52rpx;
  padding-top: 28rpx;
  border-top: 1rpx solid rgba(0,77,64,.08);
  color: #9aa7a3;
  font-size: 25rpx;
}

.empty-state {
  min-height: 100vh;
  padding-top: calc(env(safe-area-inset-top) + 120rpx);
}

.empty-card {
  margin: 40rpx 28rpx 0;
  padding: 56rpx;
  border-radius: 28rpx;
  background: rgba(255,255,255,.82);
  color: #6b7c78;
  text-align: center;
  box-shadow: 0 18rpx 52rpx rgba(0,77,64,.07);
}

@media (min-width: 768px) {
  .article-page {
    max-width: 560px;
    margin: 0 auto;
    box-shadow: 0 0 80px rgba(0,77,64,.08);
  }
}
</style>
