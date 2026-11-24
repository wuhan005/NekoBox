import { test, expect } from '@playwright/test';
import { clickSubmitWhenReady, uniqueUser, waitForMailhogEmail } from './helpers';


test.describe('Authentication', () => {

    test('password reset emails a single-use link and replaces the old password', async ({ page, request }) => {
        const user = uniqueUser('recovery');
        const newPassword = 'Replacement123!';

        await page.goto('/sign-up');
        for (const [name, value] of Object.entries({ ...user, repeatPassword: user.password })) {
            await page.locator(`input[name="${name}"]`).fill(value);
        }
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL('/sign-in');

        const requestReset = async (email: string) => {
            await page.goto('/forgot-password');
            await page.locator('input[name="email"]').fill(email);
            const response = page.waitForResponse(resp =>
                new URL(resp.url()).pathname === '/api/auth/forgot-password' && resp.request().method() === 'POST',
            );
            await clickSubmitWhenReady(page);
            const result = await response;
            expect(result.status()).toBe(200);
            expect(await result.json()).toEqual({data: '若该邮箱已注册，我们已向其发送了重置密码邮件，请前往查收。'});
            await expect(page).toHaveURL('/');
        };

        await requestReset(`missing-${user.email}`);
        await requestReset(user.email);

        const mail = await waitForMailhogEmail(user.email, 30_000);
        const recoveryURL = mail?.replace(/=\r?\n/g, '').match(/https?:\/\/[^"\s]+\/recover-password\?code(?:=3D|=)[a-zA-Z0-9]+/)?.[0];
        expect(recoveryURL, 'Password reset link in the recovery email').toBeDefined();
        const link = new URL(recoveryURL!.replace('?code=3D', '?code='));
        const code = link.searchParams.get('code');
        expect(code).toHaveLength(64);

        await requestReset(user.email);

        await page.goto(link.pathname + link.search);
        await expect(page.getByText(`${user.name}，您正在重置您的密码：`)).toBeVisible();
        await page.locator('input[name="newPassword"]').fill(newPassword);
        await page.locator('input[name="repeatPassword"]').fill(newPassword);
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL('/');

        const reused = await request.post('/api/auth/recover-password', {
            data: { code, newPassword: 'Repeated123!', repeatPassword: 'Repeated123!' },
        });
        expect(reused.status()).toBe(400);
        expect((await reused.json()).msg).toContain('邮件已过期');

        await page.goto('/sign-in');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="password"]').fill(user.password);
        await clickSubmitWhenReady(page);
        await expect(page.getByText('电子邮箱或密码错误')).toBeVisible();
        await page.locator('input[name="password"]').fill(newPassword);
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL(`/_/${user.domain}`);
    });

    test('user can sign up and is redirected to sign-in', async ({ page }) => {
        const user = uniqueUser('signup');

        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="domain"]').fill(user.domain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await clickSubmitWhenReady(page);

        // Successful registration redirects to /sign-in.
        await expect(page).toHaveURL(/\/sign-in$/);
    });

    test('user can sign in and land on their profile page', async ({ page }) => {
        const user = uniqueUser('signin');

        // Register first.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="domain"]').fill(user.domain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL(/\/sign-in$/);

        // Sign in.
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="password"]').fill(user.password);
        await clickSubmitWhenReady(page);

        // Should land on the user's own profile page.
        await expect(page).toHaveURL(new RegExp(`/_/${user.domain}$`));
        await expect(page).toHaveTitle(new RegExp(user.name));
    });

    test('sign-up with duplicate email shows an error', async ({ page }) => {
        const user = uniqueUser('dupmail');
        const duplicateDomain = `${user.domain.slice(0, 19)}2`;

        // First registration.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="domain"]').fill(user.domain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL(/\/sign-in$/);

        // Try to register again with the same email but a different domain.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);       // duplicate
        await page.locator('input[name="domain"]').fill(duplicateDomain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);

        const duplicateSignUpResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/auth/sign-up') && resp.request().method() === 'POST'
        );
        await clickSubmitWhenReady(page);

        const duplicateSignUpResponse = await duplicateSignUpResponsePromise;
        expect(duplicateSignUpResponse.status()).toBe(400);
        const duplicateSignUpPayload = await duplicateSignUpResponse.json();

        // Error toast must appear.
        await expect(duplicateSignUpPayload.msg).toContain('邮箱');
    });

    test('sign-up with duplicate domain shows an error', async ({ page }) => {
        const user = uniqueUser('dupdomain');
        const altEmail = `alt_${user.domain}@example.com`.slice(0, 40);

        // First registration.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="domain"]').fill(user.domain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL(/\/sign-in$/);

        // Second registration with a different email but the same domain.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(altEmail);
        await page.locator('input[name="domain"]').fill(user.domain);      // duplicate
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);

        const dupDomainResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/auth/sign-up') && resp.request().method() === 'POST'
        );
        await clickSubmitWhenReady(page);

        const dupDomainResponse = await dupDomainResponsePromise;
        expect(dupDomainResponse.status()).toBe(400);
        const dupDomainPayload = await dupDomainResponse.json();
        expect(dupDomainPayload.msg).toContain('域名');
    });

    test('sign-in with wrong password shows an error', async ({ page }) => {
        const user = uniqueUser('wrongpw');

        // Register.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="domain"]').fill(user.domain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await clickSubmitWhenReady(page);
        await expect(page).toHaveURL(/\/sign-in$/);

        // Sign in with wrong password.
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="password"]').fill('totally-wrong-password');

        const wrongPasswordSignInResponsePromise = page.waitForResponse(resp =>
            resp.url().includes('/auth/sign-in') && resp.request().method() === 'POST'
        );
        await clickSubmitWhenReady(page);

        const wrongPasswordSignInResponse = await wrongPasswordSignInResponsePromise;
        expect(wrongPasswordSignInResponse.status()).toBe(400);
        const wrongPasswordSignInPayload = await wrongPasswordSignInResponse.json();

        // Error toast must appear; URL must not change.
        await expect(wrongPasswordSignInPayload.msg).toContain('密码');
        await expect(page).toHaveURL('/sign-in');
    });
});

