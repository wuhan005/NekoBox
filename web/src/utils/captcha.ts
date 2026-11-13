export const CAPTCHA_TIMEOUT_MS = 15000

// E2E bypasses the captcha entirely: a fixed token is returned from the frontend, while the
// backend's e2e config uses reCAPTCHA universal test keys that accept any token.
export const E2E_BYPASS_TOKEN = 'e2e-test-token'

export const isE2EMode = import.meta.env.MODE === 'e2e'
