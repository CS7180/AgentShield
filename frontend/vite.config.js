import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      tslib: fileURLToPath(new URL('./src/vendor/tslib.js', import.meta.url)),
    },
  },
})
