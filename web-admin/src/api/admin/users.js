import request from '../request'

export function getUserList(params) {
  return request.get('/v1/admin/users', { params }).then((res) => res.data)
}

export function updateUserRole(id, role) {
  return request.put(`/v1/admin/users/${id}/role`, { role }).then((res) => res.data)
}

export function deleteUser(id) {
  return request.delete(`/v1/admin/users/${id}`).then((res) => res.data)
}
