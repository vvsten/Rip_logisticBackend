import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  // Для GitHub Pages используем VITE_BASE_URL или имя репозитория
  base: process.env.VITE_BASE_URL || (process.env.NODE_ENV === 'production' ? '/RIP-2-mod-/' : '/'),
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'mask-icon.svg'],
      manifest: {
        name: 'GruzDelivery - Грузоперевозки',
        short_name: 'GruzDelivery',
        description: 'Приложение для заказа грузоперевозок',
        theme_color: '#0d6efd',
        background_color: '#f8f9fa',
        display: 'standalone',
        orientation: 'portrait',
        scope: process.env.VITE_BASE_URL || (process.env.NODE_ENV === 'production' ? '/RIP-2-mod-/' : '/'),
        start_url: process.env.VITE_BASE_URL || (process.env.NODE_ENV === 'production' ? '/RIP-2-mod-/' : '/'),
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png'
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable'
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/api\./i,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 60 * 60 * 24 // 24 часа
              }
            }
          }
        ]
      }
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    // Прокси для обхода CORS - перенаправляем /api на бэкенд
    proxy: {
      '/api': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        ws: true,
      },
      // Auth endpoints (swagger-based): в dev должны проксироваться на бэкенд
      '/login': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        ws: true,
      },
      '/sign_up': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        ws: true,
      },
      '/logout': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        ws: true,
      },
      '/refresh': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        ws: true,
      },
      // Прокси для MinIO изображений через бэкенд
      '/lab1': {
        target: 'http://localhost:8083',
        changeOrigin: true,
      },
      // HTML-страницы, которые рендерит бэкенд (не React-роуты!)
      '/delivery-quote': {
        target: 'http://localhost:8083',
        changeOrigin: true,
      },
      // ВАЖНО: проксируем только quote-страницу, чтобы не перехватывать SPA-роут
      // `/logistic-requests` (иначе при refresh получаем пустой ответ от бэкенда).
      '/logistic-request/quote': {
        target: 'http://localhost:8083',
        changeOrigin: true,
      },
      // Прокси для статических файлов бэкенда (CSS, JS, изображения)
      '/static': {
        target: 'http://localhost:8083',
        changeOrigin: true,
      }
    }
  }
})
