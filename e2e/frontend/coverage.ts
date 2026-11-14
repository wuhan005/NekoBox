import {resolve} from 'node:path';
import {test as base} from '@playwright/test';

export const repoRoot = resolve(__dirname, '../..');
export const coverageDir = resolve(__dirname, '../test-results/frontend/coverage');

export async function createCoverageReport() {
    const {MCR} = await import('monocart-coverage-reports');
    return MCR({
        baseDir: repoRoot,
        outputDir: coverageDir,
        reports: ['lcovonly'],
        logging: 'error',
        entryFilter: ({url}) => /^https?:\/\/[^/]+\/src\/.+\.(vue|ts)(?:\?|$)/.test(url),
        // Vite's inline sourcemaps name sources relative to the served module, under /src.
        sourcePath: (sourcePath, {distFile, url}) => {
            if (!distFile) return sourcePath;
            return `web${new URL(url, `http://${distFile}`).pathname}`;
        },
        outputFilter: ({sourcePath}) => /^web\/src\/.+\.(vue|ts)$/.test(sourcePath),
    });
}

export const test = base.extend<{collectFrontendCoverage: boolean}>({
    collectFrontendCoverage: [false, {option: true}],
    page: async ({page, collectFrontendCoverage}, use) => {
        if (!collectFrontendCoverage) {
            await use(page);
            return;
        }

        await page.coverage.startJSCoverage({resetOnNavigation: false});
        try {
            await use(page);
        } finally {
            const entries = await page.coverage.stopJSCoverage();
            const coverage = await createCoverageReport();
            await coverage.add(entries);
        }
    },
});
