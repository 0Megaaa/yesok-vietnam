import request from './request'

export function loginWithTG(initData) {
  return request.post('/auth/tg', { initData }).then((res) => res.data)
}

export function getUserInfo() {
  return request.get('/user/me').then((res) => res.data)
}
