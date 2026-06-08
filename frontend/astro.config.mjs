// @ts-check

import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig, envField } from 'astro/config';

// https://astro.build/config
export default defineConfig({
  env: {
    schema: {
      PUBLIC_API_URL: envField.string({
        context: 'client',
        access: 'public',
        default: 'http://localhost:8080',
      }),
    },
  },
  vite: {
    plugins: [tailwindcss()],
  },
  integrations: [react()],
});
