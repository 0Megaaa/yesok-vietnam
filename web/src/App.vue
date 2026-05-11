<script setup>
import { onMounted } from 'vue'
import { useClientStore } from '@/store/client'
import { loginWithTG, getMe } from '@/api/client/auth'

const client = useClientStore()

onMounted(async () => {
  // Only attempt Telegram auth when running inside the Telegram WebApp environment.
  const tg = window.Telegram?.WebApp
  if (!tg) {
    // Not in TG — this is a normal browser session.
    // If we have a stored client token, refresh user info silently.
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
    // Re-hydrate user info on revisit.
    try {
      const userInfo = await getMe()
      client.setUserInfo(userInfo)
    } catch {
      client.logout()
    }
    return
  }

  // First visit — exchange Telegram initData for a JWT.
  try {
    const data = await loginWithTG(initData)
    client.setToken(data.token)
    client.setUserInfo(data.user)
  } catch (e) {
    console.error('[App] Telegram login failed', e)
  }
})
</script>

<template>
  <RouterView />
</template>
