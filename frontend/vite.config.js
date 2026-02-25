import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Vite config keeps the dev server simple for local usage.
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173
  }
})
