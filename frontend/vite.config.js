import { fileURLToPath, URL } from 'node:url';

import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  server: {
    proxy: {
      '/api': {
        target: 'https://oklahoma.aredn.mcswain.cloud',
        changeOrigin: true,
      },
      '/ws': {
        target: 'wss://oklahoma.aredn.mcswain.cloud',
        ws: true,
        changeOrigin: true,
      },
      '/ws/events': {
        target: 'wss://oklahoma.aredn.mcswain.cloud',
        ws: true,
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    reporter: ['junit', 'html', 'default'],
    outputFile: {
      junit: 'reports/unit/junit.xml',
      html: 'reports/unit/index.html',
    },
  },
});
