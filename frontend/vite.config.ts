/// <reference types="vitest" />
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [svelte()],
  server: {
    port: 5173,
    proxy: {
      // Proxy API calls to the Go backend during development.
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    // Debug mode keeps sourcemaps; release omits them.
    sourcemap: mode !== 'production',
    minify: mode === 'production',
    rollupOptions: {
      output: {
        manualChunks: undefined,
      },
    },
  },
  test: {
    environment: 'node',
    include: ['tests/**/*.test.ts'],
  },
}))
