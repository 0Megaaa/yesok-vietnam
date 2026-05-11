<script setup>
import { useClientStore } from '../../store/client'
import { RouterLink } from 'vue-router'

const client = useClientStore()
</script>

<template>
  <div>
    <!-- Profile Header -->
    <div class="prof-hd">
      <div class="prof-hd-deco"></div>
      <div class="prof-hd-deco2"></div>
      <div style="text-align:center;margin-bottom:16px;">
        <div class="prof-hd-ic">
          <img v-if="client.userInfo?.avatar_url" :src="client.userInfo.avatar_url" style="width:64px;height:64px;border-radius:50%;border:3px solid rgba(255,255,255,.4);" alt="avatar" />
          <span v-else style="font-size:52px;">👤</span>
        </div>
        <div class="prof-hd-nm">{{ client.userInfo?.username || 'Guest' }}</div>
        <div class="prof-hd-rl">{{ client.userInfo?.role === 'admin' ? '管理员' : '尊贵用户' }} · {{ client.userInfo?.language || 'zh' }}</div>
      </div>
      <!-- Wallet stats -->
      <div class="prof-wallet">
        <div class="pw-c">
          <div class="pw-v">{{ client.userInfo?.balance ?? '—' }}</div>
          <div class="pw-l">账户余额</div>
        </div>
        <div class="pw-c">
          <div class="pw-v">{{ client.activeOrders?.length ?? 0 }}</div>
          <div class="pw-l">进行中</div>
        </div>
        <div class="pw-c">
          <div class="pw-v">{{ client.completedOrders?.length ?? 0 }}</div>
          <div class="pw-l">已完成</div>
        </div>
      </div>
    </div>

    <!-- Manager card -->
    <div class="prof-mgr">
      <div class="pmgr-av">🏠</div>
      <div style="flex:1;">
        <div class="pmgr-nm">专属管家团队</div>
        <div class="pmgr-st"><div class="pmgr-dot"></div>在线 · 7×24小时</div>
      </div>
      <button style="padding:8px 14px;border-radius:18px;background:var(--bl);color:#fff;font-size:12px;font-weight:600;border:none;cursor:pointer;">发消息</button>
    </div>

    <!-- Order stats row -->
    <div style="background:#fff;margin:0 0 6px;padding:14px 16px;">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:14px;">
        <div style="font-size:15px;font-weight:700;color:#1A2340;">我的订单</div>
        <span style="font-size:12px;color:var(--bl);cursor:pointer;">全部订单 ›</span>
      </div>
      <div style="display:grid;grid-template-columns:1fr 1fr 1fr 1fr;gap:0;">
        <div style="text-align:center;cursor:pointer;padding:4px 0;">
          <div style="width:44px;height:44px;border-radius:12px;background:var(--gy);margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#4A5568" stroke-width="2"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="9" y1="7" x2="15" y2="7"/></svg>
          </div>
          <div style="font-size:11px;color:#4A5568;">全部</div>
        </div>
        <div style="text-align:center;cursor:pointer;padding:4px 0;">
          <div style="width:44px;height:44px;border-radius:12px;background:#E3F0FF;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;position:relative;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#0D47A1" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
          </div>
          <div style="font-size:11px;color:#4A5568;">进行中</div>
        </div>
        <div style="text-align:center;cursor:pointer;padding:4px 0;">
          <div style="width:44px;height:44px;border-radius:12px;background:#E8F5E9;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#2E7D32" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
          </div>
          <div style="font-size:11px;color:#4A5568;">已完成</div>
        </div>
        <div style="text-align:center;cursor:pointer;padding:4px 0;">
          <div style="width:44px;height:44px;border-radius:12px;background:#F3E5F5;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#6A1B9A" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/></svg>
          </div>
          <div style="font-size:11px;color:#4A5568;">优惠券</div>
        </div>
      </div>
    </div>

    <!-- Menu list -->
    <div class="prof-sec">账户</div>
    <div class="prof-menu">
      <div class="pm-row">
        <div class="pm-ic">💰</div>
        <div><div class="pm-nm">我的钱包</div><div class="pm-ht">查看余额与账单</div></div>
        <div class="pm-arr">›</div>
      </div>
      <div class="pm-row">
        <div class="pm-ic">📍</div>
        <div><div class="pm-nm">地址管理</div><div class="pm-ht">管理收件地址</div></div>
        <div class="pm-arr">›</div>
      </div>
      <div class="pm-row">
        <div class="pm-ic">⭐</div>
        <div><div class="pm-nm">我的评价</div><div class="pm-ht">查看服务评价</div></div>
        <div class="pm-arr">›</div>
      </div>
    </div>

    <div class="prof-sec">其他</div>
    <div class="prof-menu">
      <div class="pm-row">
        <div class="pm-ic">👥</div>
        <div><div class="pm-nm">邀请好友</div><div class="pm-ht">邀请赚积分</div></div>
        <div class="pm-badge">赚积分</div>
      </div>
      <div class="pm-row">
        <div class="pm-ic">❓</div>
        <div><div class="pm-nm">帮助中心</div><div class="pm-ht">常见问题</div></div>
        <div class="pm-arr">›</div>
      </div>
      <div class="pm-row">
        <div class="pm-ic">ℹ️</div>
        <div><div class="pm-nm">关于我们</div><div class="pm-ht">了解 Yesok</div></div>
        <div class="pm-arr">›</div>
      </div>
      <div class="pm-row">
        <div class="pm-ic">⚙️</div>
        <div><div class="pm-nm">设置</div><div class="pm-ht">账号与偏好</div></div>
        <div class="pm-arr">›</div>
      </div>
    </div>

    <!-- Admin entry -->
    <div style="margin:0 12px 12px;background:linear-gradient(135deg,#1A237E,#283593);border-radius:14px;padding:14px 16px;display:flex;align-items:center;gap:12px;cursor:pointer;">
      <div style="width:40px;height:40px;border-radius:10px;background:#F6B000;display:flex;align-items:center;justify-content:center;font-size:19px;">🔐</div>
      <div style="flex:1;">
        <div style="font-size:14px;font-weight:700;color:#F6B000;">管理后台</div>
        <div style="font-size:11px;color:rgba(255,255,255,.5);margin-top:2px;">平台运营管理（内部）</div>
      </div>
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
    </div>

    <!-- Logout -->
    <div v-if="client.isLoggedIn" style="padding:0 12px 14px;">
      <button
        style="width:100%;padding:12px;border-radius:10px;background:#FFEBEE;color:var(--rd);font-size:13px;font-weight:600;border:1px solid #FFCDD2;cursor:pointer;"
        @click="client.logout()"
      >退出登录</button>
    </div>

    <div style="height:12px;"></div>
  </div>
</template>

<style scoped>
/* All styles are inherited from style.css */
</style>
