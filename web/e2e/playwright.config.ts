import { defineConfig, devices } from '@playwright/test';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const e2eRoot = path.dirname(fileURLToPath(import.meta.url));
const webRoot = path.resolve(e2eRoot, '..');
const repoRoot = path.resolve(webRoot, '..');
const backendPort = process.env.E2E_BACKEND_PORT ?? '18080';
const frontendPort = process.env.E2E_FRONTEND_PORT ?? '15173';

export default defineConfig({
  testDir: './tests',
  testMatch: ['room/**/*.real.test.ts'],
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
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: repoRoot,
      url: `http://127.0.0.1:${backendPort}/api/health`,
      env: {
        ...process.env,
        PORT: backendPort,
      },
      reuseExistingServer: false,
      timeout: 120_000,
    },
    {
      command: `npm run dev -- --host 127.0.0.1 --port ${frontendPort} --strictPort`,
      cwd: webRoot,
      url: `http://127.0.0.1:${frontendPort}`,
      env: {
        ...process.env,
        VITE_WS_URL: `ws://127.0.0.1:${backendPort}/ws`,
      },
      reuseExistingServer: false,
      timeout: 120_000,
    },
  ],
});
