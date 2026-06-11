import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { TDesignResolver } from '@tdesign-vue-next/auto-import-resolver';
import path from 'path';

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    AutoImport({
      resolvers: [
        TDesignResolver({
          library: 'vue-next',
        }),
      ],
    }),
    Components({
      resolvers: [
        TDesignResolver({
          library: 'vue-next',
        }),
      ],
    }),
  ],
  optimizeDeps: {
    exclude: ['@xkit/hotkeys'],
  },
  resolve: {
    dedupe: ['vue'],
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@xkit/hotkeys': path.resolve(__dirname, '../../hotkeys/src/index.ts'),
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    cssMinify: 'esbuild',
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (id.includes('node_modules')) {
            if (id.includes('tdesign-vue-next')) {
              if (
                id.includes('/date-picker/') ||
                id.includes('/calendar/') ||
                id.includes('/time-picker/')
              ) {
                return 'tdesign-date';
              }
              if (id.includes('/select/') || id.includes('/input/')) {
                return 'tdesign-form';
              }
              return 'tdesign-core';
            }
            if (id.includes('marked') || id.includes('dompurify'))
              return 'markdown';
            return 'vendor';
          }
        },
      },
    },
  },
  server: {
    fs: {
      allow: [path.resolve(__dirname, '../..')],
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
