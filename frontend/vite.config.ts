import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // Ходим на бэк через прокси, чтобы в деве не упираться в CORS.
      '/v1': 'http://localhost:8080',
    },
  },
})
