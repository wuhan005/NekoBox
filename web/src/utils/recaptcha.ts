import type {IReCaptchaComposition} from 'vue-recaptcha-v3'

const E2E_RECAPTCHA_TOKEN = 'e2e-test-token'
const RECAPTCHA_TEST_SITE_KEY = '6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI'

type RecaptchaClient = Pick<IReCaptchaComposition, 'executeRecaptcha' | 'recaptchaLoaded'>

export const isE2EMode = import.meta.env.MODE === 'e2e'
export const isUsingRecaptchaTestKey = import.meta.env.VITE_RECAPTCHA_SITE_KEY === RECAPTCHA_TEST_SITE_KEY
const shouldBypassRecaptcha = isE2EMode

export async function ensureRecaptchaReady(recaptcha: RecaptchaClient) {
    if (shouldBypassRecaptcha) {
        return
    }

    await recaptcha.recaptchaLoaded()
}

export async function getRecaptchaToken(recaptcha: RecaptchaClient, action = 'submit') {
    if (shouldBypassRecaptcha) {
        return E2E_RECAPTCHA_TOKEN
    }

    await recaptcha.recaptchaLoaded()
    return await recaptcha.executeRecaptcha(action)
}
