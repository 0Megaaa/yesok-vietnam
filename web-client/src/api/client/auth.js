import request from '../request'
import { MOCK_USER } from '../mockData'

// loginWithTG 是 Telegram Mini App 登录占位函数。
// 实现步骤：
// 1. 保留 initData 入参，日后接入 Telegram WebApp initData 校验时无需改页面。
// 2. 当前演示版不联调真实后端，统一通过 Mock 请求返回演示用户。
// 3. 后续真实接入时，后端必须校验 hash、auth_date 和机器人 Token，禁止前端自行信任用户数据。
export function loginWithTG(initData = '') {
  return request.post('/v1/client/auth/tg', { initData }).then((res) => res.data)
}

// loginWithDemoAccount 执行底部登录弹窗中的演示登录。
// 实现步骤：
// 1. 不要求用户填写手机号，降低演示阻力。
// 2. 写入 Mock Token 与 Mock 用户。
// 3. 返回统一结构，供 store 更新登录态。
export function loginWithDemoAccount() {
  return Promise.resolve({
    token: 'mock-client-token-demo',
    user: MOCK_USER,
    loginType: 'demo',
  })
}

// getMe 获取当前用户信息。
// 实现步骤：
// 1. 通过统一 request 调用客户端 Mock 路由。
// 2. 返回与真实接口一致的用户对象。
export function getMe() {
  return request.get('/v1/client/user/me').then((res) => res.data)
}

// getState 获取客户端页面聚合状态。
// 实现步骤：
// 1. 聚合用户、订单、服务和攻略数据。
// 2. 页面只消费状态，不关心数据来自 Mock 还是真实后端。
export function getState() {
  return request.get('/v1/client/state').then((res) => res.data)
}

// updateState 更新客户端状态的占位方法。
// 实现步骤：
// 1. 保留 mutations 入参，日后可映射到真实后端 PATCH/PUT。
// 2. 当前演示版返回提交内容，避免出现网络依赖。
export function updateState(mutations) {
  return request.put('/v1/client/state', mutations).then((res) => res.data)
}
