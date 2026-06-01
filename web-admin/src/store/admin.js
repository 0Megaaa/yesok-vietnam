import { defineStore } from 'pinia'
import { adminLogin, adminLogout } from '@/api/admin/auth'

const readStorage = (key) => {
  if (typeof localStorage !== 'undefined') return localStorage.getItem(key) || ''
  return ''
}

const writeStorage = (key, value) => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(key, value)
  }
}

const removeStorage = (key) => {
  if (typeof localStorage !== 'undefined') localStorage.removeItem(key)
}

export const useAdminStore = defineStore('admin', {
  state: () => ({
    token: readStorage('admin_token') || '',
    userInfo: null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
  },

  actions: {
    setToken(token) {
      this.token = token
      if (token) writeStorage('admin_token', token)
      else removeStorage('admin_token')
    },

    setUserInfo(userInfo) {
      this.userInfo = userInfo
      if (userInfo) writeStorage('admin_user', JSON.stringify(userInfo))
      else removeStorage('admin_user')
    },

    setExpire(expire) {
      if (expire) writeStorage('admin_token_expire', String(expire))
      else removeStorage('admin_token_expire')
    },

    async login(username, password) {
      const data = await adminLogin(username, password)
      this.setToken(data.token)
      this.setUserInfo(data.user)
      this.setExpire(data.expire)
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
      removeStorage('admin_token')
      removeStorage('admin_user')
      removeStorage('admin_token_expire')
    },
  },
})
