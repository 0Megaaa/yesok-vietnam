import {
  MOCK_ADMIN_ORDERS,
  MOCK_ADMIN_USER,
  MOCK_NEWS,
  MOCK_NEWS_CATEGORIES,
  MOCK_ORDER_TIMELINES,
  MOCK_ORDERS,
  MOCK_PAYMENT_RECORDS,
  MOCK_SERVICES,
  MOCK_USER,
  MOCK_WORKFLOW_NODES,
} from './mockData'

const BASE_URL = '/api'
const TIMEOUT = 10000
const MOCK_DELAY = 180
const HTTP_OK_STATUS = 200

// cloneData 创建深拷贝数据，隔离页面局部修改。
// 意图：避免管理端列表、详情页或 C 端组件直接污染全局 Mock 常量。
// 实现步骤：
// 1. 将普通对象序列化为 JSON 字符串。
// 2. 再反序列化为新的对象引用。
// 3. 统一作为 Mock API 的返回值。
// 返回：与输入内容一致但引用独立的数据副本。
const cloneData = (payload) => JSON.parse(JSON.stringify(payload))

// getActionNodes 获取某个订单当前状态下可渲染的后台动作按钮。
// 意图：模拟 sys_workflow_nodes 表的状态驱动按钮能力。
// 实现步骤：
// 1. 按 serviceId 限定当前服务流程。
// 2. 按 currentStatus 限定当前状态节点。
// 3. 返回按钮名称、目标状态和是否触发支付等配置。
// 返回：后台订单看板可直接渲染的动作数组。
const getActionNodes = (serviceId, currentStatus) => MOCK_WORKFLOW_NODES.filter((node) => (
  node.serviceId === serviceId && node.currentStatus === currentStatus
))

// attachOrderRuntimeFields 补齐后台订单运行时展示字段。
// 意图：让列表页和详情页都能消费同一份订单结构，避免页面重复计算。
// 实现步骤：
// 1. 根据服务 ID 和状态码查找动态流程按钮。
// 2. 按订单号挂载时间线数据。
// 3. 过滤关联支付记录用于财务对账卡片。
// 返回：包含 actionNodes、timelines 和 payments 的订单对象。
const attachOrderRuntimeFields = (order) => ({
  ...order,
  actionNodes: getActionNodes(order.serviceId, order.currentStatus),
  timelines: MOCK_ORDER_TIMELINES[order.id] || [],
  payments: MOCK_PAYMENT_RECORDS.filter((payment) => payment.orderId === order.id),
})

// getAdminStats 汇总后台看板统计数据。
// 意图：为 B 端首页提供订单总量、待办数量和收款节点概览。
// 实现步骤：
// 1. 基于 Mock 订单数组实时计算总单量。
// 2. 统计 pending、processing、payment_pending 等关键状态。
// 3. 汇总支付记录中的待确认收款数量。
// 返回：后台统计卡片可直接展示的数据对象。
const getAdminStats = () => {
  const totalOrders = MOCK_ADMIN_ORDERS.length
  const pendingOrders = MOCK_ADMIN_ORDERS.filter((order) => order.currentStatus === 'pending').length
  const processingOrders = MOCK_ADMIN_ORDERS.filter((order) => order.currentStatus === 'processing').length
  const paymentPendingOrders = MOCK_ADMIN_ORDERS.filter((order) => order.currentStatus === 'payment_pending').length
  return {
    totalOrders,
    pendingOrders,
    processingOrders,
    paymentPendingOrders,
    todayNewOrders: 2,
    vipOrders: MOCK_ADMIN_ORDERS.filter((order) => order.priority === 'VIP').length,
  }
}

// createClientMockResponse 创建 C 端 Mock 响应对象。
// 意图：让首页、详情页、登录和下单全部脱离真实后端稳定预览。
// 实现步骤：
// 1. 先处理固定路由，例如登录、个人信息和聚合状态。
// 2. 再处理动态服务详情与下单动作。
// 3. 未命中时返回通用成功对象，方便渐进式开发。
// 返回：符合 request 封装约定的 Promise 响应。
function createClientMockResponse(method, url, data) {
  const mockRoutes = {
    '/v1/client/auth/tg': {
      token: 'mock-client-token-tg-placeholder',
      user: MOCK_USER,
      loginType: 'telegram_placeholder',
    },
    '/v1/client/auth/wechat': {
      token: 'mock-client-token-wechat-placeholder',
      user: MOCK_USER,
      loginType: 'wechat_mock',
    },
    '/v1/client/user/me': MOCK_USER,
    '/v1/client/state': {
      user: MOCK_USER,
      orders: MOCK_ORDERS,
      services: MOCK_SERVICES,
      newsCategories: MOCK_NEWS_CATEGORIES,
      news: MOCK_NEWS,
    },
  }

  if (url.startsWith('/v1/client/services/')) {
    const serviceId = decodeURIComponent(url.split('/').pop())
    const service = MOCK_SERVICES.find((item) => item.id === serviceId) || MOCK_SERVICES[0]
    return Promise.resolve({ data: cloneData(service), status: HTTP_OK_STATUS })
  }

  if (method === 'POST' && url === '/v1/client/orders') {
    const service = MOCK_SERVICES.find((item) => item.id === data?.serviceId) || MOCK_SERVICES[0]
    const mockOrderNo = `YS${Date.now()}`
    return Promise.resolve({
      data: cloneData({
        id: mockOrderNo,
        orderNo: mockOrderNo,
        serviceName: service.name,
        icon: service.icon,
        status: 'pending',
        sk: 'pending',
        managerName: 'Linh 专属管家',
        price: `${service.price}${service.unit ? ` ${service.unit}` : ''}`,
      }),
      status: HTTP_OK_STATUS,
    })
  }

  if (mockRoutes[url]) {
    return Promise.resolve({ data: cloneData(mockRoutes[url]), status: HTTP_OK_STATUS })
  }

  return Promise.resolve({ data: { ok: true, url, method }, status: HTTP_OK_STATUS })
}

// createAdminMockResponse 创建 B 端 Mock 响应对象。
// 意图：让管理端订单看板、详情页、登录与统计全部使用前端 Mock 数据。
// 实现步骤：
// 1. 处理后台认证、统计、订单列表和订单详情路由。
// 2. 使用 sys_workflow_nodes 风格数据动态挂载按钮。
// 3. PUT 更新动作返回模拟后的目标状态，便于页面即时反馈。
// 返回：符合 request 封装约定的 Promise 响应。
function createAdminMockResponse(method, url, data) {
  if (method === 'POST' && url === '/v1/admin/auth/login') {
    return Promise.resolve({
      data: cloneData({ token: 'mock-admin-token-yesok-2', user: MOCK_ADMIN_USER }),
      status: HTTP_OK_STATUS,
    })
  }

  if (url === '/v1/admin/auth/me') {
    return Promise.resolve({ data: cloneData(MOCK_ADMIN_USER), status: HTTP_OK_STATUS })
  }

  if (url === '/v1/admin/auth/logout') {
    return Promise.resolve({ data: { ok: true }, status: HTTP_OK_STATUS })
  }

  if (url === '/v1/admin/dashboard/stats') {
    return Promise.resolve({ data: cloneData(getAdminStats()), status: HTTP_OK_STATUS })
  }

  if (method === 'GET' && url === '/v1/admin/orders') {
    return Promise.resolve({
      data: cloneData({
        list: MOCK_ADMIN_ORDERS.map(attachOrderRuntimeFields),
        total: MOCK_ADMIN_ORDERS.length,
      }),
      status: HTTP_OK_STATUS,
    })
  }

  if (url.startsWith('/v1/admin/orders/')) {
    const orderId = decodeURIComponent(url.split('/').pop())
    const order = MOCK_ADMIN_ORDERS.find((item) => item.id === orderId) || MOCK_ADMIN_ORDERS[0]

    if (method === 'PUT') {
      const targetStatus = data?.targetStatus || data?.currentStatus || order.currentStatus
      const updatedOrder = attachOrderRuntimeFields({
        ...order,
        status: targetStatus,
        currentStatus: targetStatus,
        statusText: data?.statusText || '状态已更新',
        updatedAt: '刚刚',
      })
      return Promise.resolve({ data: cloneData(updatedOrder), status: HTTP_OK_STATUS })
    }

    return Promise.resolve({ data: cloneData(attachOrderRuntimeFields(order)), status: HTTP_OK_STATUS })
  }

  return Promise.resolve({ data: { ok: true, url, method }, status: HTTP_OK_STATUS })
}

// createMockResponse 按业务端创建 Mock 响应对象。
// 意图：统一拦截 /v1/client 与 /v1/admin，贯彻前端 Mock 隔离原则。
// 实现步骤：
// 1. 客户端路由交给 createClientMockResponse。
// 2. 后台路由交给 createAdminMockResponse。
// 3. 其他路由返回通用成功对象。
// 返回：Promise 包装的 Mock 响应。
function createMockResponse(method, url, data) {
  if (url.startsWith('/v1/client')) return createClientMockResponse(method, url, data)
  if (url.startsWith('/v1/admin')) return createAdminMockResponse(method, url, data)
  return Promise.resolve({ data: { ok: true, url, method }, status: HTTP_OK_STATUS })
}

// createRequest 创建统一请求实例。
// 意图：提供 get/post/put/delete/patch 标准方法，同时让演示版完全不依赖真实后端。
// 实现步骤：
// 1. /v1/client 与 /v1/admin 默认走 Mock，保证 C/B 双端都能独立验收。
// 2. 其他路由保留真实 uni.request 能力，为未来联调预留路径。
// 3. 保持方法签名与旧封装一致，降低页面迁移风险。
// 返回：包含常见 HTTP 方法的 request 实例。
function createRequest() {
  const instance = {}

  function request(method, url, data, config = {}) {
    const shouldUseMock = url && (url.startsWith('/v1/client') || url.startsWith('/v1/admin'))
    if (shouldUseMock) {
      return new Promise((resolve) => {
        setTimeout(() => {
          createMockResponse(method, url, data).then(resolve)
        }, MOCK_DELAY)
      })
    }

    return new Promise((resolve, reject) => {
      const isAdminRoute = url && url.startsWith('/v1/admin')
      const tokenKey = isAdminRoute ? 'admin_token' : 'client_token'
      const token = uni.getStorageSync(tokenKey)

      const header = {
        'Content-Type': 'application/json',
        ...(config.headers || {}),
      }
      if (token) {
        header.Authorization = `Bearer ${token}`
      }

      uni.request({
        url: BASE_URL + url,
        method,
        data,
        header,
        timeout: TIMEOUT,
        success: (res) => {
          const { statusCode, data: body } = res
          if (statusCode >= 200 && statusCode < 300) {
            resolve({ data: body, status: statusCode })
          } else if (statusCode === 401) {
            uni.removeStorageSync(tokenKey)
            reject(new Error(`Unauthorized (401): ${url}`))
          } else {
            reject({
              response: { status: statusCode, data: body },
              message: body?.message || `Request failed with status ${statusCode}`,
            })
          }
        },
        fail: (err) => {
          reject({ response: null, message: err.errMsg || 'Network request failed' })
        },
      })
    })
  }

  instance.get = (url, data, config) => request('GET', url, data, config)
  instance.post = (url, data, config) => request('POST', url, data, config)
  instance.put = (url, data, config) => request('PUT', url, data, config)
  instance.delete = (url, data, config) => request('DELETE', url, data, config)
  instance.patch = (url, data, config) => request('PATCH', url, data, config)
  instance.request = request

  return instance
}

const request = createRequest()

export default request
export { request }
