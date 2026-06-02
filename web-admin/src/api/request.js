import axios from 'axios'
import { ElMessage } from 'element-plus'

const BASE_URL = (() => {
  const val = import.meta.env.VITE_API_BASE_URL
  if (!val) {
    throw new Error('缺少 VITE_API_BASE_URL 环境变量，请配置为 http://host:7625/api 或 https://你的域名/api')
  }
  return val
})()

// 启动校验：建议以 /api 结尾，避免路径拼接错误
if (!/\/api\/?$/.test(BASE_URL)) {
  console.warn('[request] VITE_API_BASE_URL 建议以 /api 结尾，当前为：', BASE_URL)
}

// ORIGIN_URL 用于拼接静态资源 /material 和 /uploads，避免带 /api 前缀
// 例如：BASE_URL = http://127.0.0.1:7625/api → ORIGIN_URL = http://127.0.0.1:7625
const ORIGIN_URL = (() => {
  const raw = String(BASE_URL || '').replace(/\/+$/, '')
  return raw.replace(/\/api$/, '')
})()

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

// ─── 模块级状态：防止多个 401 同时触发时重复弹窗/重复跳转 ────────────────────────────
let isHandlingUnauthorized = false

// clearAuthStorage 清理对应端的 token 及用户信息。
const clearAuthStorage = (isAdminRoute) => {
  if (typeof localStorage === 'undefined') return
  if (isAdminRoute) {
    removeStorage('admin_token')
    removeStorage('admin_user')
    removeStorage('admin_token_expire')
  } else {
    removeStorage('client_token')
    removeStorage('client_user')
    removeStorage('client_token_expire')
  }
}

// handleUnauthorized 统一处理所有 401 响应：清理 token、提示用户、静默跳转。
// 1.意图 -> 无论是 TOKEN_EXPIRED / TOKEN_INVALID / TOKEN_KICKED / TOKEN_MISSING / ACCOUNT_DISABLED，
//           还是其他后端返回的 401，统一拦截并引导用户重新登录。
// 2.步骤 -> 判断 B/C 端路由，清理对应 token，ElMessage 提示一次后静默跳转。
// 3.返回 -> 无。
function handleUnauthorized(error) {
  if (isHandlingUnauthorized) return
  isHandlingUnauthorized = true

  const url = error.config?.url || ''
  const isAdminRoute = url.includes('/v1/admin')

  clearAuthStorage(isAdminRoute)

  const message =
    error.response?.data?.error ||
    error.response?.data?.message ||
    '登录状态已失效，请重新登录'

  ElMessage.closeAll()
  ElMessage.error({ message, duration: 0 })

  // 0.5s 后静默跳转，不依赖用户点击
  if (isAdminRoute) {
    window.setTimeout(() => {
      window.location.replace(`${window.location.origin}/admin/`)
    }, 500)
  } else {
    window.setTimeout(() => {
      window.location.replace('/')
    }, 500)
  }
}

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
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    },
    (error) => Promise.reject(error)
  )

  // 响应拦截器：统一错误处理
  http.interceptors.response.use(
    (response) => {
      return { data: response.data, status: response.status }
    },
    (error) => {
      const status = error.response?.status
      const responseBody = error.response?.data
      if (status === 401) {
        handleUnauthorized(error)
      }
      return Promise.reject({
        response: { status, data: responseBody },
        message: normalizeErrorMessage(
          responseBody,
          status === 401 ? '登录状态已失效，请重新登录' : `HTTP ${status || 'ERROR'}`
        ),
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
export { request, ORIGIN_URL, BASE_URL }
