import request from '../request'

export function getDashboardStats() {
  return request.get('/v1/admin/dashboard/stats').then((res) => res.data)
}
