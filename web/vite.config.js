import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  build: {
    outDir: '../server/web/dist',
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:7625',
        changeOrigin: true,
      },
    },
  },
})
