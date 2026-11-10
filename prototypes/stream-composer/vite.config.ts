import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  base: './',
  build: { target: 'safari16', outDir: 'dist' },
  server: { host: '127.0.0.1', port: 5199, strictPort: true },
});
