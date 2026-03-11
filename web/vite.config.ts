import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: '/yp/',
  build: {
    outDir: '../internal/httpd/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/yp/api': 'http://localhost:8080',
      '/yp/index.txt': 'http://localhost:8080',
    },
  },
})
