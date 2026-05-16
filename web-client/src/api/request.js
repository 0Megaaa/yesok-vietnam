const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'
const TIMEOUT = 10000

// safeUni 获取跨端可用的 uni 对象。
// 1.意图 -> 保持 AuthPopup、后台登录与 H5 预览在不同运行环境下不崩溃。
// 2.步骤 -> 判断全局 uni 是否存在，存在则返回，不存在则返回 null。
// 3.返回 -> uni 对象或 null。
const safeUni = () => (typeof uni !== 'undefined' ? uni : null)

// readStorage 读取本地 token。
// 1.意图 -> 后台与客户端共享一套请求封装，同时保护已有鉴权组件逻辑。
// 2.步骤 -> 优先调用 uni.getStorageSync，浏览器环境降级 localStorage。
// 3.返回 -> 字符串 token 或空字符串。
const readStorage = (key) => {
  const uniApi = safeUni()
  if (uniApi?.getStorageSync) return uniApi.getStorageSync(key) || ''
  if (typeof localStorage !== 'undefined') return localStorage.getItem(key) || ''
  return ''
}

// removeStorage 清理失效 token。
// 1.意图 -> 当后端返回 401 时主动退出，避免页面继续使用过期身份。
// 2.步骤 -> 同步清理 uni storage 与浏览器 localStorage。
// 3.返回 -> 无返回值。
const removeStorage = (key) => {
  const uniApi = safeUni()
  if (uniApi?.removeStorageSync) uniApi.removeStorageSync(key)
  if (typeof localStorage !== 'undefined') localStorage.removeItem(key)
}

// buildUrl 生成真实后端请求地址。
// 1.意图 -> 将前端原 /v1 路径统一转为 Go 服务 /api/v1 路径。
// 2.步骤 -> 拼接 BASE_URL、业务路径和 GET 查询参数。
// 3.返回 -> 可直接传给 uni.request 或 fetch 的完整 URL。
const buildUrl = (url, data, method) => {
  const query = method === 'GET' && data && Object.keys(data).length
    ? `?${new URLSearchParams(data.params || data).toString()}`
    : ''
  return `${BASE_URL}${url}${query}`
}

// normalizeErrorMessage 统一错误文案。
// 1.意图 -> 将后端 error/message/detail 转为后台可读中文提示。
// 2.步骤 -> 优先读取结构化错误，其次读取原始网络错误。
// 3.返回 -> 错误消息字符串。
const normalizeErrorMessage = (body, fallback) => body?.message || body?.error || body?.detail || fallback || '网络请求失败'

// createRequest 创建真实后端请求实例。
// 1.意图 -> 彻底关闭默认 Mock 拦截，让 B/C 端都走 Go API 与数据库。
// 2.步骤 -> 注入 admin_token/client_token，优先使用 uni.request，H5 环境降级 fetch。
// 3.返回 -> get/post/put/delete/patch 标准方法集合。
function createRequest() {
  const instance = {}

  function request(method, url, data = {}, config = {}) {
    const isAdminRoute = url && url.startsWith('/v1/admin')
    const tokenKey = isAdminRoute ? 'admin_token' : 'client_token'
    const token = readStorage(tokenKey)
    const header = {
      'Content-Type': 'application/json',
      ...(config.headers || {}),
    }
    if (token) header.Authorization = `Bearer ${token}`
    const fullUrl = buildUrl(url, data, method)
    const body = method === 'GET' ? undefined : data
    const uniApi = safeUni()

    if (uniApi?.request) {
      return new Promise((resolve, reject) => {
        uniApi.request({
          url: fullUrl,
          method,
          data: body,
          header,
          timeout: TIMEOUT,
          success: (res) => {
            const statusCode = res.statusCode
            const responseBody = res.data
            if (statusCode >= 200 && statusCode < 300) {
              resolve({ data: responseBody, status: statusCode })
              return
            }
            if (statusCode === 401) removeStorage(tokenKey)
            reject({ response: { status: statusCode, data: responseBody }, message: normalizeErrorMessage(responseBody, `HTTP ${statusCode}`) })
          },
          fail: (err) => reject({ response: null, message: err?.errMsg || '网络请求失败' }),
        })
      })
    }

    return fetch(fullUrl, {
      method,
      headers: header,
      body: body === undefined ? undefined : JSON.stringify(body),
    }).then(async (res) => {
      const text = await res.text()
      const responseBody = text ? JSON.parse(text) : {}
      if (res.ok) return { data: responseBody, status: res.status }
      if (res.status === 401) removeStorage(tokenKey)
      throw { response: { status: res.status, data: responseBody }, message: normalizeErrorMessage(responseBody, `HTTP ${res.status}`) }
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
