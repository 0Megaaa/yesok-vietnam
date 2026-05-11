<script setup>
import { useUserStore } from '../store'

const userStore = useUserStore()
</script>

<template>
  <div class="profile">
    <h1>个人中心</h1>
    <div v-if="userStore.isLoggedIn" class="info">
      <p><strong>用户名：</strong>{{ userStore.userInfo?.username }}</p>
      <p><strong>角色：</strong>{{ userStore.userInfo?.role }}</p>
      <p><strong>余额：</strong>{{ userStore.userInfo?.balance }}</p>
    </div>
    <div v-else>
      <p>请先登录</p>
    </div>
  </div>
</template>

<style scoped>
.profile {
  padding: 2rem;
}
.info p {
  margin: 0.5rem 0;
}
</style>
