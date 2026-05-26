<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(true)
const articles = ref([])
const dictData = ref([])

const articleForm = ref({
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
})

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const getAdminToken = () =>
  typeof localStorage !== 'undefined' ? localStorage.getItem('admin_token') || '' : ''

const articleCategories = computed(() =>
  dictData.value.filter((item) => item.dict_code === 'article_category' && item.status === 1)
)

const loadData = async () => {
  loading.value = true
  try {
    console.log('[Articles] 发起 GET /v1/admin/articles + /v1/admin/dict-data')
    const [articleRes, dictRes] = await Promise.all([
      request.get('/v1/admin/articles'),
      request.get('/v1/admin/dict-data'),
    ])
    console.log('[Articles] ✅ articles =', articleRes.data)
    articles.value = articleRes.data.list || []
    dictData.value = dictRes.data.list || []
  } catch (error) {
    console.error('[Articles] ❌ 报错：', error)
    showToast(error?.message || '资讯加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  articleForm.value = {
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

const editArticle = (item) => {
  articleForm.value = { ...item }
}

const saveArticle = async () => {
  try {
    const payload = {
      ...articleForm.value,
      sort_order: Number(articleForm.value.sort_order || 0),
      status: Number(articleForm.value.status || 1),
      view_count: Number(articleForm.value.view_count || 0),
    }
    if (payload.id) {
      await request.put(`/v1/admin/articles/${payload.id}`, payload)
    } else {
      await request.post('/v1/admin/articles', payload)
    }
    resetForm()
    showToast('资讯已保存，C端将动态更新', 'success')
    await loadData()
  } catch (error) {
    showToast(error?.message || '资讯保存失败', 'error')
  }
}

const deleteArticle = async (item) => {
  try {
    await request.delete(`/v1/admin/articles/${item.id}`)
    showToast('资讯已删除', 'success')
    await loadData()
  } catch (error) {
    showToast(error?.message || '资讯删除失败', 'error')
  }
}

const uploadArticleCover = async (event) => {
  const file = event?.target?.files?.[0]
  if (!file) return
  try {
    const formData = new FormData()
    formData.append('file', file)
    const res = await fetch('/api/v1/admin/upload', {
      method: 'POST',
      headers: { Authorization: `Bearer ${getAdminToken()}` },
      body: formData,
    })
    const body = await res.json()
    if (!res.ok) throw new Error(body?.error || '上传失败')
    articleForm.value.cover_img = body.url
    showToast('封面已上传', 'success')
  } catch (error) {
    showToast(error?.message || '封面上传失败', 'error')
  }
}

onMounted(() => {
  console.log('[Articles] 组件挂载，调用 loadData')
  loadData()
})

onUnmounted(() => {
  loading.value = false
})
</script>

<template>
  <div class="articles-page">
    <div class="page-header">
      <span class="page-title">资讯配置</span>
      <span class="page-sub">资讯封面可上传到 server/uploads 并通过 /uploads 访问</span>
    </div>

    <!-- 资讯表单 -->
    <div class="content-card">
      <div class="article-form">
        <div class="cover-box">
          <img
            class="cover-preview"
            :src="articleForm.cover_img || '/static/img.png'"
            alt="封面"
          />
          <input class="file-input" type="file" accept="image/*" @change="uploadArticleCover" />
          <span class="cover-path">当前封面：{{ articleForm.cover_img }}</span>
        </div>

        <div class="service-form two-col">
          <el-input v-model="articleForm.title" class="input span-2" placeholder="资讯标题" />
          <el-input v-model="articleForm.category" class="input" placeholder="分类，如 guide/city/notice" />
          <el-input v-model="articleForm.author" class="input" placeholder="作者" />
          <el-input v-model="articleForm.sort_order" class="input" type="number" placeholder="排序" />
          <el-input v-model="articleForm.status" class="input" type="number" placeholder="状态 1发布 0草稿" />
          <el-input v-model="articleForm.summary" class="input span-2" type="textarea" placeholder="摘要" />
          <el-input v-model="articleForm.content" class="input span-2" type="textarea" placeholder="正文内容" />
          <el-button class="primary-btn" type="default" @click="saveArticle">
            {{ articleForm.id ? '更新资讯' : '发布资讯' }}
          </el-button>
          <el-button class="ghost-btn form-btn" type="default" @click="resetForm">清空</el-button>
        </div>
      </div>
    </div>

    <!-- 分类提示 -->
    <div class="category-hint">
      <span v-for="cat in articleCategories" :key="cat.id">
        {{ cat.dict_label }}：{{ cat.dict_value }}
      </span>
    </div>

    <!-- 资讯列表 -->
    <div class="content-card">
      <div class="table-scroll">
        <div class="article-list">
          <div v-for="article in articles" :key="article.id" class="article-row">
            <img
              class="article-thumb"
              :src="article.cover_img || '/static/img.png'"
              alt="封面"
            />
            <div class="article-info">
              <span class="order-title">{{ article.title }}</span>
              <span class="muted">{{ article.category }} · {{ article.summary }}</span>
            </div>
            <el-button class="ghost-btn" type="default" @click="editArticle(article)">编辑</el-button>
            <el-button class="danger-btn" type="default" @click="deleteArticle(article)">删除</el-button>
          </div>
          <div v-if="!articles.length" class="empty">暂无资讯内容</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.articles-page {
  width: 100%;
  min-width: 800px;
}

.page-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 900;
  color: #12312c;
}

.page-sub {
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

.article-form {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.cover-box {
  margin-top: 12px;
}

.cover-preview {
  width: 100%;
  height: 180px;
  border-radius: 24px;
  background: #dfeae6;
  object-fit: cover;
  display: block;
}

.file-input {
  width: 100%;
  margin-top: 12px;
}

.cover-path {
  display: block;
  margin-top: 8px;
  color: #6b7c78;
  font-size: 12px;
  word-break: break-all;
}

.service-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.input {
  width: 100%;
}

.span-2 {
  grid-column: span 2;
}

.primary-btn {
  height: 46px;
  color: #fff;
  background: #004d40;
  border: 0;
  border-radius: 999px;
  font-weight: 900;
  font-size: 15px;
}

.ghost-btn {
  height: 46px;
  padding: 0 16px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.form-btn {
  margin-top: 0;
}

.category-hint {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
  color: #6b7c78;
  font-size: 12px;
}

.table-scroll {
  width: 100%;
  overflow-x: auto;
}

.article-list {
  min-width: 800px;
  margin-top: 16px;
}

.article-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
}

.article-thumb {
  flex-shrink: 0;
  width: 88px;
  height: 66px;
  border-radius: 16px;
  background: #dfeae6;
  object-fit: cover;
}

.article-info {
  flex: 1;
  min-width: 200px;
  display: flex;
  flex-direction: column;
}

.order-title {
  font-size: 15px;
  font-weight: 900;
  color: #12312c;
}

.muted {
  display: block;
  margin-top: 4px;
  color: #6b7c78;
  font-size: 12px;
}

.ghost-btn {
  height: 32px;
  padding: 0 14px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.danger-btn {
  height: 32px;
  padding: 0 14px;
  color: #b42318;
  background: rgba(180, 35, 24, 0.09);
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.empty {
  padding: 20px;
  color: #6b7c78;
  text-align: center;
  font-size: 13px;
}

@media (max-width: 768px) {
  .article-form {
    grid-template-columns: 1fr;
  }

  .service-form {
    grid-template-columns: 1fr;
  }

  .span-2 {
    grid-column: span 1;
  }
}
</style>
