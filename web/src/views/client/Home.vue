<script setup>
import { useClientStore } from '../../store/client'

const client = useClientStore()
</script>

<template>
  <div class="home">
    <h1>Yesok Vietnam</h1>
    <div v-if="client.isLoggedIn && client.userInfo" class="user-info">
      <p>欢迎，<strong>{{ client.userInfo.username }}</strong></p>
      <p>余额：{{ client.userInfo.balance }}</p>
      <p>角色：{{ client.userInfo.role }}</p>
      <button class="logout-btn" @click="client.logout()">退出登录</button>
    </div>
    <div v-else class="loading">
      <p>正在验证身份...</p>
      <p class="hint">请在 Telegram WebApp 中打开此页面</p>
    </div>
  </div>
</template>

<style scoped>
.home {
  max-width: 480px;
}
.user-info p {
  margin: 0.5rem 0;
  font-size: 1rem;
}
.user-info strong {
  color: var(--text-h);
}
.logout-btn {
  margin-top: 1rem;
  background: var(--accent-bg);
  border: 1px solid var(--accent-border);
  color: var(--accent);
  padding: 0.5rem 1.25rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9375rem;
}
.loading {
  color: var(--text);
}
.hint {
  font-size: 0.875rem;
  opacity: 0.6;
  margin-top: 0.25rem;
}
</style>
