import {expect} from '@playwright/test';
import {test} from './coverage';
import {avatar, image, profile, question, respond, holdNextRequest, mockFrontend} from './helpers';

test.beforeEach(mockFrontend);

test('closing a success notice does not hide the next private question link', async ({page}) => {
    await page.goto('/_/alice');
    await page.locator('textarea[name="content"]').fill('First question');
    await page.getByRole('button', {name: '发送提问', exact: true}).click();
    await expect(page.locator('.uk-alert-success a')).toHaveAttribute('href', '/_/alice/4?t=private-4');
    await page.getByRole('button', {name: '关闭提示'}).click();
    await expect(page.locator('.uk-alert-success')).not.toBeVisible();
    await page.locator('textarea[name="content"]').fill('Second question');
    await page.getByRole('button', {name: '发送提问', exact: true}).click();
    await expect(page.locator('.uk-alert-success a')).toHaveAttribute('href', '/_/alice/5?t=private-5');
});

test('an unchecked notification option sends no email address', async ({page}) => {
    await page.goto('/_/alice');
    const checkbox = page.getByRole('checkbox', {name: '我想接收回复通知'});
    await checkbox.check();
    await page.locator('input[name="receiveReplyEmail"]').fill('private@example.com');
    await checkbox.uncheck();
    await page.locator('textarea[name="content"]').fill('No email please');
    const sent = page.waitForRequest(r => r.method() === 'POST' && r.url().includes('/users/alice/questions'));
    await page.getByRole('button', {name: '发送提问', exact: true}).click();
    expect((await sent).postData()).not.toContain('private@example.com');
});

test('a failed profile load offers retry without an invalid question form', async ({page}) => {
    const slow = await holdNextRequest(page, '**/api/users/alice/profile');
    await page.goto('/_/alice');
    await respond(await slow.request, null, 500);
    await expect(page.getByRole('button', {name: '重新加载提问箱'})).toBeVisible();
    await expect(page.locator('textarea[name="content"]')).not.toBeVisible();
    await page.getByRole('button', {name: '重新加载提问箱'}).click();
    await expect(page.getByRole('button', {name: '发送提问', exact: true})).toBeEnabled();
});

for (const resource of ['profile', 'question']) {
    test(`a slow or failed ${resource} cannot enable the detail page question form`, async ({page}) => {
        await page.route(/\/api\/users\/alice\/questions\/1(?:\?|$)/, r => respond(r, {...question(1), isOwner: false}));
        const slow = await holdNextRequest(page, resource === 'profile'
            ? '**/api/users/alice/profile' : /\/api\/users\/alice\/questions\/1(?:\?|$)/);
        await page.goto('/_/alice/1');
        const route = await slow.request;
        if (resource === 'profile') await expect(page.getByRole('heading', {name: 'Question 1', exact: true})).toBeVisible();
        await expect(page.locator('textarea[name="content"]')).not.toBeVisible();
        await respond(route, null, 500);
        const retry = page.getByRole('button', {name: resource === 'profile' ? '重新加载箱主资料' : '重新加载问题', exact: true});
        await expect(retry).toBeVisible();
        await expect(page.locator('textarea[name="content"]')).not.toBeVisible();
        await retry.click();
        await expect(page.getByRole('button', {name: '发送提问', exact: true})).toBeEnabled();
        await page.locator('textarea[name="content"]').fill('A valid question');
        const sent = page.waitForRequest(r => r.method() === 'POST' && r.url().includes('/users/alice/questions'));
        await page.getByRole('button', {name: '发送提问', exact: true}).click();
        expect((await sent).url()).not.toContain('undefined');
    });
}

test('logging in from a private question preserves the token and all return parameters', async ({page}) => {
    await page.addInitScript(() => localStorage.removeItem('auth'));
    await page.route('**/api/users/bob/profile', r => respond(r, {...profile('bob'), harassmentSetting: 'register_only'}));
    await page.route(/\/api\/users\/bob\/questions\/1(?:\?|$)/, r => respond(r, {...question(1), isOwner: false},
        new URL(r.request().url()).searchParams.get('t') === 'private-token' ? 200 : 404));
    const returnTo = '/_/bob/1?t=private-token&from=mail#answer';
    await page.goto(returnTo);
    await page.getByRole('link', {name: '前往登录'}).click();
    await expect(page).toHaveURL(/\/sign-in\?/);
    expect(new URL(page.url()).searchParams.get('to')).toBe(returnTo);
    await page.locator('input[name="email"]').fill('alice@example.com');
    await page.locator('input[name="password"]').fill('Password123!');
    await page.getByRole('button', {name: '登录', exact: true}).click();
    await expect(page).toHaveURL(new RegExp(returnTo.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '$'));
    await expect(page.getByRole('heading', {name: 'Question 1', exact: true})).toBeVisible();
});

test('authentication guards keep the requested protected page', async ({page}) => {
    await page.addInitScript(() => localStorage.removeItem('auth'));
    await page.goto('/mine/questions?from=bookmark');
    await expect(page.locator('input[name="email"]')).toBeVisible();
    expect(new URL(page.url()).searchParams.get('to')).toBe('/mine/questions?from=bookmark');
    await page.locator('input[name="email"]').fill('alice@example.com');
    await page.locator('input[name="password"]').fill('Password123!');
    await page.getByRole('button', {name: '登录', exact: true}).click();
    await expect(page).toHaveURL(/\/mine\/questions\?from=bookmark$/);
});

test('an expired session keeps the current return address', async ({page}) => {
    const slow = await holdNextRequest(page, '**/api/mine/settings/profile');
    await page.goto('/settings?from=bookmark');
    await respond(await slow.request, null, 401);
    await expect(page.locator('input[name="email"]')).toBeVisible();
    expect(new URL(page.url()).searchParams.get('to')).toBe('/settings?from=bookmark');
});

test('later unauthorized responses cannot replace the original login return address', async ({page}) => {
    const first = await holdNextRequest(page, '**/api/mine/settings/profile');
    const second = await holdNextRequest(page, '**/api/mine/settings/box');
    await page.goto('/settings?from=bookmark');
    const secondRoute = await second.request;
    await respond(await first.request, null, 401);
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await respond(secondRoute, null, 401);
    await page.waitForLoadState('networkidle');
    expect(new URL(page.url()).searchParams.get('to')).toBe('/settings?from=bookmark');
});

for (const settings of [
    {path: 'profile', tab: '个人信息', save: '修改信息', field: '#name', value: 'New name'},
    {path: 'box', tab: '提问箱设置', save: '保存配置', field: '#intro', value: 'New introduction'},
    {path: 'harassment', tab: '防骚扰设置', save: '更新防骚扰设置', field: 'input[name="blockWords"]', value: 'newblock'},
]) {
    test(`a failed ${settings.path} save retains edits and supports retry`, async ({page}) => {
        await page.goto('/settings');
        await page.getByRole('tab', {name: settings.tab, exact: true}).click();
        await page.locator(settings.field).fill(settings.value);
        if (settings.path === 'profile') {
            await page.locator('#oldPassword').fill('Old password');
            await page.locator('#newPassword').fill('New password');
        }
        if (settings.path === 'box') {
            await page.locator('input[type="file"]').first().setInputFiles(image);
            await page.locator('input[type="file"]').nth(1).setInputFiles({...image, name: 'background.png'});
            await page.getByRole('checkbox').uncheck();
        }
        if (settings.path === 'harassment') await page.getByRole('checkbox').uncheck();
        const slow = await holdNextRequest(page, `**/api/mine/settings/${settings.path}`);
        const button = page.getByRole('button', {name: settings.save, exact: true});
        await button.click();
        const route = await slow.request;
        await expect(page.locator('.uk-tab-panel:visible button[type="submit"]')).toBeDisabled();
        await respond(route, null, 500);
        await expect(button).toBeEnabled();
        await expect(page.locator(settings.field)).toHaveValue(settings.value);
        if (settings.path === 'profile') {
            await expect(page.locator('#oldPassword')).toHaveValue('Old password');
            await expect(page.locator('#newPassword')).toHaveValue('New password');
        }
        if (settings.path === 'box') {
            expect(await page.locator('input[type="file"]').first().evaluate((e: HTMLInputElement) => e.files!.length)).toBe(1);
            expect(await page.locator('input[type="file"]').nth(1).evaluate((e: HTMLInputElement) => e.files!.length)).toBe(1);
            await expect(page.getByRole('checkbox')).not.toBeChecked();
        }
        if (settings.path === 'harassment') await expect(page.getByRole('checkbox')).not.toBeChecked();
        const sent = page.waitForRequest(r => r.method() === 'PUT' && r.url().endsWith(`/settings/${settings.path}`));
        await button.click();
        const body = (await sent).postData();
        expect(body).toContain(settings.value);
        if (settings.path === 'box') {
            expect(body).toContain(image.name);
            expect(body).toContain('background.png');
        }
        await expect(button).toBeEnabled();
    });
}

for (const width of [320, 375, 640, 768]) {
    test(`long nicknames, profiles and question lists fit a ${width}px viewport`, async ({page}) => {
        const name = '这是一个二十字长的合法昵称用于测试布局哦';
        await page.addInitScript(name => {
            const auth = JSON.parse(localStorage.getItem('auth')!);
            auth.profile.name = name;
            localStorage.setItem('auth', JSON.stringify(auth));
        }, name);
        await page.setViewportSize({width, height: 812});
        await page.route('**/api/users/alice/profile', r => respond(r, {...profile('alice'), name: 'W'.repeat(20), intro: 'W'.repeat(100)}));
        await page.goto('/_/alice');
        await expect(page.getByRole('heading', {name: 'W'.repeat(20), exact: true})).toBeVisible();
        const nickname = page.locator('nav .profile-name');
        const box = await nickname.boundingBox();
        expect(box!.x).toBeGreaterThanOrEqual(0);
        expect(box!.x + box!.width).toBeLessThanOrEqual(width);
        expect(await nickname.getAttribute('title')).toBe(name);
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
        for (const path of ['/api/mine/questions', '/api/mine/questions/sent']) {
            await page.route(`**${path}?*`, r => respond(r, {total: 1, cursor: '1', questions: [{...question(1), content: 'W'.repeat(1000), targetDomain: 'alice'}]}));
        }
        await page.goto('/mine/questions');
        await expect(page.locator('.uk-tab-panel:visible p.uk-text-small')).toHaveText('W'.repeat(1000));
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
        await page.getByRole('tab', {name: /我的提问/}).click();
        await expect(page.locator('.uk-tab-panel:visible p.uk-text-small')).toHaveText('W'.repeat(1000));
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    });
}

test('question details preserve user line breaks', async ({page}) => {
    await page.route(/\/api\/users\/alice\/questions\/1(?:\?|$)/, r => respond(r, {...question(1), content: 'First line\nSecond line'}));
    await page.goto('/_/alice/1');
    await expect(page.locator('h4')).toHaveText('First line\nSecond line', {useInnerText: true});
});
