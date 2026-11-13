import {CAPTCHA_TIMEOUT_MS} from '@/utils/captcha.ts'
import type {CaptchaConfig} from '@/api/api.ts'

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
export async function loadRecaptcha({siteKey, domain}: CaptchaConfig): Promise<Recaptcha> {
    if (!siteKey || !domain) throw new Error('未配置 reCAPTCHA')
    if (loading) return loading

    const scriptURL = new URL('/recaptcha/api.js', domain)
    // A new URL prevents a retry from sharing an earlier stalled browser request.
    scriptURL.search = new URLSearchParams({render: siteKey, retry: String(++attempts)}).toString()
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
        script.src = scriptURL.href
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

export async function executeRecaptcha(config: CaptchaConfig, action: string): Promise<string> {
    const {siteKey} = config
    if (!siteKey) throw new Error('未配置 reCAPTCHA')
    const api = await loadRecaptcha(config)
    let timer: ReturnType<typeof setTimeout> | undefined
    try {
        return await Promise.race([
            api.execute(siteKey, {action}),
            new Promise<never>((_, reject) => {
                timer = setTimeout(() => reject(new Error('验证码获取超时，请重试')), CAPTCHA_TIMEOUT_MS)
            }),
        ])
    } finally {
        clearTimeout(timer)
    }
}
