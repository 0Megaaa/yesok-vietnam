<template>
  <RouterView />
</template>

<script setup>
import { onMounted } from 'vue'
import { useClientStore } from '@/store/client'
import { loginWithTG, getMe } from '@/api/client/auth'

const client = useClientStore()

onMounted(async () => {
  const tg = window.Telegram?.WebApp
  if (!tg) {
    if (client.isLoggedIn) {
      try {
        const userInfo = await getMe()
        client.setUserInfo(userInfo)
      } catch {
        client.logout()
      }
    }
    return
  }

  tg.ready()
  tg.expand()
  tg.setHeaderColor('#0D47A1')

  const initData = tg.initData
  if (!initData) {
    console.warn('[App] No initData from Telegram WebApp')
    return
  }

  if (client.isLoggedIn) {
    try {
      const userInfo = await getMe()
      client.setUserInfo(userInfo)
    } catch {
      client.logout()
    }
    return
  }

  try {
    const data = await loginWithTG(initData)
    client.setToken(data.token)
    client.setUserInfo(data.user)
  } catch (e) {
    console.error('[App] Telegram login failed', e)
  }
})
</script>

<style>
/* 严禁在这里写任何样式，确保所有样式来自全局 style.css */
</style>
