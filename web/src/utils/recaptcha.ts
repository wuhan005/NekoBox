import {CAPTCHA_TIMEOUT_MS} from '@/utils/captcha.ts'
import {SiteKey} from '@/utils/consts.ts'

interface Recaptcha {
    ready(callback: () => void): void;
    execute(siteKey: string, options: {action: string}): Promise<string>;
}

declare global {
    interface Window {
        grecaptcha?: Recaptcha;
    }
}

let loading: Promise<Recaptcha> | null = null
let attempts = 0

// Share a successful load across forms, but remove failed scripts so retries can start fresh.
export function loadRecaptcha(): Promise<Recaptcha> {
    if (!SiteKey) return Promise.reject(new Error('未配置 reCAPTCHA'))
    if (loading) return loading

    loading = new Promise<Recaptcha>((resolve, reject) => {
        const script = document.createElement('script')
        const fail = () => {
            clearTimeout(timer)
            script.onload = script.onerror = null
            script.remove()
            reject(new Error('验证码加载失败，请重试'))
        }
        const timer = setTimeout(fail, CAPTCHA_TIMEOUT_MS)
        const loaded = () => {
            const api = window.grecaptcha
            if (!api) return fail()
            try {
                api.ready(() => {
                    clearTimeout(timer)
                    script.onload = script.onerror = null
                    resolve(api)
                })
            } catch {
                fail()
            }
        }
        // A new URL prevents a retry from sharing an earlier stalled browser request.
        script.src = `https://recaptcha.net/recaptcha/api.js?render=${encodeURIComponent(SiteKey)}&retry=${++attempts}`
        script.async = true
        script.onload = loaded
        script.onerror = fail
        document.head.appendChild(script)
    }).catch(error => {
        loading = null
        throw error
    })
    return loading
}

export async function executeRecaptcha(action: string): Promise<string> {
    const api = await loadRecaptcha()
    let timer: ReturnType<typeof setTimeout> | undefined
    try {
        return await Promise.race([
            api.execute(SiteKey, {action}),
            new Promise<never>((_, reject) => {
                timer = setTimeout(() => reject(new Error('验证码获取超时，请重试')), CAPTCHA_TIMEOUT_MS)
            }),
        ])
    } finally {
        clearTimeout(timer)
    }
}
