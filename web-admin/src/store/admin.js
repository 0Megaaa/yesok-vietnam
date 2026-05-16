import { defineStore } from 'pinia'
import { adminLogin, adminLogout } from '@/api/admin/auth'

export const useAdminStore = defineStore('admin', {
  state: () => ({
    token: uni.getStorageSync('admin_token') || '',
    userInfo: null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
  },

  actions: {
    setToken(token) {
      this.token = token
      if (token) {
        uni.setStorageSync('admin_token', token)
      } else {
        uni.removeStorageSync('admin_token')
      }
    },

    setUserInfo(userInfo) {
      this.userInfo = userInfo
    },

    async login(username, password) {
      const data = await adminLogin(username, password)
      this.setToken(data.token)
      this.setUserInfo(data.user)
      return data
    },

    async logout() {
      try {
        await adminLogout()
      } catch {
        // ignore — always clear local state
      }
      this.token = ''
      this.userInfo = null
      uni.removeStorageSync('admin_token')
    },
  },
})
