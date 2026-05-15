<script setup>
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'

// #ifdef H5
import { loginWithTG } from '@/api/client/auth'
// #endif

const client = useClientStore()

// onLaunch 是应用启动入口。
// 实现步骤：
// 1. 先加载 Mock 聚合状态，保证首页、服务、订单等页面无需真实后端即可渲染。
// 2. H5 环境检测 Telegram WebApp 对象，并保留无感登录入口。
// 3. 当前不校验真实 Telegram initData，仅通过 Mock 请求返回演示用户，后续接入时由后端完成签名验证。
onLaunch(() => {
  console.log('[App] onLaunch')

  ;(async () => {
    try {
      await client.initMockState()
    } catch (error) {
      console.error('[App] init mock state failed', error)
    }
  })()

  // #ifdef H5
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
        } catch (error) {
          console.error('[App] Telegram login placeholder failed', error)
        }
      }
    }
  })()
  // #endif

  // #ifndef H5
  console.log('非 H5 环境，跳过 TG 登录，保留微信小程序授权入口。')
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
@import './style.css';
</style>
