import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import tailwindcss from '@tailwindcss/vite'
import path from 'path';
import fs from 'fs';
import crypto from 'crypto';

// Custom plugin to add integrity attributes
function addIntegrity() {
  return {
    name: 'vite-plugin-integrity',
    enforce: 'post',
    apply: 'build',
    closeBundle() {
      const distPath = path.resolve(process.cwd(), 'dist/assets');
      const files = fs.readdirSync(distPath);

      files.forEach((file) => {
        if (file.endsWith('.js') || file.endsWith('.css')) {
          const filePath = path.join(distPath, file);
          const fileContent = fs.readFileSync(filePath);
          const hash = crypto.createHash('sha384').update(fileContent).digest('base64');

          const indexHtmlPath = path.resolve(process.cwd(), 'dist/index.html');
          let indexHtmlContent = fs.readFileSync(indexHtmlPath, 'utf-8');

          const scriptTag = file.endsWith('.js')
            ? `<script type="module" src="/assets/${file}" integrity="sha384-${hash}" crossorigin="anonymous">`
            : `<link rel="stylesheet" href="/assets/${file}" integrity="sha384-${hash}" crossorigin="anonymous">`;

          indexHtmlContent = indexHtmlContent.replace(
            new RegExp(`<script\\s+type="module"\\scrossorigin\\s+src="/assets/${file}"[^>]*>|<link\\s+rel="stylesheet"\\scrossorigin\\s+href="/assets/${file}"[^>]*>`, 'g'),
            scriptTag
          );

          fs.writeFileSync(indexHtmlPath, indexHtmlContent, 'utf-8');
        }
      });
    }
  };
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
    tailwindcss(),
    addIntegrity(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  build: {
    sourcemap: true
  }
})