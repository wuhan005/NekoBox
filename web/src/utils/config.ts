import {computed, shallowRef} from 'vue'
import {api} from '@/api/client.ts'
import type {GlobalConfig} from '@/api/api.ts'

export const globalConfig = shallowRef<GlobalConfig>()
export const externalURL = computed(() => globalConfig.value?.externalURL || window.location.origin)

let configPromise: Promise<GlobalConfig> | undefined

// Share the request across the app and allow retries after a failed load.
export function loadGlobalConfig(): Promise<GlobalConfig> {
    return configPromise ??= api.config.getGlobalConfig({timeout: 15000})
        .then(({data: config}) => {
            const driver = config?.captcha?.type
            if (driver !== 'recaptcha' && driver !== 'go_captcha') {
                throw new Error('未知的验证码类型')
            }
            globalConfig.value = config
            return config
        })
        .catch(error => {
            configPromise = undefined
            throw error
        })
}
