import { defineConfig, devices } from '@playwright/test';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const e2eRoot = path.dirname(fileURLToPath(import.meta.url));
const webRoot = path.resolve(e2eRoot, '..');
const frontendPort = process.env.E2E_HARNESS_FRONTEND_PORT ?? '15174';

export default defineConfig({
  testDir: './tests',
  testMatch: ['skills/**/*.mock.test.ts'],
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: `http://127.0.0.1:${frontendPort}`,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: `npm run dev -- --host 127.0.0.1 --port ${frontendPort} --strictPort`,
    cwd: webRoot,
    url: `http://127.0.0.1:${frontendPort}`,
    env: {
      ...process.env,
      VITE_WS_URL: 'ws://127.0.0.1:19090/ws',
    },
    reuseExistingServer: false,
    timeout: 120_000,
  },
});
