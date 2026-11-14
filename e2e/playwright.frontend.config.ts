import {defineConfig} from '@playwright/test';
import base from './playwright.config';

export default defineConfig<{collectFrontendCoverage: boolean}>(base, {
    testDir: '.',
    testMatch: ['**/frontend/*.spec.ts', '**/tests/frontend-state.spec.ts'],
    outputDir: 'test-results/frontend',
    reporter: [
        ['html', {outputFolder: 'playwright-report/frontend', open: 'never'}],
        ...(process.env.CI ? ([['github']] as const) : []),
    ],
    globalTeardown: './frontend/coverage.teardown.ts',
    use: {baseURL: 'http://127.0.0.1:4173', collectFrontendCoverage: true},
    webServer: {
        command: 'npm --prefix ../web run dev -- --mode frontend-test --host 127.0.0.1 --port 4173 --strictPort',
        url: 'http://127.0.0.1:4173',
        env: {
            VITE_BASE_URL: '/api',
        },
    },
});
