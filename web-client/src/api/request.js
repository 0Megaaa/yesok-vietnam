/**
 * 统一请求模块（生产级）
 * - 微信小程序 / uni-app 端：使用 uni.request
 * - Web 端：按需动态加载 axios（打包时被 tree-shaking 剔除，不污染小程序）
 */

export const BASE_URL = import.meta.env.VITE_API_BASE_URL;

if (!BASE_URL) {
  throw new Error('前端构建配置错误：缺少 VITE_API_BASE_URL 环境变量！');
}

// ORIGIN_URL 用于拼接静态资源 /material，避免带 /api 前缀
// 例如：BASE_URL = http://127.0.0.1:7625/api → ORIGIN_URL = http://127.0.0.1:7625
export const ORIGIN_URL = (() => {
  const raw = String(BASE_URL || '').replace(/\/+$/, '');
  return raw.replace(/\/api$/, '');
})();

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

// --- 模块级状态：防止多个 401 同时触发时重复提示/重复跳转 ---
let isRedirecting = false;

// handleUnauthorized 统一处理 401：清理 token、提示用户。
// B 端走 /admin/login 跳转；C 端清 token 后引导重新进入授权流程。
function handleUnauthorized(error, isAdminRoute) {
  if (isRedirecting) return
  isRedirecting = true

  const message = '登录状态已失效，请重新登录'

  // 清理对应端 token 及本地用户信息
  if (isAdminRoute) {
    removeStorage('admin_token')
    removeStorage('admin_user')
    removeStorage('admin_token_expire')
  } else {
    removeStorage('client_token')
    removeStorage('client_user')
  }

  if (isUniApp) {
    // 小程序/uni-app 环境：弹窗提示后跳转
    uni.showToast({ title: message, icon: 'none', duration: 2000 })
    setTimeout(() => {
      if (isAdminRoute) {
        uni.redirectTo({ url: '/pages/login/index' })
      } else {
        // C 端无独立登录页，引导回首页重新授权
        uni.switchTab({ url: '/pages/home/index' })
      }
      isRedirecting = false
    }, 2000)
  } else {
    // Web 环境
    if (typeof window !== 'undefined') {
      if (typeof window.$message !== 'undefined') {
        window.$message.error(message)
      } else if (window.alert) {
        window.alert(message)
      }
      if (isAdminRoute) {
        window.location.href = '/admin/login'
      } else {
        window.location.href = '/'
      }
    }
  }
}

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
            handleUnauthorized(res, isAdminRoute)
            reject({ status: 401, data: res.data, message: '登录状态已失效，请重新登录' })
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
      const status = err.response?.status
      if (status === 401) {
        handleUnauthorized(err, isAdminRoute)
      }
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