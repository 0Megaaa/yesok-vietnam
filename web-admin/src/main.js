import { createSSRApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'
// createApp 创建 UniApp 原生应用实例。
// 1.意图 -> 恢复 pages.json 原生路由与标准 Tabbar，不再使用 H5PreviewRoot 接管页面。
// 2.步骤 -> 创建 SSR App、安装 Pinia，保持 B 端管理后台独立运行。
// 3.返回 -> UniApp 要求的 app 实例。
export function createApp() {
  const app = createSSRApp(App)
  const pinia = createPinia()
  app.use(pinia)
  return { app }
}
