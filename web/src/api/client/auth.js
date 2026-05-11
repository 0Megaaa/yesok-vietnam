import request from '../request'

export function loginWithTG(initData) {
  return request.post('/v1/client/auth/tg', { initData }).then((res) => res.data)
}

export function getMe() {
  return request.get('/v1/client/user/me').then((res) => res.data)
}

export function getState() {
  return request.get('/v1/client/state').then((res) => res.data)
}

export function updateState(mutations) {
  return request.put('/v1/client/state', mutations).then((res) => res.data)
}
