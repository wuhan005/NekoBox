import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './tests',

    // Tests are sequential to avoid inter-test race conditions against shared DB/mail state.
    fullyParallel: false,
    workers: 1,

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

