<script setup>
import { computed, onMounted, ref } from 'vue'
import request from '@/api/request'

const loading = ref(true)
const loggedIn = ref(false)
const submitting = ref(false)
const activePanel = ref('dashboard')
const activeFilter = ref('all')
const loginForm = ref({ username: 'admin', password: '123456' })
const stats = ref({})
const orders = ref([])
const services = ref([])
const payments = ref([])
const appUsers = ref([])
const sysUsers = ref([])
const serviceForm = ref({ service_code: '', service_name: '', display_name: '', icon: '🌴', description: '', base_price: 0, currency: 'VND', unit: '次', sort_order: 10, status: 1, is_hot: false, form_schema: '{"fields":[]}' })

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

const filteredOrders = computed(() => {
  if (activeFilter.value === 'all') return orders.value
  return orders.value.filter((item) => item.current_status === activeFilter.value)
})

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
// 1.意图 -> 复用既有 admin_token 鉴权存储，避免破坏 AuthPopup 等底层逻辑。
// 2.步骤 -> 同步写入 uni storage 与浏览器 localStorage。
// 3.返回 -> 无返回值。
const saveAdminToken = (token) => {
  if (typeof uni !== 'undefined' && uni?.setStorageSync) uni.setStorageSync('admin_token', token)
  if (typeof localStorage !== 'undefined') localStorage.setItem('admin_token', token)
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
// 1.意图 -> 建立数据看板、订单、服务、财务、用户五个后台模块的真实闭环。
// 2.步骤 -> 并发读取统计、订单、服务、支付流水、C 端客户和后台员工。
// 3.返回 -> Promise<void>。
const loadAll = async () => {
  loading.value = true
  try {
    const [statsRes, ordersRes, servicesRes, paymentsRes, appUsersRes, sysUsersRes] = await Promise.all([
      request.get('/v1/admin/dashboard/stats'),
      request.get('/v1/admin/orders'),
      request.get('/v1/admin/services'),
      request.get('/v1/admin/payments'),
      request.get('/v1/admin/app-users'),
      request.get('/v1/admin/sys-users'),
    ])
    stats.value = statsRes.data
    orders.value = (ordersRes.data.list || ordersRes.data.orders || []).map(normalizeOrder)
    services.value = servicesRes.data.list || []
    payments.value = (paymentsRes.data.list || []).map(normalizePayment)
    appUsers.value = appUsersRes.data.list || []
    sysUsers.value = sysUsersRes.data.list || []
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
    serviceForm.value.service_code = ''
    serviceForm.value.service_name = ''
    serviceForm.value.display_name = ''
    serviceForm.value.description = ''
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

onMounted(async () => {
  loggedIn.value = Boolean((typeof localStorage !== 'undefined' && localStorage.getItem('admin_token')) || (typeof uni !== 'undefined' && uni?.getStorageSync?.('admin_token')))
  if (loggedIn.value) await loadAll()
  loading.value = false
})
</script>

<template>
  <view class="admin-shell">
    <view v-if="!loggedIn" class="login-card">
      <text class="eyebrow">YESOK COMMAND CENTER</text>
      <text class="login-title">管家后台登录</text>
      <text class="login-desc">使用 sys_users 种子账号进入商业闭环后台。</text>
      <input v-model="loginForm.username" class="input" placeholder="账号 admin" />
      <input v-model="loginForm.password" class="input" password placeholder="密码 123456" />
      <button class="primary-btn" :disabled="submitting" @click="login">进入后台</button>
    </view>

    <template v-else>
      <aside class="side-nav">
        <text class="brand">Yesok 2.0</text>
        <button v-for="panel in panels" :key="panel.key" class="nav-item" :class="{ active: activePanel === panel.key }" @click="activePanel = panel.key">{{ panel.label }}</button>
      </aside>

      <main class="workspace">
        <section class="hero-card">
          <view>
            <text class="eyebrow">REAL DATA BUSINESS LOOP</text>
            <text class="hero-title">越南奢华生活服务管家后台</text>
            <text class="hero-desc">后台配置服务与流程，C 端动态展示，下单后由状态节点自动沉淀时间线与财务流水。</text>
          </view>
          <button class="refresh-btn" @click="loadAll">刷新数据</button>
        </section>

        <section class="stats-grid">
          <view class="stat-card green"><text>{{ stats.total_orders || stats.totalOrders || 0 }}</text><span>总订单</span></view>
          <view class="stat-card"><text>{{ stats.pending_orders || stats.pendingOrders || 0 }}</text><span>待受理</span></view>
          <view class="stat-card"><text>{{ stats.completed_orders || stats.completedOrders || 0 }}</text><span>已完成</span></view>
          <view class="stat-card gold"><text>{{ stats.total_revenue_text || stats.total_revenue || 0 }}</text><span>确认收入</span></view>
        </section>

        <section v-if="activePanel === 'dashboard'" class="panel-grid">
          <view class="glass-panel wide"><text class="panel-title">今日履约雷达</text><view class="radar-line" v-for="filter in filters" :key="filter.key"><span>{{ filter.label }}</span><b>{{ filter.count }}</b></view></view>
          <view class="glass-panel"><text class="panel-title">服务上架</text><text class="big-number">{{ services.filter((s) => s.status === 1).length }}</text><span>项启用服务正在驱动 C 端</span></view>
          <view class="glass-panel"><text class="panel-title">客户矩阵</text><text class="big-number">{{ appUsers.length }}</text><span>位 C 端客户画像</span></view>
        </section>

        <section v-if="activePanel === 'orders'" class="content-card">
          <view class="section-head"><text class="panel-title">订单中心</text><view class="chip-row"><button v-for="filter in filters" :key="filter.key" class="chip" :class="{ active: activeFilter === filter.key }" @click="activeFilter = filter.key">{{ filter.label }} {{ filter.count }}</button></view></view>
          <view v-if="loading" class="empty">正在加载真实订单...</view>
          <view v-else class="order-list">
            <view v-for="order in filteredOrders" :key="order.id" class="order-card">
              <view class="order-top"><view><text class="order-title">{{ order.service_name }}</text><text class="muted">{{ order.order_no }} · {{ order.contact_name }} {{ order.contact_phone }}</text></view><span class="status">{{ statusMap[order.current_status] || order.current_status }}</span></view>
              <view class="order-meta"><span>金额：{{ order.totalAmountText }}</span><span>支付：{{ order.payment_status }}</span><span>时间：{{ order.created_at }}</span></view>
              <view class="json-box"><text v-for="(value, key) in order.form_data" :key="key">{{ key }}：{{ value }}</text></view>
              <view class="actions"><button v-for="node in order.actionNodes" :key="node.id" class="action-btn" :class="{ payment: node.trigger_payment }" @click="applyWorkflowAction(order, node)">{{ node.button_name }}</button><span v-if="!order.actionNodes.length" class="muted">暂无下一步动作</span></view>
            </view>
          </view>
        </section>

        <section v-if="activePanel === 'services'" class="content-card">
          <view class="section-head"><text class="panel-title">服务配置</text><span>服务价格与热门标签会立即驱动 C 端首页</span></view>
          <view class="service-form"><input v-model="serviceForm.service_code" class="input" placeholder="service_code" /><input v-model="serviceForm.service_name" class="input" placeholder="服务名称" /><input v-model="serviceForm.display_name" class="input" placeholder="展示名称" /><input v-model="serviceForm.icon" class="input" placeholder="图标" /><input v-model="serviceForm.base_price" class="input" type="number" placeholder="价格（分）" /><input v-model="serviceForm.unit" class="input" placeholder="单位" /><input v-model="serviceForm.description" class="input span-2" placeholder="服务描述" /><button class="primary-btn" @click="saveService">保存服务</button></view>
          <view class="table-list"><view v-for="service in services" :key="service.id" class="table-row"><span>{{ service.icon }} {{ service.display_name || service.service_name }}</span><span>{{ Math.round(service.base_price / 100) }} {{ service.currency }}/{{ service.unit }}</span><button class="ghost-btn" @click="toggleService(service)">{{ service.status === 1 ? '下架' : '上架' }}</button></view></view>
        </section>

        <section v-if="activePanel === 'finance'" class="content-card">
          <view class="section-head"><text class="panel-title">财务流水</text><span>订单状态推进到收款节点后自动生成 payment_records</span></view>
          <view class="table-list"><view v-for="payment in payments" :key="payment.id" class="table-row"><span>{{ payment.payment_no }} · 订单 #{{ payment.order_id }}</span><span>{{ payment.amountText }}</span><span>{{ payment.pay_status }}</span></view><view v-if="!payments.length" class="empty">暂无财务流水，推进订单“确认收款”后自动生成。</view></view>
        </section>

        <section v-if="activePanel === 'users'" class="panel-grid">
          <view class="glass-panel wide"><text class="panel-title">C 端客户矩阵</text><view v-for="user in appUsers" :key="user.id" class="user-line"><span>{{ user.nickname || '未命名客户' }}</span><b>{{ user.phone || user.wechat_open_id || '无联系方式' }}</b></view></view>
          <view class="glass-panel wide"><text class="panel-title">B 端员工矩阵</text><view v-for="user in sysUsers" :key="user.id" class="user-line"><span>{{ user.real_name || user.username }}</span><b>{{ user.role }}</b></view></view>
        </section>
      </main>
    </template>
  </view>
</template>

<style scoped>
.admin-shell { min-height: 100vh; background: #f2f6f5; color: #12312c; }
.login-card { width: min(440px, calc(100% - 32px)); margin: 10vh auto; padding: 34px; border-radius: 36px; background: rgba(255,255,255,.82); box-shadow: 0 26px 80px rgba(0,77,64,.12); backdrop-filter: blur(18px); }
.eyebrow, .login-title, .login-desc, .hero-title, .hero-desc, .panel-title, .order-title, .muted { display: block; }
.eyebrow { color: #c5a059; font-size: 11px; font-weight: 900; letter-spacing: 1.8px; }
.login-title, .hero-title { margin-top: 10px; font-size: 28px; font-weight: 900; }
.login-desc, .hero-desc, .muted, .section-head span, .glass-panel span { color: #6b7c78; font-size: 13px; line-height: 1.7; }
.input { box-sizing: border-box; width: 100%; height: 44px; margin-top: 12px; padding: 0 16px; border: 1px solid rgba(0,77,64,.1); border-radius: 16px; background: #fff; }
.primary-btn, .refresh-btn, .ghost-btn, .nav-item, .chip, .action-btn { border: 0; border-radius: 999px; font-weight: 900; }
.primary-btn { width: 100%; height: 46px; margin-top: 16px; color: #fff; background: #004d40; }
.side-nav { position: fixed; inset: 0 auto 0 0; width: 218px; padding: 28px 18px; background: linear-gradient(180deg,#07362f,#004d40); box-shadow: 24px 0 70px rgba(0,77,64,.16); }
.brand { display: block; margin-bottom: 28px; color: #fff; font-size: 24px; font-weight: 900; }
.nav-item { display: block; width: 100%; height: 44px; margin-bottom: 10px; color: rgba(255,255,255,.72); background: transparent; text-align: left; padding-left: 18px; }
.nav-item.active { color: #12312c; background: #f5d98f; }
.workspace { margin-left: 254px; padding: 24px; }
.hero-card { display: flex; align-items: center; justify-content: space-between; padding: 30px; border-radius: 36px; color: #fff; background: radial-gradient(circle at 85% 15%, rgba(245,217,143,.46), transparent 26%), linear-gradient(135deg,#004d40,#0f3d3e); box-shadow: 0 28px 80px rgba(0,77,64,.18); }
.refresh-btn { height: 42px; padding: 0 20px; color: #12312c; background: #f5d98f; }
.stats-grid, .panel-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; margin-top: 18px; }
.stat-card, .glass-panel, .content-card { border: 1px solid rgba(255,255,255,.72); border-radius: 30px; background: rgba(255,255,255,.78); box-shadow: 0 20px 60px rgba(0,77,64,.08); backdrop-filter: blur(15px); }
.stat-card { padding: 22px; }
.stat-card text, .big-number { display: block; font-size: 28px; font-weight: 900; }
.stat-card.green { color: #fff; background: #004d40; }
.stat-card.gold { background: linear-gradient(135deg,#f5d98f,#fff); }
.glass-panel, .content-card { padding: 22px; }
.panel-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.wide { grid-column: span 1; }
.radar-line, .user-line, .table-row, .order-top, .order-meta, .section-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.radar-line, .user-line, .table-row { padding: 12px 0; border-bottom: 1px solid rgba(0,77,64,.08); }
.content-card { margin-top: 18px; }
.chip-row, .actions { display: flex; flex-wrap: wrap; gap: 10px; }
.chip { height: 34px; padding: 0 14px; color: #6b7c78; background: #eef5f2; }
.chip.active, .action-btn { color: #fff; background: #004d40; }
.order-list { display: grid; gap: 14px; margin-top: 16px; }
.order-card { padding: 18px; border-radius: 24px; background: #fff; }
.order-title { font-size: 17px; font-weight: 900; }
.status { padding: 7px 12px; border-radius: 999px; color: #004d40; background: rgba(0,77,64,.08); font-size: 12px; font-weight: 900; }
.order-meta { margin: 12px 0; color: #6b7c78; font-size: 12px; }
.json-box { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; padding: 12px; border-radius: 18px; background: #f2f6f5; color: #4c5d59; font-size: 12px; }
.action-btn { height: 34px; padding: 0 16px; }
.action-btn.payment { color: #12312c; background: #f5d98f; }
.service-form { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; margin-top: 12px; }
.span-2 { grid-column: span 2; }
.table-list { margin-top: 16px; }
.ghost-btn { height: 32px; padding: 0 14px; color: #004d40; background: rgba(0,77,64,.08); }
.empty { padding: 20px; color: #6b7c78; text-align: center; }
@media (max-width: 820px) { .side-nav { position: static; width: auto; display: flex; overflow-x: auto; gap: 8px; padding: 14px; } .brand { display: none; } .nav-item { flex: 0 0 auto; width: auto; padding: 0 16px; text-align: center; } .workspace { margin-left: 0; padding: 14px; } .hero-card, .section-head { flex-direction: column; align-items: flex-start; } .stats-grid, .panel-grid, .service-form { grid-template-columns: repeat(2, minmax(0, 1fr)); } .json-box { grid-template-columns: 1fr; } }
@media (max-width: 520px) { .stats-grid, .panel-grid, .service-form { grid-template-columns: 1fr; } .span-2 { grid-column: span 1; } .order-meta { flex-direction: column; align-items: flex-start; } }
</style>
