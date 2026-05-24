import {type CaptchaConfig, type CaptchaType, getCaptchaConfig} from '@/api/captcha.ts'

// E2E bypasses the captcha entirely: a fixed token is returned from the frontend, while the
// backend's e2e config uses reCAPTCHA universal test keys that accept any token.
export const E2E_BYPASS_TOKEN = 'e2e-test-token'

export const isE2EMode = import.meta.env.MODE === 'e2e'

const RECAPTCHA_TEST_SITE_KEY = '6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI'
export const isUsingRecaptchaTestKey = import.meta.env.VITE_RECAPTCHA_SITE_KEY === RECAPTCHA_TEST_SITE_KEY

let configPromise: Promise<CaptchaConfig> | null = null

// loadCaptchaConfig fetches the active captcha driver from the backend and caches the result for
// the lifetime of the page. On failure it falls back to "recaptcha" to preserve the legacy UX.
export function loadCaptchaConfig(): Promise<CaptchaConfig> {
    if (configPromise) {
        return configPromise
    }
    configPromise = getCaptchaConfig()
        .then(config => config ?? {type: 'recaptcha' as CaptchaType})
        .catch(() => ({type: 'recaptcha' as CaptchaType}))
    return configPromise
}

// resetCaptchaConfigCache clears the cached config; used by tests and by stores that need to
// re-fetch the driver after a configuration change.
export function resetCaptchaConfigCache() {
    configPromise = null
}
