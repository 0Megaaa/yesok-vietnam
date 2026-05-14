<template>
  <view id="app"></view>
</template>

<script setup>
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'
import { loginWithTG, getMe } from '@/api/client/auth'

const client = useClientStore()

onLaunch(async () => {
  console.log('[App] onLaunch')

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

onShow(() => {
  console.log('[App] onShow')
})

onHide(() => {
  console.log('[App] onHide')
})
</script>

<style>
/* All global styles are sourced from style.css */
</style>
