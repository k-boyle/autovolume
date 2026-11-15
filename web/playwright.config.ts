import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: 'e2e',
  forbidOnly: !!process.env.CI,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: 'http://localhost:4173',
  },
  webServer: {
    command: 'npx vite build --mode mock --outDir e2e-dist && npx vite preview --outDir e2e-dist --port 4173 --strictPort',
    url: 'http://localhost:4173',
  },
})
