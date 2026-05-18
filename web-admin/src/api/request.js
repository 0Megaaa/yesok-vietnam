import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'
const TIMEOUT = 10000

// readStorage 读取本地 token。
// 1.意图 -> 后台与客户端共享一套请求封装，同时保护已有鉴权组件逻辑。
// 2.步骤 -> 优先调用浏览器 localStorage。
// 3.返回 -> 字符串 token 或空字符串。
const readStorage = (key) => {
  if (typeof localStorage !== 'undefined') return localStorage.getItem(key) || ''
  return ''
}

// removeStorage 清理失效 token。
// 1.意图 -> 当后端返回 401 时主动退出，避免页面继续使用过期身份。
// 2.步骤 -> 同步清理浏览器 localStorage。
// 3.返回 -> 无返回值。
const removeStorage = (key) => {
  if (typeof localStorage !== 'undefined') localStorage.removeItem(key)
}

// buildUrl 生成真实后端请求地址。
// 1.意图 -> 将前端原 /v1 路径统一转为 Go 服务 /api/v1 路径。
// 2.步骤 -> 拼接 BASE_URL、业务路径和 GET 查询参数。
// 3.返回 -> 可直接传给 axios 的完整 URL。
const buildUrl = (url, data, method) => {
  const query =
    method === 'GET' && data && Object.keys(data).length
      ? `?${new URLSearchParams(data.params || data).toString()}`
      : ''
  return `${BASE_URL}${url}${query}`
}

// normalizeErrorMessage 统一错误文案。
// 1.意图 -> 将后端 error/message/detail 转为后台可读中文提示。
// 2.步骤 -> 优先读取结构化错误，其次读取原始网络错误。
// 3.返回 -> 错误消息字符串。
const normalizeErrorMessage = (body, fallback) =>
  body?.message || body?.error || body?.detail || fallback || '网络请求失败'

// createRequest 创建真实后端请求实例。
// 1.意图 -> 彻底关闭默认 Mock 拦截，让 B/C 端都走 Go API 与数据库。
// 2.步骤 -> 注入 admin_token/client_token，使用 axios 发送请求。
// 3.返回 -> get/post/put/delete/patch 标准方法集合。
function createRequest() {
  const instance = {}

  const http = axios.create({
    baseURL: BASE_URL,
    timeout: TIMEOUT,
    headers: { 'Content-Type': 'application/json' },
  })

  // 请求拦截器：注入 Bearer token
  http.interceptors.request.use(
    (config) => {
      const isAdminRoute = config.url && config.url.startsWith('/v1/admin')
      const tokenKey = isAdminRoute ? 'admin_token' : 'client_token'
      const token = readStorage(tokenKey)
      if (token) config.headers.Authorization = `Bearer ${token}`
      // 规范化 GET 参数
      if (config.method === 'get' && config.params && Object.keys(config.params).length) {
        config.url = buildUrl(config.url, { params: config.params }, 'GET')
        config.params = undefined
      }
      return config
    },
    (error) => Promise.reject(error)
  )

  // 响应拦截器：统一错误处理
  http.interceptors.response.use(
    (response) => ({ data: response.data, status: response.status }),
    (error) => {
      if (error.response) {
        const status = error.response.status
        const responseBody = error.response.data
        if (status === 401) {
          const isAdminRoute = error.config?.url?.startsWith('/v1/admin')
          removeStorage(isAdminRoute ? 'admin_token' : 'client_token')
        }
        return Promise.reject({
          response: { status, data: responseBody },
          message: normalizeErrorMessage(responseBody, `HTTP ${status}`),
        })
      }
      return Promise.reject({
        response: null,
        message: error.message || '网络请求失败',
      })
    }
  )

  function request(method, url, data = {}, config = {}) {
    return http({
      method,
      url,
      data: method === 'GET' ? undefined : data,
      ...config,
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
