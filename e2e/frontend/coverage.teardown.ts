import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {extname, resolve} from 'node:path';
import {coverageDir, createCoverageReport, repoRoot} from './coverage';

export default async function generateCoverage() {
    const coverage = await createCoverageReport();
    await coverage.generate();

    const lcov = await readFile(resolve(coverageDir, 'lcov.info'), 'utf8');
    const coveredExtensions = new Set<string>();
    for (const record of lcov.split(/^end_of_record\r?$/m)) {
        const path = record.match(/^SF:(.+)$/m)?.[1];
        if (!path) continue;

        assert.match(path, /^web\/src\/.+\.(vue|ts)$/);
        const lineCount = (await readFile(resolve(repoRoot, path), 'utf8')).split('\n').length;
        for (const [, line, hits] of record.matchAll(/^DA:(\d+),(\d+)/gm)) {
            assert.ok(Number(line) > 0 && Number(line) <= lineCount, `${path}:${line} is out of range`);
            if (Number(hits) > 0) coveredExtensions.add(extname(path));
        }
    }
    assert.deepEqual(coveredExtensions, new Set(['.vue', '.ts']), 'LCOV must cover both Vue and TypeScript sources');
}
