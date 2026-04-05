import { test, expect } from '@playwright/test';
import { mockRecaptcha, uniqueUser } from './helpers';

test.describe('Authentication', () => {
    // Each test sets up its own reCAPTCHA mock before navigating.
    test.beforeEach(async ({ page }) => {
        await mockRecaptcha(page);
    });

    test('user can sign up and is redirected to sign-in', async ({ page }) => {
        const user = uniqueUser('signup');

        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="domain"]').fill(user.domain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await page.locator('button[type="submit"]').click();

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
        await page.locator('button[type="submit"]').click();
        await expect(page).toHaveURL(/\/sign-in$/);

        // Sign in.
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('button[type="submit"]').click();

        // Should land on the user's own profile page.
        await expect(page).toHaveURL(new RegExp(`/_/${user.domain}$`));
        await expect(page).toHaveTitle(new RegExp(user.name));
    });

    test('sign-up with duplicate email shows an error', async ({ page }) => {
        const user = uniqueUser('dupmail');

        // First registration.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="domain"]').fill(user.domain);
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await page.locator('button[type="submit"]').click();
        await expect(page).toHaveURL(/\/sign-in$/);

        // Try to register again with the same email but a different domain.
        await page.goto('/sign-up');
        await page.locator('input[name="email"]').fill(user.email);       // duplicate
        await page.locator('input[name="domain"]').fill(user.domain + '2');
        await page.locator('input[name="name"]').fill(user.name);
        await page.locator('input[name="password"]').fill(user.password);
        await page.locator('input[name="repeatPassword"]').fill(user.password);
        await page.locator('button[type="submit"]').click();

        // Error toast must appear.
        await expect(page.locator('.Toastify')).toContainText('邮箱', { timeout: 8_000 });
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
        await page.locator('button[type="submit"]').click();
        await expect(page).toHaveURL(/\/sign-in$/);

        // Sign in with wrong password.
        await page.locator('input[name="email"]').fill(user.email);
        await page.locator('input[name="password"]').fill('totally-wrong-password');
        await page.locator('button[type="submit"]').click();

        // Error toast must appear; URL must not change.
        await expect(page.locator('.Toastify')).toContainText('密码', { timeout: 8_000 });
        await expect(page).toHaveURL('/sign-in');
    });
});

