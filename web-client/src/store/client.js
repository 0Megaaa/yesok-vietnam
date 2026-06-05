import { defineStore } from 'pinia'
import { getState, loginWithDemoAccount, loginWithWechat } from '@/api/client/auth'
import { getClientMe, updateClientProfile, getClientOrders, uploadClientAvatar } from '@/api/client/profile'

const getUniApi = () => (typeof uni !== 'undefined' ? uni : null)

const readStorage = (key, fallback = null) => {
  try {
    const uniApi = getUniApi()
    if (uniApi?.getStorageSync) return uniApi.getStorageSync(key) || fallback
    if (typeof localStorage !== 'undefined') {
      const rawValue = localStorage.getItem(key)
      if (!rawValue) return fallback
      try {
        return JSON.parse(rawValue)
      } catch (error) {
        return rawValue
      }
    }
  } catch (error) {
    console.warn('[YesOK Storage Read Failed]', key, error)
  }
  return fallback
}

const writeStorage = (key, value) => {
  try {
    const uniApi = getUniApi()
    if (uniApi?.setStorageSync) {
      uniApi.setStorageSync(key, value)
      return
    }
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(key, typeof value === 'string' ? value : JSON.stringify(value))
    }
  } catch (error) {
    console.warn('[YesOK Storage Write Failed]', key, error)
  }
}

const removeStorage = (key) => {
  try {
    const uniApi = getUniApi()
    if (uniApi?.removeStorageSync) {
      uniApi.removeStorageSync(key)
      return
    }
    if (typeof localStorage !== 'undefined') localStorage.removeItem(key)
  } catch (error) {
    console.warn('[YesOK Storage Remove Failed]', key, error)
  }
}

const getClientPlatform = () => 'mp_weixin'

const getOrCreateDevIdentity = () => {
  const key = 'yesok_dev_identity'
  try {
    const uniApi = getUniApi()
    if (uniApi?.getStorageSync) {
      const existing = uniApi.getStorageSync(key)
      if (existing) return existing
      const value = `mpdev_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
      uniApi.setStorageSync(key, value)
      return value
    }
    if (typeof localStorage !== 'undefined') {
      const existing = localStorage.getItem(key)
      if (existing) return existing
      const value = `webdev_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
      localStorage.setItem(key, value)
      return value
    }
  } catch (error) {
    console.warn('[DevIdentity] failed:', error)
  }
  return `fallback_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
}

const getWechatProfileSafely = async () => {
  const uniApi = getUniApi()
  if (!uniApi?.getUserProfile) return {}

  try {
    const res = await new Promise((resolve, reject) => {
      uniApi.getUserProfile({
        desc: '用于完善 YesOK 用户资料展示',
        success: resolve,
        fail: reject,
      })
    })

    const info = res.userInfo || {}
    return {
      nickname: info.nickName || '',
      avatar_url: info.avatarUrl || '',
    }
  } catch (error) {
    console.warn('[WechatProfile] user cancelled or failed:', error)
    return {}
  }
}

const isLocalAvatarPath = (url = '') => {
  const value = String(url || '')
  return (
    value.startsWith('wxfile://') ||
    value.startsWith('http://tmp/') ||
    value.startsWith('file://') ||
    value.includes('/tmp/') ||
    value.includes('/temp/')
  )
}

const isDefaultWechatNickname = (name = '') => {
  const value = String(name || '').trim()
  return !value || value === '微信用户' || value === '用户昵称'
}

const showSafeToast = (options) => {
  const uniApi = getUniApi()
  if (uniApi?.showToast) {
    uniApi.showToast(options)
    return
  }
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
    profileSheetVisible: false,
    profileForm: {
      nickname: '',
      avatar_url: '',
    },
    pendingActionText: '继续使用该服务',
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.userInfo?.role === 'admin',
    activeOrders: (state) =>
      state.orders.filter((order) => {
        const s = order.macro_status || order.current_stage || order.status
        return s !== 'completed'
      }),
    completedOrders: (state) =>
      state.orders.filter((order) => {
        const s = order.macro_status || order.current_stage || order.status
        return s === 'completed'
      }),
  },

  actions: {
    _cleanupMockToken() {
      if (this.token === 'mock-client-token-demo') {
        this.setToken('')
        this.setUserInfo(null)
      }
    },

    setToken(token) {
      this.token = token
      if (token) {
        writeStorage('client_token', token)
      } else {
        removeStorage('client_token')
      }
    },

    setUserInfo(userInfo) {
      this.userInfo = userInfo
      if (userInfo) {
        writeStorage('client_user', userInfo)
      } else {
        removeStorage('client_user')
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

    async fetchMe() {
      if (!this.token) return null
      try {
        const data = await getClientMe()
        const user = data.user || data.data?.user || data
        if (user) this.setUserInfo(user)
        return user
      } catch {
        return null
      }
    },

    async updateProfile(profile) {
      const data = await updateClientProfile(profile)
      const user = data.user || data.data?.user || data
      if (user) this.setUserInfo(user)
      return user
    },

    async fetchOrders(status = 'all') {
      if (!this.token) {
        this.setOrders([])
        return []
      }
      try {
        const data = await getClientOrders({ status })
        const list = data.list || data.orders || []
        this.setOrders(list)
        return list
      } catch {
        this.setOrders([])
        return []
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

    openProfileSheet(profile = {}) {
      this.profileForm = {
        nickname: profile.nickname || this.userInfo?.nickname || '',
        avatar_url: profile.avatar_url || this.userInfo?.avatar_url || '',
      }
      this.profileSheetVisible = true
    },

    closeProfileSheet() {
      this.profileSheetVisible = false
    },

    shouldCompleteProfile(user = this.userInfo) {
      if (!user) return true
      return isDefaultWechatNickname(user.nickname) || !user.avatar_url
    },

    setProfileForm(patch = {}) {
      this.profileForm = {
        ...this.profileForm,
        ...patch,
      }
    },

    async completeProfile(profile = {}) {
      if (!this.token) {
        throw new Error('请先登录')
      }

      let nickname = String(profile.nickname || this.profileForm.nickname || '').trim()
      let avatarUrl = String(profile.avatar_url || this.profileForm.avatar_url || '').trim()

      if (!nickname) {
        throw new Error('请输入昵称')
      }

      if (!avatarUrl) {
        throw new Error('请选择头像')
      }

      if (isLocalAvatarPath(avatarUrl)) {
        const uploadRes = await uploadClientAvatar(avatarUrl)
        avatarUrl = uploadRes.url || uploadRes.data?.url || ''
      }

      if (!avatarUrl) {
        throw new Error('头像上传失败')
      }

      const user = await this.updateProfile({
        nickname,
        avatar_url: avatarUrl,
      })

      this.profileSheetVisible = false
      showSafeToast({ title: '资料已保存', icon: 'success' })
      return user
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

    async loginByWechat(phoneCode = '') {
      const uniApi = getUniApi()
      if (!uniApi?.login) {
        throw new Error('当前环境不支持微信登录')
      }

      const loginRes = await new Promise((resolve, reject) => {
        uniApi.login({
          provider: 'weixin',
          success: resolve,
          fail: reject,
        })
      })

      if (!loginRes.code) {
        throw new Error('微信登录失败：缺少 code')
      }

      if (this.token === 'mock-client-token-demo') {
        this.setToken('')
      }

      const profile = await getWechatProfileSafely()

      const data = await loginWithWechat({
        code: loginRes.code,
        phone_code: phoneCode || '',
        nickname: profile.nickname || this.userInfo?.nickname || '',
        avatar_url: profile.avatar_url || this.userInfo?.avatar_url || '',
        login_provider: 'wechat',
        client_platform: getClientPlatform(),
        dev_identity: getOrCreateDevIdentity(),
      })

      const token = data.token || data.data?.token
      const user = data.user || data.data?.user

      if (!token) {
        throw new Error('登录失败：后端未返回 token')
      }

      this.setToken(token)
      this.setUserInfo(user)
      this.closeLoginSheet()

      if (this.shouldCompleteProfile(user)) {
        this.openProfileSheet({
          nickname: profile.nickname || user?.nickname || '',
          avatar_url: profile.avatar_url || user?.avatar_url || '',
        })
        showSafeToast({ title: '请完善头像昵称', icon: 'none' })
      } else {
        showSafeToast({ title: '登录成功', icon: 'success' })
      }

      return data
    },

    logout() {
      this.token = ''
      this.userInfo = null
      this.orders = []
      this.profileSheetVisible = false
      this.profileForm = { nickname: '', avatar_url: '' }
      removeStorage('client_token')
      removeStorage('client_user')
    },
  },
})
