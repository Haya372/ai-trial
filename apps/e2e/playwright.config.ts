import { defineConfig, devices } from '@playwright/test'

const webPort = 4173
const backendPort = 8080
const isCI = !!process.env.CI

export default defineConfig({
  testDir: './tests',
  forbidOnly: isCI,
  retries: isCI ? 1 : 0,
  reporter: isCI ? [['github'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: `http://localhost:${webPort}`,
    locale: 'ja-JP',
    trace: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: '../backend',
      url: `http://localhost:${backendPort}/health`,
      env: { PORT: String(backendPort) },
      reuseExistingServer: !isCI,
    },
    {
      command: `pnpm --filter web exec vite preview --port ${webPort} --strictPort`,
      url: `http://localhost:${webPort}`,
      reuseExistingServer: !isCI,
    },
  ],
})
