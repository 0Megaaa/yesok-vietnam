import { defineStore } from 'pinia'

export const useClientStore = defineStore('client', {
  state: () => ({
    token: localStorage.getItem('client_token') || '',
    userInfo: null,
    orders: [],
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.userInfo?.role === 'admin',
    activeOrders: (state) => state.orders.filter(o => o.status !== 'completed'),
    completedOrders: (state) => state.orders.filter(o => o.status === 'completed'),
  },

  actions: {
    setToken(token) {
      this.token = token
      if (token) {
        localStorage.setItem('client_token', token)
      } else {
        localStorage.removeItem('client_token')
      }
    },

    setUserInfo(userInfo) {
      this.userInfo = userInfo
    },

    setOrders(orders) {
      this.orders = orders
    },

    addOrder(order) {
      this.orders.unshift(order)
    },

    updateOrder(orderId, updates) {
      const idx = this.orders.findIndex(o => o.id === orderId)
      if (idx !== -1) {
        this.orders[idx] = { ...this.orders[idx], ...updates }
      }
    },

    logout() {
      this.token = ''
      this.userInfo = null
      this.orders = []
      localStorage.removeItem('client_token')
    },
  },
})
