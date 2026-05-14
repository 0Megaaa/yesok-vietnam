const BASE_URL = '/api'
const TIMEOUT = 10000

// ── 核心请求封装 ──────────────────────────────────────────────
function createRequest() {
  const instance = {}

  function request(method, url, data, config = {}) {
    return new Promise((resolve, reject) => {
      // ── 请求拦截：注入 Token ───────────────────────────────
      const isAdminRoute = url && url.startsWith('/v1/admin')
      const tokenKey = isAdminRoute ? 'admin_token' : 'client_token'
      const token = uni.getStorageSync(tokenKey)

      const header = {
        'Content-Type': 'application/json',
        ...(config.headers || {}),
      }
      if (token) {
        header['Authorization'] = `Bearer ${token}`
      }

      uni.request({
        url: BASE_URL + url,
        method,
        data,
        header,
        timeout: TIMEOUT,
        success: (res) => {
          const { statusCode, data: body } = res

          // ── 响应拦截：统一错误处理 ─────────────────────────
          if (statusCode >= 200 && statusCode < 300) {
            // 正常响应 — 透传，与 axios 行为一致
            resolve({ data: body, status: statusCode })
          } else if (statusCode === 401) {
            // Token 过期或无效，清除对应 token 并跳转
            uni.removeStorageSync(tokenKey)
            if (isAdminRoute) {
              uni.reLaunch({ url: '/pages/admin/login' })
            } else {
              uni.reLaunch({ url: '/pages/home/index' })
            }
            reject(new Error(`Unauthorized (401): ${url}`))
          } else {
            // 其他 HTTP 错误：包装为与 axios 一致的格式
            reject({
              response: { status: statusCode, data: body },
              message: body?.message || `Request failed with status ${statusCode}`,
            })
          }
        },
        fail: (err) => {
          reject({
            response: null,
            message: err.errMsg || 'Network request failed',
          })
        },
      })
    })
  }

  // ── 对外暴露的方法（与 axios-request 签名完全一致）─────────────
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
