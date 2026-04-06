import { test, expect } from '@playwright/test';
import { clickSubmitWhenReady, uniqueUser } from './helpers';

test.describe('Authentication', () => {

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

