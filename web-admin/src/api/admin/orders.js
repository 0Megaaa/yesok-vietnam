import request from '../request'

export function getOrderList(params) {
  return request.get('/v1/admin/orders', { params }).then((res) => res.data)
}

export function getOrderDetail(id) {
  return request.get(`/v1/admin/orders/${id}`).then((res) => res.data)
}

export function updateOrder(id, data) {
  return request.put(`/v1/admin/orders/${id}`, data).then((res) => res.data)
}

export function getOrderActions(id) {
  return request.get(`/v1/admin/orders/${id}/actions`).then((res) => res.data)
}

// performOrderAction 调用 POST /api/v1/orders/:id/action
// 专用于工作流动作推进，与 updateOrder (PUT) 语义区分
export function performOrderAction(id, data) {
  return request.post(`/v1/admin/orders/${id}/action`, data).then((res) => res.data)
}
