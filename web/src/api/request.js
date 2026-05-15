import { MOCK_NEWS, MOCK_NEWS_CATEGORIES, MOCK_ORDERS, MOCK_SERVICES, MOCK_USER } from './mockData'

const BASE_URL = '/api'
const TIMEOUT = 10000
const MOCK_DELAY = 180

// createMockResponse 创建与 axios/uni.request 兼容的 Mock 响应对象。
// 实现步骤：
// 1. 根据请求地址判断客户端业务场景。
// 2. 返回 Promise，模拟真实接口的异步行为。
// 3. 用深拷贝隔离页面修改，避免污染全局 Mock 数据。
function createMockResponse(method, url, data) {
  const cloned = (payload) => JSON.parse(JSON.stringify(payload))

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

  // 复杂逻辑：服务详情根据动态 URL 返回单个服务。
  // 实现步骤：
  // 1. 判断是否命中服务详情路径。
  // 2. 提取最后一个路径片段作为服务 ID。
  // 3. 如果未找到服务，则返回第一个服务作为演示兜底。
  if (url.startsWith('/v1/client/services/')) {
    const serviceId = decodeURIComponent(url.split('/').pop())
    const service = MOCK_SERVICES.find((item) => item.id === serviceId) || MOCK_SERVICES[0]
    return Promise.resolve({ data: cloned(service), status: 200 })
  }

  // 复杂逻辑：咨询与下单在演示版中只生成 Mock 订单，不访问真实后端。
  // 实现步骤：
  // 1. 判断请求方法和 URL 是否为下单动作。
  // 2. 从请求体读取服务信息。
  // 3. 返回符合订单卡片渲染结构的数据。
  if (method === 'POST' && url === '/v1/client/orders') {
    const service = MOCK_SERVICES.find((item) => item.id === data?.serviceId) || MOCK_SERVICES[0]
    return Promise.resolve({
      data: cloned({
        id: `YS${Date.now()}`,
        orderNo: `YS${Date.now()}`,
        serviceName: service.name,
        icon: service.icon,
        status: 'pending',
        sk: 'pending',
        managerName: 'Linh 专属管家',
        price: `${service.price}${service.unit ? ` ${service.unit}` : ''}`,
      }),
      status: 200,
    })
  }

  if (mockRoutes[url]) {
    return Promise.resolve({ data: cloned(mockRoutes[url]), status: 200 })
  }

  return Promise.resolve({ data: { ok: true, url, method }, status: 200 })
}

// createRequest 创建统一请求实例。
// 实现步骤：
// 1. 客户端 /v1/client 请求默认走 Mock，保证演示版不依赖后端。
// 2. 后台 /v1/admin 请求仍保留真实请求能力，避免管理端后续联调路径丢失。
// 3. 保持 get/post/put/delete/patch 签名与旧封装一致，降低改造风险。
function createRequest() {
  const instance = {}

  function request(method, url, data, config = {}) {
    const shouldUseMock = url && url.startsWith('/v1/client')
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
