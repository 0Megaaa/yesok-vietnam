/**
 * 统一请求模块（生产级）
 * - 微信小程序 / uni-app 端：使用 uni.request
 * - Web 端：按需动态加载 axios（打包时被 tree-shaking 剔除，不污染小程序）
 */

const BASE_URL = import.meta.env.VITE_API_BASE_URL;

if (!BASE_URL) {
  throw new Error('前端构建配置错误：缺少 VITE_API_BASE_URL 环境变量！');
}

const TIMEOUT = 10000;
const isUniApp = typeof uni !== 'undefined' && typeof uni.request === 'function';

// --- Token 辅助函数 ---
const readStorage = (key) => {
  if (isUniApp) return uni.getStorageSync(key) || '';
  return localStorage.getItem(key) || '';
};

const removeStorage = (key) => {
  if (isUniApp) uni.removeStorageSync(key);
  else localStorage.removeItem(key);
};

// --- 请求适配逻辑 ---
async function request(method, url, data, config = {}) {
  const isAdminRoute = url.startsWith('/v1/admin');
  const tokenKey = isAdminRoute ? 'admin_token' : 'client_token';
  const token = readStorage(tokenKey);
  const header = { 'Content-Type': 'application/json', ...((config && config.headers) || {}) };
  if (token) header.Authorization = `Bearer ${token}`;

  if (isUniApp) {
    // 微信小程序/uni-app 原生请求
    return new Promise((resolve, reject) => {
      uni.request({
        url: BASE_URL + url,
        method: method,
        data: method === 'GET' ? (data?.params || data) : data,
        header: header,
        timeout: TIMEOUT,
        success: (res) => {
          if (res.statusCode === 401) {
            removeStorage(tokenKey)
            reject(res.data || { message: '登录已失效，请重新登录' })
            return
          }
          if (res.statusCode < 200 || res.statusCode >= 300) {
            reject(res.data || { message: '请求失败' })
            return
          }
          resolve({ data: res.data, status: res.statusCode })
        },
        fail: (err) => reject({ message: err.errMsg || '网络请求失败' })
      });
    });
  } else {
    // Web 环境：延迟加载 axios
    const axios = (await import('axios')).default;
    try {
      const res = await axios({
        baseURL: BASE_URL,
        method,
        url,
        data: method !== 'GET' ? data : undefined,
        params: method === 'GET' ? (data?.params || data) : undefined,
        headers: header,
        timeout: TIMEOUT
      });
      return { data: res.data, status: res.status };
    } catch (err) {
      if (err.response?.status === 401) removeStorage(tokenKey)
      throw err
    }
  }
}

// 统一导出接口
export const get = (url, data, config) => request('GET', url, data, config);
export const post = (url, data, config) => request('POST', url, data, config);
export const put = (url, data, config) => request('PUT', url, data, config);
export const del = (url, data, config) => request('DELETE', url, data, config);
export const patch = (url, data, config) => request('PATCH', url, data, config);

export default request;