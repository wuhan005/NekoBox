import { test, expect } from '@playwright/test';
import { image, question, respond, holdNextRequest, mockFrontend } from '../frontend/helpers';

const questionLink = (page: import('@playwright/test').Page, id: number) => page.locator(`#question-list a[href="/_/alice/${id}"]`);

test.beforeEach(mockFrontend);

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
