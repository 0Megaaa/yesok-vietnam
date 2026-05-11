<script setup>
import { useClientStore } from '../../store/client'

const client = useClientStore()
</script>

<template>
  <div class="profile">
    <h1>个人中心</h1>
    <div v-if="client.isLoggedIn && client.userInfo" class="info">
      <div class="field"><span class="label">用户名</span><span class="value">{{ client.userInfo.username }}</span></div>
      <div class="field"><span class="label">角色</span><span class="value">{{ client.userInfo.role }}</span></div>
      <div class="field"><span class="label">余额</span><span class="value">{{ client.userInfo.balance }}</span></div>
      <div class="field"><span class="label">语言</span><span class="value">{{ client.userInfo.language }}</span></div>
      <div class="field"><span class="label">头像</span>
        <img v-if="client.userInfo.avatar_url" :src="client.userInfo.avatar_url" class="avatar" alt="avatar" />
        <span v-else class="no-avatar">无</span>
      </div>
    </div>
    <div v-else>
      <p class="not-logged-in">请先通过 Telegram 登录</p>
    </div>
  </div>
</template>

<style scoped>
.profile {
  max-width: 480px;
}
.info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1.5rem;
}
.field {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border);
}
.label {
  width: 80px;
  font-size: 0.875rem;
  color: var(--text);
  flex-shrink: 0;
}
.value {
  color: var(--text-h);
  font-size: 0.9375rem;
}
.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
}
.no-avatar {
  color: var(--text);
  opacity: 0.5;
  font-size: 0.875rem;
}
.not-logged-in {
  color: var(--text);
  opacity: 0.6;
  margin-top: 1rem;
}
</style>
