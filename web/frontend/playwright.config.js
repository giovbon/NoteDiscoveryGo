import { defineConfig, devices } from '@playwright/test';

const PORT = Number(process.env.E2E_PORT || 8123);
const baseURL = `http://127.0.0.1:${PORT}`;

// Smoke tests for the real Go server + bundled frontend. They deliberately run
// serially against one throwaway vault (see e2e/start-server.mjs) because the
// tests create and edit notes.
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL,
    serviceWorkers: 'block',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: 'node e2e/start-server.mjs',
    url: `${baseURL}/health`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
