import { defineStore } from 'pinia'
import { getState, loginWithDemoAccount } from '@/api/client/auth'

export const useClientStore = defineStore('client', {
  state: () => ({
    token: uni.getStorageSync('client_token') || '',
    userInfo: uni.getStorageSync('client_user') || null,
    orders: [],
    services: [],
    news: [],
    newsCategories: [],
    loginSheetVisible: false,
    pendingActionText: '继续使用该服务',
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.userInfo?.role === 'admin',
    activeOrders: (state) => state.orders.filter((order) => order.status !== 'completed'),
    completedOrders: (state) => state.orders.filter((order) => order.status === 'completed'),
  },

  actions: {
    // setToken 统一维护本地 Token。
    // 实现步骤：
    // 1. 更新 Pinia 内存状态。
    // 2. 有值时写入本地缓存，刷新页面仍保持登录。
    // 3. 无值时清理缓存，避免残留无效凭证。
    setToken(token) {
      this.token = token
      if (token) {
        uni.setStorageSync('client_token', token)
      } else {
        uni.removeStorageSync('client_token')
      }
    },

    // setUserInfo 统一维护用户资料。
    // 实现步骤：
    // 1. 更新用户对象。
    // 2. 持久化到本地缓存，适配 H5 刷新和小程序冷启动。
    // 3. 用户为空时清理缓存，防止个人中心展示旧资料。
    setUserInfo(userInfo) {
      this.userInfo = userInfo
      if (userInfo) {
        uni.setStorageSync('client_user', userInfo)
      } else {
        uni.removeStorageSync('client_user')
      }
    },

    setOrders(orders) {
      this.orders = Array.isArray(orders) ? orders : []
    },

    setServices(services) {
      this.services = Array.isArray(services) ? services : []
    },

    setNews(payload) {
      this.news = Array.isArray(payload?.news) ? payload.news : []
      this.newsCategories = Array.isArray(payload?.newsCategories) ? payload.newsCategories : []
    },

    addOrder(order) {
      this.orders.unshift(order)
    },

    updateOrder(orderId, updates) {
      const idx = this.orders.findIndex((order) => order.id === orderId)
      if (idx !== -1) {
        this.orders[idx] = { ...this.orders[idx], ...updates }
      }
    },

    // initMockState 初始化演示版聚合数据。
    // 实现步骤：
    // 1. 从 api/client/auth 读取 Mock 聚合状态。
    // 2. 将服务、攻略和订单同步进 store。
    // 3. 如果已经有 Token 但没有用户资料，则用 Mock 用户补齐展示。
    async initMockState() {
      const state = await getState()
      this.setOrders(state.orders)
      this.setServices(state.services)
      this.setNews({ news: state.news, newsCategories: state.newsCategories })
      if (this.isLoggedIn && !this.userInfo) {
        this.setUserInfo(state.user)
      }
      return state
    },

    // openLoginSheet 打开底部登录弹窗。
    // 实现步骤：
    // 1. 记录触发登录的业务动作，方便弹窗文案解释。
    // 2. 切换显示状态，由页面组件统一渲染。
    openLoginSheet(actionText = '继续使用该服务') {
      this.pendingActionText = actionText
      this.loginSheetVisible = true
    },

    closeLoginSheet() {
      this.loginSheetVisible = false
    },

    // ensureLogin 执行业务入口鉴权。
    // 实现步骤：
    // 1. 已登录时直接返回 true，业务继续执行。
    // 2. 未登录时打开底部授权弹窗。
    // 3. 返回 false，让调用方停止跳转或下单动作。
    ensureLogin(actionText = '继续使用该服务') {
      if (this.isLoggedIn) return true
      this.openLoginSheet(actionText)
      return false
    },

    // loginByDemo 执行演示版一键登录。
    // 实现步骤：
    // 1. 调用 Mock 登录 API 获取 token 和用户。
    // 2. 写入本地状态并关闭弹窗。
    // 3. 给用户展示成功提示，证明登录链路已经打通。
    async loginByDemo() {
      const data = await loginWithDemoAccount()
      this.setToken(data.token)
      this.setUserInfo(data.user)
      this.closeLoginSheet()
      uni.showToast({ title: '登录成功', icon: 'success' })
      return data
    },

    logout() {
      this.token = ''
      this.userInfo = null
      this.orders = []
      uni.removeStorageSync('client_token')
      uni.removeStorageSync('client_user')
    },
  },
})
