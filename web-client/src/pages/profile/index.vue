<script setup>
import { useClientStore } from '@/store/client'

const client = useClientStore()

// sendMessage 触发管家消息提示。
// 1.意图 -> 替代原生 button，保持 Uni-app 组件规范。
// 2.步骤 -> 调用 uni.showToast 展示占位提示。
// 3.返回 -> 无返回值。
const sendMessage = () => {
  uni.showToast({ title: '管家稍后联系您', icon: 'none' })
}
</script>

<template>
  <view class="profile-page">
    <view class="profile-hero">
      <view class="hero-orb left"></view>
      <view class="hero-orb right"></view>
      <view class="avatar-wrap">
        <image v-if="client.userInfo?.avatar_url" class="avatar-img" :src="client.userInfo.avatar_url" mode="aspectFill" />
        <text v-else class="avatar-text">👤</text>
      </view>
      <text class="user-name">{{ client.userInfo?.username || client.userInfo?.nickname || 'Guest' }}</text>
      <text class="user-role">{{ client.userInfo?.role === 'admin' ? '管理员' : '尊贵用户' }} · {{ client.userInfo?.language || 'zh' }}</text>
      <view class="wallet-row">
        <view class="wallet-item"><text class="wallet-value">{{ client.userInfo?.balance ?? '—' }}</text><text class="wallet-label">账户余额</text></view>
        <view class="wallet-item"><text class="wallet-value">{{ client.activeOrders?.length ?? 0 }}</text><text class="wallet-label">进行中</text></view>
        <view class="wallet-item"><text class="wallet-value">{{ client.completedOrders?.length ?? 0 }}</text><text class="wallet-label">已完成</text></view>
      </view>
    </view>

    <view class="manager-card">
      <view class="manager-avatar">管</view>
      <view class="manager-main"><text class="manager-name">专属管家团队</text><view class="manager-status"><view class="dot"></view><text>在线 · 7×24小时</text></view></view>
      <view class="message-btn" @click="sendMessage">发消息</view>
    </view>

    <view class="order-card">
      <view class="card-head"><text class="card-title">我的订单</text><text class="card-link">全部订单 ›</text></view>
      <view class="order-grid">
        <view class="order-item"><view class="order-icon">全</view><text>全部</text></view>
        <view class="order-item"><view class="order-icon blue">进</view><text>进行中</text></view>
        <view class="order-item"><view class="order-icon green">完</view><text>已完成</text></view>
        <view class="order-item"><view class="order-icon purple">券</view><text>优惠券</text></view>
      </view>
    </view>

    <text class="section-label">账户</text>
    <view class="menu-card">
      <view class="menu-row"><view class="menu-icon">¥</view><view class="menu-main"><text class="menu-name">我的钱包</text><text class="menu-desc">查看余额与账单</text></view><text class="arrow">›</text></view>
      <view class="menu-row"><view class="menu-icon">址</view><view class="menu-main"><text class="menu-name">地址管理</text><text class="menu-desc">管理收件地址</text></view><text class="arrow">›</text></view>
      <view class="menu-row"><view class="menu-icon">评</view><view class="menu-main"><text class="menu-name">我的评价</text><text class="menu-desc">查看服务评价</text></view><text class="arrow">›</text></view>
    </view>

    <text class="section-label">其他</text>
    <view class="menu-card">
      <view class="menu-row"><view class="menu-icon">邀</view><view class="menu-main"><text class="menu-name">邀请好友</text><text class="menu-desc">邀请赚积分</text></view><view class="badge">赚积分</view></view>
      <view class="menu-row"><view class="menu-icon">问</view><view class="menu-main"><text class="menu-name">帮助中心</text><text class="menu-desc">常见问题</text></view><text class="arrow">›</text></view>
      <view class="menu-row"><view class="menu-icon">介</view><view class="menu-main"><text class="menu-name">关于我们</text><text class="menu-desc">了解 Yesok</text></view><text class="arrow">›</text></view>
      <view class="menu-row"><view class="menu-icon">设</view><view class="menu-main"><text class="menu-name">设置</text><text class="menu-desc">账号与偏好</text></view><text class="arrow">›</text></view>
    </view>


    <view v-if="client.isLoggedIn" class="logout-wrap">
      <view class="logout-btn" @click="client.logout()">退出登录</view>
    </view>
  </view>
</template>

<style scoped>
.profile-page { min-height: 100vh; padding-bottom: 96px; background: #f2f6f5; color: #12312c; }
.profile-hero { position: relative; overflow: hidden; padding: 58px 18px 22px; border-radius: 0 0 34px 34px; background: linear-gradient(135deg,#004d40,#0f766e); color: #fff; text-align: center; }
.hero-orb { position: absolute; width: 160px; height: 160px; border-radius: 50%; background: rgba(245,217,143,.18); filter: blur(2px); }
.hero-orb.left { left: -62px; top: -48px; }
.hero-orb.right { right: -70px; bottom: -58px; }
.avatar-wrap { position: relative; z-index: 1; display: flex; align-items: center; justify-content: center; width: 78px; height: 78px; margin: 0 auto 14px; border-radius: 50%; background: rgba(255,255,255,.18); border: 3px solid rgba(255,255,255,.34); }
.avatar-img { width: 72px; height: 72px; border-radius: 50%; }
.avatar-text { font-size: 42px; }
.user-name, .user-role, .wallet-value, .wallet-label, .manager-name, .card-title, .menu-name, .menu-desc { display: block; }
.user-name { position: relative; z-index: 1; font-size: 22px; font-weight: 900; }
.user-role { position: relative; z-index: 1; margin-top: 6px; color: rgba(255,255,255,.76); font-size: 12px; }
.wallet-row { position: relative; z-index: 1; display: flex; margin-top: 22px; padding: 14px 0; border-radius: 24px; background: rgba(255,255,255,.15); backdrop-filter: blur(16px); }
.wallet-item { flex: 1; }
.wallet-value { font-size: 18px; font-weight: 900; }
.wallet-label { margin-top: 4px; color: rgba(255,255,255,.72); font-size: 11px; }
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
.card-link { color: #004d40; font-size: 12px; font-weight: 800; }
.order-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; }
.order-item { text-align: center; color: #4a5f5a; font-size: 11px; }
.order-icon { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; margin: 0 auto 7px; border-radius: 16px; background: #eef5f2; color: #004d40; font-weight: 900; }
.order-icon.blue { background: #e3f0ff; color: #0d47a1; }
.order-icon.green { background: #e8f5e9; color: #2e7d32; }
.order-icon.purple { background: #f3e5f5; color: #6a1b9a; }
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
