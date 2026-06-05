<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { ORIGIN_URL } from '@/api/request'
import AuthPopup from '@/components/AuthPopup.vue'
import { useClientStore } from '@/store/client'

const client = useClientStore()

const toAvatarUrl = (url) => {
  if (!url) return ''
  const value = String(url || '').trim()
  if (!value) return ''

  if (
    value.startsWith('wxfile://') ||
    value.startsWith('file://') ||
    value.startsWith('http://tmp/') ||
    value.includes('/tmp/') ||
    value.includes('/temp/')
  ) {
    return value
  }

  if (/^https?:\/\//.test(value)) return value

  if (value.startsWith('/static/') || value.startsWith('static/')) {
    return value.startsWith('/') ? value : `/${value}`
  }

  if (
    value.startsWith('/uploads/') ||
    value.startsWith('uploads/') ||
    value.startsWith('/material/') ||
    value.startsWith('material/')
  ) {
    const path = value.startsWith('/') ? value : `/${value}`
    return `${ORIGIN_URL}${path}`
  }

  return `${ORIGIN_URL}${value.startsWith('/') ? '' : '/'}${value}`
}

const loadingProfile = ref(false)
const loadingOrders = ref(false)
const activeOrderTab = ref('all')

const orderTabs = [
  { key: 'all', label: '全部', icon: '全' },
  { key: 'active', label: '进行中', icon: '进' },
  { key: 'completed', label: '已完成', icon: '完' },
]

const displayName = computed(() => {
  return client.userInfo?.nickname || client.userInfo?.username || '微信用户'
})

const avatarUrl = computed(() => {
  const raw = client.userInfo?.avatar_url || client.userInfo?.avatarUrl || ''
  return toAvatarUrl(raw)
})

const needCompleteProfile = computed(() => {
  return client.isLoggedIn && (!avatarUrl.value || displayName.value === '微信用户')
})

const orderCountAll = computed(() => client.orders.length)
const orderCountActive = computed(() => client.activeOrders.length)
const orderCountCompleted = computed(() => client.completedOrders.length)

const currentOrders = computed(() => {
  if (activeOrderTab.value === 'completed') return client.completedOrders
  if (activeOrderTab.value === 'active') return client.activeOrders
  return client.orders
})

const formatMoney = (amount) => {
  if (amount === null || amount === undefined || amount === '') return '—'
  const n = Number(amount || 0)
  return `¥${n.toLocaleString('zh-CN')}`
}

const formatTime = (t) => {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const statusLabel = (order) => {
  if (order.macro_status_text) return order.macro_status_text
  const code = order.macro_status || order.current_stage || order.status
  const map = {
    pending: '等待受理',
    wait_quote: '待报价',
    quoted: '已报价',
    wait_pay: '待支付',
    paid: '已支付',
    dispatching: '派车中',
    in_progress: '服务中',
    completed: '已完成',
    cancelled: '已取消',
  }
  return map[code] || code || '未知'
}

const ensureProfileLogin = () => {
  if (client.isLoggedIn) return true

  client.setOrders?.([])
  client.openLoginSheet('查看我的页面')
  return false
}

const loadProfile = async () => {
  if (!client.isLoggedIn) return
  loadingProfile.value = true
  try {
    await client.fetchMe()
  } catch (error) {
    console.warn('[profile] fetchMe failed:', error)
  } finally {
    loadingProfile.value = false
  }
}

const loadOrders = async (status = activeOrderTab.value) => {
  if (!client.isLoggedIn) {
    client.openLoginSheet('查看我的订单')
    return
  }
  loadingOrders.value = true
  try {
    await client.fetchOrders(status)
  } catch (error) {
    const uniApi = typeof uni !== 'undefined' ? uni : null
    if (uniApi?.showToast) uniApi.showToast({ title: '订单加载失败', icon: 'none' })
  } finally {
    loadingOrders.value = false
  }
}

const switchOrderTab = async (tabKey) => {
  activeOrderTab.value = tabKey
  await loadOrders(tabKey)
}

const goOrderDetail = (order) => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateTo) {
    uniApi.navigateTo({ url: `/subpkg/order-detail/index?id=${order.id}` })
  }
}

const goAllOrders = () => {
  activeOrderTab.value = 'all'
  loadOrders('all')
}

const openProfileEditor = () => {
  if (!client.isLoggedIn) {
    client.openLoginSheet('完善个人资料')
    return
  }

  client.openProfileSheet({
    nickname: client.userInfo?.nickname || '',
    avatar_url: client.userInfo?.avatar_url || client.userInfo?.avatarUrl || '',
  })
}

const sendMessage = () => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.showToast) {
    uniApi.showToast({ title: '专属管家功能建设中', icon: 'none' })
  }
}

onMounted(async () => {
  if (!ensureProfileLogin()) return
  await loadProfile()
  await loadOrders('all')
})

onShow(async () => {
  if (!ensureProfileLogin()) return
  await loadProfile()
  await loadOrders(activeOrderTab.value)
})

watch(
  () => client.isLoggedIn,
  async (loggedIn) => {
    if (!loggedIn) return
    await loadProfile()
    await loadOrders(activeOrderTab.value)
  }
)
</script>

<template>
  <view class="profile-page">
    <view class="profile-hero">
      <view class="hero-orb left"></view>
      <view class="hero-orb right"></view>
      <view
        class="avatar-wrap"
        @click="!client.isLoggedIn && client.openLoginSheet('登录后查看我的页面')"
      >
        <image v-if="avatarUrl" class="avatar-img" :src="avatarUrl" mode="aspectFill" />
        <text v-else class="avatar-text">👤</text>
      </view>
      <text class="user-name">{{ displayName }}</text>
      <text class="user-role">
        {{ client.isLoggedIn ? `尊贵用户 · ${client.userInfo?.vip_level || 1}` : '点击登录后查看个人信息' }}
      </text>
      <view v-if="needCompleteProfile" class="complete-profile-btn" @click="openProfileEditor">
        完善微信头像昵称
      </view>
      <view class="wallet-row">
        <view class="wallet-item">
          <text class="wallet-value">{{ client.userInfo?.balance ?? '—' }}</text>
          <text class="wallet-label">账户余额</text>
        </view>
        <view class="wallet-item">
          <text class="wallet-value">{{ orderCountActive }}</text>
          <text class="wallet-label">进行中</text>
        </view>
        <view class="wallet-item">
          <text class="wallet-value">{{ orderCountCompleted }}</text>
          <text class="wallet-label">已完成</text>
        </view>
      </view>
    </view>

    <view class="manager-card">
      <view class="manager-avatar">管</view>
      <view class="manager-main">
        <text class="manager-name">专属管家团队</text>
        <view class="manager-status">
          <view class="dot"></view>
          <text>在线 · 7×24小时</text>
        </view>
      </view>
      <view class="message-btn" @click="sendMessage">发消息</view>
    </view>

    <view class="order-card">
      <view class="card-head">
        <text class="card-title">我的订单</text>
        <text class="card-link" @click="goAllOrders">全部订单 ›</text>
      </view>

      <view class="order-grid three">
        <view
          v-for="tab in orderTabs"
          :key="tab.key"
          class="order-item"
          :class="{ active: activeOrderTab === tab.key }"
          @click="switchOrderTab(tab.key)"
        >
          <view class="order-icon" :class="tab.key">{{ tab.icon }}</view>
          <text>{{ tab.label }}</text>
          <text class="order-count">
            {{
              tab.key === 'all'
                ? orderCountAll
                : tab.key === 'active'
                  ? orderCountActive
                  : orderCountCompleted
            }}
          </text>
        </view>
      </view>

      <view class="order-list">
        <view v-if="loadingOrders" class="order-empty">订单加载中...</view>

        <view
          v-else-if="!client.isLoggedIn"
          class="order-empty"
          @click="client.openLoginSheet('查看我的订单')"
        >
          登录后查看我的订单
        </view>

        <view v-else-if="!currentOrders.length" class="order-empty">
          暂无{{ orderTabs.find(t => t.key === activeOrderTab)?.label || '' }}订单
        </view>

        <view
          v-else
          v-for="item in currentOrders"
          :key="item.id"
          class="mini-order"
          @click="goOrderDetail(item)"
        >
          <view class="mini-order-main">
            <text class="mini-order-title">{{ item.service_name || '服务订单' }}</text>
            <text class="mini-order-no">{{ item.order_no }}</text>
            <text class="mini-order-time">{{ formatTime(item.created_at) }}</text>
          </view>
          <view class="mini-order-side">
            <text class="mini-order-status">{{ statusLabel(item) }}</text>
            <text class="mini-order-amount">{{ formatMoney(item.total_amount || item.amount) }}</text>
          </view>
        </view>
      </view>
    </view>

    <text class="section-label">账户</text>
    <view class="menu-card">
      <view class="menu-row">
        <view class="menu-icon">¥</view>
        <view class="menu-main"><text class="menu-name">我的钱包</text><text class="menu-desc">查看余额与账单</text></view>
        <text class="arrow">›</text>
      </view>
      <view class="menu-row">
        <view class="menu-icon">址</view>
        <view class="menu-main"><text class="menu-name">地址管理</text><text class="menu-desc">管理收件地址</text></view>
        <text class="arrow">›</text>
      </view>
      <view class="menu-row">
        <view class="menu-icon">评</view>
        <view class="menu-main"><text class="menu-name">我的评价</text><text class="menu-desc">查看服务评价</text></view>
        <text class="arrow">›</text>
      </view>
    </view>

    <text class="section-label">其他</text>
    <view class="menu-card">
      <view class="menu-row">
        <view class="menu-icon">邀</view>
        <view class="menu-main"><text class="menu-name">邀请好友</text><text class="menu-desc">邀请赚积分</text></view>
        <view class="badge">赚积分</view>
      </view>
      <view class="menu-row">
        <view class="menu-icon">问</view>
        <view class="menu-main"><text class="menu-name">帮助中心</text><text class="menu-desc">常见问题</text></view>
        <text class="arrow">›</text>
      </view>
      <view class="menu-row">
        <view class="menu-icon">介</view>
        <view class="menu-main"><text class="menu-name">关于我们</text><text class="menu-desc">了解 Yesok</text></view>
        <text class="arrow">›</text>
      </view>
      <view class="menu-row" @click="openProfileEditor">
        <view class="menu-icon">设</view>
        <view class="menu-main"><text class="menu-name">设置</text><text class="menu-desc">修改头像与昵称</text></view>
        <text class="arrow">›</text>
      </view>
    </view>

    <view v-if="client.isLoggedIn" class="logout-wrap">
      <view class="logout-btn" @click="client.logout()">退出登录</view>
    </view>

    <AuthPopup />
  </view>
</template>

<style scoped>
.profile-page { min-height: 100vh; padding-bottom: 96px; background: #f2f6f5; color: #12312c; }
.profile-hero { position: relative; overflow: hidden; padding: 58px 18px 22px; border-radius: 0 0 34px 34px; background: linear-gradient(135deg,#004d40,#0f766e); color: #fff; text-align: center; }
.hero-orb { position: absolute; width: 160px; height: 160px; border-radius: 50%; background: rgba(245,217,143,.18); filter: blur(2px); }
.hero-orb.left { left: -62px; top: -48px; }
.hero-orb.right { right: -70px; bottom: -58px; }
.avatar-wrap { position: relative; z-index: 1; display: flex; align-items: center; justify-content: center; width: 78px; height: 78px; margin: 0 auto 14px; border-radius: 50%; background: rgba(255,255,255,.18); border: 3px solid rgba(255,255,255,.34); cursor: pointer; }
.avatar-img { width: 72px; height: 72px; border-radius: 50%; }
.avatar-text { font-size: 42px; }
.user-name, .user-role, .wallet-value, .wallet-label, .manager-name, .card-title, .menu-name, .menu-desc { display: block; }
.user-name { position: relative; z-index: 1; font-size: 22px; font-weight: 900; }
.user-role { position: relative; z-index: 1; margin-top: 6px; color: rgba(255,255,255,.76); font-size: 12px; }
.wallet-row { position: relative; z-index: 1; display: flex; margin-top: 22px; padding: 14px 0; border-radius: 24px; background: rgba(255,255,255,.15); backdrop-filter: blur(16px); }
.wallet-item { flex: 1; }
.wallet-value { font-size: 18px; font-weight: 900; }
.wallet-label { margin-top: 4px; color: rgba(255,255,255,.72); font-size: 11px; }
.complete-profile-btn { position: relative; z-index: 1; display: inline-flex; margin-top: 10px; padding: 6px 12px; border-radius: 999px; background: rgba(255, 255, 255, 0.18); color: #fff; font-size: 12px; font-weight: 800; }
.manager-card, .order-card, .menu-card { margin: 14px 14px 0; border-radius: 26px; background: rgba(255,255,255,.88); box-shadow: 0 18px 52px rgba(0,77,64,.08); }
.manager-card { display: flex; align-items: center; gap: 12px; padding: 14px; }
.manager-avatar { display: flex; align-items: center; justify-content: center; width: 46px; height: 46px; border-radius: 18px; background: linear-gradient(135deg,#f5d98f,#e97832); color: #12312c; font-size: 18px; font-weight: 900; }
.manager-main, .menu-main { flex: 1; min-width: 0; }
.manager-name { font-size: 15px; font-weight: 900; }
.manager-status { display: flex; align-items: center; gap: 6px; margin-top: 4px; color: #6b7c78; font-size: 11px; }
.dot { width: 7px; height: 7px; border-radius: 50%; background: #36c66f; }
.message-btn { padding: 8px 14px; border-radius: 999px; background: #004d40; color: #fff; font-size: 12px; font-weight: 900; }
.order-card { padding: 16px; }
.card-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
.card-title { font-size: 16px; font-weight: 900; }
.card-link { color: #004d40; font-size: 12px; font-weight: 800; cursor: pointer; }
.order-grid { display: grid; gap: 10px; }
.order-grid.three { grid-template-columns: repeat(3, 1fr); }
.order-item { position: relative; padding: 8px 0; border-radius: 18px; text-align: center; color: #4a5f5a; font-size: 11px; }
.order-item.active { background: rgba(0, 77, 64, 0.06); }
.order-item.active .order-icon { background: #004d40; color: #fff; }
.order-icon { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; margin: 0 auto 7px; border-radius: 16px; font-weight: 900; }
.order-icon.all { background: #eef5f2; color: #004d40; }
.order-icon.active { background: #e3f0ff; color: #0d47a1; }
.order-icon.completed { background: #e8f5e9; color: #2e7d32; }
.order-count { display: block; margin-top: 3px; color: #8a9996; font-size: 10px; }
.order-list { margin-top: 14px; padding-top: 12px; border-top: 1px solid rgba(0, 77, 64, 0.06); }
.order-empty { padding: 18px 0 4px; color: #8a9996; font-size: 12px; text-align: center; cursor: pointer; }
.mini-order { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 12px 0; border-bottom: 1px solid rgba(0, 77, 64, 0.06); cursor: pointer; }
.mini-order:last-child { border-bottom: 0; }
.mini-order-main { flex: 1; min-width: 0; }
.mini-order-title { display: block; color: #12312c; font-size: 13px; font-weight: 900; }
.mini-order-no, .mini-order-time { display: block; margin-top: 3px; color: #8a9996; font-size: 10px; }
.mini-order-side { display: flex; flex-direction: column; align-items: flex-end; gap: 5px; }
.mini-order-status { padding: 4px 8px; border-radius: 999px; background: #fff7e6; color: #9a5b00; font-size: 10px; font-weight: 900; }
.mini-order-amount { color: #e97832; font-size: 12px; font-weight: 900; }
.section-label { display: block; margin: 20px 20px 8px; color: #7a8b87; font-size: 12px; font-weight: 900; letter-spacing: 1px; }
.menu-card { overflow: hidden; }
.menu-row { display: flex; align-items: center; gap: 12px; padding: 15px 16px; border-bottom: 1px solid rgba(0,77,64,.06); }
.menu-row:last-child { border-bottom: 0; }
.menu-icon { display: flex; align-items: center; justify-content: center; width: 38px; height: 38px; border-radius: 14px; background: #eef5f2; color: #004d40; font-weight: 900; }
.menu-name { font-size: 14px; font-weight: 900; }
.menu-desc { margin-top: 3px; color: #8a9996; font-size: 11px; }
.arrow { color: #9aa7a3; font-size: 22px; }
.badge { padding: 5px 9px; border-radius: 999px; background: rgba(233,120,50,.12); color: #e97832; font-size: 11px; font-weight: 900; }
.logout-wrap { padding: 14px; }
.logout-btn { height: 46px; border-radius: 18px; background: #ffebee; color: #c62828; font-size: 13px; font-weight: 900; line-height: 46px; text-align: center; border: 1px solid #ffcdd2; }
@media (min-width: 768px) { .profile-page { max-width: 560px; margin: 0 auto; box-shadow: 0 0 80px rgba(0,77,64,.08); } }
</style>
