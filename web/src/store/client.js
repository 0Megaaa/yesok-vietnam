import { defineStore } from 'pinia'

export const useClientStore = defineStore('client', {
  state: () => ({
    token: localStorage.getItem('client_token') || '',
    userInfo: null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.userInfo?.role === 'admin',
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

    logout() {
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('client_token')
    },
  },
})
