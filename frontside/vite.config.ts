import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    tailwindcss(),
    svelte(),
  ],
  build: {
    outDir: '../backend/internal/web/static',
    emptyOutDir: true,
    minify: true,
    cssMinify: true,
  },
  server: {
    port: 5174,
    proxy: {
      // 所有 /api 請求轉發到 Go
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        //rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
