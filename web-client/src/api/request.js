/**
 * 统一请求模块（生产级）
 * - 微信小程序 / uni-app 端：uni.request
 * - Web 端：axios（按需动态加载，绝不进入 uni-app 产物）
 *
 * 所有 IP/域名通过 VITE_API_BASE_URL 环境变量注入，源码中无任何硬编码。
 */

// ─── 环境检测（模块加载时求值，打包工具可识别 dead-code branch）───────────────────

const BASE_URL = import.meta.env.VITE_API_BASE_URL;
const isUniApp = typeof uni !== 'undefined' && typeof uni.request === 'function';

// ─── BASE_URL 兜底校验 ─────────────────────────────────────────────────────────

if (!BASE_URL) {
  if (isUniApp && uni.showModal) {
    uni.showModal({ title: '配置缺失', content: 'API 配置缺失：请在 .env 中设置 VITE_API_BASE_URL', showCancel: false });
  }
  throw new Error('缺少 VITE_API_BASE_URL 环境变量！');
}

// ─── 顶层分支：uni-app 与 Web 逻辑完全分离，确保 bundler tree-shaking 生效 ─────

if (isUniApp) {
  // ══════════════════════════════════════════════════════════════════════════
  //  uni-app / 微信小程序 分支
  //  此分支内不引用任何 Web 端模块，打包工具可将 axios 完全剔除出产物
  // ══════════════════════════════════════════════════════════════════════════

  const TIMEOUT = 10000;

  const readStorage = (key) =>
    uni.getStorageSync ? (uni.getStorageSync(key) || '') : '';

  const removeStorage = (key) => {
    uni.removeStorageSync?.({ key });
  };

  const normalizeError = (err) => ({
    status: err?.response?.status ?? null,
    message: err?.response?.data?.message || err?.response?.data?.error || err?.response?.data?.detail || err?.message || '网络请求失败',
    data: err?.response?.data ?? null,
  });

  const buildUrl = (url, data, method) => {
    if (method === 'GET' && data && typeof data === 'object') {
      const params = data.params || data;
      const parts = [];
      for (const key in params) {
        if (Object.prototype.hasOwnProperty.call(params, key) && params[key] != null) {
          parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`);
        }
      }
      if (parts.length) return `${BASE_URL}${url}?${parts.join('&')}`;
    }
    return `${BASE_URL}${url}`;
  };

  const wrapRequest = (options) =>
    new Promise((resolve, reject) => {
      uni.request({
        ...options,
        success: (res) => {
          if (res.statusCode >= 200 && res.statusCode < 300) {
            resolve({ data: res.data, status: res.statusCode });
          } else {
            reject({ response: { status: res.statusCode, data: res.data } });
          }
        },
        fail: (err) => reject({ response: null, message: err?.errMsg || '网络请求失败' }),
      });
    });

  function createRequest() {
    const instance = {};

    function request(method, url, data, config) {
      // 实时读取 token，不缓存 header
      const isAdminRoute = url.startsWith('/v1/admin');
      const tokenKey = isAdminRoute ? 'admin_token' : 'client_token';
      const token = readStorage(tokenKey);

      const header = { 'Content-Type': 'application/json', ...((config && config.headers) || {}) };
      if (token) header.Authorization = `Bearer ${token}`;

      const reqOptions = {
        url: buildUrl(url, data || {}, method),
        method,
        header,
        timeout: (config && config.timeout) || TIMEOUT,
      };
      // 【防拦截】：GET 请求不传 body data
      if (method !== 'GET' && data) reqOptions.data = data;

      return wrapRequest(reqOptions).catch((err) => {
        const normalized = normalizeError(err);
        if (normalized.status === 401) removeStorage(tokenKey);
        throw normalized;
      });
    }

    instance.get = (url, data, config) => request('GET', url, data, config);
    instance.post = (url, data, config) => request('POST', url, data, config);
    instance.put = (url, data, config) => request('PUT', url, data, config);
    instance.delete = (url, data, config) => request('DELETE', url, data, config);
    instance.patch = (url, data, config) => request('PATCH', url, data, config);
    instance.request = request;
    return instance;
  }

  module.exports = createRequest();

} else {
  // ══════════════════════════════════════════════════════════════════════════
  //  Web 端分支（浏览器 / H5）
  //  axios 在首次调用时动态加载，不参与 uni-app 构建
  // ══════════════════════════════════════════════════════════════════════════

  const TIMEOUT = 10000;

  let service = null;

  const getService = () => {
    if (service) return Promise.resolve(service);
    return import('axios').then(({ default: axios }) => {
      service = axios.create({ baseURL: BASE_URL, timeout: TIMEOUT });
      service.interceptors.response.use(
        (res) => res,
        (err) => Promise.reject(err)
      );
      return service;
    });
  };

  const readStorage = (key) =>
    typeof localStorage !== 'undefined' ? (localStorage.getItem(key) || '') : '';

  const removeStorage = (key) => {
    if (typeof localStorage !== 'undefined') localStorage.removeItem(key);
  };

  const normalizeError = (err) => ({
    status: err?.response?.status ?? null,
    message: err?.response?.data?.message || err?.response?.data?.error || err?.response?.data?.detail || err?.message || '网络请求失败',
    data: err?.response?.data ?? null,
  });

  function createRequest() {
    const instance = {};

    function request(method, url, data, config) {
      // 实时读取 token，不缓存 header
      const isAdminRoute = url.startsWith('/v1/admin');
      const tokenKey = isAdminRoute ? 'admin_token' : 'client_token';
      const token = readStorage(tokenKey);

      const header = { 'Content-Type': 'application/json', ...((config && config.headers) || {}) };
      if (token) header.Authorization = `Bearer ${token}`;

      return getService().then((svc) =>
        svc
          .request({
            method,
            url,
            data,
            headers: header,
            params: method === 'GET' ? (data && (data.params || data)) : undefined,
          })
          .then((res) => {
            if (res.status === 401) removeStorage(tokenKey);
            return { data: res.data, status: res.status };
          })
          .catch((err) => {
            const normalized = normalizeError(err);
            if (normalized.status === 401) removeStorage(tokenKey);
            throw normalized;
          })
      );
    }

    instance.get = (url, data, config) => request('GET', url, data, config);
    instance.post = (url, data, config) => request('POST', url, data, config);
    instance.put = (url, data, config) => request('PUT', url, data, config);
    instance.delete = (url, data, config) => request('DELETE', url, data, config);
    instance.patch = (url, data, config) => request('PATCH', url, data, config);
    instance.request = request;
    return instance;
  }

  module.exports = createRequest();
}
