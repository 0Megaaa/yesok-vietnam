import { post } from '../request'

export function loginWithWeChat(code) {
  return post('/v1/client/auth/wechat', { code }).then((res) => res.data)
}
