import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    watch: {
      ignored: ['**/node_modules/**', '**/build/**']
    }
  },
  build: {
    rollupOptions: {
      external: ['/wails/ipc.js', '/wails/runtime.js']
    }
  }
})
