const BASE_URL = (import.meta.env.VITE_API_URL) || 'http://127.0.0.1:7625/api'
const TIMEOUT = 10000

const safeUni = () => (typeof uni !== 'undefined' ? uni : null)

const readStorage = (key) => {
  const uniApi = safeUni()
  if (uniApi?.getStorageSync) return uniApi.getStorageSync(key) || ''
  if (typeof localStorage !== 'undefined') return localStorage.getItem(key) || ''
  return ''
}

const removeStorage = (key) => {
  const uniApi = safeUni()
  if (uniApi?.removeStorageSync) uniApi.removeStorageSync(key)
  if (typeof localStorage !== 'undefined') localStorage.removeItem(key)
}

const buildUrl = (url, data, method) => {
  let queryStr = '';
  // 极度安全的参数拼接，绝不使用小程序的死穴 URLSearchParams
  if (method === 'GET' && data && typeof data === 'object') {
    const params = data.params || data;
    const queryParts = [];
    for (const key in params) {
      if (Object.prototype.hasOwnProperty.call(params, key) && params[key] !== undefined && params[key] !== null) {
        queryParts.push(`${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`);
      }
    }
    if (queryParts.length > 0) {
      queryStr = `?${queryParts.join('&')}`;
    }
  }
  return `${BASE_URL}${url}${queryStr}`;
}

const normalizeErrorMessage = (body, fallback) => body?.message || body?.error || body?.detail || fallback || '网络请求失败'

function createRequest() {
  const instance = {}

  function request(method, url, data, config = {}) {
    const isAdminRoute = url && url.startsWith('/v1/admin')
    const tokenKey = isAdminRoute ? 'admin_token' : 'client_token'
    const token = readStorage(tokenKey)
    const header = {
      'Content-Type': 'application/json',
      ...(config.headers || {}),
    }
    if (token) header.Authorization = `Bearer ${token}`

    const fullUrl = buildUrl(url, data || {}, method)
    const uniApi = safeUni()

    if (uniApi?.request) {
      return new Promise((resolve, reject) => {
        const reqOptions = {
          url: fullUrl,
          method,
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
        }

        // 【核心防拦截机制】：小程序环境中，如果是 GET 请求，绝不允许向底层传递 data 字段！
        if (method !== 'GET' && data) {
          reqOptions.data = data;
        }

        uniApi.request(reqOptions)
      })
    }

    return fetch(fullUrl, {
      method,
      headers: header,
      body: method === 'GET' ? null : JSON.stringify(data || {}),
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
