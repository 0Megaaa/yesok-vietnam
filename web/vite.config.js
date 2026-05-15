import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

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
    uni()
  ]
})
