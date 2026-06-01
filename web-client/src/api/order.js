/**
 * C端订单相关 API
 */
import { BASE_URL } from './request'

const readStorage = (key) => {
  if (typeof uni !== 'undefined' && uni.getStorageSync) {
    return uni.getStorageSync(key) || ''
  }
  return ''
}

const normalizeBaseUrl = (base) => {
  return String(base || '').replace(/\/+$/, '')
}

export const uploadOrderMaterial = (orderId, filePath, fieldKey) => {
  const token = readStorage('client_token')
  const base = normalizeBaseUrl(BASE_URL)

  if (!base || !/^https?:\/\//.test(base)) {
    return Promise.reject(new Error('上传地址配置错误：VITE_API_BASE_URL 必须是完整 http/https 地址'))
  }

  if (!orderId) {
    return Promise.reject(new Error('订单ID不能为空'))
  }

  if (!filePath) {
    return Promise.reject(new Error('请选择要上传的图片'))
  }

  const url = `${base}/v1/client/orders/${orderId}/materials/upload`

  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url,
      filePath,
      name: 'file',
      formData: {
        field_key: fieldKey || '',
      },
      header: {
        Authorization: token ? `Bearer ${token}` : '',
      },
      success: (res) => {
        let data = null
        try {
          data = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
        } catch {
          reject(new Error('上传响应解析失败'))
          return
        }

        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(data?.data || data)
          return
        }

        reject(new Error(data?.error || data?.message || '上传失败'))
      },
      fail: (err) => {
        reject(new Error(err?.errMsg || '上传失败'))
      },
    })
  })
}
