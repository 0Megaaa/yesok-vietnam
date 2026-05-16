import request from '../request'

export function adminLogin(username, password) {
  return request.post('/v1/admin/auth/login', { username, password }).then((res) => res.data)
}

export function adminLogout() {
  return request.post('/v1/admin/auth/logout').then((res) => res.data)
}

export function adminMe() {
  return request.get('/v1/admin/auth/me').then((res) => res.data)
}
