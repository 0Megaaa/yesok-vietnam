<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAdminStore } from '@/store/admin'

const admin = useAdminStore()
const route = useRoute()
</script>

<template>
  <div id="adm-ov" class="on">
    <div class="adm-shell">
      <!-- Sidebar -->
      <div class="adm-side">
        <div class="adm-logo">
          <div class="adm-logo-ic">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="#0D47A1"/>
              <path d="M2 17l10 5 10-5M2 12l10 5 10-5" stroke="#0D47A1" fill="none" stroke-width="2"/>
            </svg>
          </div>
          <div>
            <div class="adm-logo-nm">Yes<span>ok</span></div>
            <div class="adm-logo-sb">管理后台</div>
          </div>
        </div>

        <div class="adm-nav">
          <RouterLink to="/admin" :class="['adm-ni', { on: route.path === '/admin' }]">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7"/>
              <rect x="14" y="3" width="7" height="7"/>
              <rect x="14" y="14" width="7" height="7"/>
              <rect x="3" y="14" width="7" height="7"/>
            </svg>
            数据看板
          </RouterLink>
          <RouterLink to="/admin/orders" :class="['adm-ni', { on: route.path === '/admin/orders' }]">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="5" y="2" width="14" height="20" rx="2"/>
              <line x1="9" y1="7" x2="15" y2="7"/>
              <line x1="9" y1="12" x2="15" y2="12"/>
            </svg>
            订单管理
          </RouterLink>
          <RouterLink to="/admin/users" :class="['adm-ni', { on: route.path === '/admin/users' }]">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
            </svg>
            用户管理
          </RouterLink>
        </div>

        <RouterLink to="/" class="adm-exit">
          ← 返回客户端
        </RouterLink>
      </div>

      <!-- Main content -->
      <main class="adm-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
/* Styles from style.css are inherited globally — adm-* classes are defined there */
</style>
