import { test, expect } from '@playwright/test';
import { registerAndLogin, waitForMailhogEmail } from './helpers';

// ─── Post a question ──────────────────────────────────────────────────────────

test('can post a question to a profile box', async ({ page }) => {
    const user = await registerAndLogin(page, 'qpost');

    await page.goto(`/_/${user.domain}`);
    await page.locator('textarea[name="content"]').fill('What is your favorite color?');
    await page.locator('button[type="submit"]').click();

    // A success banner with the private link appears.
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });
    await expect(page.locator('.uk-alert-success')).toContainText('发送问题成功');
});

// ─── Full question lifecycle ───────────────────────────────────────────────────

test('owner can answer a question and the answer is shown publicly', async ({ page }) => {
    const user = await registerAndLogin(page, 'qanswer');

    // 1. Post a question (as the signed-in owner – valid for testing).
    await page.goto(`/_/${user.domain}`);
    await page.locator('textarea[name="content"]').fill('What is your favorite programming language?');
    await page.locator('button[type="submit"]').click();
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

    // 2. Navigate to "my questions" and open the first one.
    await page.goto('/mine/questions');
    await expect(page.locator('p.uk-text-small').first()).toContainText(
        'What is your favorite programming language?',
        { timeout: 10_000 },
    );
    await page.locator('p.uk-text-small').first().click();

    // 3. On the question-detail page, the owner sees the answer form.
    await page.waitForURL(`/_/${user.domain}/**`);
    await expect(page.locator('textarea[name="answer"]')).toBeVisible({ timeout: 10_000 });

    // 4. Submit the answer.
    await page.locator('textarea[name="answer"]').fill('Go is my favorite!');
    await page.locator('button[type="submit"]').click();

    // 5. Success toast appears.
    await expect(page.locator('.Toastify')).toContainText('提问回复成功', { timeout: 10_000 });

    // 6. The answer text is now rendered in the card body.
    await expect(page.locator('.uk-card-body p.uk-text-small')).toContainText('Go is my favorite!');
});

// ─── Email notifications ──────────────────────────────────────────────────────

test('owner receives a "new question" email when someone posts a question', async ({ page }) => {
    const owner = await registerAndLogin(page, 'mailowner');

    // Post a question to the owner's box (as the signed-in owner for simplicity).
    await page.goto(`/_/${owner.domain}`);
    await page.locator('textarea[name="content"]').fill('Will you get an email about this?');
    await page.locator('button[type="submit"]').click();
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

    // MailHog should receive the "new question" notification.
    const received = await waitForMailhogEmail(owner.email);
    expect(received).toBe(true);
});

test('questioner receives a reply email when their question is answered', async ({ page }) => {
    const owner = await registerAndLogin(page, 'replyowner');
    const replyEmail = `reply-${owner.domain}-${Date.now()}@example.com`;

    // 1. Post a question with an email-reply address.
    await page.goto(`/_/${owner.domain}`);
    await page.locator('textarea[name="content"]').fill('Will I get a reply notification?');

    // Check the "接收回复通知" checkbox (the second checkbox in the form).
    await page.locator('label:has-text("我想接收回复通知") input[type="checkbox"]').check();
    await page.locator('input[name="receiveReplyEmail"]').fill(replyEmail);

    await page.locator('button[type="submit"]').click();
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

    // 2. Open the question from the mine list.
    await page.goto('/mine/questions');
    await page.locator('p.uk-text-small').first().click();
    await page.waitForURL(`/_/${owner.domain}/**`);

    // 3. Answer the question – this triggers the reply email.
    await expect(page.locator('textarea[name="answer"]')).toBeVisible({ timeout: 10_000 });
    await page.locator('textarea[name="answer"]').fill('Yes, you will receive a reply!');
    await page.locator('button[type="submit"]').click();
    await expect(page.locator('.Toastify')).toContainText('提问回复成功', { timeout: 10_000 });

    // 4. MailHog must have delivered the reply notification to replyEmail.
    const received = await waitForMailhogEmail(replyEmail);
    expect(received).toBe(true);
});

// ─── Image upload (MinIO / S3) ────────────────────────────────────────────────

test('can post a question with an image (MinIO upload)', async ({ page }) => {
    // mockRecaptcha is already set by registerAndLogin via helpers.
    const user = await registerAndLogin(page, 'imgupload');

    await page.goto(`/_/${user.domain}`);

    // Create a minimal 1×1 PNG as a Buffer and attach it via the file chooser.
    const minimalPNG = Buffer.from([
        0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
        0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
        0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
        0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
        0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc, 0x33, 0x00, 0x00, 0x00,
        0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
    ]);

    await page.locator('textarea[name="content"]').fill('Check out this image!');

    // Attach via the hidden file input.
    const fileInput = page.locator('input[name="images"][type="file"]');
    await fileInput.setInputFiles({
        name: 'test.png',
        mimeType: 'image/png',
        buffer: minimalPNG,
    });

    await page.locator('button[type="submit"]').click();

    // Success message must appear.
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 15_000 });
    await expect(page.locator('.uk-alert-success')).toContainText('发送问题成功');
});

