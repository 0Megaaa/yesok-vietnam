<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useClientStore } from '../../store/client'

const router = useRouter()
const client = useClientStore()
const orders = computed(() => client.orders || [])

const STATUS_CONFIG = {
  pending:      { i: '📋', l: '等待审核',    b: 'stb-bl', p: 10 },
  requirement_submitted: { i: '📝', l: '需求已提交', b: 'stb-bl', p: 20 },
  processing:   { i: '⚙️', l: '服务进行中',  b: 'stb-gd', p: 45 },
  supplementing:{ i: '⚠️', l: '需要补充材料', b: 'stb-rd', p: 65 },
  completed:    { i: '✅', l: '服务已完成',  b: 'stb-gr', p: 100 },
}
const getStatus = (sk) => STATUS_CONFIG[sk] || STATUS_CONFIG.pending
const getOrderClass = (sk) => {
  if (sk === 'supplementing') return 'oc-w'
  if (sk === 'completed') return 'oc-d'
  return 'oc-a'
}
</script>

<template>
  <div class="hero">
    <div class="hero-deco"></div>
    <div class="hero-badge">
      <div class="hero-dot"></div>
      一站式管家服务
    </div>
    <h1 class="hero-h1">Yes<span class="yk">ok</span> Vietnam</h1>
    <p class="hero-sub">在越南，所有事我们帮你搞定</p>
    <div class="hero-search">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.8)" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
      <span style="font-size:13px;color:rgba(255,255,255,.75);">搜索服务关键词…</span>
    </div>
    <div class="hero-mgr">
      <div class="hero-mgr-av">🏠</div>
      <div>
        <div style="font-size:13px;font-weight:700;color:#fff;">专属管家团队</div>
        <div style="font-size:11px;color:rgba(255,255,255,.7);margin-top:2px;">7×24小时 · 全程中文陪同</div>
      </div>
      <button style="margin-left:auto;padding:7px 13px;border-radius:18px;background:rgba(255,255,255,.2);border:1.5px solid rgba(255,255,255,.35);color:#fff;font-size:11px;font-weight:600;cursor:pointer;">立即咨询</button>
    </div>
  </div>

  <div class="stats-row">
    <div class="st-item">
      <div class="st-num">3000+</div>
      <div class="st-lbl">服务用户</div>
    </div>
    <div class="st-item">
      <div class="st-num">98%</div>
      <div class="st-lbl">好评率</div>
    </div>
    <div class="st-item">
      <div class="st-num">5min</div>
      <div class="st-lbl">平均响应</div>
    </div>
    <div class="st-item">
      <div class="st-num">12+</div>
      <div class="st-lbl">服务项目</div>
    </div>
  </div>

  <div style="padding-top:14px;">
    <div class="cat-row">
      <div class="cat-item" @click="router.push('/service')">
        <div class="cat-ic">✈️</div>
        <div class="cat-lb">签证旅游</div>
      </div>
      <div class="cat-item" @click="router.push('/service')">
        <div class="cat-ic">🏢</div>
        <div class="cat-lb">房产投资</div>
      </div>
      <div class="cat-item" @click="router.push('/service')">
        <div class="cat-ic">📋</div>
        <div class="cat-lb">公司注册</div>
      </div>
      <div class="cat-item" @click="router.push('/service')">
        <div class="cat-ic">🚗</div>
        <div class="cat-lb">接机包车</div>
      </div>
      <div class="cat-item" @click="router.push('/service')">
        <div class="cat-ic">🗣️</div>
        <div class="cat-lb">专业翻译</div>
      </div>
      <div class="cat-item" @click="router.push('/service')">
        <div class="cat-ic">⊞</div>
        <div class="cat-lb">全部服务</div>
      </div>
    </div>
  </div>

  <div class="sec-hd" style="padding-top:4px;">
    <div class="sec-t"><span style="font-size:16px;">🔥</span> 热门服务</div>
    <div class="sec-more" @click="router.push('/service')">查看更多 ›</div>
  </div>
  <div class="hot-scroll">
    <div class="hot-card" @click="router.push('/service')">
      <div class="hot-thumb">📋</div>
      <div class="hot-body">
        <div class="hot-nm">公司注册</div>
        <div class="hot-pr">¥3000起</div>
        <div class="hot-tg">全程代办</div>
      </div>
    </div>
    <div class="hot-card" @click="router.push('/service')">
      <div class="hot-thumb">✈️</div>
      <div class="hot-body">
        <div class="hot-nm">商务签证</div>
        <div class="hot-pr">¥1200起</div>
        <div class="hot-tg">极速出签</div>
      </div>
    </div>
    <div class="hot-card" @click="router.push('/service')">
      <div class="hot-thumb">🚗</div>
      <div class="hot-body">
        <div class="hot-nm">机场接送</div>
        <div class="hot-pr">¥200起</div>
        <div class="hot-tg">中文司机</div>
      </div>
    </div>
  </div>

  <div class="vip-hd" @click="router.push('/service')">
    <div class="vip-badge">👑</div>
    <div style="flex:1;">
      <div class="vip-ti">高端通道服务 <span class="vi-bg gold" style="margin-left:6px; padding:1px 6px; border-radius:10px; font-size:9px;">私密·专属</span></div>
      <div class="vip-sb">解决复杂问题，一对一资源对接</div>
    </div>
    <div style="color:var(--gd);font-size:18px;">›</div>
  </div>

  <div class="sec-hd">
    <div class="sec-t"><span style="font-size:16px;">📰</span> 越南早知道</div>
    <div class="sec-more">查看更多 ›</div>
  </div>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:10px;padding:0 12px 14px;">
    <div class="hnc">
      <div class="hnc-em">🇻🇳</div>
      <div class="hnc-ov"></div>
      <div class="hnc-badge" style="background:var(--rd);">热门</div>
      <div class="hnc-body">
        <div class="hnc-title">2026越南最新签证政策解读</div>
        <div class="hnc-meta">
          <span>👁 1.2k</span>
          <span style="margin-left:4px;">⭐ 4.9</span>
        </div>
      </div>
    </div>
    <div class="hnc">
      <div class="hnc-em">🏢</div>
      <div class="hnc-ov"></div>
      <div class="hnc-badge" style="background:#2E7D32;">最新</div>
      <div class="hnc-body">
        <div class="hnc-title">外资注册公司避坑指南</div>
        <div class="hnc-meta">
          <span>👁 856</span>
          <span style="margin-left:4px;">⭐ 4.8</span>
        </div>
      </div>
    </div>
  </div>

  <div class="sec-hd" style="border-top: 1px solid var(--br); padding-top: 18px;">
    <div class="sec-t"><span style="font-size:16px;">📋</span> 订单动态</div>
    <span class="sec-more" @click="router.push('/orders')">查看全部 ›</span>
  </div>

  <div class="ord-list">
    <template v-if="orders.length === 0">
      <div class="empty">
        <div class="empty-ic">📭</div>
        <div class="empty-ti">暂无订单</div>
      </div>
    </template>
    <template v-else>
      <div
        v-for="order in orders"
        :key="order.id"
        :class="['ord-card', getOrderClass(order.sk)]"
      >
        <div class="oc-top">
          <div>
            <div class="oc-svc">{{ order.icon || '📋' }} {{ order.serviceName || order.svc }}</div>
            <div class="oc-no">订单号 {{ order.id }}</div>
          </div>
          <span :class="['stb', getStatus(order.status || order.sk).b]">
            {{ getStatus(order.status || order.sk).l }}
          </span>
        </div>
        <div style="margin:9px 0;">
          <div class="oc-pbar">
            <div class="oc-pfill" :style="{ width: getStatus(order.status || order.sk).p + '%' }"></div>
          </div>
          <div class="oc-plbl">
            <span>{{ getStatus(order.status || order.sk).l }}</span>
            <span>{{ getStatus(order.status || order.sk).p }}% 完成</span>
          </div>
        </div>
        <div v-if="order.status === 'supplementing'" class="oc-warn">
          <div class="oc-wdot"></div>
          需要操作：请补充材料
        </div>
        <div class="oc-bot">
          <div class="oc-mgr">
            <div class="oc-mav">{{ (order.managerName || order.mg || '管')[0] }}</div>
            {{ order.managerName || order.mg || '专属管家' }}
          </div>
          <div style="font-size:11px;color:var(--tx3);">{{ order.price || order.pr || '—' }}</div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
/* 严禁在这里写任何样式，确保所有样式来自全局 style.css 保持甲方原汁原味 */
</style>