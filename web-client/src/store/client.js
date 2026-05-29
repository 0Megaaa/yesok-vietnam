import { defineStore } from 'pinia'
import { getState, loginWithDemoAccount, loginWithWechat } from '@/api/client/auth'
import { getClientMe, updateClientProfile, getClientOrders } from '@/api/client/profile'

// getUniApi 安全获取 UniApp 运行时对象。
// 实现步骤：
// 1. 判断全局 uni 是否存在，兼容微信小程序、H5 与 App 端。
// 2. 普通浏览器预览环境下返回 null，避免 setup 阶段直接引用 uni 抛错。
// 3. 由下方缓存与提示封装统一降级到 localStorage 或 console 输出。
const getUniApi = () => (typeof uni !== 'undefined' ? uni : null)

// readStorage 安全读取本地缓存。
// 实现步骤：
// 1. 优先使用 UniApp 官方 storage API。
// 2. 浏览器预览环境降级使用 localStorage。
// 3. 读取失败时返回兜底值，避免影响页面渲染。
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

// writeStorage 安全写入本地缓存。
// 实现步骤：
// 1. 优先写入 UniApp 官方 storage。
// 2. 浏览器预览环境写入 localStorage。
// 3. 捕获写入异常，避免隐私模式或低版本浏览器导致页面崩溃。
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

// removeStorage 安全移除本地缓存。
// 实现步骤：
// 1. 优先调用 UniApp 官方 removeStorageSync。
// 2. 浏览器预览环境降级移除 localStorage。
// 3. 捕获异常保证退出登录等动作不会中断页面。
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

// getClientPlatform 识别当前客户端平台。
const getClientPlatform = () => 'mp_weixin'

// getWechatProfileSafely 安全获取微信用户资料，用户拒绝授权也不阻止登录。
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

// showSafeToast 安全展示轻提示。
// 实现步骤：
// 1. UniApp 环境使用 showToast。
// 2. 普通浏览器预览环境仅输出日志。
// 3. 避免登录成功反馈调用导致 H5 预览崩溃。
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
    // 初始化时清理 mock token，避免干扰真实后端鉴权
    _cleanupMockToken() {
      if (this.token === 'mock-client-token-demo') {
        this.setToken('')
        this.setUserInfo(null)
      }
    },
    // setToken 统一维护本地 Token。
    // 实现步骤：
    // 1. 更新 Pinia 内存状态。
    // 2. 有值时写入本地缓存，刷新页面仍保持登录。
    // 3. 无值时清理缓存，避免残留无效凭证。
    setToken(token) {
      this.token = token
      if (token) {
        writeStorage('client_token', token)
      } else {
        removeStorage('client_token')
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

    // fetchMe 从后端拉取最新用户资料并更新 store。
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

    // updateProfile 更新用户资料（昵称、头像）。
    async updateProfile(profile) {
      const data = await updateClientProfile(profile)
      const user = data.user || data.data?.user || data
      if (user) this.setUserInfo(user)
      return user
    },

    // fetchOrders 从后端拉取当前用户的订单列表并更新 store。
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

    // checkAuth 执行业务入口全局鉴权。
    // 意图：为咨询、下单、支付等关键动作提供统一登录拦截能力。
    // 实现步骤：
    // 1. 已登录时直接返回 true，业务动作继续执行。
    // 2. 未登录时记录触发动作，并打开 AuthPopup 底部授权弹窗。
    // 3. 返回 false，让调用方停止跳转或下单动作。
    // 返回：布尔值，表示当前动作是否允许继续。
    checkAuth(actionText = '继续使用该服务') {
      if (this.isLoggedIn) return true
      this.openLoginSheet(actionText)
      return false
    },

    // ensureLogin 保留旧调用名并转发到 checkAuth。
    // 意图：兼容第一轮页面代码，避免批量迁移时遗漏旧方法导致运行错误。
    // 实现步骤：
    // 1. 接收旧页面传入的业务动作文案。
    // 2. 调用新的全局 checkAuth 方法。
    // 3. 原样返回鉴权判断结果。
    // 返回：布尔值，表示业务动作是否允许继续。
    ensureLogin(actionText = '继续使用该服务') {
      return this.checkAuth(actionText)
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
      showSafeToast({ title: '登录成功', icon: 'success' })
      return data
    },

    // loginByWechat 执行微信小程序登录（含昵称头像授权）。
    // 实现步骤：
    // 1. 调用 uni.login 获取微信 code。
    // 2. 调用 getWechatProfileSafely 获取昵称头像（用户拒绝不阻止登录）。
    // 3. 调用后端 /v1/client/auth/wechat 接口换取 JWT token。
    // 4. 写入本地状态并关闭弹窗。
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

      // 清理可能残留的 mock token
      if (this.token === 'mock-client-token-demo') {
        this.setToken('')
      }

      // 获取微信昵称头像（用户拒绝授权不阻止登录）
      const profile = await getWechatProfileSafely()

      const data = await loginWithWechat({
        code: loginRes.code,
        phone_code: phoneCode || '',
        nickname: profile.nickname || this.userInfo?.nickname || '',
        avatar_url: profile.avatar_url || this.userInfo?.avatar_url || '',
        login_provider: 'wechat',
        client_platform: getClientPlatform(),
      })

      const token = data.token || data.data?.token
      const user = data.user || data.data?.user

      if (!token) {
        throw new Error('登录失败：后端未返回 token')
      }

      this.setToken(token)
      this.setUserInfo(user)
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
