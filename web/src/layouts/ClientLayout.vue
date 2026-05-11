<script setup>
import { RouterLink, useRoute } from 'vue-router'
import { useClientStore } from '@/store/client'

const client = useClientStore()
const route = useRoute()
</script>

<template>
  <div id="app">
    <!-- Header -->
    <div class="tb-main">
      <div class="logo-wrap">
        <div class="logo-ic">
          <svg viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2">
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
          </svg>
        </div>
        <div class="logo-tx-wrap">
          <span class="logo-tx"><span class="yes">Yes</span><span class="ok">ok</span></span>
        </div>
      </div>
      <div class="tb-ics">
        <div class="tb-btn">🔔</div>
      </div>
    </div>

    <!-- Scrollable content area -->
    <div id="scroll">
      <RouterView />
    </div>

    <!-- Bottom Navigation -->
    <nav id="botnav">
      <RouterLink to="/" :class="['bni', { on: route.path === '/' }]">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
        </svg>
        <span class="bni-lb">首页</span>
        <div class="bni-dot"></div>
      </RouterLink>

      <RouterLink to="/profile" :class="['bni', { on: route.path === '/profile' }]">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
          <circle cx="12" cy="7" r="4"/>
        </svg>
        <span class="bni-lb">我的</span>
        <div class="bni-dot"></div>
      </RouterLink>
    </nav>

    <!-- FAB: New Order -->
    <RouterLink to="/profile" id="fab" title="提交新需求">+</RouterLink>
  </div>
</template>

<style scoped>
/* Styles from style.css are inherited globally */
</style>
