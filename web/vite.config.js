import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

export default defineConfig({
  resolve: {
    alias: [
      {
        // 兼容说明：
        // 1. 仅拦截裸导入 vue，避免影响 vue 子路径和编译器包解析。
        // 2. 通过本地适配层补齐 UniApp 当前版本需要的私有 SSR 导出。
        // 3. 保持业务代码继续按标准 Vue API 编写，后续升级依赖可平滑移除。
        find: /^vue$/,
        replacement: fileURLToPath(new URL('./src/shims/vue-compat.js', import.meta.url))
      }
    ]
  },
  plugins: [
    uni()
  ]
})
