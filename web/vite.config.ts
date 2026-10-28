import { fileURLToPath, URL } from 'node:url'
import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

// The dev server proxies the Abudemy REST API so the browser talks to the
// backend on the same origin (avoids CORS during development).
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const target = env.VITE_PROXY_TARGET || 'http://localhost:8080'

  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
server: {
      port: 5173,
      strictPort: false,
      allowedHosts: ['.ngrok-free.dev'],
      hmr: {
        clientPort: 443,
      },
      proxy: {
        '/api': { target, changeOrigin: true },
        '/healthz': { target, changeOrigin: true },
        '/readyz': { target, changeOrigin: true },
      },
    },

    build: {
      outDir: 'dist',
      sourcemap: false,
      rollupOptions: {
        output: {
          manualChunks: {
            'vendor-react': ['react', 'react-dom', 'react-router-dom'],
            'vendor-query': ['@tanstack/react-query'],
            'vendor-charts': ['recharts'],
            'vendor-maps': ['leaflet', 'react-leaflet', 'leaflet.heat'],
            'vendor-pdf': ['jspdf', 'jspdf-autotable'],
          },
        },
      },
    },
    test: {
      environment: 'node',
      include: ['src/**/*.test.ts'],
    },
  }
})
