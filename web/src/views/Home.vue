<script setup>
import { useUserStore } from '../store'

const userStore = useUserStore()
</script>

<template>
  <div class="home">
    <h1>Yesok Vietnam</h1>
    <div v-if="userStore.isLoggedIn && userStore.userInfo" class="user-info">
      <p>欢迎，{{ userStore.userInfo.username }}</p>
      <p>余额：{{ userStore.userInfo.balance }}</p>
      <button @click="userStore.logout()">退出登录</button>
    </div>
    <div v-else class="loading">
      <p>正在验证身份...</p>
    </div>
  </div>
</template>

<style scoped>
.home {
  padding: 2rem;
}
.user-info p {
  margin: 0.5rem 0;
}
.loading {
  color: #888;
}
</style>
