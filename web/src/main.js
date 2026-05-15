import { createSSRApp, createApp as createBrowserApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import App from './App.vue'
import H5PreviewRoot from './H5PreviewRoot.vue'

export function createApp() {
  const app = createSSRApp(App)
  app.use(createPinia())
  return { app }
}

// H5 浏览器预览兜底挂载。
// 实现步骤：
// 1. 判断当前是否运行在浏览器环境，避免影响微信小程序与后续 App 端编译。
// 2. 检查 #app 是否尚未被 UniApp 官方运行时挂载，避免重复挂载产生副作用。
// 3. 使用 H5PreviewRoot 直接渲染首页与全局登录弹窗，保证演示预览能看到真实 UI 模块。
if (typeof document !== 'undefined') {
  const appContainer = document.querySelector('#app')
  const hasMountedContent = Boolean(appContainer?.innerHTML?.trim())

  if (appContainer && !hasMountedContent) {
    const previewApp = createBrowserApp(H5PreviewRoot)
    // H5 预览错误捕获器。
    // 实现步骤：
    // 1. 捕获 Vue 组件 setup/render 阶段异常。
    // 2. 将错误输出到控制台，便于本地预览定位白屏。
    // 3. 不影响微信小程序与 App 端正式编译链路。
    previewApp.config.errorHandler = (error, instance, info) => {
      console.error('[YesOK H5 Preview Error]', info, error)
    }
    previewApp.use(createPinia())
    previewApp.mount(appContainer)
  }
}
