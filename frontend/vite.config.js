import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      // 将以 /api 开头的请求代理到后端服务器
      '/api': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        // 重写路径：去掉 /api 前缀（如果后端路由也带 /api 就不需要重写）
        // 我们的后端路由是 /api/xxx，所以不需要重写
        // rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})