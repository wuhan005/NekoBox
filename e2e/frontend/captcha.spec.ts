import {expect, type Page} from '@playwright/test';
import {test} from './coverage';
import {avatar, globalConfig, respond, holdNextRequest, mockFrontend} from './helpers';

test.beforeEach(mockFrontend);

async function anonymous(page: Page) {
    await page.addInitScript(() => localStorage.removeItem('auth'));
}

const captchaCalls = (page: Page) => page.evaluate(() => (window as any).__captchaCalls || 0);
const releaseCaptcha = (page: Page, token = 'fresh-captcha-token') => page.evaluate(token => (window as any).__resolveCaptcha(token), token);
const configRoute = '**/api/config';
const scriptRoute = `${globalConfig.captcha.domain}/**`;
const slideConfig = {...globalConfig, captcha: {type: 'go_captcha'}};

test('a failed global configuration can be retried without losing form input', async ({page}) => {
    await anonymous(page);
    const slow = await holdNextRequest(page, configRoute);
    await page.goto('/sign-in');
    await page.locator('input[name="email"]').fill('alice@example.com');
    await respond(await slow.request, null, 503);
    await expect(page.locator('button[type="submit"]')).toBeDisabled();
    await page.getByRole('button', {name: '重新加载验证码'}).click();
    await expect(page.getByRole('button', {name: '登录', exact: true})).toBeEnabled();
    await expect(page.locator('input[name="email"]')).toHaveValue('alice@example.com');
});

test('a failed config is not cached as the wrong driver across page navigation', async ({page}) => {
    await anonymous(page);
    const slow = await holdNextRequest(page, configRoute);
    await page.goto('/sign-in');
    await expect(page.getByRole('button', {name: '加载中...'})).toBeDisabled();
    await respond(await slow.request, null, 503);
    await expect(page.getByRole('button', {name: '重新加载验证码'})).toBeVisible();
    await page.route(configRoute, r => respond(r, slideConfig));
    await page.locator('nav').getByRole('link', {name: '注册', exact: true}).click();
    await expect(page.getByRole('button', {name: '注册', exact: true})).toBeEnabled();
    await expect(page.locator('script[src*="/recaptcha/api.js"]')).toHaveCount(0);
});

test('a failed reCAPTCHA script can be loaded again in the same page', async ({page}) => {
    await anonymous(page);
    await page.route(scriptRoute, r => r.abort(), {times: 1});
    await page.goto('/sign-in');
    await page.locator('input[name="email"]').fill('alice@example.com');
    await expect(page.getByRole('button', {name: '重新加载验证码'})).toBeVisible();
    await page.getByRole('button', {name: '重新加载验证码'}).click();
    await expect(page.getByRole('button', {name: '登录', exact: true})).toBeEnabled();
    await expect(page.locator('input[name="email"]')).toHaveValue('alice@example.com');
    await expect(page.locator('script[src*="/recaptcha/api.js"]')).toHaveCount(1);
});

test('a stalled reCAPTCHA script times out and can be retried', async ({page}) => {
    await anonymous(page);
    await page.clock.install();
    const slow = await holdNextRequest(page, scriptRoute);
    await page.goto('/sign-in');
    await slow.request;
    await page.clock.fastForward(15001);
    await page.getByRole('button', {name: '重新加载验证码'}).click();
    await expect(page.getByRole('button', {name: '登录', exact: true})).toBeEnabled();
});

test('site details and reCAPTCHA share one configuration across pages', async ({page}) => {
    await anonymous(page);
    const requests: string[] = [];
    page.on('request', request => {
        if (new URL(request.url()).pathname === '/api/config') requests.push(request.url());
    });
    const script = page.waitForRequest(request => request.url().startsWith(globalConfig.captcha.domain));
    await page.goto('/sign-up');
    const scriptURL = new URL((await script).url());
    expect(scriptURL.searchParams.get('render')).toBe(globalConfig.captcha.siteKey);
    await page.locator('input[name="domain"]').fill('alice');
    await expect(page.locator('code')).toHaveText(`${globalConfig.externalURL}/_/alice`);
    await expect(page.getByText(globalConfig.icp)).toBeVisible();

    await page.locator('nav').getByRole('link', {name: '登录', exact: true}).click();
    const submit = page.getByRole('button', {name: '登录', exact: true});
    await expect(submit).toBeEnabled();
    await page.locator('input[name="email"]').fill('alice@example.com');
    await page.locator('input[name="password"]').fill('Password123!');
    await submit.click();
    await expect.poll(() => page.evaluate(() => (window as any).__captchaSiteKey)).toBe(globalConfig.captcha.siteKey);
    expect(requests).toHaveLength(1);
});

for (const form of [
    {path: '/sign-in', endpoint: '/api/auth/sign-in', button: '登录', fields: {email: 'alice@example.com', password: 'Password123!'}},
    {path: '/sign-up', endpoint: '/api/auth/sign-up', button: '注册', fields: {email: 'alice@example.com', domain: 'alice', name: 'Alice', password: 'Password123!', repeatPassword: 'Password123!'}},
    {path: '/forgot-password', endpoint: '/api/auth/forgot-password', button: '找回密码', fields: {email: 'alice@example.com'}},
    {path: '/_/bob', endpoint: '/api/users/bob/questions', button: '发送提问', fields: {content: 'Only one question'}},
]) {
    test(`${form.path} locks submission during captcha acquisition and the business request`, async ({page}) => {
        await anonymous(page);
        await page.addInitScript(() => { (window as any).__holdCaptcha = true; });
        let posts = 0;
        page.on('request', r => { if (r.method() === 'POST' && r.url().endsWith(form.endpoint)) posts++; });
        const slow = await holdNextRequest(page, `**${form.endpoint}`);
        await page.goto(form.path);
        for (const [name, value] of Object.entries(form.fields)) {
            await page.locator(`[name="${name}"]`).fill(value!);
        }
        const button = page.getByRole('button', {name: form.button, exact: true});
        await expect(button).toBeEnabled();
        await button.dblclick();
        await expect.poll(() => captchaCalls(page)).toBe(1);
        await expect(page.locator('button[type="submit"]')).toBeDisabled();
        expect(posts).toBe(0);
        await releaseCaptcha(page);
        const route = await slow.request;
        expect(route.request().postData()).toContain('fresh-captcha-token');
        expect(posts).toBe(1);
        await expect(page.locator('button[type="submit"]')).toBeDisabled();
        await respond(route, null, 400);
        await expect(button).toBeEnabled();
        expect(posts).toBe(1);
    });
}

test('an empty captcha token restores the button without submitting', async ({page}) => {
    await anonymous(page);
    await page.addInitScript(() => { (window as any).__holdCaptcha = true; });
    let posts = 0;
    page.on('request', r => { if (r.method() === 'POST') posts++; });
    await page.goto('/forgot-password');
    await page.locator('input[name="email"]').fill('alice@example.com');
    const button = page.getByRole('button', {name: '找回密码', exact: true});
    await button.click();
    await expect.poll(() => captchaCalls(page)).toBe(1);
    await releaseCaptcha(page, '');
    await expect(button).toBeEnabled();
    expect(posts).toBe(0);
});

test('a captcha execution timeout restores the button and permits a fresh token', async ({page}) => {
    await anonymous(page);
    await page.clock.install();
    await page.addInitScript(() => { (window as any).__holdCaptcha = true; });
    await page.goto('/forgot-password');
    await page.locator('input[name="email"]').fill('alice@example.com');
    const button = page.getByRole('button', {name: '找回密码', exact: true});
    await button.click();
    await expect.poll(() => captchaCalls(page)).toBe(1);
    await page.clock.fastForward(15001);
    await expect(button).toBeEnabled();
    await button.click();
    await expect.poll(() => captchaCalls(page)).toBe(2);
    const sent = page.waitForRequest(r => r.url().endsWith('/auth/forgot-password'));
    await releaseCaptcha(page, 'after-timeout-token');
    expect((await sent).postData()).toContain('after-timeout-token');
});

async function mockSlide(page: Page) {
    let challenge = 0;
    await page.route(configRoute, r => respond(r, slideConfig));
    await page.route('**/api/captcha/challenge', r => respond(r, {
        key: `slide-${++challenge}`, image: avatar, thumb: avatar, thumbX: 0, thumbY: 50, thumbWidth: 40, thumbHeight: 40,
    }));
    await page.route('**/api/captcha/verify', r => respond(r, {token: 'slide-token'}));
}

async function solveSlide(page: Page) {
    const block = page.locator('.captcha-popover-visible .gc-drag-block');
    await expect(block).toBeVisible();
    const box = (await block.boundingBox())!;
    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
    await page.mouse.down();
    await page.mouse.move(box.x + box.width / 2 + 100, box.y + box.height / 2, {steps: 5});
    await page.mouse.up();
}

test('a cancelled slide verification releases the form and the next attempt succeeds', async ({page}) => {
    await mockSlide(page);
    await page.goto('/_/alice');
    await page.locator('textarea[name="content"]').fill('A slide question');
    const button = page.getByRole('button', {name: '发送提问', exact: true});
    await button.click();
    await expect(page.locator('.captcha-popover-visible')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeDisabled();
    await page.locator('.captcha-popover-mask').click({position: {x: 5, y: 5}});
    await expect(button).toBeEnabled();
    await button.click();
    const sent = page.waitForRequest(r => r.method() === 'POST' && r.url().includes('/users/alice/questions'));
    await solveSlide(page);
    expect((await sent).postData()).toContain('slide-token');
    await expect(page.locator('.uk-alert-success')).toBeVisible();
});

test('a failed slide challenge can be retried without discarding the question', async ({page}) => {
    await mockSlide(page);
    const slow = await holdNextRequest(page, '**/api/captcha/challenge');
    await page.goto('/_/alice');
    await page.locator('textarea[name="content"]').fill('Keep this question');
    const button = page.getByRole('button', {name: '发送提问', exact: true});
    await button.click();
    await respond(await slow.request, null, 503);
    await page.getByRole('button', {name: '重新加载验证码'}).click();
    await expect(button).toBeEnabled();
    await expect(page.locator('textarea[name="content"]')).toHaveValue('Keep this question');
    await button.click();
    await solveSlide(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible();
});

test('a cancelled refresh failure cannot close a newer slide challenge', async ({page}) => {
    await mockSlide(page);
    await page.goto('/_/alice');
    await page.locator('textarea[name="content"]').fill('Keep the current challenge');
    const button = page.getByRole('button', {name: '发送提问', exact: true});
    await button.click();
    await expect(page.locator('.captcha-popover-visible')).toBeVisible();
    const slow = await holdNextRequest(page, '**/api/captcha/challenge');
    await page.locator('.captcha-popover-visible .gc-icon-block svg').last().click();
    const stale = await slow.request;
    await page.locator('.captcha-popover-mask').click({position: {x: 5, y: 5}});
    await expect(button).toBeEnabled();
    await button.click();
    await expect(page.locator('.captcha-popover-visible')).toBeVisible();
    await respond(stale, null, 503);
    await page.waitForLoadState('networkidle');
    await expect(page.locator('.captcha-popover-visible')).toBeVisible();
    await solveSlide(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible();
});

test('a rejected slide answer refreshes the challenge and can then succeed', async ({page}) => {
    await mockSlide(page);
    await page.route('**/api/captcha/verify', r => respond(r, null, 400), {times: 1});
    await page.goto('/_/alice');
    await page.locator('textarea[name="content"]').fill('Try the puzzle again');
    await page.getByRole('button', {name: '发送提问', exact: true}).click();
    const refreshed = page.waitForResponse(r => r.url().endsWith('/captcha/challenge'));
    await solveSlide(page);
    await refreshed;
    await page.waitForLoadState('networkidle');
    await solveSlide(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible();
});

test('cancelling an in-flight verification does not block a newer challenge', async ({page}) => {
    await mockSlide(page);
    const slow = await holdNextRequest(page, '**/api/captcha/verify');
    await page.goto('/_/alice');
    await page.locator('textarea[name="content"]').fill('Use the newer verification');
    const button = page.getByRole('button', {name: '发送提问', exact: true});
    await button.click();
    await solveSlide(page);
    const stale = await slow.request;
    await page.locator('.captcha-popover-mask').click({position: {x: 5, y: 5}});
    await expect(button).toBeEnabled();
    await button.click();
    await solveSlide(page);
    await expect(page.locator('.uk-alert-success')).toBeVisible();
    await respond(stale, null, 503);
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('button', {name: '重新加载验证码'})).not.toBeVisible();
    await expect(button).toBeEnabled();
});
