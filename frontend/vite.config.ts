import path from "path"
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  css: {
    lightningcss: {
      customAtRules: {
        theme: {
          prelude: '<custom-ident>',
          body: 'style-block',
        },
        utility: {
          prelude: '*',
          body: 'style-block',
        },
        'custom-variant': {
          prelude: '<custom-ident>',
          body: 'style-block',
        },
        slot: {
          prelude: null,
          body: null,
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
