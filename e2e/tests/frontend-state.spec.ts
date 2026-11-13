import { test, expect } from '@playwright/test';
import { image, question, respond, holdNextRequest, mockFrontend } from '../frontend/helpers';

const questionLink = (page: import('@playwright/test').Page, id: number) => page.locator(`#question-list a[href="/_/alice/${id}"]`);

test.beforeEach(mockFrontend);

test('browser back to a question box restores its loaded answers and scroll position', async ({ page }) => {
    await page.addInitScript(() => localStorage.removeItem('auth'));
    await page.setViewportSize({ width: 390, height: 844 });
    const cursors: string[] = [];
    await page.route('**/api/users/alice/questions?*', route => {
        const cursor = new URL(route.request().url()).searchParams.get('cursor') ?? '';
        cursors.push(cursor);
        const first = cursor ? Number(cursor) - 1 : 45;
        const questions = Array.from({ length: Math.min(first, 20) }, (_, index) => question(first - index));
        return respond(route, { total: 45, cursor: String(questions.at(-1)?.id ?? ''), questions });
    });
    await page.route(/\/api\/users\/alice\/questions\/7(?:\?|$)/, route =>
        respond(route, { ...question(7), isOwner: false }));

    await page.goto('/_/alice');
    await expect(page.getByText('以前回答过的问题', { exact: false })).toBeVisible();
    await page.getByRole('button', { name: '加载更多' }).click();
    const answerLink = questionLink(page, 7);
    await expect(answerLink).toHaveText('查看回答');
    await answerLink.scrollIntoViewIfNeeded();
    const scrollY = await page.evaluate(() => window.scrollY);
    expect(scrollY).toBeGreaterThan(0);
    await answerLink.click();
    await expect(page.getByRole('heading', { name: 'Question 7', exact: true })).toBeVisible();
    await expect(questionLink(page, 45)).toBeVisible();
    await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(0);
    const cursorsBeforeBack = [...cursors];

    await page.goBack();
    await expect(page).toHaveURL(/\/_\/alice$/);
    await expect(page).toHaveTitle('Alice - NekoBox');
    await expect(answerLink).toBeInViewport();
    await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(scrollY);
    expect(cursors).toEqual(cursorsBeforeBack);

    await page.getByRole('button', { name: '加载更多' }).click();
    await expect(questionLink(page, 1)).toBeVisible();
    expect(cursors.at(-1)).toBe('6');
});

test('browser back to a question box omits an answer just made private', async ({ page }) => {
    let isPrivate = false;
    await page.route('**/api/users/alice/questions?*', route => respond(route, {
        total: isPrivate ? 1 : 2,
        cursor: '2',
        questions: isPrivate ? [question(2)] : [question(3), question(2)],
    }));
    await page.route('**/api/mine/questions/3/visible', route => {
        isPrivate = true;
        return respond(route, 'Saved');
    });

    await page.goto('/_/alice');
    await questionLink(page, 3).click();
    await page.getByRole('button', { name: '设为私密', exact: true }).click();
    await expect(page.getByRole('button', { name: '设为公开', exact: true })).toBeEnabled();

    await page.goBack();
    await expect(page).toHaveURL(/\/_\/alice$/);
    await expect(questionLink(page, 2)).toBeVisible();
    await expect(questionLink(page, 3)).toHaveCount(0);
});

test('returning from a question restores the selected tab, loaded pages and scroll position', async ({ page }) => {
    const cursors: string[] = [];
    let receivedRequests = 0;
    page.on('request', request => {
        if (new URL(request.url()).pathname === '/api/mine/questions') receivedRequests++;
    });
    await page.route('**/api/mine/questions/sent?*', route => {
        const cursor = new URL(route.request().url()).searchParams.get('cursor') ?? '';
        cursors.push(cursor);
        const first = cursor ? Number(cursor) - 1 : 45;
        const questions = Array.from({ length: Math.min(first, 20) }, (_, index) => ({
            ...question(first - index), isAnswered: true, targetDomain: 'alice',
        }));
        return respond(route, { total: 45, cursor: String(questions.at(-1)?.id ?? ''), questions });
    });

    await page.goto('/mine/questions');
    await page.getByRole('tab', { name: '我的提问 (45)' }).click();
    await page.getByRole('button', { name: '加载更多' }).click();
    const loadedQuestion = page.getByRole('link', { name: /Question 7$/ });
    await loadedQuestion.scrollIntoViewIfNeeded();
    const scrollY = await page.evaluate(() => window.scrollY);
    expect(scrollY).toBeGreaterThan(0);
    for (const returnToList of [
        () => page.goBack(),
        () => page.getByRole('navigation').getByRole('link', { name: '提问', exact: true }).click(),
    ]) {
        await loadedQuestion.click();
        await expect(page.getByRole('heading', { name: 'Question 7', exact: true })).toBeVisible();
        await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(0);

        await returnToList();
        await expect(page).toHaveURL(/\/mine\/questions$/);
        await expect(page.getByRole('tab', { name: '我的提问 (45)' })).toHaveAttribute('aria-selected', 'true');
        await expect(loadedQuestion).toBeInViewport();
        await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(scrollY);
    }
    expect(receivedRequests).toBe(1);
    expect(cursors).toEqual(['', '26']);

    await page.getByRole('button', { name: '加载更多' }).click();
    await expect(page.getByRole('link', { name: /Question 1$/ })).toBeVisible();
    expect(cursors).toEqual(['', '26', '6']);
});

test('back to the question list reflects a visibility change made in the detail page', async ({ page }) => {
    let isPrivate = true;
    await page.route('**/api/mine/questions?*', route => respond(route, {
        total: 1, cursor: '1', questions: [{ ...question(1), isAnswered: true, isPrivate }],
    }));
    await page.route('**/api/mine/questions/1/visible', route => {
        isPrivate = false;
        return respond(route, 'Saved');
    });

    await page.goto('/mine/questions');
    const item = page.getByRole('link', { name: /Question 1$/ });
    await expect(item.getByText('私密')).toBeVisible();
    await item.click();
    await page.getByRole('button', { name: '设为公开', exact: true }).click();
    await expect(page.getByRole('button', { name: '设为私密', exact: true })).toBeEnabled();

    await page.goBack();
    await expect(item).toBeVisible();
    await expect(item.getByText('私密')).toHaveCount(0);
});

test('switching box owners reloads the recipient and clears the previous draft', async ({ page }) => {
    await page.goto('/_/bob');
    await expect(page.getByRole('heading', { name: 'Bob', exact: true })).toBeVisible();
    await page.locator('textarea[name="content"]').fill('A draft for Bob');
    await page.locator('input[name="images"]').setInputFiles(image);

    await page.locator('nav').getByRole('link', { name: 'Alice', exact: true }).click();
    await expect(page).toHaveURL(/\/_\/alice$/);
    await expect(page.getByRole('heading', { name: 'Alice', exact: true })).toBeVisible();
    await expect(page.locator('textarea[name="content"]')).toHaveValue('');
    expect(await page.locator('input[name="images"]').evaluate((input: HTMLInputElement) => input.files!.length)).toBe(0);

    await page.locator('textarea[name="content"]').fill('A question for Alice');
    const sent = page.waitForRequest(request => request.method() === 'POST' && request.url().includes('/users/alice/questions'));
    await page.getByRole('button', { name: '发送提问', exact: true }).click();
    expect((await sent).postData()).toContain('A question for Alice');
});

test('a missing box response cannot redirect after switching owners', async ({ page }) => {
    const slow = await holdNextRequest(page, '**/api/users/bob/profile');
    await page.goto('/_/bob');
    const route = await slow.request;
    await page.locator('nav').getByRole('link', { name: 'Alice', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Alice', exact: true })).toBeVisible();
    await respond(route, null, 404);
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/\/_\/alice$/);
});

for (const status of [200, 404]) {
    test(`a stale question ${status} response cannot replace or redirect the current question`, async ({ page }) => {
        await page.goto('/_/alice/1');
        await expect(page.getByRole('heading', { name: 'Question 1', exact: true })).toBeVisible();
        const slow = await holdNextRequest(page, /\/api\/users\/alice\/questions\/2(?:\?|$)/);
        await questionLink(page, 2).click();
        const route = await slow.request;
        await questionLink(page, 3).click();
        await expect(page.getByRole('heading', { name: 'Question 3', exact: true })).toBeVisible();

        await respond(route, status === 200 ? question(2) : null, status);
        await page.waitForLoadState('networkidle');
        await expect(page).toHaveURL(/\/_\/alice\/3$/);
        await expect(page.getByRole('heading', { name: 'Question 3', exact: true })).toBeVisible();
        await expect(page.locator('textarea[name="answer"]')).toHaveValue('Answer 3');

        await page.locator('textarea[name="answer"]').fill('The answer to question 3');
        const sent = page.waitForRequest(request => request.method() === 'PUT' && request.url().includes('/mine/questions/3/answer'));
        await page.getByRole('button', { name: '更新回答', exact: true }).click();
        expect((await sent).postData()).toContain('The answer to question 3');
    });
}

test('a stale question error cannot redirect away from settings', async ({ page }) => {
    const slow = await holdNextRequest(page, /\/api\/users\/alice\/questions\/1(?:\?|$)/);
    await page.goto('/_/alice/1');
    const route = await slow.request;
    await page.locator('nav').getByRole('link', { name: '设置', exact: true }).click();
    await expect(page.locator('#name')).toHaveValue('Alice');
    await respond(route, null, 404);
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/\/settings$/);
});

test('visibility changes are locked until completion and use the requested state', async ({ page }) => {
    const slow = await holdNextRequest(page, '**/api/mine/questions/1/visible');
    let changes = 0;
    page.on('request', request => { if (request.url().endsWith('/1/visible')) changes++; });
    await page.goto('/_/alice/1');
    const button = page.getByRole('button', { name: '设为公开', exact: true });
    await button.dblclick();
    const route = await slow.request;
    expect(route.request().postDataJSON()).toEqual({ visible: true });
    await expect(button).toBeDisabled();
    await expect(page.getByRole('button', { name: '更新回答', exact: true })).toBeDisabled();
    expect(changes).toBe(1);

    await respond(route, 'Saved');
    await expect(page.getByRole('button', { name: '设为私密', exact: true })).toBeEnabled();
    await expect(page.getByRole('button', { name: '更新回答', exact: true })).toBeEnabled();
    expect(changes).toBe(1);
});

test('a failed visibility change preserves the state and allows retry', async ({ page }) => {
    const slow = await holdNextRequest(page, '**/api/mine/questions/1/visible');
    await page.goto('/_/alice/1');
    const button = page.getByRole('button', { name: '设为公开', exact: true });
    await button.click();
    await respond(await slow.request, null, 500);
    await expect(button).toBeEnabled();
    await button.click();
    await expect(page.getByRole('button', { name: '设为私密', exact: true })).toBeEnabled();
});

test('a visibility response for the previous question cannot change the next question', async ({ page }) => {
    const slow = await holdNextRequest(page, '**/api/mine/questions/1/visible');
    await page.goto('/_/alice/1');
    await page.getByRole('button', { name: '设为公开', exact: true }).click();
    const route = await slow.request;
    await questionLink(page, 2).click();
    await expect(page.getByRole('heading', { name: 'Question 2', exact: true })).toBeVisible();
    await respond(route, 'Saved');
    await expect(page.getByRole('button', { name: '设为私密', exact: true })).toBeEnabled();
});

for (const settings of [
    { path: 'profile', tab: '个人信息', save: '修改信息', value: 'Alice' },
    { path: 'box', tab: '提问箱设置', save: '保存配置', value: 'Existing introduction' },
    { path: 'harassment', tab: '防骚扰设置', save: '更新防骚扰设置', value: 'register_only' },
]) {
    test(`${settings.path} settings cannot save before loading or after a failed read`, async ({ page }) => {
        const slow = await holdNextRequest(page, `**/api/mine/settings/${settings.path}`);
        const writes: string[] = [];
        page.on('request', request => {
            if (request.method() === 'PUT' && request.url().endsWith(`/settings/${settings.path}`)) writes.push(request.postData()!);
        });
        await page.goto('/settings');
        await page.getByRole('tab', { name: settings.tab, exact: true }).click();
        const route = await slow.request;
        await expect(page.getByRole('status')).toContainText('正在加载');
        await expect(page.getByRole('button', { name: settings.save, exact: true })).not.toBeVisible();
        expect(writes).toHaveLength(0);

        await respond(route, null, 500);
        await expect(page.locator('.uk-tab-panel:visible').getByRole('alert')).toContainText('加载失败');
        await expect(page.getByRole('button', { name: settings.save, exact: true })).not.toBeVisible();
        await page.getByRole('button', { name: '重新加载', exact: true }).click();
        await expect(page.getByRole('button', { name: settings.save, exact: true })).toBeEnabled();
        if (settings.path === 'harassment') {
            await expect(page.getByRole('checkbox')).toBeChecked();
            await expect(page.locator('input[name="blockWords"]')).toHaveValue('blocked');
        }
        if (settings.path === 'box') await expect(page.getByRole('checkbox')).toBeChecked();

        const sent = page.waitForRequest(request => request.method() === 'PUT' && request.url().endsWith(`/settings/${settings.path}`));
        await page.getByRole('button', { name: settings.save, exact: true }).click();
        expect((await sent).postData()).toContain(settings.value);
        expect(writes).toHaveLength(1);
    });
}

test('switching questions clears the previous answer attachment', async ({ page }) => {
    await page.goto('/_/alice/1');
    await page.locator('input[name="images"]').setInputFiles(image);
    await questionLink(page, 2).click();
    await expect(page.getByRole('heading', { name: 'Question 2', exact: true })).toBeVisible();
    expect(await page.locator('input[name="images"]').evaluate((input: HTMLInputElement) => input.files!.length)).toBe(0);
    await expect(page.getByText('first-answer.png', { exact: true })).not.toBeVisible();
    await page.locator('textarea[name="answer"]').fill('An answer without an attachment');
    const sent = page.waitForRequest(request => request.url().endsWith('/2/answer'));
    await page.getByRole('button', { name: '更新回答', exact: true }).click();
    expect((await sent).postData()).not.toContain('filename=');
});

test('answer attachments survive failure but are cleared after a successful retry', async ({ page }) => {
    const slow = await holdNextRequest(page, '**/api/mine/questions/1/answer');
    await page.goto('/_/alice/1');
    await page.locator('input[name="images"]').setInputFiles(image);
    const button = page.getByRole('button', { name: '更新回答', exact: true });
    await button.click();
    await expect(page.getByRole('button', { name: '设为公开', exact: true })).toBeDisabled();
    await respond(await slow.request, null, 500);
    await expect(button).toBeEnabled();
    await expect(page.getByText('first-answer.png', { exact: true })).toBeVisible();

    const sent = page.waitForRequest(request => request.url().endsWith('/1/answer'));
    await button.click();
    expect((await sent).postData()).toContain('first-answer.png');
    await expect(button).toBeEnabled();
    expect(await page.locator('input[name="images"]').evaluate((input: HTMLInputElement) => input.files!.length)).toBe(0);
    await expect(page.getByText('first-answer.png', { exact: true })).not.toBeVisible();

    const update = page.waitForRequest(request => request.url().endsWith('/1/answer'));
    await page.locator('textarea[name="answer"]').fill('A text-only update');
    await button.click();
    expect((await update).postData()).not.toContain('filename=');
});

test('a previous answer response cannot clear the next question attachment', async ({ page }) => {
    const slow = await holdNextRequest(page, '**/api/mine/questions/1/answer');
    await page.goto('/_/alice/1');
    await page.locator('input[name="images"]').setInputFiles(image);
    await page.getByRole('button', { name: '更新回答', exact: true }).click();
    const route = await slow.request;
    await questionLink(page, 2).click();
    await expect(page.getByRole('heading', { name: 'Question 2', exact: true })).toBeVisible();
    await page.locator('input[name="images"]').setInputFiles({ ...image, name: 'second-answer.png' });
    await respond(route, 'Saved');
    await expect(page.getByRole('button', { name: '更新回答', exact: true })).toBeEnabled();
    await expect(page.getByText('second-answer.png', { exact: true })).toBeVisible();
});

test('saving an answer only clears the attachment that was submitted', async ({ page }) => {
    const slow = await holdNextRequest(page, '**/api/mine/questions/1/answer');
    await page.goto('/_/alice/1');
    await page.locator('input[name="images"]').setInputFiles(image);
    const button = page.getByRole('button', { name: '更新回答', exact: true });
    await button.click();
    const route = await slow.request;
    await page.locator('input[name="images"]').setInputFiles({ ...image, name: 'next-upload.png' });
    await respond(route, 'Saved');
    await expect(button).toBeEnabled();
    await expect(page.getByText('next-upload.png', { exact: true })).toBeVisible();

    const sent = page.waitForRequest(request => request.url().endsWith('/1/answer'));
    await button.click();
    const body = (await sent).postData();
    expect(body).toContain('next-upload.png');
    expect(body).not.toContain('first-answer.png');
});
