import axios, {type AxiosResponse} from 'axios'
import router from '@/router'
import {useAuthStore} from '@/store'
import {ToastError} from '@/utils/notify.ts'
import {Api} from './api.ts'

interface ErrorResponse {
    error: number;
    msg: string;
}

class ApiClient extends Api<unknown> {
    // The generated client would send the absent values as "undefined" or "null" strings in the form data.
    protected override createFormData(input: Record<string, unknown>): FormData {
        if (input instanceof FormData) {
            return input
        }
        return super.createFormData(Object.fromEntries(
            Object.entries(input).filter(([, value]) => value !== undefined && value !== null),
        ))
    }
}

export const api = new ApiClient({
    baseURL: import.meta.env.VITE_BASE_URL || '/api',
})

// errorMessage returns the message of an error response, whose body is a Blob for the requests of files such as exporting the data.
async function errorMessage(data: unknown): Promise<string | undefined> {
    if (data instanceof Blob) {
        try {
            return (JSON.parse(await data.text()) as Partial<ErrorResponse>).msg
        } catch {
            return undefined
        }
    }
    return (data as Partial<ErrorResponse> | undefined)?.msg
}

api.instance.interceptors.request.use((config) => {
    const authStore = useAuthStore()
    if (authStore.isSignedIn) {
        config.headers.Authorization = `Token ${authStore.sessionID}`
    }
    return config
})

api.instance.interceptors.response.use(
    (response: AxiosResponse) => {
        const contentType = response.headers['content-type']
        // Files such as the exported data are returned as they are, so `.data` of the result is the Blob.
        if (typeof contentType !== 'string' || !contentType.includes('application/json')) {
            return response
        }
        // Return the whole `{ data }` body as the AxiosResponse, so `.data` of the result is the type declared in Swagger.
        return response.data as AxiosResponse
    },
    async (error: unknown) => {
        if (!axios.isAxiosError(error) || !error.response) {
            ToastError('网络错误，请检查网络连接')
            return Promise.reject(error)
        }

        if (error.response.status === 401) {
            const authStore = useAuthStore()
            authStore.signOut()

            if (router.currentRoute.value.name !== 'sign-in') {
                router.push({name: 'sign-in', query: {to: router.currentRoute.value.fullPath}})
            }
            return Promise.reject(error)
        }

        // GET 404s are page-level (e.g. profile/question pages) and the calling component
        // redirects on its own, so we suppress the toast there. For mutating requests a 404
        // means the business operation failed (e.g. forgot-password with an unknown email),
        // which the user must be told about.
        const isGetRequest = (error.config?.method ?? '').toLowerCase() === 'get'
        if (error.response.status !== 404 || !isGetRequest) {
            ToastError(await errorMessage(error.response.data) || '未知错误')
        }
        return Promise.reject(error)
    }
)
