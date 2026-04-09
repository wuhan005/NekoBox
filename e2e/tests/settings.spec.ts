import { test, expect } from '@playwright/test';
import { registerAndLogin, clickSubmitWhenReady, authHeaderFromPage } from './helpers';

// ─── Profile settings ─────────────────────────────────────────────────────────

test.describe('Profile Settings', () => {

    test('settings page loads correct name and email', async ({ page }) => {
        const user = await registerAndLogin(page, 'settload');

        await page.goto('/settings');

        // Name field is pre-populated with the registered name.
        await expect(page.locator('input#name')).toHaveValue(user.name, { timeout: 10_000 });
        // Email field (disabled) shows the registered email.
        await expect(page.getByLabel('电子邮箱')).toHaveValue(user.email);
    });

    test('user can update their display name', async ({ page }) => {
        const user = await registerAndLogin(page, 'nameupd');
        const newName = `${user.name} Upd`;

        await page.goto('/settings');
        await expect(page.locator('input#name')).toBeVisible({ timeout: 10_000 });
        await page.locator('input#name').fill(newName);

        const updateProfileResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/api/mine/settings/profile') && resp.request().method() === 'PUT'
        );
        await page.locator('button[type="submit"]:has-text("修改信息")').click();
        const updateProfileResponse = await updateProfileResponsePromise;

        expect(updateProfileResponse.status()).toBe(200);
        const payload = await updateProfileResponse.json();
        expect(payload.data).toContain('个人信息更新成功');

        // The field should reflect the new name after the save.
        await expect(page.locator('input#name')).toHaveValue(newName, { timeout: 10_000 });
    });

    test('user can change their password and sign in with the new password', async ({ page }) => {
        const user = await registerAndLogin(page, 'pwdchg');
        const newPassword = 'NewPassword456!';

        await page.goto('/settings');
        await expect(page.locator('input#name')).toBeVisible({ timeout: 10_000 });
        await page.locator('input#oldPassword').fill(user.password);
        await page.locator('input#newPassword').fill(newPassword);

        const updateProfileResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/api/mine/settings/profile') && resp.request().method() === 'PUT'
        );
        await page.locator('button[type="submit"]:has-text("修改信息")').click();
        const updateProfileResponse = await updateProfileResponsePromise;
        expect(updateProfileResponse.status()).toBe(200);

        // Sign out via the settings-page button.
        await page.locator('button:has-text("退出登录")').click();
        await expect(page).toHaveURL('/');

        // Sign in with the new password should succeed.
        await page.goto('/sign-in');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="password"]').fill(newPassword);
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL(new RegExp(`/_/${user.domain}$`));
    });

    test('wrong old password is rejected when changing password', async ({ page }) => {
        const user = await registerAndLogin(page, 'badoldpw');

        await page.goto('/settings');
        await expect(page.locator('input#name')).toBeVisible({ timeout: 10_000 });
        await page.locator('input#oldPassword').fill('totally-wrong-old-password');
        await page.locator('input#newPassword').fill('SomeNewPassword789!');

        const updateProfileResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/api/mine/settings/profile') && resp.request().method() === 'PUT'
        );
        await page.locator('button[type="submit"]:has-text("修改信息")').click();
        const updateProfileResponse = await updateProfileResponsePromise;

        expect(updateProfileResponse.status()).toBe(400);
        const payload = await updateProfileResponse.json();
        expect(payload.msg).toContain('旧密码');
    });
});

// ─── Box settings ─────────────────────────────────────────────────────────────

test.describe('Box Settings', () => {

    test('user can update the box intro text', async ({ page }) => {
        const user = await registerAndLogin(page, 'boxintro');

        await page.goto('/settings');
        // Switch to "提问箱设置" tab.
        await page.locator('ul[uk-tab] li a:has-text("提问箱设置")').click();

        const newIntro = '这是我全新的个人介绍！';
        await expect(page.locator('input#intro')).toBeVisible({ timeout: 10_000 });
        await page.locator('input#intro').fill(newIntro);

        const updateBoxResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/api/mine/settings/box') && resp.request().method() === 'PUT'
        );
        await page.locator('button[type="submit"]:has-text("保存配置")').click();
        const updateBoxResponse = await updateBoxResponsePromise;

        expect(updateBoxResponse.status()).toBe(200);
        const payload = await updateBoxResponse.json();
        expect(payload.data).toContain('提问箱设置更新成功');

        // The updated intro must appear on the public profile page.
        await page.goto(`/_/${user.domain}`);
        await expect(page.locator(`text=${newIntro}`)).toBeVisible({ timeout: 10_000 });
    });

    test('invalid notify type is rejected by the server', async ({ page }) => {
        const user = await registerAndLogin(page, 'badnotify');
        const authHeader = await authHeaderFromPage(page);

        // Send a PUT request directly with an invalid notifyType.
        const response = await page.request.put('/api/mine/settings/box', {
            headers: { 'Content-Type': 'application/json', ...authHeader },
            data: { intro: 'test', notifyType: 'sms' },
        });
        // The server should reject unknown notify types.
        expect(response.status()).toBe(400);
    });
});

// ─── Harassment settings ──────────────────────────────────────────────────────

test.describe('Harassment Settings', () => {

    test('block words prevent matching questions from being posted', async ({ page }) => {
        const user = await registerAndLogin(page, 'blkwrd');

        await page.goto('/settings');
        await page.locator('ul[uk-tab] li a:has-text("防骚扰设置")').click();
        await expect(page.locator('input[name="blockWords"]')).toBeVisible({ timeout: 10_000 });
        await page.locator('input[name="blockWords"]').fill('forbidden');

        const updateHarassmentResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/api/mine/settings/harassment') && resp.request().method() === 'PUT'
        );
        await page.locator('button[type="submit"]:has-text("更新防骚扰设置")').click();
        const updateHarassmentResponse = await updateHarassmentResponsePromise;
        expect(updateHarassmentResponse.status()).toBe(200);

        // Post a question that contains the blocked word.
        await page.goto(`/_/${user.domain}`);
        await page.locator('textarea[name="content"]').fill('This contains forbidden word');

        const postResponsePromise = page.waitForResponse(resp =>
            resp.request().method() === 'POST' &&
            new RegExp(`/api/users/${user.domain}/questions$`).test(new URL(resp.url()).pathname)
        );
        await clickSubmitWhenReady(page);
        const postResponse = await postResponsePromise;

        expect(postResponse.status()).toBe(400);
        const payload = await postResponse.json();
        expect(payload.msg).toContain('屏蔽词');
    });

    test('register-only setting shows a login prompt to anonymous visitors', async ({ page, browser }) => {
        const owner = await registerAndLogin(page, 'regonly');

        await page.goto('/settings');
        await page.locator('ul[uk-tab] li a:has-text("防骚扰设置")').click();
        await expect(
            page.locator('label:has-text("仅允许注册用户向我提问")')
        ).toBeVisible({ timeout: 10_000 });
        await page.locator('label:has-text("仅允许注册用户向我提问") input[type="checkbox"]').check();

        const updateHarassmentResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/api/mine/settings/harassment') && resp.request().method() === 'PUT'
        );
        await page.locator('button[type="submit"]:has-text("更新防骚扰设置")').click();
        await updateHarassmentResponsePromise;

        // Open an incognito context (anonymous visitor).
        const incognitoContext = await browser.newContext();
        const guestPage = await incognitoContext.newPage();
        await guestPage.goto(`/_/${owner.domain}`);

        // Anonymous visitors should see the "前往登录" prompt.
        await expect(guestPage.locator('a:has-text("前往登录")')).toBeVisible({ timeout: 10_000 });
        // The question textarea must NOT be visible.
        await expect(guestPage.locator('textarea[name="content"]')).not.toBeVisible();

        await incognitoContext.close();
    });

    test('block words with more than 10 entries are rejected', async ({ page }) => {
        const user = await registerAndLogin(page, 'toomanyblk');
        const authHeader = await authHeaderFromPage(page);

        const tooManyWords = 'a,b,c,d,e,f,g,h,i,j,k'; // 11 words
        const response = await page.request.put('/api/mine/settings/harassment', {
            headers: { 'Content-Type': 'application/json', ...authHeader },
            data: { harassmentSettingType: 'none', blockWords: tooManyWords },
        });
        expect(response.status()).toBe(400);
        const payload = await response.json();
        expect(payload.msg).toContain('10');
    });
});

// ─── Account deactivation ─────────────────────────────────────────────────────

test.describe('Account Deactivation', () => {

    test('user can deactivate account and is redirected to home; subsequent sign-in fails', async ({ page }) => {
        const user = await registerAndLogin(page, 'deact');

        await page.goto('/mine/deactivate');
        await expect(page.locator('button:has-text("我确认停用账号")')).toBeVisible({ timeout: 10_000 });

        const deactivateResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/api/mine/settings/deactivate') && resp.request().method() === 'POST'
        );
        await page.locator('button:has-text("我确认停用账号")').click();
        const deactivateResponse = await deactivateResponsePromise;

        expect(deactivateResponse.status()).toBe(200);
        const deactivatePayload = await deactivateResponse.json();
        expect(deactivatePayload.data).toContain('停用');

        // Should redirect to home.
        await expect(page).toHaveURL('/', { timeout: 10_000 });

        // Attempting to sign in with the now-deleted account must fail.
        await page.goto('/sign-in');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="password"]').fill(user.password);

        const signInResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/auth/sign-in') && resp.request().method() === 'POST'
        );
        await clickSubmitWhenReady(page);
        const signInResponse = await signInResponsePromise;

        expect(signInResponse.status()).toBe(400);
        const signInPayload = await signInResponse.json();
        expect(signInPayload.msg).toContain('密码');
    });
});

