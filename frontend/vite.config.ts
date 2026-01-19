import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    watch: {
      ignored: ['**/node_modules/**', '**/build/**']
    },
    hmr: {
      protocol: 'ws',
      host: 'localhost',
      port: 5173
    }
  },
  build: {
    rollupOptions: {
      input: {
        main: new URL('./index.html', import.meta.url).pathname
      },
      external: ['/wails/ipc.js', '/wails/runtime.js']
    }
  }
})
