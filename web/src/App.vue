<script setup>
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'

// --- 只有在 H5 (Telegram) 环境下才引入这些 API ---
// #ifdef H5
import { loginWithTG, getMe } from '@/api/client/auth'
// #endif

const client = useClientStore()

onLaunch(async () => {
  console.log('[App] onLaunch')

  // ── H5 / Telegram 环境 ────────────────────────────────────────
  // #ifdef H5
  if (typeof window !== 'undefined' && window.Telegram?.WebApp) {
    const tg = window.Telegram.WebApp
    tg.ready()
    tg.expand()
    tg.setHeaderColor('#0D47A1')

    const initData = tg.initData
    if (initData && !client.isLoggedIn) {
      try {
        const data = await loginWithTG(initData)
        client.setToken(data.token)
        client.setUserInfo(data.user)
      } catch (e) {
        console.error('[App] Telegram login failed', e)
      }
    }
  } else {
    console.log('[App] 当前为 H5 环境但未检测到 Telegram WebApp，跳过 TG 登录')
  }
  // #endif

  // ── 非 H5 环境（微信小程序 / iOS 原生等）───────────────────────
  // #ifndef H5
  console.log('非 H5 环境，直接进入页面渲染')
  // #endif
})

onShow(() => {
  console.log('[App] onShow')
})

onHide(() => {
  console.log('[App] onHide')
})
</script>

<style>
/* 引入全局皮肤 */
@import "./style.css";
</style>