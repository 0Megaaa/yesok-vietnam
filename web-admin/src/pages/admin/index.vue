<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const router = useRouter()

const loading = ref(true)
const loggedIn = ref(false)
const submitting = ref(false)
const sidebarCollapsed = ref(false)
const activePanel = ref('dashboard')
const activeFilter = ref('all')
const loginForm = ref({ username: 'admin', password: '123456' })
const stats = ref({})
const orders = ref([])
const services = ref([])
const payments = ref([])
const appUsers = ref([])
const sysUsers = ref([])
const dictTypes = ref([])
const dictData = ref([])
const articles = ref([])
const serviceForm = ref({ service_code: '', service_name: '', display_name: '', icon: '🌴', description: '', base_price: 0, currency: 'VND', unit: '次', sort_order: 10, status: 1, is_hot: false, form_schema: '{"fields":[]}' })
const dictTypeForm = ref({ id: null, dict_name: '', dict_code: '', remark: '', status: 1 })
const dictDataForm = ref({ id: null, dict_code: 'article_category', dict_label: '', dict_value: '', sort_order: 10, status: 1, remark: '' })
const articleForm = ref({ id: null, title: '', cover_img: '/static/img.png', summary: '', content: '', category: 'guide', author: 'Yesok Vietnam', status: 1, sort_order: 10, view_count: 0 })

const statusMap = {
  pending: '待受理',
  quoted: '已报价',
  paid: '已收款',
  in_progress: '履约中',
  completed: '已完成',
  cancelled: '已取消',
}

const panels = [
  { key: 'dashboard', label: '数据看板' },
  { key: 'orders', label: '订单中心' },
  { key: 'services', label: '服务配置' },
  { key: 'dicts', label: '字典管理' },
  { key: 'articles', label: '资讯配置' },
  { key: 'finance', label: '财务流水' },
  { key: 'users', label: '用户矩阵' },
]

const filters = computed(() => [
  { key: 'all', label: '全部', count: orders.value.length },
  { key: 'pending', label: '待受理', count: orders.value.filter((item) => item.current_status === 'pending').length },
  { key: 'quoted', label: '已报价', count: orders.value.filter((item) => item.current_status === 'quoted').length },
  { key: 'paid', label: '已收款', count: orders.value.filter((item) => item.current_status === 'paid').length },
  { key: 'in_progress', label: '履约中', count: orders.value.filter((item) => item.current_status === 'in_progress').length },
])
const filteredOrders = computed(() => (activeFilter.value === 'all' ? orders.value : orders.value.filter((item) => item.current_status === activeFilter.value)))
const articleCategories = computed(() => dictData.value.filter((item) => item.dict_code === 'article_category' && item.status === 1))

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const normalizeOrder = (order) => ({
  ...order,
  order_no: order.order_no || order.orderNo,
  service_name: order.service_name || order.serviceName,
  current_status: order.current_status || order.currentStatus,
  total_amount: order.total_amount || order.amount || 0,
  totalAmountText: order.price || `${Math.round((order.total_amount || 0) / 100)} ${order.currency || 'VND'}`,
  form_data: order.form_data || order.formData || {},
  actionNodes: (order.actionNodes || []).map((node) => ({
    ...node,
    button_name: node.button_name || node.buttonName,
    target_status: node.target_status || node.targetStatus,
    trigger_payment: node.trigger_payment || node.triggerPayment,
    required_material: node.required_material || node.requiredMaterial,
  })),
})

const normalizePayment = (payment) => {
  const payAmount = Number(payment.pay_amount ?? payment.payAmount ?? payment.amount ?? 0)
  return {
    ...payment,
    payment_no: payment.third_trade_no || payment.payment_no || payment.paymentNo || `PAY-${payment.id}`,
    amount: payAmount,
    amountText: `${Math.round(payAmount / 100)} ${payment.currency || 'VND'}`,
    pay_status: payment.status || payment.pay_status || payment.payStatus || 'pending',
  }
}

const saveAdminToken = (token) => {
  if (typeof localStorage !== 'undefined') localStorage.setItem('admin_token', token)
}

const getAdminToken = () => {
  if (typeof localStorage !== 'undefined') return localStorage.getItem('admin_token') || ''
  return ''
}

const login = async () => {
  submitting.value = true
  try {
    const res = await request.post('/v1/admin/auth/login', loginForm.value)
    saveAdminToken(res.data.token)
    loggedIn.value = true
    showToast('管家登录成功', 'success')
    await loadAll()
  } catch (error) {
    showToast(error?.message || '登录失败', 'error')
  } finally {
    submitting.value = false
  }
}

const loadAll = async () => {
  loading.value = true
  try {
    const [statsRes, ordersRes, servicesRes, paymentsRes, appUsersRes, sysUsersRes, dictTypeRes, dictDataRes, articleRes] = await Promise.all([
      request.get('/v1/admin/dashboard/stats'),
      request.get('/v1/admin/orders'),
      request.get('/v1/admin/services'),
      request.get('/v1/admin/payments'),
      request.get('/v1/admin/app-users'),
      request.get('/v1/admin/sys-users'),
      request.get('/v1/admin/dict-types'),
      request.get('/v1/admin/dict-data'),
      request.get('/v1/admin/articles'),
    ])
    stats.value = statsRes.data
    orders.value = (ordersRes.data.list || ordersRes.data.orders || []).map(normalizeOrder)
    services.value = servicesRes.data.list || []
    payments.value = (paymentsRes.data.list || []).map(normalizePayment)
    appUsers.value = appUsersRes.data.list || []
    sysUsers.value = sysUsersRes.data.list || []
    dictTypes.value = dictTypeRes.data.list || []
    dictData.value = dictDataRes.data.list || []
    articles.value = articleRes.data.list || []
  } catch (error) {
    if (error?.response?.status === 401) loggedIn.value = false
    showToast(error?.message || '后台数据加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const applyWorkflowAction = async (order, node) => {
  try {
    const res = await request.put(`/v1/admin/orders/${order.id}`, {
      target_status: node.target_status,
      remark: `${node.button_name}：由后台管家执行`,
    })
    orders.value = orders.value.map((item) => (item.id === order.id ? normalizeOrder(res.data.order || res.data) : item))
    await loadAll()
    showToast(`${node.button_name}已执行`, 'success')
  } catch (error) {
    showToast(error?.message || '流程推进失败', 'error')
  }
}

const saveService = async () => {
  try {
    await request.post('/v1/admin/services', { ...serviceForm.value, base_price: Number(serviceForm.value.base_price), sort_order: Number(serviceForm.value.sort_order) })
    serviceForm.value = { service_code: '', service_name: '', display_name: '', icon: '🌴', description: '', base_price: 0, currency: 'VND', unit: '次', sort_order: 10, status: 1, is_hot: false, form_schema: '{"fields":[]}' }
    showToast('服务配置已保存，C端将动态更新', 'success')
    await loadAll()
  } catch (error) {
    showToast(error?.message || '服务保存失败', 'error')
  }
}

const toggleService = async (service) => {
  try {
    await request.put(`/v1/admin/services/${service.id}`, { ...service, status: service.status === 1 ? 0 : 1 })
    await loadAll()
  } catch (error) {
    showToast(error?.message || '服务状态更新失败', 'error')
  }
}

const selectPanel = (key) => {
  activePanel.value = key
  if (typeof window !== 'undefined' && window.innerWidth < 768) sidebarCollapsed.value = true
}

const resetDictTypeForm = () => {
  dictTypeForm.value = { id: null, dict_name: '', dict_code: '', remark: '', status: 1 }
}

const editDictType = (item) => {
  dictTypeForm.value = { ...item }
}

const saveDictType = async () => {
  try {
    const payload = { ...dictTypeForm.value, status: Number(dictTypeForm.value.status) }
    if (payload.id) await request.put(`/v1/admin/dict-types/${payload.id}`, payload)
    else await request.post('/v1/admin/dict-types', payload)
    resetDictTypeForm()
    await loadAll()
    showToast('字典类型已保存', 'success')
  } catch (error) {
    showToast(error?.message || '字典类型保存失败', 'error')
  }
}

const deleteDictType = async (item) => {
  try {
    await request.delete(`/v1/admin/dict-types/${item.id}`)
    await loadAll()
    showToast('字典类型已删除', 'success')
  } catch (error) {
    showToast(error?.message || '字典类型删除失败', 'error')
  }
}

const resetDictDataForm = () => {
  dictDataForm.value = { id: null, dict_code: 'article_category', dict_label: '', dict_value: '', sort_order: 10, status: 1, remark: '' }
}

const editDictData = (item) => {
  dictDataForm.value = { ...item }
}

const saveDictData = async () => {
  try {
    const payload = { ...dictDataForm.value, sort_order: Number(dictDataForm.value.sort_order), status: Number(dictDataForm.value.status) }
    if (payload.id) await request.put(`/v1/admin/dict-data/${payload.id}`, payload)
    else await request.post('/v1/admin/dict-data', payload)
    resetDictDataForm()
    await loadAll()
    showToast('字典数据已保存', 'success')
  } catch (error) {
    showToast(error?.message || '字典数据保存失败', 'error')
  }
}

const deleteDictData = async (item) => {
  try {
    await request.delete(`/v1/admin/dict-data/${item.id}`)
    await loadAll()
    showToast('字典数据已删除', 'success')
  } catch (error) {
    showToast(error?.message || '字典数据删除失败', 'error')
  }
}

const resetArticleForm = () => {
  articleForm.value = { id: null, title: '', cover_img: '/static/img.png', summary: '', content: '', category: 'guide', author: 'Yesok Vietnam', status: 1, sort_order: 10, view_count: 0 }
}

const editArticle = (item) => {
  articleForm.value = { ...item }
}

const saveArticle = async () => {
  try {
    const payload = { ...articleForm.value, sort_order: Number(articleForm.value.sort_order), status: Number(articleForm.value.status), view_count: Number(articleForm.value.view_count || 0) }
    if (payload.id) await request.put(`/v1/admin/articles/${payload.id}`, payload)
    else await request.post('/v1/admin/articles', payload)
    resetArticleForm()
    await loadAll()
    showToast('资讯已保存，C端将动态更新', 'success')
  } catch (error) {
    showToast(error?.message || '资讯保存失败', 'error')
  }
}

const deleteArticle = async (item) => {
  try {
    await request.delete(`/v1/admin/articles/${item.id}`)
    await loadAll()
    showToast('资讯已删除', 'success')
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

const handleResize = () => {
  if (typeof window === 'undefined') return
  sidebarCollapsed.value = window.innerWidth < 768
}

onMounted(async () => {
  handleResize()
  if (typeof window !== 'undefined') window.addEventListener('resize', handleResize)
  loggedIn.value = Boolean(typeof localStorage !== 'undefined' && localStorage.getItem('admin_token'))
  if (loggedIn.value) await loadAll()
  loading.value = false
})

onUnmounted(() => {
  if (typeof window !== 'undefined') window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="admin-shell" :class="{ collapsed: sidebarCollapsed }">
    <div v-if="!loggedIn" class="login-card">
      <span class="eyebrow">YESOK COMMAND CENTER</span>
      <span class="login-title">管家后台登录</span>
      <span class="login-desc">使用 sys_users 种子账号进入商业闭环后台。</span>
      <el-input v-model="loginForm.username" class="input" placeholder="账号 admin" />
      <el-input v-model="loginForm.password" class="input" type="password" placeholder="密码 123456" show-password />
      <el-button class="primary-btn" :loading="submitting" type="default" @click="login">进入后台</el-button>
    </div>

    <template v-else>
      <el-button class="collapse-toggle" type="default" @click="sidebarCollapsed = !sidebarCollapsed">{{ sidebarCollapsed ? '展开菜单' : '收起菜单' }}</el-button>
      <div class="side-nav">
        <span class="brand">Yesok 2.0</span>
        <el-button v-for="panel in panels" :key="panel.key" class="nav-item" :class="{ active: activePanel === panel.key }" type="default" @click="selectPanel(panel.key)">{{ panel.label }}</el-button>
      </div>

      <div class="workspace">
        <div class="hero-card">
          <div>
            <span class="eyebrow">REAL DATA BUSINESS LOOP</span>
            <span class="hero-title">越南奢华生活服务管家后台</span>
            <span class="hero-desc">后台配置服务、字典、资讯与流程，C 端动态展示，下单后由状态节点自动沉淀时间线与财务流水。</span>
          </div>
          <el-button class="refresh-btn" type="default" @click="loadAll">刷新数据</el-button>
        </div>

        <div class="stats-grid">
          <div class="stat-card green"><span>{{ stats.total_orders || stats.totalOrders || 0 }}</span><span>总订单</span></div>
          <div class="stat-card"><span>{{ stats.pending_orders || stats.pendingOrders || 0 }}</span><span>待受理</span></div>
          <div class="stat-card"><span>{{ articles.length }}</span><span>资讯内容</span></div>
          <div class="stat-card gold"><span>{{ stats.total_revenue_text || stats.total_revenue || 0 }}</span><span>确认收入</span></div>
        </div>

        <div v-if="activePanel === 'dashboard'" class="panel-grid">
          <div class="glass-panel wide"><span class="panel-title">今日履约雷达</span><div v-for="filter in filters" :key="filter.key" class="radar-line"><span>{{ filter.label }}</span><span class="strong">{{ filter.count }}</span></div></div>
          <div class="glass-panel"><span class="panel-title">服务上架</span><span class="big-number">{{ services.filter((s) => s.status === 1).length }}</span><span>项启用服务正在驱动 C 端</span></div>
          <div class="glass-panel"><span class="panel-title">字典枚举</span><span class="big-number">{{ dictData.length }}</span><span>条可配置业务字典</span></div>
          <div class="glass-panel"><span class="panel-title">客户矩阵</span><span class="big-number">{{ appUsers.length }}</span><span>位 C 端客户画像</span></div>
        </div>

        <div v-if="activePanel === 'orders'" class="content-card">
          <div class="section-head"><span class="panel-title">订单中心</span><div class="chip-row"><el-button v-for="filter in filters" :key="filter.key" class="chip" :class="{ active: activeFilter === filter.key }" type="default" @click="activeFilter = filter.key">{{ filter.label }} {{ filter.count }}</el-button></div></div>
          <div v-if="loading" class="empty">正在加载真实订单...</div>
          <div v-else class="order-list">
            <div v-for="order in filteredOrders" :key="order.id" class="order-card">
              <div class="order-top"><div><span class="order-title">{{ order.service_name }}</span><span class="muted">{{ order.order_no }} · {{ order.contact_name }} {{ order.contact_phone }}</span></div><span class="status">{{ statusMap[order.current_status] || order.current_status }}</span></div>
              <div class="order-meta"><span>金额：{{ order.totalAmountText }}</span><span>支付：{{ order.payment_status }}</span><span>时间：{{ order.created_at }}</span></div>
              <div class="json-box"><span v-for="(value, key) in order.form_data" :key="key">{{ key }}：{{ value }}</span></div>
              <div class="actions"><el-button v-for="node in order.actionNodes" :key="node.id" class="action-btn" :class="{ payment: node.trigger_payment }" type="default" @click="applyWorkflowAction(order, node)">{{ node.button_name }}</el-button><span v-if="!order.actionNodes.length" class="muted">暂无下一步动作</span></div>
            </div>
          </div>
        </div>

        <div v-if="activePanel === 'services'" class="content-card">
          <div class="section-head"><span class="panel-title">服务配置</span><span>服务价格与热门标签会立即驱动 C 端首页</span></div>
          <div class="service-form">
            <el-input v-model="serviceForm.service_code" class="input" placeholder="service_code" />
            <el-input v-model="serviceForm.service_name" class="input" placeholder="服务名称" />
            <el-input v-model="serviceForm.display_name" class="input" placeholder="展示名称" />
            <el-input v-model="serviceForm.icon" class="input" placeholder="图标" />
            <el-input v-model="serviceForm.base_price" class="input" type="number" placeholder="价格（分）" />
            <el-input v-model="serviceForm.unit" class="input" placeholder="单位" />
            <el-input v-model="serviceForm.description" class="input span-2" placeholder="服务描述" />
            <el-button class="primary-btn" type="default" @click="saveService">保存服务</el-button>
          </div>
          <div class="table-scroll"><div class="table-list"><div v-for="service in services" :key="service.id" class="table-row"><span>{{ service.icon }} {{ service.display_name || service.service_name }}</span><span>{{ Math.round(service.base_price / 100) }} {{ service.currency }}/{{ service.unit }}</span><el-button class="ghost-btn" type="default" @click="toggleService(service)">{{ service.status === 1 ? '下架' : '上架' }}</el-button></div></div></div>
        </div>

        <div v-if="activePanel === 'dicts'" class="content-card">
          <div class="section-head"><span class="panel-title">字典管理</span><span>服务分类、资讯分类与订单状态统一维护</span></div>
          <div class="dual-grid">
            <div class="sub-panel">
              <span class="sub-title">字典类型</span>
              <div class="service-form two-col">
                <el-input v-model="dictTypeForm.dict_name" class="input" placeholder="字典名称" />
                <el-input v-model="dictTypeForm.dict_code" class="input" placeholder="dict_code" />
                <el-input v-model="dictTypeForm.remark" class="input span-2" placeholder="备注" />
                <el-button class="primary-btn" type="default" @click="saveDictType">{{ dictTypeForm.id ? '更新类型' : '新增类型' }}</el-button>
                <el-button class="ghost-btn form-btn" type="default" @click="resetDictTypeForm">清空</el-button>
              </div>
              <div class="table-scroll"><div class="table-list"><div v-for="item in dictTypes" :key="item.id" class="table-row"><span>{{ item.dict_name }} · {{ item.dict_code }}</span><span>{{ item.status === 1 ? '启用' : '停用' }}</span><el-button class="ghost-btn" type="default" @click="editDictType(item)">编辑</el-button><el-button class="danger-btn" type="default" @click="deleteDictType(item)">删除</el-button></div></div></div>
            </div>
            <div class="sub-panel">
              <span class="sub-title">字典数据</span>
              <div class="service-form two-col">
                <el-input v-model="dictDataForm.dict_code" class="input" placeholder="dict_code" />
                <el-input v-model="dictDataForm.dict_label" class="input" placeholder="标签" />
                <el-input v-model="dictDataForm.dict_value" class="input" placeholder="值" />
                <el-input v-model="dictDataForm.sort_order" class="input" type="number" placeholder="排序" />
                <el-input v-model="dictDataForm.remark" class="input span-2" placeholder="备注" />
                <el-button class="primary-btn" type="default" @click="saveDictData">{{ dictDataForm.id ? '更新数据' : '新增数据' }}</el-button>
                <el-button class="ghost-btn form-btn" type="default" @click="resetDictDataForm">清空</el-button>
              </div>
              <div class="table-scroll"><div class="table-list"><div v-for="item in dictData" :key="item.id" class="table-row"><span>{{ item.dict_code }} · {{ item.dict_label }}</span><span>{{ item.dict_value }}</span><el-button class="ghost-btn" type="default" @click="editDictData(item)">编辑</el-button><el-button class="danger-btn" type="default" @click="deleteDictData(item)">删除</el-button></div></div></div>
            </div>
          </div>
        </div>

        <div v-if="activePanel === 'articles'" class="content-card">
          <div class="section-head"><span class="panel-title">资讯配置</span><span>资讯封面可上传到 server/uploads 并通过 /uploads 访问</span></div>
          <div class="article-form">
            <div class="cover-box">
              <img class="cover-preview" :src="articleForm.cover_img || '/static/img.png'" alt="封面" />
              <input class="file-input" type="file" accept="image/*" @change="uploadArticleCover" />
              <span class="muted">当前封面：{{ articleForm.cover_img }}</span>
            </div>
            <div class="service-form two-col">
              <el-input v-model="articleForm.title" class="input span-2" placeholder="资讯标题" />
              <el-input v-model="articleForm.category" class="input" placeholder="分类，如 guide/city/notice" />
              <el-input v-model="articleForm.author" class="input" placeholder="作者" />
              <el-input v-model="articleForm.sort_order" class="input" type="number" placeholder="排序" />
              <el-input v-model="articleForm.status" class="input" type="number" placeholder="状态 1发布 0草稿" />
              <el-input v-model="articleForm.summary" class="input span-2" type="textarea" placeholder="摘要" />
              <el-input v-model="articleForm.content" class="input span-2" type="textarea" placeholder="正文内容" />
              <el-button class="primary-btn" type="default" @click="saveArticle">{{ articleForm.id ? '更新资讯' : '发布资讯' }}</el-button>
              <el-button class="ghost-btn form-btn" type="default" @click="resetArticleForm">清空</el-button>
            </div>
          </div>
          <div class="category-hint"><span v-for="cat in articleCategories" :key="cat.id">{{ cat.dict_label }}：{{ cat.dict_value }}</span></div>
          <div class="table-scroll"><div class="article-list"><div v-for="article in articles" :key="article.id" class="article-row"><img class="article-thumb" :src="article.cover_img || '/static/img.png'" alt="封面" /><div class="article-info"><span class="order-title">{{ article.title }}</span><span class="muted">{{ article.category }} · {{ article.summary }}</span></div><el-button class="ghost-btn" type="default" @click="editArticle(article)">编辑</el-button><el-button class="danger-btn" type="default" @click="deleteArticle(article)">删除</el-button></div></div></div>
        </div>

        <div v-if="activePanel === 'finance'" class="content-card">
          <div class="section-head"><span class="panel-title">财务流水</span><span>订单状态推进到收款节点后自动生成 payment_records</span></div>
          <div class="table-scroll"><div class="table-list"><div v-for="payment in payments" :key="payment.id" class="table-row"><span>{{ payment.payment_no }} · 订单 #{{ payment.order_id }}</span><span>{{ payment.amountText }}</span><span>{{ payment.pay_status }}</span></div><div v-if="!payments.length" class="empty">暂无财务流水，推进订单"确认收款"后自动生成。</div></div></div>
        </div>

        <div v-if="activePanel === 'users'" class="panel-grid">
          <div class="glass-panel wide"><span class="panel-title">C 端客户矩阵</span><div v-for="user in appUsers" :key="user.id" class="user-line"><span>{{ user.nickname || '未命名客户' }}</span><span class="strong">{{ user.phone || user.wechat_open_id || '无联系方式' }}</span></div></div>
          <div class="glass-panel wide"><span class="panel-title">B 端员工矩阵</span><div v-for="user in sysUsers" :key="user.id" class="user-line"><span>{{ user.real_name || user.username }}</span><span class="strong">{{ user.role }}</span></div></div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.admin-shell { display: flex; width: 100%; height: 100vh; overflow: hidden; background: #f2f6f5; color: #12312c; }
.login-card { width: min(440px, calc(100% - 32px)); margin: 10vh auto; padding: 34px; border-radius: 36px; background: rgba(255,255,255,.82); box-shadow: 0 26px 80px rgba(0,77,64,.12); backdrop-filter: blur(18px); }
.eyebrow, .login-title, .login-desc, .hero-title, .hero-desc, .panel-title, .order-title, .muted, .sub-title { display: block; }
.eyebrow { color: #c5a059; font-size: 11px; font-weight: 900; letter-spacing: 1.8px; }
.login-title, .hero-title { margin-top: 10px; font-size: 28px; font-weight: 900; }
.login-desc, .hero-desc, .muted, .section-head span + span, .glass-panel span:last-child { color: #6b7c78; font-size: 13px; line-height: 1.7; }
.input, .textarea { box-sizing: border-box; width: 100%; margin-top: 12px; }
.textarea { min-height: 92px; padding: 12px 16px; line-height: 1.6; }
.primary-btn, .refresh-btn, .ghost-btn, .danger-btn, .nav-item, .chip, .action-btn, .collapse-toggle { border: 0; border-radius: 999px; font-weight: 900; }
.primary-btn { width: 100%; height: 46px; margin-top: 16px; color: #fff; background: #004d40; }
.side-nav { position: fixed; inset: 0 auto 0 0; z-index: 8; width: 218px; padding: 72px 18px 28px; overflow-y: auto; background: linear-gradient(180deg,#07362f,#004d40); box-shadow: 24px 0 70px rgba(0,77,64,.16); transition: transform .24s ease; }
.admin-shell.collapsed .side-nav { transform: translateX(-256px); }
.brand { display: block; margin-bottom: 28px; color: #fff; font-size: 24px; font-weight: 900; }
.nav-item { display: block; width: 100%; height: 44px; margin-bottom: 10px; color: rgba(255,255,255,.72); background: transparent; text-align: left; padding-left: 18px; }
.nav-item.active { color: #12312c; background: #f5d98f; }
.collapse-toggle { position: fixed; top: 18px; left: 18px; z-index: 12; height: 38px; padding: 0 16px; color: #12312c; background: #f5d98f; box-shadow: 0 12px 32px rgba(0,77,64,.16); }
.workspace { box-sizing: border-box; flex: 1; width: 100%; height: 100vh; min-width: 0; margin-left: 254px; overflow-x: auto; overflow-y: auto; padding: 24px; transition: margin-left .24s ease; }
.admin-shell.collapsed .workspace { margin-left: 0; }
.hero-card { display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 30px; border-radius: 36px; color: #fff; background: radial-gradient(circle at 85% 15%, rgba(245,217,143,.46), transparent 26%), linear-gradient(135deg,#004d40,#0f3d3e); box-shadow: 0 28px 80px rgba(0,77,64,.18); }
.refresh-btn { height: 42px; padding: 0 20px; color: #12312c; background: #f5d98f; }
.stats-grid, .panel-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; margin-top: 18px; }
.stat-card, .glass-panel, .content-card, .sub-panel { border: 1px solid rgba(255,255,255,.72); border-radius: 30px; background: rgba(255,255,255,.78); box-shadow: 0 20px 60px rgba(0,77,64,.08); backdrop-filter: blur(15px); }
.stat-card { padding: 22px; }
.stat-card span:first-child, .big-number { display: block; font-size: 28px; font-weight: 900; }
.stat-card.green { color: #fff; background: #004d40; }
.stat-card.gold { background: linear-gradient(135deg,#f5d98f,#fff); }
.glass-panel, .content-card, .sub-panel { padding: 22px; }
.content-card { flex: 1; overflow-x: auto; box-sizing: border-box; margin-top: 18px; }
.panel-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.wide { grid-column: span 1; }
.radar-line, .user-line, .table-row, .order-top, .order-meta, .section-head, .article-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.radar-line, .user-line, .table-row { min-width: 1200px !important; padding: 12px 0; border-bottom: 1px solid rgba(0,77,64,.08); }
.strong { font-weight: 900; }
.chip-row, .actions, .category-hint { display: flex; flex-wrap: wrap; gap: 10px; }
.chip { height: 34px; padding: 0 14px; color: #6b7c78; background: #eef5f2; }
.chip.active, .action-btn { color: #fff; background: #004d40; }
.order-list { display: grid; gap: 14px; margin-top: 16px; }
.order-card { min-width: 1200px !important; padding: 18px; border-radius: 24px; background: #fff; }
.order-title { font-size: 17px; font-weight: 900; }
.status { padding: 7px 12px; border-radius: 999px; color: #004d40; background: rgba(0,77,64,.08); font-size: 12px; font-weight: 900; }
.order-meta { margin: 12px 0; color: #6b7c78; font-size: 12px; }
.json-box { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; padding: 12px; border-radius: 18px; background: #f2f6f5; color: #4c5d59; font-size: 12px; }
.action-btn { height: 34px; padding: 0 16px; }
.action-btn.payment { color: #12312c; background: #f5d98f; }
.service-form { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; margin-top: 12px; }
.two-col { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.span-2 { grid-column: span 2; }
.table-scroll { width: 100%; min-width: 100%; overflow-x: auto; }
.table-list, .article-list, .order-list { min-width: 1200px !important; margin-top: 16px; }
.ghost-btn, .danger-btn { height: 32px; padding: 0 14px; color: #004d40; background: rgba(0,77,64,.08); }
.danger-btn { color: #b42318; background: rgba(180,35,24,.09); }
.form-btn { height: 46px; margin-top: 16px; }
.dual-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; }
.sub-title { color: #12312c; font-size: 18px; font-weight: 900; }
.article-form { display: grid; grid-template-columns: 260px minmax(0, 1fr); gap: 18px; align-items: start; }
.cover-box { margin-top: 12px; }
.cover-preview { width: 100%; height: 180px; border-radius: 24px; background: #dfeae6; object-fit: cover; display: block; }
.file-input { width: 100%; margin-top: 12px; }
.category-hint { margin-top: 12px; color: #6b7c78; font-size: 12px; }
.article-row { min-width: 1200px !important; padding: 12px 0; border-bottom: 1px solid rgba(0,77,64,.08); }
.article-thumb { flex: 0 0 88px; width: 88px; height: 66px; border-radius: 16px; background: #dfeae6; object-fit: cover; }
.article-info { flex: 1; min-width: 260px; }
.empty { padding: 20px; color: #6b7c78; text-align: center; }
@media (max-width: 1024px) { .stats-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .dual-grid, .article-form { grid-template-columns: 1fr; } }
@media (max-width: 820px) { .side-nav { width: 226px; } .workspace { margin-left: 0; padding: 74px 14px 14px; } .hero-card, .section-head { flex-direction: column; align-items: flex-start; } .panel-grid, .service-form, .two-col { grid-template-columns: repeat(2, minmax(0, 1fr)); } .json-box { grid-template-columns: 1fr; } }
@media (max-width: 520px) { .stats-grid, .panel-grid, .service-form, .two-col { grid-template-columns: 1fr; } .span-2 { grid-column: span 1; } .order-meta { flex-direction: column; align-items: flex-start; } .login-card { padding: 24px; } }
</style>
