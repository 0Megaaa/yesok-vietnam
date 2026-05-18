import { defineStore } from 'pinia'
import { getState, loginWithDemoAccount } from '@/api/client/auth'

const readStorage = (key, fallback = null) => {
  try {
    if (typeof localStorage !== 'undefined') {
      const rawValue = localStorage.getItem(key)
      if (!rawValue) return fallback
      try {
        return JSON.parse(rawValue)
      } catch {
        return rawValue
      }
    }
  } catch {
    console.warn('[YesOK Storage Read Failed]', key)
  }
  return fallback
}

const writeStorage = (key, value) => {
  try {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(key, typeof value === 'string' ? value : JSON.stringify(value))
    }
  } catch {
    console.warn('[YesOK Storage Write Failed]', key)
  }
}

const removeStorage = (key) => {
  try {
    if (typeof localStorage !== 'undefined') localStorage.removeItem(key)
  } catch {
    console.warn('[YesOK Storage Remove Failed]', key)
  }
}

const showSafeToast = (options) => {
  console.info('[YesOK Toast]', options?.title || '')
}

export const useClientStore = defineStore('client', {
  state: () => ({
    token: readStorage('client_token', ''),
    userInfo: readStorage('client_user', null),
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
    setToken(token) {
      this.token = token
      if (token) writeStorage('client_token', token)
      else removeStorage('client_token')
    },

    setUserInfo(userInfo) {
      this.userInfo = userInfo
      if (userInfo) writeStorage('client_user', userInfo)
      else removeStorage('client_user')
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

    openLoginSheet(actionText = '继续使用该服务') {
      this.pendingActionText = actionText
      this.loginSheetVisible = true
    },

    closeLoginSheet() {
      this.loginSheetVisible = false
    },

    checkAuth(actionText = '继续使用该服务') {
      if (this.isLoggedIn) return true
      this.openLoginSheet(actionText)
      return false
    },

    ensureLogin(actionText = '继续使用该服务') {
      return this.checkAuth(actionText)
    },

    async loginByDemo() {
      const data = await loginWithDemoAccount()
      this.setToken(data.token)
      this.setUserInfo(data.user)
      this.closeLoginSheet()
      showSafeToast({ title: '登录成功', icon: 'success' })
      return data
    },

    logout() {
      this.token = ''
      this.userInfo = null
      this.orders = []
      removeStorage('client_token')
      removeStorage('client_user')
    },
  },
})
