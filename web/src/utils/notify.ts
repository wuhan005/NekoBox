import {toast, type ToastOptions} from 'vue3-toastify';
import {currentTheme} from "@/utils/theme.ts";

const defaultOptions: ToastOptions & { clearOnUrlChange: boolean } = {
    "position": "top-center",
    "pauseOnHover": false,
    "autoClose": 1500,
    "hideProgressBar": true,
    "transition": "slide",
    "clearOnUrlChange": false,
}

// The theme is read per toast so that a manually selected theme is respected ('auto' would only follow the system).
function toastOptions(type: ToastOptions['type']): ToastOptions {
    return {
        type,
        theme: currentTheme(),
        ...defaultOptions,
    }
}

export function ToastInfo(message: string) {
    toast(message, toastOptions("info"));
}

export function ToastSuccess(message: string) {
    toast(message, toastOptions("success"));
}

export function ToastError(message: string) {
    toast(message, toastOptions("error"));
}
