import { createSSRApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'
import { useClientStore } from './store/client'

// installGlobalAuth 注入全局鉴权方法。
// 1.意图 -> 保留 AuthPopup 与历史页面依赖的 $checkAuth，不破坏登录底层逻辑。
// 2.步骤 -> 从 Pinia 客户端 store 读取 checkAuth 并挂载到 Vue 全局属性。
// 3.返回 -> 无返回值。
function installGlobalAuth(app, pinia) {
  const client = useClientStore(pinia)
  app.config.globalProperties.$checkAuth = client.checkAuth
}

// createApp 创建 UniApp 原生应用实例。
// 1.意图 -> 恢复 pages.json 原生路由与标准 Tabbar，不再使用 H5PreviewRoot 接管页面。
// 2.步骤 -> 创建 SSR App、安装 Pinia、注入全局鉴权。
// 3.返回 -> UniApp 要求的 app 实例。
export function createApp() {
  const app = createSSRApp(App)
  const pinia = createPinia()
  app.use(pinia)
  installGlobalAuth(app, pinia)
  return { app }
}
