import {expect} from '@playwright/test';
import {test} from './coverage';
import {mockFrontend, respond} from './helpers';

test.beforeEach(mockFrontend);

test('password reset loads the account and submits the new password with the recovery code', async ({page}) => {
    await page.addInitScript(() => localStorage.removeItem('auth'));
    await page.route('**/api/auth/recover-password?code=recovery-code', route => respond(route, {name: 'Alice'}));
    await page.route('**/api/auth/recover-password', route => respond(route, '密码重置成功'));

    await page.goto('/recover-password?code=recovery-code');
    await expect(page.getByText('Alice，您正在重置您的密码：')).toBeVisible();
    const password = 'Replacement123!';
    await page.locator('input[name="newPassword"]').fill(password);
    const confirmation = page.locator('input[name="repeatPassword"]');
    await confirmation.fill('Different123!');
    const submit = page.getByRole('button', {name: '重置密码', exact: true});
    await submit.click();
    const confirmationError = confirmation.locator('..').locator('.field-error-message');
    await expect(confirmationError).toBeVisible();

    await confirmation.fill(password);
    await confirmation.blur();
    await expect(confirmationError).toBeHidden();
    const request = page.waitForRequest(r => r.method() === 'POST' && r.url().endsWith('/api/auth/recover-password'));
    await submit.click();
    expect((await request).postDataJSON()).toEqual({code: 'recovery-code', newPassword: password, repeatPassword: password});
    await expect(page).toHaveURL('/');
    await expect(page.getByText('密码重置成功')).toBeVisible();
});

test('missing and expired recovery codes return to the home page', async ({page}) => {
    await page.addInitScript(() => localStorage.removeItem('auth'));
    await page.goto('/recover-password');
    await expect(page).toHaveURL('/');

    await page.route('**/api/auth/recover-password?code=expired', route => route.fulfill({
        status: 400,
        json: {error: 400, msg: '邮件已过期，请重新发送'},
    }));
    await page.goto('/recover-password?code=expired');
    await expect(page).toHaveURL('/');
    await expect(page.getByText('邮件已过期，请重新发送')).toBeVisible();
});

test('account deactivation can be cancelled and requires an explicit confirmation', async ({page}) => {
    let deactivations = 0;
    await page.route('**/api/mine/settings/deactivate', route => {
        expect(route.request().method()).toBe('POST');
        deactivations++;
        return respond(route, '您的账号已停用');
    });

    await page.goto('/settings');
    const openDeactivation = async () => {
        await page.getByRole('tab', {name: '账号设置'}).click();
        await page.getByRole('button', {name: '停用我的账号'}).click();
        await expect(page).toHaveURL('/mine/deactivate');
    };
    await openDeactivation();
    await page.getByRole('button', {name: '返回', exact: true}).click();
    await expect(page).toHaveURL('/settings');
    expect(deactivations).toBe(0);

    await openDeactivation();
    await page.getByRole('button', {name: '我确认停用账号'}).click();
    await expect(page).toHaveURL('/');
    await expect(page.locator('nav').getByRole('link', {name: '登录', exact: true})).toBeVisible();
    expect(deactivations).toBe(1);
});

test('exporting account data saves the file returned by the API', async ({page}) => {
    const data = Buffer.from('test workbook');
    await page.route('**/api/mine/settings/export-data', route => {
        expect(route.request().method()).toBe('POST');
        expect(route.request().headers().authorization).toBe('Token test-session');
        return route.fulfill({
            contentType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
            body: data,
        });
    });

    await page.goto('/settings');
    await page.getByRole('tab', {name: '账号设置'}).click();
    const button = page.getByRole('button', {name: '导出我的数据'});
    const downloadReady = page.waitForEvent('download');
    await button.click();
    const download = await downloadReady;
    const chunks: Buffer[] = [];
    for await (const chunk of await download.createReadStream()) chunks.push(chunk);
    expect(Buffer.concat(chunks)).toEqual(data);
    await expect(button).toBeEnabled();
});
