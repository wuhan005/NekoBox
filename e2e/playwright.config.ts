import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './tests',

    // Run tests in parallel, but keep worker count conservative for CI stability.
    fullyParallel: true,
    workers: process.env.CI ? 2 : undefined,
    timeout: process.env.CI ? 90_000 : 60_000,

    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 1 : 0,

    reporter: [
        ['html'],
        ...(process.env.CI ? ([['github']] as any) : []),
    ],

    use: {
        baseURL: process.env.E2E_BASE_URL || 'http://localhost:3000',
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
        // Give Vue time to hydrate and animations to settle.
        actionTimeout: 10_000,
        navigationTimeout: 30_000,
    },

    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
});

