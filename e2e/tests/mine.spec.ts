import { test, expect, type Page } from '@playwright/test';
import { registerAndLogin, clickSubmitWhenReady } from './helpers';

// ─── Helpers ──────────────────────────────────────────────────────────────────

/** Post a question, then navigate from the mine list to the question detail page. */
async function postAndOpenQuestion(page: Page, domain: string, content: string): Promise<void> {
    await page.goto(`/_/${domain}`);
    await page.locator('textarea[name="content"]').fill(content);
    await clickSubmitWhenReady(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

    await page.goto('/mine/questions');
    await expect(page.locator('p.uk-text-small').first()).toContainText(content, { timeout: 10_000 });
    await page.locator('p.uk-text-small').first().click();
    await expect(page).toHaveURL(new RegExp(`/_/${domain}/\\d+`));
}

function waitForVisibleResponse(page: Page) {
    return page.waitForResponse(resp =>
        /\/api\/mine\/questions\/\d+\/visible$/.test(resp.url()) &&
        resp.request().method() === 'PUT'
    );
}

// ─── Question management ──────────────────────────────────────────────────────

test.describe('Mine – Question Management', () => {

    test('owner can delete a question', async ({ page }) => {
        const user = await registerAndLogin(page, 'qdelete');

        await postAndOpenQuestion(page, user.domain, 'Please delete me!');

        // Open the delete drop-down then confirm.
        await page.locator('a:has-text("删除提问")').click();
        const deleteResponsePromise = page.waitForResponse(resp =>
            /\/api\/mine\/questions\/\d+$/.test(new URL(resp.url()).pathname) &&
            resp.request().method() === 'DELETE'
        );
        await page.locator('button:has-text("确认删除")').click();
        const deleteResponse = await deleteResponsePromise;

        expect(deleteResponse.status()).toBe(200);
        const payload = await deleteResponse.json();
        expect(payload.data).toContain('提问删除成功');

        // After deletion the router navigates back to the owner's profile page.
        await expect(page).toHaveURL(new RegExp(`/_/${user.domain}$`), { timeout: 10_000 });
    });

    test('deleted question no longer appears in the mine list', async ({ page }) => {
        const user = await registerAndLogin(page, 'qdeldel');
        const content = 'Deleted question should vanish';

        await postAndOpenQuestion(page, user.domain, content);

        await page.locator('a:has-text("删除提问")').click();
        const deleteResponsePromise = page.waitForResponse(resp =>
            /\/api\/mine\/questions\/\d+$/.test(new URL(resp.url()).pathname) &&
            resp.request().method() === 'DELETE'
        );
        await page.locator('button:has-text("确认删除")').click();
        await deleteResponsePromise;

        // Reload the mine list and verify the question is gone.
        await page.goto('/mine/questions');
        await expect(page.locator(`text=${content}`)).not.toBeVisible();
    });

    test('owner can toggle question visibility (public → private → public)', async ({ page }) => {
        const user = await registerAndLogin(page, 'qtoggle');

        await postAndOpenQuestion(page, user.domain, 'Toggle visibility test');

        // Answer the question so it has an initial public state.
        await expect(page.locator('textarea[name="answer"]')).toBeVisible({ timeout: 10_000 });
        await page.locator('textarea[name="answer"]').fill('Toggle visibility answer');
        const answerResponsePromise = page.waitForResponse(resp =>
            /\/api\/mine\/questions\/\d+\/answer$/.test(resp.url()) &&
            resp.request().method() === 'PUT'
        );
        await clickSubmitWhenReady(page);
        await answerResponsePromise;

        // ── Set to private ──
        await expect(page.locator('button:has-text("设为私密")')).toBeVisible({ timeout: 10_000 });
        const setPrivateResponsePromise = waitForVisibleResponse(page);
        await page.locator('button:has-text("设为私密")').click();
        const setPrivateResponse = await setPrivateResponsePromise;

        expect(setPrivateResponse.status()).toBe(200);
        const setPrivatePayload = await setPrivateResponse.json();
        expect(setPrivatePayload.data).toContain('私密');

        // Button text should flip to "设为公开".
        await expect(page.locator('button:has-text("设为公开")')).toBeVisible({ timeout: 10_000 });

        // ── Set back to public ──
        const setPublicResponsePromise = waitForVisibleResponse(page);
        await page.locator('button:has-text("设为公开")').click();
        const setPublicResponse = await setPublicResponsePromise;

        expect(setPublicResponse.status()).toBe(200);
        const setPublicPayload = await setPublicResponse.json();
        expect(setPublicPayload.data).toContain('公开');

        // Button text should flip back to "设为私密".
        await expect(page.locator('button:has-text("设为私密")')).toBeVisible({ timeout: 10_000 });
    });

    test('mine question list shows "未回答" badge for unanswered questions', async ({ page }) => {
        const user = await registerAndLogin(page, 'listbadge');

        // Post an unanswered question.
        await page.goto(`/_/${user.domain}`);
        await page.locator('textarea[name="content"]').fill('Unanswered badge test');
        await clickSubmitWhenReady(page);
        await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

        // Open mine/questions and verify the "未回答" label is present.
        await page.goto('/mine/questions');
        await expect(page.locator('span.uk-label:has-text("未回答")').first()).toBeVisible({ timeout: 10_000 });
    });

    test('mine question list shows "私密" badge for private questions', async ({ page }) => {
        const user = await registerAndLogin(page, 'privbadge');

        // Post a private question.
        await page.goto(`/_/${user.domain}`);
        await page.locator('textarea[name="content"]').fill('Private badge test');
        await page.locator('label:has-text("回复后不公开提问") input[type="checkbox"]').check();
        await clickSubmitWhenReady(page);
        await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

        // Open mine/questions and verify the "私密" label is present.
        await page.goto('/mine/questions');
        await expect(page.locator('span.uk-label-warning:has-text("私密")').first()).toBeVisible({ timeout: 10_000 });
    });

    // ─── Mine API boundary checks ─────────────────────────────────────────────

    test('unauthenticated requests to mine APIs return 403', async ({ page }) => {
        // Use a fresh browser context with no session cookie.
        const ctx = page.context();
        const freshPage = await ctx.newPage();
        await freshPage.goto('/');

        // Clear all cookies to simulate an unauthenticated state.
        await ctx.clearCookies();

        const response = await freshPage.request.get('/api/mine/questions');
        // The server requires sign-in; it returns 403 for anonymous access.
        expect([401, 403]).toContain(response.status());

        await freshPage.close();
    });

    test('answering someone else\'s question returns 404', async ({ page, browser }) => {
        const owner = await registerAndLogin(page, 'qother');

        // Post a question as the owner.
        await page.goto(`/_/${owner.domain}`);
        await page.locator('textarea[name="content"]').fill('Owner question');
        const postResponsePromise = page.waitForResponse(resp =>
            resp.request().method() === 'POST' &&
            new RegExp(`/api/users/${owner.domain}/questions$`).test(new URL(resp.url()).pathname)
        );
        await clickSubmitWhenReady(page);
        const postResponse = await postResponsePromise;
        const postPayload = await postResponse.json();

        // Extract the question ID from the private URL in the success message.
        const match = (postPayload.data as string).match(/\/(\d+)\?t=/);
        expect(match).not.toBeNull();
        const questionID = match![1];

        // Register a second user in an isolated context, then try to answer owner's question.
        const otherContext = await browser.newContext();
        const otherPage = await otherContext.newPage();
        try {
            await registerAndLogin(otherPage, 'qother2');

            // page.request does not automatically mirror app-level axios auth headers from Pinia.
            const sessionID = await otherPage.evaluate(() => {
                const raw = window.localStorage.getItem('auth');
                if (!raw) return '';
                try {
                    return (JSON.parse(raw) as { sessionID?: string }).sessionID ?? '';
                } catch {
                    return '';
                }
            });
            expect(sessionID).not.toBe('');

            const answerResponse = await otherPage.request.put(
                `/api/mine/questions/${questionID}/answer`,
                {
                    headers: { Authorization: `Token ${sessionID}` },
                    multipart: { answer: 'Hacked!' },
                }
            );
            // The Questioner middleware should return 404 because the question doesn't belong to this user.
            expect(answerResponse.status()).toBe(404);
        } finally {
            await otherContext.close();
        }
    });
});

