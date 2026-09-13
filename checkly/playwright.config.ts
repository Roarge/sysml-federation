import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './__checks__',
  testMatch: 'us*.spec.ts',
  fullyParallel: false,
  workers: process.env.CHECKLY ? 4 : 1,
  reporter: [['html', { outputFolder: 'playwright-report', open: 'never' }], ['list']],
  outputDir: 'test-results',
  use: {
    baseURL: process.env.DEMO_URL,
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
