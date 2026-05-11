<script setup>
import { computed } from 'vue'
import { useClientStore } from '../../store/client'

const client = useClientStore()

const orders = computed(() => client.orders || [])

// Status config: maps order status to icon, label, badge class, pct, header bg, header color
const STATUS_CONFIG = {
  pending:      { i: '📋', l: '等待审核',    b: 'stb-bl', p: 10, hBg: 'var(--blp)',   hCl: 'var(--bl)'  },
  requirement_submitted: { i: '📝', l: '需求已提交', b: 'stb-bl', p: 20, hBg: 'var(--blp)',   hCl: 'var(--bl)'  },
  processing:   { i: '⚙️', l: '服务进行中',  b: 'stb-gd', p: 45, hBg: 'var(--gdl)',   hCl: '#795500'  },
  supplementing:{ i: '⚠️', l: '需要补充材料', b: 'stb-rd', p: 65, hBg: 'var(--rdl)',   hCl: 'var(--rd)'  },
  completed:    { i: '✅', l: '服务已完成',  b: 'stb-gr', p: 100, hBg: '#E8F5E9',    hCl: '#2E7D32'  },
}
const getStatus = (sk) => STATUS_CONFIG[sk] || STATUS_CONFIG.pending

const getOrderClass = (sk) => {
  if (sk === 'supplementing') return 'oc-w'
  if (sk === 'completed') return 'oc-d'
  return 'oc-a'
}

const isLoggedIn = computed(() => client.isLoggedIn)
const userInfo = computed(() => client.userInfo)
</script>

<template>
  <div id="scroll">

    <!-- ── HERO ── -->
    <div class="hero">
      <div class="hero-deco"></div>
      <div class="hero-badge">
        <div class="hero-dot"></div>
        一站式管家服务
      </div>
      <h1 class="hero-h1">
        Yes<span class="yk">ok</span> Vietnam
      </h1>
      <p class="hero-sub">在越南，所有事我们帮你搞定</p>

      <!-- quick search -->
      <div class="hero-search">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.8)" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <span style="font-size:13px;color:rgba(255,255,255,.75);">搜索服务关键词…</span>
      </div>

      <!-- manager card -->
      <div class="hero-mgr">
        <div class="hero-mgr-av">🏠</div>
        <div>
          <div style="font-size:13px;font-weight:700;color:#fff;">专属管家团队</div>
          <div style="font-size:11px;color:rgba(255,255,255,.7);margin-top:2px;">7×24小时 · 全程中文陪同</div>
        </div>
        <button style="margin-left:auto;padding:7px 13px;border-radius:18px;background:rgba(255,255,255,.2);border:1.5px solid rgba(255,255,255,.35);color:#fff;font-size:11px;font-weight:600;cursor:pointer;">立即咨询</button>
      </div>
    </div>

    <!-- ── STATS ROW ── -->
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

    <!-- ── 订单动态 ── -->
    <div style="padding:14px 16px 10px;">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
        <div style="display:flex;align-items:center;gap:6px;">
          <span style="font-size:15px;">📋</span>
          <span style="font-size:15px;font-weight:700;color:var(--tx);">订单动态</span>
        </div>
        <span style="font-size:12px;color:var(--bl);cursor:pointer;">查看全部 ›</span>
      </div>
    </div>

    <!-- orders list (dynamic) -->
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

    <!-- placeholder for bottom nav spacing -->
    <div style="height:12px;"></div>
  </div>
</template>

<style scoped>
/* All styles are inherited from style.css which contains the exact yesok-final.html CSS */
</style>
