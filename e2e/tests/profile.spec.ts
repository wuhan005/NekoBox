import { test, expect } from '@playwright/test';
import { registerAndLogin, clickSubmitWhenReady } from './helpers';

// ─── Public profile page ──────────────────────────────────────────────────────

test.describe('Public Profile Page', () => {

    test('profile page title includes the user name', async ({ page }) => {
        const user = await registerAndLogin(page, 'pubprof');

        await page.goto(`/_/${user.domain}`);
        await expect(page).toHaveTitle(new RegExp(user.name), { timeout: 10_000 });
    });

    test('profile card shows the user name', async ({ page }) => {
        const user = await registerAndLogin(page, 'cardname');

        await page.goto(`/_/${user.domain}`);
        await expect(page.locator('.uk-card-body h3')).toContainText(user.name, { timeout: 10_000 });
    });

    test('non-existent domain redirects to home', async ({ page }) => {
        await page.goto('/_/this-domain-should-not-exist-xyz-abc');
        // The Vue router catches the 404 from the profile API and pushes to '/'.
        await expect(page).toHaveURL('/', { timeout: 15_000 });
    });

    // ─── Question visibility ──────────────────────────────────────────────────

    test('answered questions are visible on the public profile', async ({ page }) => {
        const user = await registerAndLogin(page, 'pubans');

        // Post a question.
        await page.goto(`/_/${user.domain}`);
        await page.locator('textarea[name="content"]').fill('A publicly visible question?');
        await clickSubmitWhenReady(page);
        await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

        // Answer it from mine/questions.
        await page.goto('/mine/questions');
        await expect(page.locator('p.uk-text-small').first()).toContainText(
            'A publicly visible question?', { timeout: 10_000 }
        );
        await page.locator('p.uk-text-small').first().click();
        await expect(page.locator('textarea[name="answer"]')).toBeVisible({ timeout: 10_000 });
        await page.locator('textarea[name="answer"]').fill('Yes, this is the public answer!');

        const answerResponsePromise = page.waitForResponse(resp =>
            /\/api\/mine\/questions\/\d+\/answer$/.test(resp.url()) &&
            resp.request().method() === 'PUT'
        );
        await clickSubmitWhenReady(page);
        await answerResponsePromise;

        // Visit the public profile and confirm the answered question is listed.
        const questionsResponsePromise = page.waitForResponse(resp =>
            new RegExp(`/api/users/${user.domain}/questions`).test(new URL(resp.url()).pathname)
        );
        await page.goto(`/_/${user.domain}`);
        await questionsResponsePromise;

        await expect(
            page.locator('#question-list p.uk-text-small.uk-text-break')
                .filter({ hasText: 'A publicly visible question?' })
        ).toBeVisible({ timeout: 15_000 });
    });

    test('unanswered questions do NOT appear on the public profile', async ({ page }) => {
        const user = await registerAndLogin(page, 'unans');

        // Post a question but never answer it.
        await page.goto(`/_/${user.domain}`);
        await page.locator('textarea[name="content"]').fill('This unanswered question should stay hidden');
        await clickSubmitWhenReady(page);
        await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

        // Visit the public profile and wait for the questions API to respond.
        const questionsResponsePromise = page.waitForResponse(resp =>
            new RegExp(`/api/users/${user.domain}/questions`).test(new URL(resp.url()).pathname)
        );
        await page.goto(`/_/${user.domain}`);
        await questionsResponsePromise;

        await expect(
            page.locator('text=This unanswered question should stay hidden')
        ).not.toBeVisible();
    });

    test('private answered questions are not listed on the public profile', async ({ page }) => {
        const user = await registerAndLogin(page, 'priv');

        // Post a question with the "回复后不公开" flag.
        await page.goto(`/_/${user.domain}`);
        await page.locator('textarea[name="content"]').fill('This is a private question');
        await page.locator('label:has-text("回复后不公开提问") input[type="checkbox"]').check();
        await clickSubmitWhenReady(page);
        await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

        // Answer it from mine/questions.
        await page.goto('/mine/questions');
        await expect(page.locator('p.uk-text-small').first()).toContainText(
            'This is a private question', { timeout: 10_000 }
        );
        await page.locator('p.uk-text-small').first().click();
        await expect(page.locator('textarea[name="answer"]')).toBeVisible({ timeout: 10_000 });
        await page.locator('textarea[name="answer"]').fill('Private answer here');

        const answerResponsePromise = page.waitForResponse(resp =>
            /\/api\/mine\/questions\/\d+\/answer$/.test(resp.url()) &&
            resp.request().method() === 'PUT'
        );
        await clickSubmitWhenReady(page);
        await answerResponsePromise;

        // Private question must NOT appear on the public profile.
        const questionsResponsePromise = page.waitForResponse(resp =>
            new RegExp(`/api/users/${user.domain}/questions`).test(new URL(resp.url()).pathname)
        );
        await page.goto(`/_/${user.domain}`);
        await questionsResponsePromise;

        await expect(page.locator('text=This is a private question')).not.toBeVisible();
    });

    // ─── User profile API ─────────────────────────────────────────────────────

    test('GET /api/users/:domain/profile returns expected fields', async ({ page }) => {
        const user = await registerAndLogin(page, 'profapi');

        const response = await page.request.get(`/api/users/${user.domain}/profile`);
        expect(response.status()).toBe(200);
        const payload = await response.json();
        expect(payload.data.name).toBe(user.name);
        expect(payload.data.domain).toBe(user.domain);
    });

    test('GET /api/users/:domain/profile for non-existent domain returns 404', async ({ page }) => {
        const response = await page.request.get('/api/users/domain-that-does-not-exist-ever');
        expect(response.status()).toBe(404);
    });
});

