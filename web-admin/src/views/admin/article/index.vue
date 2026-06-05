<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request, { BASE_URL, ORIGIN_URL } from '@/api/request'

const loading = ref(true)
const saving = ref(false)
const uploadLoading = ref(false)
const articles = ref([])
const dictData = ref([])

const dialogVisible = ref(false)
const dialogMode = ref('create')

const filters = ref({
  keyword: '',
  category: 'all',
  status: 'all',
})

const articleForm = ref(defaultArticleForm())

function defaultArticleForm() {
  return {
    id: null,
    title: '',
    cover_img: '/static/img.png',
    summary: '',
    content: '',
    category: 'guide',
    author: 'Yesok Vietnam',
    status: 1,
    sort_order: 10,
    view_count: 0,
  }
}

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const getAdminToken = () =>
  typeof localStorage !== 'undefined' ? localStorage.getItem('admin_token') || '' : ''

const uploadApiUrl = `${String(BASE_URL || '').replace(/\/+$/, '')}/v1/admin/upload`

const articleCategories = computed(() =>
  dictData.value.filter((item) => item.dict_code === 'article_category' && Number(item.status) === 1)
)

const categoryOptions = computed(() => [
  { dict_label: '全部分类', dict_value: 'all' },
  ...articleCategories.value,
])

const stripHtml = (html = '') =>
  String(html || '')
    .replace(/<[^>]+>/g, '')
    .replace(/&nbsp;/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

const articleExcerpt = (article) => {
  const text = String(article.summary || '').trim() || stripHtml(article.content || '')
  return text ? `${text.slice(0, 90)}${text.length > 90 ? '...' : ''}` : '暂无摘要'
}

const filteredArticles = computed(() => {
  const keyword = filters.value.keyword.trim().toLowerCase()

  return articles.value.filter((item) => {
    const matchKeyword =
      !keyword ||
      `${item.title || ''}${item.content || ''}${item.author || ''}${item.category || ''}${
        item.summary || ''
      }`
        .toLowerCase()
        .includes(keyword)

    const matchCategory =
      filters.value.category === 'all' || item.category === filters.value.category

    const matchStatus =
      filters.value.status === 'all' || String(item.status) === String(filters.value.status)

    return matchKeyword && matchCategory && matchStatus
  })
})

const categoryLabel = (value) => {
  const found = articleCategories.value.find((item) => item.dict_value === value)
  return found?.dict_label || value || '未分类'
}

const statusText = (status) => (Number(status) === 1 ? '已发布' : '草稿')

const resolveImageUrl = (url) => {
  if (!url) return '/static/img.png'
  if (/^https?:\/\//.test(url)) return url

  if (url.startsWith('/static/') || url.startsWith('static/')) {
    return url.startsWith('/') ? url : `/${url}`
  }

  if (url.startsWith('/uploads/') || url.startsWith('uploads/')) {
    const path = url.startsWith('/') ? url : `/${url}`
    return `${ORIGIN_URL}${path}`
  }

  if (url.startsWith('/material/') || url.startsWith('material/')) {
    const path = url.startsWith('/') ? url : `/${url}`
    return `${ORIGIN_URL}${path}`
  }

  return url.startsWith('/') ? `${ORIGIN_URL}${url}` : `${ORIGIN_URL}/${url}`
}

const loadArticles = async () => {
  loading.value = true
  try {
    const res = await request.get('/v1/admin/articles')
    const payload = res.data || res
    articles.value = payload.list || []
  } catch (error) {
    console.error('[Articles] load articles failed:', error)
    showToast(error?.message || '资讯加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const loadDicts = async () => {
  try {
    const res = await request.get('/v1/admin/dict-data', {
      params: { dict_type: 'article_category', pageSize: 200 },
    })
    const payload = res.data || res
    dictData.value = payload.data || payload.list || []
  } catch (error) {
    console.warn('[Articles] load dict failed:', error)
    dictData.value = []
  }
}

const openCreateDialog = () => {
  dialogMode.value = 'create'
  articleForm.value = defaultArticleForm()
  dialogVisible.value = true
}

const openEditDialog = (item) => {
  dialogMode.value = 'edit'
  articleForm.value = {
    ...defaultArticleForm(),
    ...item,
    cover_img: item.cover_img || '/static/img.png',
    content: item.content || '',
    summary: item.summary || '',
  }
  dialogVisible.value = true
}

const closeDialog = () => {
  dialogVisible.value = false
}

const validateArticle = () => {
  if (!String(articleForm.value.category || '').trim()) {
    showToast('请选择资讯分类', 'warning')
    return false
  }

  if (!String(articleForm.value.title || '').trim()) {
    showToast('请输入资讯标题', 'warning')
    return false
  }

  const summary = String(articleForm.value.summary || '').trim()
  if (!summary) {
    showToast('请输入资讯摘要', 'warning')
    return false
  }

  if (summary.length > 100) {
    showToast('资讯摘要不能超过100字', 'warning')
    return false
  }

  if (!String(articleForm.value.content || '').trim()) {
    showToast('请输入资讯正文', 'warning')
    return false
  }

  return true
}

const saveArticle = async () => {
  if (!validateArticle()) return

  saving.value = true
  try {
    const payload = {
      ...articleForm.value,
      title: articleForm.value.title.trim(),
      cover_img: articleForm.value.cover_img || '/static/img.png',
      summary: String(articleForm.value.summary || '').trim().slice(0, 100),
      content: articleForm.value.content.trim(),
      category: articleForm.value.category || 'guide',
      author: articleForm.value.author || 'Yesok Vietnam',
      sort_order: Number(articleForm.value.sort_order || 0),
      status: Number(articleForm.value.status || 1),
      view_count: Number(articleForm.value.view_count || 0),
    }

    if (payload.id) {
      await request.put(`/v1/admin/articles/${payload.id}`, payload)
    } else {
      await request.post('/v1/admin/articles', payload)
    }

    showToast(payload.id ? '资讯已更新' : '资讯已发布', 'success')
    dialogVisible.value = false
    await loadArticles()
  } catch (error) {
    console.error('[Articles] save failed:', error)
    showToast(error?.message || '资讯保存失败', 'error')
  } finally {
    saving.value = false
  }
}

const deleteArticle = async (item) => {
  try {
    await ElMessageBox.confirm(
      `确认删除资讯「${item.title}」吗？删除后 C 端将不再展示。`,
      '删除确认',
      { type: 'warning' }
    )

    await request.delete(`/v1/admin/articles/${item.id}`)
    showToast('资讯已删除', 'success')
    await loadArticles()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    console.error('[Articles] delete failed:', error)
    showToast(error?.message || '资讯删除失败', 'error')
  }
}

const uploadArticleCover = async (options) => {
  const file = options?.file
  if (!file) return

  uploadLoading.value = true

  try {
    const formData = new FormData()
    formData.append('file', file)

    const res = await fetch(uploadApiUrl, {
      method: 'POST',
      headers: { Authorization: `Bearer ${getAdminToken()}` },
      body: formData,
    })

    const body = await res.json()

    if (!res.ok) {
      throw new Error(body?.error || '上传失败')
    }

    articleForm.value.cover_img = body.url
    options?.onSuccess?.(body)
    showToast('封面已上传', 'success')
  } catch (error) {
    options?.onError?.(error)
    showToast(error?.message || '封面上传失败', 'error')
  } finally {
    uploadLoading.value = false
  }
}

const insertRichTag = (tag) => {
  const current = articleForm.value.content || ''

  const map = {
    h2: '<h2>请输入小标题</h2>',
    p: '<p>请输入段落内容</p>',
    strong: '<p><strong>请输入加粗内容</strong></p>',
    ul: '<ul><li>请输入列表项</li></ul>',
    quote: '<blockquote>请输入引用内容</blockquote>',
  }

  articleForm.value.content = `${current}\n${map[tag] || ''}`.trim()
}

const buildArticleImageHtml = (url) => {
  const safeUrl = String(url || '').trim()
  if (!safeUrl) return ''

  return `<img src="${safeUrl}" alt="资讯图片" style="display:block;width:100%;max-width:100%;height:auto;margin:20px auto;border-radius:14px;" />`
}

const normalizePreviewHtml = (html = '') => {
  let text = String(html || '').trim()
  if (!text) return ''

  text = text.replace(/<img\b([^>]*)>/gi, (match, attrs = '') => {
    let nextAttrs = attrs

    nextAttrs = nextAttrs.replace(/src=(["'])([^"']+)\1/i, (srcMatch, quote, rawSrc) => {
      const fullUrl = resolveImageUrl(rawSrc)
      return `src=${quote}${fullUrl}${quote}`
    })

    if (!/src=(["'])[^"']+\1/i.test(nextAttrs)) {
      return match
    }

    nextAttrs = nextAttrs
      .replace(/\sstyle=(["']).*?\1/gi, '')
      .replace(/\swidth=(["']).*?\1/gi, '')
      .replace(/\sheight=(["']).*?\1/gi, '')

    return `<img${nextAttrs} style="display:block;width:100%;max-width:100%;height:auto;margin:20px auto;border-radius:14px;" />`
  })

  return text
}

const previewHtml = computed(() => normalizePreviewHtml(articleForm.value.content))

const insertImageToContent = async (options) => {
  const file = options?.file
  if (!file) return

  try {
    const formData = new FormData()
    formData.append('file', file)

    const res = await fetch(uploadApiUrl, {
      method: 'POST',
      headers: { Authorization: `Bearer ${getAdminToken()}` },
      body: formData,
    })

    const body = await res.json()

    if (!res.ok) {
      throw new Error(body?.error || '图片上传失败')
    }

    const imgHtml = buildArticleImageHtml(body.url)
    articleForm.value.content = `${articleForm.value.content || ''}\n${imgHtml}`.trim()

    options?.onSuccess?.(body)
    showToast('正文图片已插入', 'success')
  } catch (error) {
    options?.onError?.(error)
    showToast(error?.message || '正文图片上传失败', 'error')
  }
}

const loadData = async () => {
  await Promise.all([loadArticles(), loadDicts()])
}

onMounted(loadData)
</script>

<template>
  <div class="articles-page">
    <div class="page-header">
      <div>
        <div class="page-title">资讯管理</div>
        <div class="page-sub">管理 C 端资讯列表和资讯详情，支持富文本与图片上传。</div>
      </div>
      <el-button class="primary-btn" type="primary" @click="openCreateDialog">
        + 新增资讯
      </el-button>
    </div>

    <div class="content-card filter-card">
      <el-input
        v-model="filters.keyword"
        class="filter-input"
        placeholder="搜索标题、正文、作者"
        clearable
      />

      <el-select v-model="filters.category" class="filter-select" placeholder="分类">
        <el-option
          v-for="cat in categoryOptions"
          :key="cat.dict_value"
          :label="cat.dict_label"
          :value="cat.dict_value"
        />
      </el-select>

      <el-select v-model="filters.status" class="filter-select" placeholder="状态">
        <el-option label="全部状态" value="all" />
        <el-option label="已发布" value="1" />
        <el-option label="草稿" value="0" />
      </el-select>
    </div>

    <div v-if="loading" class="content-card empty">正在加载资讯...</div>

    <div v-else class="article-grid">
      <div v-for="article in filteredArticles" :key="article.id" class="article-card">
        <img class="article-cover" :src="resolveImageUrl(article.cover_img)" alt="封面" />

        <div class="article-main">
          <div class="article-head">
            <div>
              <div class="article-title">{{ article.title }}</div>
              <div class="article-meta">
                {{ categoryLabel(article.category) }} · {{ article.author || 'Yesok Vietnam' }} ·
                {{ article.view_count || 0 }} 浏览
              </div>
            </div>
            <span class="status-pill" :class="{ draft: Number(article.status) !== 1 }">
              {{ statusText(article.status) }}
            </span>
          </div>

          <div class="article-summary">{{ articleExcerpt(article) }}</div>

          <div class="article-actions">
            <el-button class="ghost-btn" @click="openEditDialog(article)">编辑</el-button>
            <el-button class="danger-btn" @click="deleteArticle(article)">删除</el-button>
          </div>
        </div>
      </div>

      <div v-if="!filteredArticles.length" class="content-card empty">暂无资讯内容</div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增资讯' : '编辑资讯'"
      width="980px"
      destroy-on-close
      class="article-dialog"
    >
      <div class="article-editor">
        <div class="cover-panel">
          <div class="cover-title">资讯封面</div>
          <img class="cover-preview" :src="resolveImageUrl(articleForm.cover_img)" alt="封面" />

          <el-upload
            class="cover-uploader"
            action="#"
            :show-file-list="false"
            :http-request="uploadArticleCover"
            accept="image/*"
          >
            <el-button :loading="uploadLoading" class="ghost-btn"> 上传封面 </el-button>
          </el-upload>
        </div>

        <div class="form-panel">
          <el-form label-position="top">
            <el-form-item label="标题">
              <el-input v-model="articleForm.title" placeholder="请输入资讯标题" />
            </el-form-item>

            <div class="two-col">
              <el-form-item label="分类">
                <el-select
                  v-model="articleForm.category"
                  filterable
                  allow-create
                  placeholder="选择分类"
                >
                  <el-option
                    v-for="cat in articleCategories"
                    :key="cat.id"
                    :label="cat.dict_label"
                    :value="cat.dict_value"
                  />
                </el-select>
              </el-form-item>

              <el-form-item label="作者">
                <el-input v-model="articleForm.author" placeholder="作者" />
              </el-form-item>

              <el-form-item label="排序">
                <el-input-number v-model="articleForm.sort_order" :min="0" />
              </el-form-item>

              <el-form-item label="状态">
                <el-radio-group v-model="articleForm.status">
                  <el-radio-button :label="1">发布</el-radio-button>
                  <el-radio-button :label="0">草稿</el-radio-button>
                </el-radio-group>
              </el-form-item>
            </div>

            <el-form-item label="摘要">
              <el-input
                v-model="articleForm.summary"
                type="textarea"
                :rows="3"
                maxlength="100"
                show-word-limit
                placeholder="请输入资讯摘要，最多100字，将展示在C端资讯列表"
              />
            </el-form-item>

            <el-form-item label="富文本正文">
              <div class="rich-toolbar">
                <el-button size="small" @click="insertRichTag('h2')">小标题</el-button>
                <el-button size="small" @click="insertRichTag('p')">段落</el-button>
                <el-button size="small" @click="insertRichTag('strong')">加粗</el-button>
                <el-button size="small" @click="insertRichTag('ul')">列表</el-button>
                <el-button size="small" @click="insertRichTag('quote')">引用</el-button>

                <el-upload
                  action="#"
                  :show-file-list="false"
                  :http-request="insertImageToContent"
                  accept="image/*"
                >
                  <el-button size="small">插入图片</el-button>
                </el-upload>
              </div>

              <el-input
                v-model="articleForm.content"
                type="textarea"
                :rows="14"
                placeholder="请输入 HTML 富文本内容，例如 <h2>标题</h2><p>正文</p>，也可以点击“插入图片”自动插入图片"
              />

              <div class="rich-preview-title">预览</div>
              <div class="rich-preview" v-html="previewHtml"></div>
            </el-form-item>
          </el-form>
        </div>
      </div>

      <template #footer>
        <el-button @click="closeDialog">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveArticle">
          {{ dialogMode === 'create' ? '发布资讯' : '保存修改' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.articles-page {
  width: 100%;
  min-width: 960px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.page-title {
  color: #12312c;
  font-size: 26px;
  font-weight: 900;
}

.page-sub {
  margin-top: 6px;
  color: #6b7c78;
  font-size: 13px;
}

.content-card {
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
  margin-bottom: 16px;
}

.filter-card {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-input {
  max-width: 360px;
}

.filter-select {
  width: 180px;
}

.article-grid {
  display: grid;
  gap: 16px;
}

.article-card {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 20px;
  padding: 18px;
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 18px 52px rgba(0, 77, 64, 0.07);
}

.article-cover {
  width: 220px;
  height: 140px;
  border-radius: 24px;
  object-fit: cover;
  background: #dfeae6;
}

.article-main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.article-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.article-title {
  color: #12312c;
  font-size: 18px;
  font-weight: 900;
}

.article-meta {
  margin-top: 6px;
  color: #6b7c78;
  font-size: 12px;
}

.article-summary {
  margin-top: 12px;
  color: #4c5d59;
  font-size: 13px;
  line-height: 1.7;
}

.article-actions {
  display: flex;
  gap: 10px;
  margin-top: auto;
  padding-top: 16px;
}

.status-pill {
  flex-shrink: 0;
  height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(0, 77, 64, 0.1);
  color: #004d40;
  font-size: 12px;
  font-weight: 900;
  line-height: 30px;
}

.status-pill.draft {
  background: rgba(180, 35, 24, 0.08);
  color: #b42318;
}

.primary-btn {
  height: 44px;
  padding: 0 22px;
  border: 0;
  border-radius: 999px;
  background: #004d40;
  color: #fff;
  font-weight: 900;
}

.ghost-btn {
  height: 40px;
  padding: 0 16px;
  border: 0;
  border-radius: 999px;
  background: rgba(0, 77, 64, 0.08);
  color: #004d40;
  font-weight: 800;
}

.danger-btn {
  height: 40px;
  padding: 0 16px;
  border: 0;
  border-radius: 999px;
  background: rgba(180, 35, 24, 0.08);
  color: #b42318;
  font-weight: 800;
}

.empty {
  color: #6b7c78;
  text-align: center;
}

.article-editor {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 24px;
}

.cover-panel {
  padding: 18px;
  border-radius: 24px;
  background: #f2f6f5;
}

.cover-title {
  margin-bottom: 12px;
  color: #12312c;
  font-weight: 900;
}

.cover-preview {
  width: 100%;
  height: 180px;
  border-radius: 22px;
  object-fit: cover;
  background: #dfeae6;
}

.cover-uploader {
  margin-top: 12px;
}

.form-panel {
  min-width: 0;
}

.two-col {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.rich-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.rich-preview-title {
  margin-top: 14px;
  margin-bottom: 8px;
  color: #12312c;
  font-size: 13px;
  font-weight: 900;
}

.rich-preview {
  min-height: 120px;
  padding: 16px;
  border: 1px solid rgba(0, 77, 64, 0.08);
  border-radius: 18px;
  background: #f8fbfa;
  color: #12312c;
  line-height: 1.8;
}

.rich-preview :deep(img) {
  display: block;
  width: 100%;
  max-width: 100%;
  height: auto;
  margin: 20px auto;
  border-radius: 14px;
  object-fit: contain;
}

.rich-preview :deep(h2) {
  margin: 16px 0 8px;
  font-size: 20px;
}

.rich-preview :deep(blockquote) {
  margin: 12px 0;
  padding: 10px 14px;
  border-left: 4px solid #004d40;
  background: rgba(0, 77, 64, 0.06);
  color: #4c5d59;
}

@media (max-width: 1180px) {
  .article-card {
    grid-template-columns: 180px minmax(0, 1fr);
  }

  .article-cover {
    width: 180px;
  }
}
</style>
