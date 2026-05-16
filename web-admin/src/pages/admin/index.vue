<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import request from '@/api/request'

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

// 1.意图 -> 基于订单、字典和资讯状态派生后台统计与筛选列表。
// 2.步骤 -> 使用 computed 对响应式数组进行状态计数和分类筛选，避免模板内复杂计算。
// 3.返回 -> 订单筛选项、当前筛选订单和资讯分类字典。
const filters = computed(() => [
  { key: 'all', label: '全部', count: orders.value.length },
  { key: 'pending', label: '待受理', count: orders.value.filter((item) => item.current_status === 'pending').length },
  { key: 'quoted', label: '已报价', count: orders.value.filter((item) => item.current_status === 'quoted').length },
  { key: 'paid', label: '已收款', count: orders.value.filter((item) => item.current_status === 'paid').length },
  { key: 'in_progress', label: '履约中', count: orders.value.filter((item) => item.current_status === 'in_progress').length },
])
const filteredOrders = computed(() => (activeFilter.value === 'all' ? orders.value : orders.value.filter((item) => item.current_status === activeFilter.value)))
const articleCategories = computed(() => dictData.value.filter((item) => item.dict_code === 'article_category' && item.status === 1))

// showSafeToast 展示跨端提示。
// 1.意图 -> B 端操作反馈在浏览器、H5 与小程序容器中均可见。
// 2.步骤 -> 优先使用 uni.showToast，缺失时降级 console.info。
// 3.返回 -> 无返回值。
const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Yesok Admin]', title)
}

// normalizeOrder 规范化后端订单字段。
// 1.意图 -> 兼容 Go 后端 snake_case 与页面驼峰展示需求。
// 2.步骤 -> 解析金额、状态、时间线和动态流程按钮字段。
// 3.返回 -> 后台页面可直接渲染的订单对象。
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

// normalizePayment 规范化后端财务流水字段。
// 1.意图 -> 让 B 端财务页准确展示 payment_records 的真实字段，避免 pay_amount 被旧 amount 字段映射成 NaN。
// 2.步骤 -> 兼容 snake_case 与 camelCase，补齐流水号、金额、状态、币种和展示文本。
// 3.返回 -> 财务流水表格可直接渲染的标准对象。
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

// saveAdminToken 保存后台令牌。
// 1.意图 -> 复用既有 admin_token 鉴权存储，避免破坏 AuthPopup 等底层登录逻辑。
// 2.步骤 -> 同步写入 uni storage 与浏览器 localStorage。
// 3.返回 -> 无返回值。
const saveAdminToken = (token) => {
  if (typeof uni !== 'undefined' && uni?.setStorageSync) uni.setStorageSync('admin_token', token)
  if (typeof localStorage !== 'undefined') localStorage.setItem('admin_token', token)
}

// getAdminToken 读取后台令牌。
// 1.意图 -> 给原生 fetch 上传图片时补齐与 request 封装一致的鉴权头。
// 2.步骤 -> 优先读取 uni storage，浏览器环境读取 localStorage。
// 3.返回 -> admin_token 字符串或空字符串。
const getAdminToken = () => {
  if (typeof uni !== 'undefined' && uni?.getStorageSync) return uni.getStorageSync('admin_token') || ''
  if (typeof localStorage !== 'undefined') return localStorage.getItem('admin_token') || ''
  return ''
}

// login 登录管家后台。
// 1.意图 -> 使用 sys_users 种子账号 admin/123456 获取真实 JWT。
// 2.步骤 -> 调用后端登录接口，保存 token，再加载全量经营数据。
// 3.返回 -> Promise<void>。
const login = async () => {
  submitting.value = true
  try {
    const res = await request.post('/v1/admin/auth/login', loginForm.value)
    saveAdminToken(res.data.token)
    loggedIn.value = true
    showSafeToast('管家登录成功')
    await loadAll()
  } catch (error) {
    showSafeToast(error?.message || '登录失败')
  } finally {
    submitting.value = false
  }
}

// loadAll 加载 B 端全量业务数据。
// 1.意图 -> 建立数据看板、订单、服务、财务、用户、字典和资讯七个后台模块的真实闭环。
// 2.步骤 -> 并发读取统计、订单、服务、支付流水、客户、员工、字典类型、字典数据和资讯列表。
// 3.返回 -> Promise<void>。
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
    showSafeToast(error?.message || '后台数据加载失败')
  } finally {
    loading.value = false
  }
}

// applyWorkflowAction 执行动态流程按钮。
// 1.意图 -> 让订单状态完全由 sys_workflow_nodes 驱动，而不是前端写死按钮。
// 2.步骤 -> 提交 target_status 给后端，后端生成时间线和必要财务流水。
// 3.返回 -> Promise<void>。
const applyWorkflowAction = async (order, node) => {
  try {
    const res = await request.put(`/v1/admin/orders/${order.id}`, {
      target_status: node.target_status,
      remark: `${node.button_name}：由后台管家执行`,
    })
    orders.value = orders.value.map((item) => (item.id === order.id ? normalizeOrder(res.data.order || res.data) : item))
    await loadAll()
    showSafeToast(`${node.button_name}已执行`)
  } catch (error) {
    showSafeToast(error?.message || '流程推进失败')
  }
}

// saveService 保存服务配置。
// 1.意图 -> 让 C 端服务分类、价格与热门卡片由 B 端动态驱动。
// 2.步骤 -> 将表单数据写入 sys_services，再刷新服务与看板数据。
// 3.返回 -> Promise<void>。
const saveService = async () => {
  try {
    await request.post('/v1/admin/services', { ...serviceForm.value, base_price: Number(serviceForm.value.base_price), sort_order: Number(serviceForm.value.sort_order) })
    serviceForm.value = { service_code: '', service_name: '', display_name: '', icon: '🌴', description: '', base_price: 0, currency: 'VND', unit: '次', sort_order: 10, status: 1, is_hot: false, form_schema: '{"fields":[]}' }
    showSafeToast('服务配置已保存，C端将动态更新')
    await loadAll()
  } catch (error) {
    showSafeToast(error?.message || '服务保存失败')
  }
}

// toggleService 切换服务启停状态。
// 1.意图 -> 支持后台一键控制 C 端是否展示某服务。
// 2.步骤 -> 将 status 在 0 与 1 之间切换并提交后端。
// 3.返回 -> Promise<void>。
const toggleService = async (service) => {
  try {
    await request.put(`/v1/admin/services/${service.id}`, { ...service, status: service.status === 1 ? 0 : 1 })
    await loadAll()
  } catch (error) {
    showSafeToast(error?.message || '服务状态更新失败')
  }
}

// selectPanel 切换后台模块。
// 1.意图 -> 支持响应式侧边栏在移动端点选模块后自动收起。
// 2.步骤 -> 设置 activePanel，小屏宽度下将 sidebarCollapsed 置为 true。
// 3.返回 -> 无返回值。
const selectPanel = (key) => {
  activePanel.value = key
  if (typeof window !== 'undefined' && window.innerWidth < 768) sidebarCollapsed.value = true
}

// resetDictTypeForm 重置字典类型表单。
// 1.意图 -> 在新增和编辑模式间快速恢复空表单。
// 2.步骤 -> 清空 id、名称、编码和备注并默认启用。
// 3.返回 -> 无返回值。
const resetDictTypeForm = () => {
  dictTypeForm.value = { id: null, dict_name: '', dict_code: '', remark: '', status: 1 }
}

// editDictType 编辑字典类型。
// 1.意图 -> 将已有字典类型回填到表单以支持修改。
// 2.步骤 -> 克隆当前行数据到 dictTypeForm。
// 3.返回 -> 无返回值。
const editDictType = (item) => {
  dictTypeForm.value = { ...item }
}

// saveDictType 保存字典类型。
// 1.意图 -> 让后台可新增或更新业务枚举分组。
// 2.步骤 -> 根据表单是否存在 id 选择 POST 或 PUT，再刷新全量数据。
// 3.返回 -> Promise<void>。
const saveDictType = async () => {
  try {
    const payload = { ...dictTypeForm.value, status: Number(dictTypeForm.value.status) }
    if (payload.id) await request.put(`/v1/admin/dict-types/${payload.id}`, payload)
    else await request.post('/v1/admin/dict-types', payload)
    resetDictTypeForm()
    await loadAll()
    showSafeToast('字典类型已保存')
  } catch (error) {
    showSafeToast(error?.message || '字典类型保存失败')
  }
}

// deleteDictType 删除字典类型。
// 1.意图 -> 支持后台清理无效枚举分组。
// 2.步骤 -> 调用 DELETE 接口并刷新字典列表。
// 3.返回 -> Promise<void>。
const deleteDictType = async (item) => {
  try {
    await request.delete(`/v1/admin/dict-types/${item.id}`)
    await loadAll()
    showSafeToast('字典类型已删除')
  } catch (error) {
    showSafeToast(error?.message || '字典类型删除失败')
  }
}

// resetDictDataForm 重置字典数据表单。
// 1.意图 -> 在新增和编辑模式间快速恢复默认枚举项表单。
// 2.步骤 -> 清空 id、标签、值和备注，默认使用资讯分类编码。
// 3.返回 -> 无返回值。
const resetDictDataForm = () => {
  dictDataForm.value = { id: null, dict_code: 'article_category', dict_label: '', dict_value: '', sort_order: 10, status: 1, remark: '' }
}

// editDictData 编辑字典数据。
// 1.意图 -> 将已有枚举项回填到表单以支持修改。
// 2.步骤 -> 克隆当前行数据到 dictDataForm。
// 3.返回 -> 无返回值。
const editDictData = (item) => {
  dictDataForm.value = { ...item }
}

// saveDictData 保存字典数据。
// 1.意图 -> 让后台可新增或更新具体枚举项。
// 2.步骤 -> 根据表单是否存在 id 选择 POST 或 PUT，再刷新全量数据。
// 3.返回 -> Promise<void>。
const saveDictData = async () => {
  try {
    const payload = { ...dictDataForm.value, sort_order: Number(dictDataForm.value.sort_order), status: Number(dictDataForm.value.status) }
    if (payload.id) await request.put(`/v1/admin/dict-data/${payload.id}`, payload)
    else await request.post('/v1/admin/dict-data', payload)
    resetDictDataForm()
    await loadAll()
    showSafeToast('字典数据已保存')
  } catch (error) {
    showSafeToast(error?.message || '字典数据保存失败')
  }
}

// deleteDictData 删除字典数据。
// 1.意图 -> 支持后台清理无效枚举项。
// 2.步骤 -> 调用 DELETE 接口并刷新字典数据列表。
// 3.返回 -> Promise<void>。
const deleteDictData = async (item) => {
  try {
    await request.delete(`/v1/admin/dict-data/${item.id}`)
    await loadAll()
    showSafeToast('字典数据已删除')
  } catch (error) {
    showSafeToast(error?.message || '字典数据删除失败')
  }
}

// resetArticleForm 重置资讯表单。
// 1.意图 -> 在新增和编辑模式间快速恢复默认文章表单。
// 2.步骤 -> 清空标题、摘要、正文和 id，保留默认封面与分类。
// 3.返回 -> 无返回值。
const resetArticleForm = () => {
  articleForm.value = { id: null, title: '', cover_img: '/static/img.png', summary: '', content: '', category: 'guide', author: 'Yesok Vietnam', status: 1, sort_order: 10, view_count: 0 }
}

// editArticle 编辑资讯文章。
// 1.意图 -> 将已有文章回填到表单以支持修改。
// 2.步骤 -> 克隆当前行数据到 articleForm。
// 3.返回 -> 无返回值。
const editArticle = (item) => {
  articleForm.value = { ...item }
}

// saveArticle 保存资讯文章。
// 1.意图 -> 让后台发布或更新 C 端首页和资讯 Tab 的动态内容。
// 2.步骤 -> 根据表单是否存在 id 选择 POST 或 PUT，再刷新文章列表。
// 3.返回 -> Promise<void>。
const saveArticle = async () => {
  try {
    const payload = { ...articleForm.value, sort_order: Number(articleForm.value.sort_order), status: Number(articleForm.value.status), view_count: Number(articleForm.value.view_count || 0) }
    if (payload.id) await request.put(`/v1/admin/articles/${payload.id}`, payload)
    else await request.post('/v1/admin/articles', payload)
    resetArticleForm()
    await loadAll()
    showSafeToast('资讯已保存，C端将动态更新')
  } catch (error) {
    showSafeToast(error?.message || '资讯保存失败')
  }
}

// deleteArticle 删除资讯文章。
// 1.意图 -> 支持后台清理无效资讯内容。
// 2.步骤 -> 调用 DELETE 接口并刷新资讯列表。
// 3.返回 -> Promise<void>。
const deleteArticle = async (item) => {
  try {
    await request.delete(`/v1/admin/articles/${item.id}`)
    await loadAll()
    showSafeToast('资讯已删除')
  } catch (error) {
    showSafeToast(error?.message || '资讯删除失败')
  }
}

// uploadArticleCover 上传资讯封面。
// 1.意图 -> 将本地图片真实保存到 server/uploads，并把 /uploads 静态 URL 写入文章表单。
// 2.步骤 -> 读取文件选择事件，使用 FormData 和 admin_token 调用 /api/v1/admin/upload。
// 3.返回 -> Promise<void>。
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
    showSafeToast('封面已上传')
  } catch (error) {
    showSafeToast(error?.message || '封面上传失败')
  }
}

// handleResize 处理响应式侧边栏状态。
// 1.意图 -> 在移动端默认折叠侧边栏，桌面端默认展开。
// 2.步骤 -> 读取 window.innerWidth，小于 768px 时折叠，否则展开。
// 3.返回 -> 无返回值。
const handleResize = () => {
  if (typeof window === 'undefined') return
  sidebarCollapsed.value = window.innerWidth < 768
}

onMounted(async () => {
  handleResize()
  if (typeof window !== 'undefined') window.addEventListener('resize', handleResize)
  loggedIn.value = Boolean((typeof localStorage !== 'undefined' && localStorage.getItem('admin_token')) || (typeof uni !== 'undefined' && uni?.getStorageSync?.('admin_token')))
  if (loggedIn.value) await loadAll()
  loading.value = false
})

onUnmounted(() => {
  if (typeof window !== 'undefined') window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <view class="admin-shell" :class="{ collapsed: sidebarCollapsed }">
    <view v-if="!loggedIn" class="login-card">
      <text class="eyebrow">YESOK COMMAND CENTER</text>
      <text class="login-title">管家后台登录</text>
      <text class="login-desc">使用 sys_users 种子账号进入商业闭环后台。</text>
      <input v-model="loginForm.username" class="input" placeholder="账号 admin" />
      <input v-model="loginForm.password" class="input" password placeholder="密码 123456" />
      <button class="primary-btn" :disabled="submitting" @click="login">进入后台</button>
    </view>

    <template v-else>
      <button class="collapse-toggle" @click="sidebarCollapsed = !sidebarCollapsed">{{ sidebarCollapsed ? '展开菜单' : '收起菜单' }}</button>
      <view class="side-nav">
        <text class="brand">Yesok 2.0</text>
        <button v-for="panel in panels" :key="panel.key" class="nav-item" :class="{ active: activePanel === panel.key }" @click="selectPanel(panel.key)">{{ panel.label }}</button>
      </view>

      <view class="workspace">
        <view class="hero-card">
          <view>
            <text class="eyebrow">REAL DATA BUSINESS LOOP</text>
            <text class="hero-title">越南奢华生活服务管家后台</text>
            <text class="hero-desc">后台配置服务、字典、资讯与流程，C 端动态展示，下单后由状态节点自动沉淀时间线与财务流水。</text>
          </view>
          <button class="refresh-btn" @click="loadAll">刷新数据</button>
        </view>

        <view class="stats-grid">
          <view class="stat-card green"><text>{{ stats.total_orders || stats.totalOrders || 0 }}</text><text>总订单</text></view>
          <view class="stat-card"><text>{{ stats.pending_orders || stats.pendingOrders || 0 }}</text><text>待受理</text></view>
          <view class="stat-card"><text>{{ articles.length }}</text><text>资讯内容</text></view>
          <view class="stat-card gold"><text>{{ stats.total_revenue_text || stats.total_revenue || 0 }}</text><text>确认收入</text></view>
        </view>

        <view v-if="activePanel === 'dashboard'" class="panel-grid">
          <view class="glass-panel wide"><text class="panel-title">今日履约雷达</text><view v-for="filter in filters" :key="filter.key" class="radar-line"><text>{{ filter.label }}</text><text class="strong">{{ filter.count }}</text></view></view>
          <view class="glass-panel"><text class="panel-title">服务上架</text><text class="big-number">{{ services.filter((s) => s.status === 1).length }}</text><text>项启用服务正在驱动 C 端</text></view>
          <view class="glass-panel"><text class="panel-title">字典枚举</text><text class="big-number">{{ dictData.length }}</text><text>条可配置业务字典</text></view>
          <view class="glass-panel"><text class="panel-title">客户矩阵</text><text class="big-number">{{ appUsers.length }}</text><text>位 C 端客户画像</text></view>
        </view>

        <view v-if="activePanel === 'orders'" class="content-card">
          <view class="section-head"><text class="panel-title">订单中心</text><view class="chip-row"><button v-for="filter in filters" :key="filter.key" class="chip" :class="{ active: activeFilter === filter.key }" @click="activeFilter = filter.key">{{ filter.label }} {{ filter.count }}</button></view></view>
          <view v-if="loading" class="empty">正在加载真实订单...</view>
          <view v-else class="order-list">
            <view v-for="order in filteredOrders" :key="order.id" class="order-card">
              <view class="order-top"><view><text class="order-title">{{ order.service_name }}</text><text class="muted">{{ order.order_no }} · {{ order.contact_name }} {{ order.contact_phone }}</text></view><text class="status">{{ statusMap[order.current_status] || order.current_status }}</text></view>
              <view class="order-meta"><text>金额：{{ order.totalAmountText }}</text><text>支付：{{ order.payment_status }}</text><text>时间：{{ order.created_at }}</text></view>
              <view class="json-box"><text v-for="(value, key) in order.form_data" :key="key">{{ key }}：{{ value }}</text></view>
              <view class="actions"><button v-for="node in order.actionNodes" :key="node.id" class="action-btn" :class="{ payment: node.trigger_payment }" @click="applyWorkflowAction(order, node)">{{ node.button_name }}</button><text v-if="!order.actionNodes.length" class="muted">暂无下一步动作</text></view>
            </view>
          </view>
        </view>

        <view v-if="activePanel === 'services'" class="content-card">
          <view class="section-head"><text class="panel-title">服务配置</text><text>服务价格与热门标签会立即驱动 C 端首页</text></view>
          <view class="service-form">
            <input v-model="serviceForm.service_code" class="input" placeholder="service_code" />
            <input v-model="serviceForm.service_name" class="input" placeholder="服务名称" />
            <input v-model="serviceForm.display_name" class="input" placeholder="展示名称" />
            <input v-model="serviceForm.icon" class="input" placeholder="图标" />
            <input v-model="serviceForm.base_price" class="input" type="number" placeholder="价格（分）" />
            <input v-model="serviceForm.unit" class="input" placeholder="单位" />
            <input v-model="serviceForm.description" class="input span-2" placeholder="服务描述" />
            <button class="primary-btn" @click="saveService">保存服务</button>
          </view>
          <view class="table-scroll"><view class="table-list"><view v-for="service in services" :key="service.id" class="table-row"><text>{{ service.icon }} {{ service.display_name || service.service_name }}</text><text>{{ Math.round(service.base_price / 100) }} {{ service.currency }}/{{ service.unit }}</text><button class="ghost-btn" @click="toggleService(service)">{{ service.status === 1 ? '下架' : '上架' }}</button></view></view></view>
        </view>

        <view v-if="activePanel === 'dicts'" class="content-card">
          <view class="section-head"><text class="panel-title">字典管理</text><text>服务分类、资讯分类与订单状态统一维护</text></view>
          <view class="dual-grid">
            <view class="sub-panel">
              <text class="sub-title">字典类型</text>
              <view class="service-form two-col"><input v-model="dictTypeForm.dict_name" class="input" placeholder="字典名称" /><input v-model="dictTypeForm.dict_code" class="input" placeholder="dict_code" /><input v-model="dictTypeForm.remark" class="input span-2" placeholder="备注" /><button class="primary-btn" @click="saveDictType">{{ dictTypeForm.id ? '更新类型' : '新增类型' }}</button><button class="ghost-btn form-btn" @click="resetDictTypeForm">清空</button></view>
              <view class="table-scroll"><view class="table-list"><view v-for="item in dictTypes" :key="item.id" class="table-row"><text>{{ item.dict_name }} · {{ item.dict_code }}</text><text>{{ item.status === 1 ? '启用' : '停用' }}</text><button class="ghost-btn" @click="editDictType(item)">编辑</button><button class="danger-btn" @click="deleteDictType(item)">删除</button></view></view></view>
            </view>
            <view class="sub-panel">
              <text class="sub-title">字典数据</text>
              <view class="service-form two-col"><input v-model="dictDataForm.dict_code" class="input" placeholder="dict_code" /><input v-model="dictDataForm.dict_label" class="input" placeholder="标签" /><input v-model="dictDataForm.dict_value" class="input" placeholder="值" /><input v-model="dictDataForm.sort_order" class="input" type="number" placeholder="排序" /><input v-model="dictDataForm.remark" class="input span-2" placeholder="备注" /><button class="primary-btn" @click="saveDictData">{{ dictDataForm.id ? '更新数据' : '新增数据' }}</button><button class="ghost-btn form-btn" @click="resetDictDataForm">清空</button></view>
              <view class="table-scroll"><view class="table-list"><view v-for="item in dictData" :key="item.id" class="table-row"><text>{{ item.dict_code }} · {{ item.dict_label }}</text><text>{{ item.dict_value }}</text><button class="ghost-btn" @click="editDictData(item)">编辑</button><button class="danger-btn" @click="deleteDictData(item)">删除</button></view></view></view>
            </view>
          </view>
        </view>

        <view v-if="activePanel === 'articles'" class="content-card">
          <view class="section-head"><text class="panel-title">资讯配置</text><text>资讯封面可上传到 server/uploads 并通过 /uploads 访问</text></view>
          <view class="article-form">
            <view class="cover-box"><image class="cover-preview" :src="articleForm.cover_img || '/static/img.png'" mode="aspectFill" /><input class="file-input" type="file" accept="image/*" @change="uploadArticleCover" /><text class="muted">当前封面：{{ articleForm.cover_img }}</text></view>
            <view class="service-form two-col"><input v-model="articleForm.title" class="input span-2" placeholder="资讯标题" /><input v-model="articleForm.category" class="input" placeholder="分类，如 guide/city/notice" /><input v-model="articleForm.author" class="input" placeholder="作者" /><input v-model="articleForm.sort_order" class="input" type="number" placeholder="排序" /><input v-model="articleForm.status" class="input" type="number" placeholder="状态 1发布 0草稿" /><textarea v-model="articleForm.summary" class="textarea span-2" placeholder="摘要" /><textarea v-model="articleForm.content" class="textarea span-2" placeholder="正文内容" /><button class="primary-btn" @click="saveArticle">{{ articleForm.id ? '更新资讯' : '发布资讯' }}</button><button class="ghost-btn form-btn" @click="resetArticleForm">清空</button></view>
          </view>
          <view class="category-hint"><text v-for="cat in articleCategories" :key="cat.id">{{ cat.dict_label }}：{{ cat.dict_value }}</text></view>
          <view class="table-scroll"><view class="article-list"><view v-for="article in articles" :key="article.id" class="article-row"><image class="article-thumb" :src="article.cover_img || '/static/img.png'" mode="aspectFill" /><view class="article-info"><text class="order-title">{{ article.title }}</text><text class="muted">{{ article.category }} · {{ article.summary }}</text></view><button class="ghost-btn" @click="editArticle(article)">编辑</button><button class="danger-btn" @click="deleteArticle(article)">删除</button></view></view></view>
        </view>

        <view v-if="activePanel === 'finance'" class="content-card">
          <view class="section-head"><text class="panel-title">财务流水</text><text>订单状态推进到收款节点后自动生成 payment_records</text></view>
          <view class="table-scroll"><view class="table-list"><view v-for="payment in payments" :key="payment.id" class="table-row"><text>{{ payment.payment_no }} · 订单 #{{ payment.order_id }}</text><text>{{ payment.amountText }}</text><text>{{ payment.pay_status }}</text></view><view v-if="!payments.length" class="empty">暂无财务流水，推进订单“确认收款”后自动生成。</view></view></view>
        </view>

        <view v-if="activePanel === 'users'" class="panel-grid">
          <view class="glass-panel wide"><text class="panel-title">C 端客户矩阵</text><view v-for="user in appUsers" :key="user.id" class="user-line"><text>{{ user.nickname || '未命名客户' }}</text><text class="strong">{{ user.phone || user.wechat_open_id || '无联系方式' }}</text></view></view>
          <view class="glass-panel wide"><text class="panel-title">B 端员工矩阵</text><view v-for="user in sysUsers" :key="user.id" class="user-line"><text>{{ user.real_name || user.username }}</text><text class="strong">{{ user.role }}</text></view></view>
        </view>
      </view>
    </template>
  </view>
</template>

<style scoped>
.admin-shell { display: flex; width: 100%; height: 100vh; overflow: hidden; background: #f2f6f5; color: #12312c; }
.login-card { width: min(440px, calc(100% - 32px)); margin: 10vh auto; padding: 34px; border-radius: 36px; background: rgba(255,255,255,.82); box-shadow: 0 26px 80px rgba(0,77,64,.12); backdrop-filter: blur(18px); }
.eyebrow, .login-title, .login-desc, .hero-title, .hero-desc, .panel-title, .order-title, .muted, .sub-title { display: block; }
.eyebrow { color: #c5a059; font-size: 11px; font-weight: 900; letter-spacing: 1.8px; }
.login-title, .hero-title { margin-top: 10px; font-size: 28px; font-weight: 900; }
.login-desc, .hero-desc, .muted, .section-head text + text, .glass-panel text:last-child { color: #6b7c78; font-size: 13px; line-height: 1.7; }
.input, .textarea { box-sizing: border-box; width: 100%; margin-top: 12px; padding: 0 16px; border: 1px solid rgba(0,77,64,.1); border-radius: 16px; background: #fff; color: #12312c; }
.input { height: 44px; }
.textarea { min-height: 92px; padding-top: 12px; line-height: 1.6; }
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
.stat-card text:first-child, .big-number { display: block; font-size: 28px; font-weight: 900; }
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
.cover-preview { width: 100%; height: 180px; border-radius: 24px; background: #dfeae6; }
.file-input { width: 100%; margin-top: 12px; }
.category-hint { margin-top: 12px; color: #6b7c78; font-size: 12px; }
.article-row { min-width: 1200px !important; padding: 12px 0; border-bottom: 1px solid rgba(0,77,64,.08); }
.article-thumb { flex: 0 0 88px; width: 88px; height: 66px; border-radius: 16px; background: #dfeae6; }
.article-info { flex: 1; min-width: 260px; }
.empty { padding: 20px; color: #6b7c78; text-align: center; }
@media (max-width: 1024px) { .stats-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .dual-grid, .article-form { grid-template-columns: 1fr; } }
@media (max-width: 820px) { .side-nav { width: 226px; } .workspace { margin-left: 0; padding: 74px 14px 14px; } .hero-card, .section-head { flex-direction: column; align-items: flex-start; } .panel-grid, .service-form, .two-col { grid-template-columns: repeat(2, minmax(0, 1fr)); } .json-box { grid-template-columns: 1fr; } }
@media (max-width: 520px) { .stats-grid, .panel-grid, .service-form, .two-col { grid-template-columns: 1fr; } .span-2 { grid-column: span 1; } .order-meta { flex-direction: column; align-items: flex-start; } .login-card { padding: 24px; } }
</style>
