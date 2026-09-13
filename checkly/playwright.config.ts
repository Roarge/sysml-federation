import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './__checks__',
  testMatch: 'us*.spec.ts',
  fullyParallel: false,
  // The demo has one shared in-memory state, and two specs editing it at once
  // would see each other's edits, so the suite runs one file at a time
  // everywhere.
  workers: 1,
  reporter: [['html', { outputFolder: 'playwright-report', open: 'never' }], ['list']],
  outputDir: 'test-results',
  use: {
    baseURL: process.env.DEMO_URL,
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
