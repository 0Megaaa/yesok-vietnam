<script setup>
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'

// #ifdef H5
import { loginWithTG, getMe } from '@/api/client/auth'
// #endif

const client = useClientStore()

// 关键点：这里绝对不能写 onLaunch(async () => {...})
onLaunch(() => {
  console.log('[App] onLaunch')

  // #ifdef H5
  // H5 和 TG 环境的异步逻辑，用自执行函数隔离，绝对不阻塞主线程
  ;(async () => {
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
    }
  })();
  // #endif

  // #ifndef H5
  console.log('非 H5 环境，跳过 TG 登录，瞬间放行！')
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