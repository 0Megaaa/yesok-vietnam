<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAdminStore } from '@/store/admin'

const admin = useAdminStore()
const route = useRoute()
</script>

<template>
  <div class="admin-layout">
    <!-- Sidebar -->
    <aside class="admin-sidebar">
      <div class="sidebar-brand">Yesok Admin</div>
      <nav class="sidebar-nav">
        <RouterLink to="/admin" :class="{ active: route.path === '/admin' }">
          仪表盘
        </RouterLink>
        <RouterLink to="/admin/users" :class="{ active: route.path === '/admin/users' }">
          用户管理
        </RouterLink>
        <RouterLink to="/admin/orders" :class="{ active: route.path === '/admin/orders' }">
          订单管理
        </RouterLink>
      </nav>
      <div class="sidebar-footer">
        <div v-if="admin.isLoggedIn" class="admin-info">
          <span>{{ admin.userInfo?.username }}</span>
          <button class="logout-btn" @click="admin.logout()">退出</button>
        </div>
        <RouterLink v-else to="/admin/login" class="login-link">登录</RouterLink>
      </div>
    </aside>

    <!-- Main content -->
    <main class="admin-main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100svh;
}
.admin-sidebar {
  width: 220px;
  flex-shrink: 0;
  background: var(--bg);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 0;
}
.sidebar-brand {
  padding: 1.5rem 1.25rem 1rem;
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-h);
  border-bottom: 1px solid var(--border);
}
.sidebar-nav {
  flex: 1;
  padding: 1rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.sidebar-nav a {
  display: block;
  padding: 0.625rem 1.25rem;
  color: var(--text);
  text-decoration: none;
  font-size: 0.9375rem;
  border-left: 3px solid transparent;
  transition: color 0.15s, background 0.15s, border-color 0.15s;
}
.sidebar-nav a:hover {
  color: var(--accent);
  background: var(--accent-bg);
}
.sidebar-nav a.active {
  color: var(--accent);
  border-left-color: var(--accent);
  background: var(--accent-bg);
}
.sidebar-footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid var(--border);
}
.admin-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.875rem;
}
.logout-btn {
  background: none;
  border: 1px solid var(--border);
  color: var(--text);
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8125rem;
  transition: background 0.2s;
}
.logout-btn:hover {
  background: var(--code-bg);
}
.login-link {
  color: var(--accent);
  text-decoration: none;
  font-size: 0.9375rem;
}
.admin-main {
  flex: 1;
  overflow: auto;
  padding: 2rem;
}
</style>
