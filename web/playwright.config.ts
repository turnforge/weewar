import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for GameViewerPage integration tests
 * Tests the real production GameViewerPage via actual server endpoints
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false, // Sequential execution for game state isolation
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // Single worker to avoid test game conflicts
  reporter: [
    ['html', {open: 'never'}],
    ['list'], // Console output for debugging
  ],
  
  use: {
    baseURL: 'http://localhost:8080', // Go server
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure', // Video for complex failure debugging
    // Enable both head and headless modes
    headless: process.env.HEAD !== 'true',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    // TODO: Add more browsers once basic tests are stable
  ],

  // Note: No webServer config - assumes Go server is running
  // Run manually: ./weewar-server --port=8080
  // Or add server management later if needed
});
