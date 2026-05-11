<script setup>
import { onMounted } from 'vue'
import { useUserStore } from './store'
import { loginWithTG, getUserInfo } from './api'

const userStore = useUserStore()

onMounted(async () => {
  const tg = window.Telegram?.WebApp
  if (!tg) {
    console.warn('[App] Telegram WebApp not available')
    return
  }

  tg.ready()
  tg.expand()

  const initData = tg.initData
  if (!initData) {
    console.warn('[App] No initData from Telegram')
    return
  }

  if (userStore.isLoggedIn) {
    try {
      const userInfo = await getUserInfo()
      userStore.setUserInfo(userInfo)
    } catch {
      userStore.logout()
    }
    return
  }

  try {
    const data = await loginWithTG(initData)
    userStore.setToken(data.token)
    userStore.setUserInfo(data.user)
  } catch (e) {
    console.error('[App] Telegram login failed', e)
  }
})
</script>

<template>
  <div id="app">
    <router-view />
  </div>
</template>
