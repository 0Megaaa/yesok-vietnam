import request from '../request'

export function getDashboardStats() {
  console.log('[API] 正在发起 API 请求：getDashboardStats')
  return request.get('/v1/admin/dashboard/stats')
    .then((res) => {
      console.log('[API] getDashboardStats 成功返回：', res.data)
      return res.data
    })
    .catch((error) => {
      console.error('[API] getDashboardStats 调用报错：', error)
      throw error
    })
}
