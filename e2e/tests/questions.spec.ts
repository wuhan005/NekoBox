import { test, expect, type Page } from '@playwright/test';
import { randomBytes } from 'node:crypto';
import { clickSubmitWhenReady, registerAndLogin, waitForMailhogEmail } from './helpers';

function waitForAnswerResponse(page: Page) {
    return page.waitForResponse(resp =>
        /\/api\/mine\/questions\/\d+\/answer$/.test(resp.url()) && resp.request().method() === 'PUT'
    );
}

function waitForPostQuestionResponse(page: Page, domain: string) {
    return page.waitForResponse(resp =>
        resp.request().method() === 'POST' &&
        new RegExp(`/api/users/${domain}/questions$`).test(new URL(resp.url()).pathname)
    );
}

function waitForGetQuestionResponse(page: Page, domain: string, questionID: number) {
    return page.waitForResponse(resp =>
        resp.request().method() === 'GET' &&
        new RegExp(`/api/users/${domain}/questions/${questionID}$`).test(new URL(resp.url()).pathname)
    );
}

function buildAclUser(prefix: string) {
    const ts = Date.now().toString(36);
    const nonce = randomBytes(2).toString('hex');
    const domainPrefix = prefix.toLowerCase().replace(/[^a-z0-9_]/g, '').slice(0, 8) || 'user';
    let domain = `${domainPrefix}_${ts}${nonce}`.slice(0, 20);
    if (domain.endsWith('_')) {
        domain = `${domain.slice(0, -1)}a`;
    }

    return {
        email: `${domain}@example.com`,
        domain,
        name: `${prefix}-${nonce}`,
        password: 'Password123!',
    };
}

async function registerAndLoginAcl(page: Page, prefix: string) {
    const user = buildAclUser(prefix);

    await page.goto('/sign-up');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="domain"]').fill(user.domain);
    await page.locator('input[name="name"]').fill(user.name);
    await page.locator('input[name="password"]').fill(user.password);
    await page.locator('input[name="repeatPassword"]').fill(user.password);

    const signUpRespPromise = page.waitForResponse(resp =>
        new URL(resp.url()).pathname.endsWith('/api/auth/sign-up') && resp.request().method() === 'POST',
    );
    await clickSubmitWhenReady(page);
    const signUpResp = await signUpRespPromise;
    expect(signUpResp.status()).toBe(200);
    await expect(page).toHaveURL(/\/sign-in$/);

    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);

    const signInRespPromise = page.waitForResponse(resp =>
        new URL(resp.url()).pathname.endsWith('/api/auth/sign-in') && resp.request().method() === 'POST',
    );
    await clickSubmitWhenReady(page);
    const signInResp = await signInRespPromise;
    expect(signInResp.status()).toBe(200);
    await expect(page).toHaveURL(new RegExp(`/_/${user.domain}$`));

    return user;
}

// ─── Post a question ──────────────────────────────────────────────────────────

test('can post a question to a profile box', async ({ page }) => {
    const user = await registerAndLogin(page, 'qpost');

    await page.goto(`/_/${user.domain}`);
    await page.locator('textarea[name="content"]').fill('What is your favorite color?');
    await clickSubmitWhenReady(page);

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
    await clickSubmitWhenReady(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

    // 2. Navigate to "my questions" and open the first one.
    await page.goto('/mine/questions');
    await expect(page.locator('p.uk-text-small').first()).toContainText(
        'What is your favorite programming language?',
        { timeout: 10_000 },
    );
    await page.locator('p.uk-text-small').first().click();

    // 3. On the question-detail page, the owner sees the answer form.
    await expect(page).toHaveURL(new RegExp(`/_/${user.domain}/\\d+`));
    await expect(page.locator('textarea[name="answer"]')).toBeVisible({ timeout: 10_000 });

    // 4. Submit the answer.
    await page.locator('textarea[name="answer"]').fill('Go is my favorite!');
    const answerResponsePromise = waitForAnswerResponse(page);
    await clickSubmitWhenReady(page);
    const answerResponse = await answerResponsePromise;
    expect(answerResponse.status()).toBe(200);
    const answerPayload = await answerResponse.json();
    expect(answerPayload.data).toContain('提问回复成功');

    // 5. The answer text is now rendered in the card body.
    await expect(page.locator('.uk-card-body p.uk-text-small').first()).toContainText('Go is my favorite!');
});

test('access control for unanswered private question is enforced for owner/asker/token holder', async ({ browser }) => {
    const ownerContext = await browser.newContext();
    const askerContext = await browser.newContext();
    const otherContext = await browser.newContext();
    const anonymousContext = await browser.newContext();
    const tokenHolderContext = await browser.newContext();

    const ownerPage = await ownerContext.newPage();
    const askerPage = await askerContext.newPage();
    const otherPage = await otherContext.newPage();
    const anonymousPage = await anonymousContext.newPage();
    const tokenHolderPage = await tokenHolderContext.newPage();

    try {
        const owner = await registerAndLoginAcl(ownerPage, 'owneracl');
        await registerAndLoginAcl(askerPage, 'askeracl');
        await registerAndLoginAcl(otherPage, 'otheracl');

        // Asker posts an unanswered private question to the owner's box.
        await askerPage.goto(`/_/${owner.domain}`);
        await askerPage.locator('textarea[name="content"]').fill('private-access-control-check');
        await askerPage.locator('input[name="private"]').check();
        const postQuestionResponsePromise = waitForPostQuestionResponse(askerPage, owner.domain);
        await clickSubmitWhenReady(askerPage);
        const postQuestionResponse = await postQuestionResponsePromise;
        expect(postQuestionResponse.status()).toBe(200);

        const privateLink = await askerPage.locator('.uk-alert-success a[target="_blank"]').getAttribute('href');
        expect(privateLink).toBeTruthy();

        const privateURL = new URL(privateLink!);
        const questionID = Number(privateURL.pathname.split('/').at(-1));
        const token = privateURL.searchParams.get('t');
        expect(Number.isFinite(questionID)).toBe(true);
        expect(token).toBeTruthy();

        const noTokenPath = `/_/${owner.domain}/${questionID}`;
        const withTokenPath = `${noTokenPath}?t=${token}`;

        // Owner can access without token.
        const ownerRespPromise = waitForGetQuestionResponse(ownerPage, owner.domain, questionID);
        await ownerPage.goto(noTokenPath);
        expect((await ownerRespPromise).status()).toBe(200);

        // Asker can access without token (the behavior introduced in this branch).
        const askerRespPromise = waitForGetQuestionResponse(askerPage, owner.domain, questionID);
        await askerPage.goto(noTokenPath);
        expect((await askerRespPromise).status()).toBe(200);

        // Another signed-in user cannot access without token.
        const otherRespPromise = waitForGetQuestionResponse(otherPage, owner.domain, questionID);
        await otherPage.goto(noTokenPath);
        expect((await otherRespPromise).status()).toBe(404);

        // Anonymous user cannot access without token.
        const anonymousRespPromise = waitForGetQuestionResponse(anonymousPage, owner.domain, questionID);
        await anonymousPage.goto(noTokenPath);
        expect((await anonymousRespPromise).status()).toBe(404);

        // Token holder can access even without sign-in.
        const tokenHolderRespPromise = waitForGetQuestionResponse(tokenHolderPage, owner.domain, questionID);
        await tokenHolderPage.goto(withTokenPath);
        expect((await tokenHolderRespPromise).status()).toBe(200);
    } finally {
        await Promise.all([
            ownerContext.close(),
            askerContext.close(),
            otherContext.close(),
            anonymousContext.close(),
            tokenHolderContext.close(),
        ]);
    }
});

// ─── Email notifications ──────────────────────────────────────────────────────

test('owner receives a "new question" email when someone posts a question', async ({ page }) => {
    const owner = await registerAndLogin(page, 'mailowner');

    // Post a question to the owner's box (as the signed-in owner for simplicity).
    await page.goto(`/_/${owner.domain}`);
    await page.locator('textarea[name="content"]').fill('Will you get an email about this?');
    await clickSubmitWhenReady(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

    // MailHog should receive the "new question" notification.
    const received = await waitForMailhogEmail(owner.email, 30_000);
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

    await clickSubmitWhenReady(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 10_000 });

    // 2. Open the question from the mine list.
    await page.goto('/mine/questions');
    await page.locator('p.uk-text-small').first().click();
    await expect(page).toHaveURL(new RegExp(`/_/${owner.domain}/\\d+`));

    // 3. Answer the question – this triggers the reply email.
    await expect(page.locator('textarea[name="answer"]')).toBeVisible({ timeout: 10_000 });
    await page.locator('textarea[name="answer"]').fill('Yes, you will receive a reply!');
    const replyAnswerResponsePromise = waitForAnswerResponse(page);
    await clickSubmitWhenReady(page);
    const replyAnswerResponse = await replyAnswerResponsePromise;
    expect(replyAnswerResponse.status()).toBe(200);
    const replyAnswerPayload = await replyAnswerResponse.json();
    expect(replyAnswerPayload.data).toContain('提问回复成功');

    // 4. MailHog must have delivered the reply notification to replyEmail.
    const received = await waitForMailhogEmail(replyEmail, 30_000);
    expect(received).toBe(true);
});

// ─── Image upload (MinIO / S3) ────────────────────────────────────────────────

test('can post a question with an image (MinIO upload)', async ({ page }) => {
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

    const postQuestionResponsePromise = waitForPostQuestionResponse(page, user.domain);
    await clickSubmitWhenReady(page);
    const postQuestionResponse = await postQuestionResponsePromise;

    const postQuestionBody = await postQuestionResponse.json();
    expect(postQuestionResponse.status(), `Post question failed: ${JSON.stringify(postQuestionBody)}`).toBe(200);
    expect(postQuestionBody?.data ?? '').toContain('发送问题成功');

    // Success message must appear (image upload in CI can take longer due to MinIO initialization).
    await expect(page.locator('.uk-alert-success')).toBeVisible({ timeout: 30_000 });
    await expect(page.locator('.uk-alert-success')).toContainText('发送问题成功');
});
