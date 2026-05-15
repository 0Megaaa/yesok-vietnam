<script setup>
import { useClientStore } from '@/store/client'

const client = useClientStore()
</script>

<template>
  <view>
    <!-- Profile Header -->
    <view class="prof-hd">
      <view class="prof-hd-deco"></view>
      <view class="prof-hd-deco2"></view>
      <view style="text-align:center;margin-bottom:16px;">
        <view class="prof-hd-ic">
          <img v-if="client.userInfo?.avatar_url" :src="client.userInfo.avatar_url" style="width:64px;height:64px;border-radius:50%;border:3px solid rgba(255,255,255,.4);" alt="avatar" />
          <text v-else style="font-size:52px;">👤</text>
        </view>
        <view class="prof-hd-nm">{{ client.userInfo?.username || 'Guest' }}</view>
        <view class="prof-hd-rl">{{ client.userInfo?.role === 'admin' ? '管理员' : '尊贵用户' }} · {{ client.userInfo?.language || 'zh' }}</view>
      </view>
      <!-- Wallet stats -->
      <view class="prof-wallet">
        <view class="pw-c">
          <view class="pw-v">{{ client.userInfo?.balance ?? '—' }}</view>
          <view class="pw-l">账户余额</view>
        </view>
        <view class="pw-c">
          <view class="pw-v">{{ client.activeOrders?.length ?? 0 }}</view>
          <view class="pw-l">进行中</view>
        </view>
        <view class="pw-c">
          <view class="pw-v">{{ client.completedOrders?.length ?? 0 }}</view>
          <view class="pw-l">已完成</view>
        </view>
      </view>
    </view>

    <!-- Manager card -->
    <view class="prof-mgr">
      <view class="pmgr-av">🏠</view>
      <view style="flex:1;">
        <view class="pmgr-nm">专属管家团队</view>
        <view class="pmgr-st"><view class="pmgr-dot"></view>在线 · 7×24小时</view>
      </view>
      <button style="padding:8px 14px;border-radius:18px;background:var(--bl);color:#fff;font-size:12px;font-weight:600;border:none;cursor:pointer;">发消息</button>
    </view>

    <!-- Order stats row -->
    <view style="background:#fff;margin:0 0 6px;padding:14px 16px;">
      <view style="display:flex;align-items:center;justify-content:space-between;margin-bottom:14px;">
        <view style="font-size:15px;font-weight:700;color:#1A2340;">我的订单</view>
        <text style="font-size:12px;color:var(--bl);cursor:pointer;">全部订单 ›</text>
      </view>
      <view style="display:grid;grid-template-columns:1fr 1fr 1fr 1fr;gap:0;">
        <view style="text-align:center;cursor:pointer;padding:4px 0;">
          <view style="width:44px;height:44px;border-radius:12px;background:var(--gy);margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#4A5568" stroke-width="2"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="9" y1="7" x2="15" y2="7"/></svg>
          </view>
          <view style="font-size:11px;color:#4A5568;">全部</view>
        </view>
        <view style="text-align:center;cursor:pointer;padding:4px 0;">
          <view style="width:44px;height:44px;border-radius:12px;background:#E3F0FF;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;position:relative;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#0D47A1" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
          </view>
          <view style="font-size:11px;color:#4A5568;">进行中</view>
        </view>
        <view style="text-align:center;cursor:pointer;padding:4px 0;">
          <view style="width:44px;height:44px;border-radius:12px;background:#E8F5E9;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#2E7D32" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
          </view>
          <view style="font-size:11px;color:#4A5568;">已完成</view>
        </view>
        <view style="text-align:center;cursor:pointer;padding:4px 0;">
          <view style="width:44px;height:44px;border-radius:12px;background:#F3E5F5;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#6A1B9A" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/></svg>
          </view>
          <view style="font-size:11px;color:#4A5568;">优惠券</view>
        </view>
      </view>
    </view>

    <!-- Menu list -->
    <view class="prof-sec">账户</view>
    <view class="prof-menu">
      <view class="pm-row">
        <view class="pm-ic">💰</view>
        <view><view class="pm-nm">我的钱包</view><view class="pm-ht">查看余额与账单</view></view>
        <view class="pm-arr">›</view>
      </view>
      <view class="pm-row">
        <view class="pm-ic">📍</view>
        <view><view class="pm-nm">地址管理</view><view class="pm-ht">管理收件地址</view></view>
        <view class="pm-arr">›</view>
      </view>
      <view class="pm-row">
        <view class="pm-ic">⭐</view>
        <view><view class="pm-nm">我的评价</view><view class="pm-ht">查看服务评价</view></view>
        <view class="pm-arr">›</view>
      </view>
    </view>

    <view class="prof-sec">其他</view>
    <view class="prof-menu">
      <view class="pm-row">
        <view class="pm-ic">👥</view>
        <view><view class="pm-nm">邀请好友</view><view class="pm-ht">邀请赚积分</view></view>
        <view class="pm-badge">赚积分</view>
      </view>
      <view class="pm-row">
        <view class="pm-ic">❓</view>
        <view><view class="pm-nm">帮助中心</view><view class="pm-ht">常见问题</view></view>
        <view class="pm-arr">›</view>
      </view>
      <view class="pm-row">
        <view class="pm-ic">ℹ️</view>
        <view><view class="pm-nm">关于我们</view><view class="pm-ht">了解 Yesok</view></view>
        <view class="pm-arr">›</view>
      </view>
      <view class="pm-row">
        <view class="pm-ic">⚙️</view>
        <view><view class="pm-nm">设置</view><view class="pm-ht">账号与偏好</view></view>
        <view class="pm-arr">›</view>
      </view>
    </view>

    <!-- Admin entry -->
    <view style="margin:0 12px 12px;background:linear-gradient(135deg,#1A237E,#283593);border-radius:14px;padding:14px 16px;display:flex;align-items:center;gap:12px;cursor:pointer;">
      <view style="width:40px;height:40px;border-radius:10px;background:#F6B000;display:flex;align-items:center;justify-content:center;font-size:19px;">🔐</view>
      <view style="flex:1;">
        <view style="font-size:14px;font-weight:700;color:#F6B000;">管理后台</view>
        <view style="font-size:11px;color:rgba(255,255,255,.5);margin-top:2px;">平台运营管理（内部）</view>
      </view>
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
    </view>

    <!-- Logout -->
    <view v-if="client.isLoggedIn" style="padding:0 12px 14px;">
      <button
        style="width:100%;padding:12px;border-radius:10px;background:#FFEBEE;color:var(--rd);font-size:13px;font-weight:600;border:1px solid #FFCDD2;cursor:pointer;"
        @click="client.logout()"
      >退出登录</button>
    </view>

    <view style="height:12px;"></view>
  </view>
</template>

<style scoped>
/* All styles are inherited from style.css */
</style>
