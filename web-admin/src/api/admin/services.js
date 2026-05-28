import request from '../request'

// 获取服务基础信息（含工作流节点）
export function getServiceDetail(id) {
  return request.get(`/v1/admin/services/${id}`).then((res) => res.data)
}

// 获取指定服务在某个 stage 下的动作列表
// role 默认为 admin，可选 client/both
export function getServiceActions(serviceId, stage, role = 'admin') {
  return request.get(`/v1/admin/services/${serviceId}/actions`, { params: { stage, role } }).then((res) => res.data)
}
