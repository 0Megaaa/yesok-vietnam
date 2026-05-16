import request from '../request'

export function loginWithWeChat(code) {
  return request.post('/v1/client/auth/wechat', { code }).then((res) => res.data)
}
