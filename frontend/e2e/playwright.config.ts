import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30 * 1000,
  use: {
    headless: true,
    baseURL: 'http://localhost:8080/ui',
    viewport: { width: 1280, height: 720 },
  },
});
