import { get, put } from '@/api/request'

// getClientMe 获取当前 C 端用户资料。
export function getClientMe() {
  return get('/v1/client/user/me').then((res) => res.data || res)
}

// updateClientProfile 更新 C 端用户资料（昵称、头像）。
export function updateClientProfile(data) {
  return put('/v1/client/user/me', data).then((res) => res.data || res)
}

// getClientOrders 获取当前用户的订单列表，支持 status=all|active|completed。
export function getClientOrders(params = {}) {
  return get('/v1/client/orders', params).then((res) => res.data || res)
}
