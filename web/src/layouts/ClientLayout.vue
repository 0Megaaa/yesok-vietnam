<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useClientStore } from '@/store/client'

const client = useClientStore()
const route = useRoute()
</script>

<template>
  <div class="client-layout">
    <header class="client-header">
      <nav class="client-nav">
        <RouterLink to="/" :class="{ active: route.path === '/' }">首页</RouterLink>
        <RouterLink to="/profile" :class="{ active: route.path === '/profile' }">个人中心</RouterLink>
        <button v-if="client.isLoggedIn" class="logout-btn" @click="client.logout()">退出</button>
      </nav>
      <div v-if="client.isLoggedIn && client.userInfo" class="user-badge">
        {{ client.userInfo.username }} · {{ client.userInfo.role }}
      </div>
    </header>
    <main class="client-main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.client-layout {
  min-height: 100svh;
  display: flex;
  flex-direction: column;
}
.client-header {
  padding: 1rem 2rem;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}
.client-nav {
  display: flex;
  gap: 1.5rem;
  align-items: center;
}
.client-nav a {
  color: var(--text);
  text-decoration: none;
  font-size: 1rem;
  padding: 0.25rem 0;
  border-bottom: 2px solid transparent;
  transition: color 0.2s, border-color 0.2s;
}
.client-nav a:hover,
.client-nav a.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}
.logout-btn {
  background: none;
  border: 1px solid var(--border);
  color: var(--text);
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background 0.2s;
}
.logout-btn:hover {
  background: var(--code-bg);
}
.user-badge {
  font-size: 0.875rem;
  color: var(--text);
  opacity: 0.7;
}
.client-main {
  flex: 1;
  padding: 2rem;
}
</style>
