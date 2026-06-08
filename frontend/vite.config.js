import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      // '@' теперь указывает на папку src — удобно для импортов
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    allowedHosts: ['jaronimo.ru'],
    port: 5173,
    hmr: {
      // для WebSocket HMR через Caddy
      clientPort: 443,
      protocol: 'wss'
    }
  }
})
