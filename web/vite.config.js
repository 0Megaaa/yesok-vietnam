import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'
import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

// yesokUniH5MountFix 提供 Uni H5 入口兜底转换。
// 1.意图 -> 解决当前脚手架 src 目录结构下官方 main.js 转换器未命中导致的 H5 静默白屏。
// 2.步骤 -> 在 H5 开发与构建时为 src/main.js 注入 pages.json 路由、uni-h5 插件与标准挂载语句。
// 3.返回 -> Vite 插件对象，保持源码 main.js 只导出 createApp，不手写 app.mount。
function yesokUniH5MountFix() {
  let mainEntry = ''

  return {
    name: 'yesok:uni-h5-mount-fix',
    enforce: 'post',
    configResolved() {
      const inputDir = process.env.UNI_INPUT_DIR || path.resolve(process.cwd(), 'src')
      mainEntry = path.resolve(inputDir, 'main.js').split(path.sep).join('/')
    },
    transform(code, id) {
      const normalizedId = id.split('?')[0].split(path.sep).join('/')
      if (normalizedId !== mainEntry || code.includes('createApp().app.use(__plugin).mount("#app")')) {
        return null
      }

      const patchedCode = [
        "import './pages-json-js'",
        "import { plugin as __plugin } from '@dcloudio/uni-h5'",
        code.replace('createSSRApp', 'createVueApp as createSSRApp'),
        'createApp().app.use(__plugin).mount("#app")'
      ].join('\n')

      return {
        code: patchedCode,
        map: null
      }
    }
  }
}

export default defineConfig({
  resolve: {
    alias: [
      {
        // 源码路径别名：
        // 1. 统一支持 @/store、@/api 等业务模块引用。
        // 2. 兼容 UniApp 官方构建与普通 H5 浏览器预览。
        // 3. 避免在多端源码中出现过长的相对路径。
        find: '@',
        replacement: fileURLToPath(new URL('./src', import.meta.url))
      },
      {
        // Vue 兼容说明：
        // 1. 仅拦截裸导入 vue，避免影响 @vue/shared、编译器和子路径解析。
        // 2. 通过本地适配层补齐 UniApp 当前发行包需要的私有 SSR/生命周期导出。
        // 3. 保持业务代码继续按标准 Vue API 编写，后续升级 UniApp/Vue 后可删除此别名。
        find: /^vue$/,
        replacement: fileURLToPath(new URL('./src/shims/vue-compat.js', import.meta.url))
      }
    ]
  },
  plugins: [
    // UniApp 官方 Vite 插件：
    // 1. 读取 pages.json 生成 H5/小程序路由。
    // 2. 转译 <view>/<text> 等跨端组件。
    // 3. 保持微信小程序、H5 与后续 App 端共用同一套页面源码。
    uni(),
    yesokUniH5MountFix()
  ],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      // 本地联调代理：
      // 1.意图 -> H5 预览保持 /api 相对路径，避免跨域与硬编码公网地址。
      // 2.步骤 -> 将 /api 请求转发到 Go 后端实际监听端口。
      // 3.返回 -> Vite 代理配置对象。
      '/api': {
        target: process.env.VITE_PROXY_TARGET || 'http://127.0.0.1:8080',
        changeOrigin: true
      }
    }
  }
})
