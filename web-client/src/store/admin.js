import { defineStore } from 'pinia'

// 注意：此文件已不再引用 web-admin 的 API 模块。
// web-client 的管理后台功能统一在 web-admin 项目中实现。
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
      // web-client 不承载管理后台登录，此方法保留但不再调用后端
      throw new Error('web-client 不支持管理后台登录，请使用 web-admin')
    },

    async logout() {
      this.token = ''
      this.userInfo = null
      uni.removeStorageSync('admin_token')
    },
  },
})
