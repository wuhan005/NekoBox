import type { Page } from '@playwright/test';
import { randomBytes } from 'node:crypto';

// ─── reCAPTCHA mock ──────────────────────────────────────────────────────────

/**
 * Intercepts the reCAPTCHA v3 script request and injects a lightweight mock so
 * the test-runner never needs network access to Google/recaptcha.net.
 *
 * Call this before navigating to any page that uses reCAPTCHA.
 */
export async function mockRecaptcha(page: Page): Promise<void> {
    const mockScript = `
        window.grecaptcha = {
            ready: function(cb) { if (typeof cb === 'function') cb(); },
            execute: function(_siteKey, _opts) { return Promise.resolve('e2e-mock-token'); },
            render: function() { return 0; },
            reset: function() {},
            getResponse: function() { return 'e2e-mock-token'; },
        };
    `;

    // 1. Intercept the script download and return the mock inline.
    await page.route(/recaptcha\.net\/recaptcha\/api\.js|google\.com\/recaptcha\/api\.js/, async route => {
        await route.fulfill({ contentType: 'application/javascript', body: mockScript });
    });

    // 2. Also pre-populate the global before any page script runs (belt-and-suspenders).
    await page.addInitScript(`(function(){ ${mockScript} })()`);
}

// ─── Unique test data ─────────────────────────────────────────────────────────

/**
 * Returns a unique user object based on the given prefix + current timestamp.
 * Each test run gets a different domain/email so repeated local runs do not clash.
 */
export function uniqueUser(prefix: string) {
    const ts = Date.now().toString(36);
    const nonce = randomBytes(3).toString('hex');
    // Keep domain <= 20 chars to satisfy form/backend constraints.
    const domainPrefix = prefix.toLowerCase().replace(/[^a-z0-9-]/g, '').slice(0, 8) || 'user';
    const domain = `${domainPrefix}-${ts}${nonce}`.slice(0, 20);

    return {
        email: `${domain}@example.com`,
        domain,
        name: `${prefix} ${nonce}`,
        password: 'Password123!',
    };
}

// ─── MailHog ──────────────────────────────────────────────────────────────────

const MAILHOG_BASE = process.env.E2E_MAILHOG_URL ?? 'http://localhost:8025';

interface MailhogResponse {
    items: Array<{
        Content: {
            Headers: Record<string, string[]>;
            Body: string;
        };
    }>;
}

/**
 * Polls the MailHog HTTP API until an email addressed to `to` appears,
 * or the timeout is reached.
 *
 * @returns true if the email was found within the timeout.
 */
export async function waitForMailhogEmail(to: string, timeoutMs = 15_000): Promise<boolean> {
    const deadline = Date.now() + timeoutMs;

    while (Date.now() < deadline) {
        try {
            const res = await fetch(`${MAILHOG_BASE}/api/v2/messages`);
            if (res.ok) {
                const data = (await res.json()) as MailhogResponse;
                for (const item of data.items) {
                    const toHeaders = item.Content?.Headers?.['To'] ?? [];
                    if (toHeaders.some(addr => addr.includes(to))) {
                        return true;
                    }
                }
            }
        } catch {
            // MailHog not yet ready – keep polling.
        }
        await new Promise(r => setTimeout(r, 500));
    }
    return false;
}

// ─── Auth flow helper ─────────────────────────────────────────────────────────

type User = ReturnType<typeof uniqueUser>;

/**
 * Registers and signs in a test user, leaving the browser on the user's
 * profile page.  Returns the user object for later use.
 */
export async function registerAndLogin(page: Page, prefix: string): Promise<User> {
    await mockRecaptcha(page);
    const user = uniqueUser(prefix);

    // ── Sign up
    await page.goto('/sign-up');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="domain"]').fill(user.domain);
    await page.locator('input[name="name"]').fill(user.name);
    await page.locator('input[name="password"]').fill(user.password);
    await page.locator('input[name="repeatPassword"]').fill(user.password);
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('/sign-in');

    // ── Sign in
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);
    await page.locator('button[type="submit"]').click();
    await page.waitForURL(`/_/${user.domain}`);

    return user;
}

