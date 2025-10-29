import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react-swc'
import { URL, fileURLToPath } from 'node:url'
import { visualizer } from 'rollup-plugin-visualizer'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    visualizer({
      filename: 'stats.html',
      open: true,
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    watch: {
      ignored: ['**/e2e/**', '**/tests/**'],
    },
    host: '0.0.0.0',
    port: 3000,
    open: true,
    proxy: {
      '/api': 'http://localhost:8000',
      '/shared': 'http://localhost:8000',
      '/static': 'http://localhost:8000',
      '/uploads': 'http://localhost:8000',
    },
  },
})
