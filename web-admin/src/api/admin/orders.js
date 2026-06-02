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

// performOrderAction 调用 POST /api/v1/admin/orders/:id/action
// 专用于工作流动作推进，与 updateOrder (PUT) 语义区分
// 注意：禁止调用 audit_approve / audit_reject，这些必须走 auditOrder
export function performOrderAction(id, data) {
  if (data?.action_name === 'audit_approve' || data?.action_name === 'audit_reject' || data?.action_name === 'audit_rejected') {
    return Promise.reject(new Error('审核动作必须调用 auditOrder，不允许走普通工作流动作'))
  }
  return request.post(`/v1/admin/orders/${id}/action`, data).then((res) => res.data)
}

// auditOrder 调用 POST /api/v1/admin/orders/:id/audit
// 专用于审核动作（审核通过 / 审核不通过）
export function auditOrder(orderId, data) {
  return request.post(`/v1/admin/orders/${orderId}/audit`, data).then((res) => res.data)
}

// uploadAdminOrderMaterial 上传订单资料图片
// 调用 POST /api/v1/admin/orders/:id/materials/upload
// 文件保存到 /material/{service_code}/ 目录，返回 /material/... 相对路径
export function uploadAdminOrderMaterial(orderId, file, fieldKey) {
  const fd = new FormData()
  fd.append('file', file)
  if (fieldKey) {
    fd.append('field_key', fieldKey)
  }
  return request.post(`/v1/admin/orders/${orderId}/materials/upload`, fd, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
