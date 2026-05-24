import type {AxiosResponse} from 'axios';
import axios from 'axios';
import router from '@/router'
import {useAuthStore} from "@/store";
import {ToastError} from "@/utils/notify.ts";

export interface HttpResponse<T = unknown> {
    msg: any;
    data: T;
}

axios.defaults.baseURL = import.meta.env.VITE_BASE_URL || '/api';

axios.interceptors.request.use(
    (config) => {
        const authStore = useAuthStore()
        if (authStore.isSignedIn) {
            config.headers.Authorization = `Token ${authStore.sessionID}`;
        }

        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
)
axios.interceptors.response.use(
    (response: AxiosResponse) => {
        const contentType = response.headers['content-type'];
        if (!contentType || !contentType.includes('application/json')) {
            return response.data;
        }

        const res = response.data;
        return res.data;
    },
    (error) => {
        if (!error.response) {
            ToastError('网络错误，请检查网络连接')
            return Promise.reject(error);
        }

        if (error.response.status === 401) {
            const authStore = useAuthStore()
            authStore.signOut()

            router.push({name: 'sign-in'})
            return Promise.reject(error);
        }

        // GET 404s are page-level (e.g. profile/question pages) and the calling component
        // redirects on its own, so we suppress the toast there. For mutating requests a 404
        // means the business operation failed (e.g. forgot-password with an unknown email),
        // which the user must be told about.
        const isGetRequest = (error.config?.method ?? '').toLowerCase() === 'get'
        if (error.response.status !== 404 || !isGetRequest) {
            ToastError(error.response?.data?.msg || '未知错误')
        }
        return Promise.reject(error);
    }
);
