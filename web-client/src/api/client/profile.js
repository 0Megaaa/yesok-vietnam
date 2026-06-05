import { get, put, BASE_URL } from '@/api/request'

const readStorage = (key) => {
  try {
    if (typeof uni !== 'undefined' && uni?.getStorageSync) {
      return uni.getStorageSync(key) || ''
    }
    if (typeof localStorage !== 'undefined') {
      return localStorage.getItem(key) || ''
    }
  } catch (error) {
    console.warn('[profile api] read storage failed:', error)
  }
  return ''
}

// getClientMe 获取当前 C 端用户资料。
export function getClientMe() {
  return get('/v1/client/user/me').then((res) => res.data || res)
}

// updateClientProfile 更新 C 端用户资料（昵称、头像）。
export function updateClientProfile(data) {
  return put('/v1/client/user/me', data).then((res) => res.data || res)
}

export function uploadClientAvatar(filePath) {
  const token = readStorage('client_token')
  const base = String(BASE_URL || '').replace(/\/+$/, '')

  if (!filePath) {
    return Promise.reject(new Error('请选择头像图片'))
  }

  if (!base || !/^https?:\/\//.test(base)) {
    return Promise.reject(new Error('上传地址配置错误'))
  }

  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: `${base}/v1/client/user/avatar/upload`,
      filePath,
      name: 'file',
      header: {
        Authorization: token ? `Bearer ${token}` : '',
      },
      success: (res) => {
        let body = null
        try {
          body = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
        } catch {
          reject(new Error('头像上传响应解析失败'))
          return
        }

        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(body?.data || body)
          return
        }

        reject(new Error(body?.error || body?.message || '头像上传失败'))
      },
      fail: (error) => {
        reject(new Error(error?.errMsg || '头像上传失败'))
      },
    })
  })
}

// getClientOrders 获取当前用户的订单列表，支持 status=all|active|completed。
export function getClientOrders(params = {}) {
  return get('/v1/client/orders', params).then((res) => res.data || res)
}
