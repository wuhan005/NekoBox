import {watch} from "vue";
import {useAppStore, useAuthStore} from "@/store";
import {api} from "@/api/client.ts";

// ThemePreference is the user's choice; 'auto' follows the system color scheme.
export type ThemePreference = 'auto' | 'light' | 'dark'
export type ResolvedTheme = 'light' | 'dark'

export const THEME_PREFERENCES: ThemePreference[] = ['auto', 'light', 'dark']

export const THEME_LABELS: Record<ThemePreference, string> = {
    auto: '跟随系统',
    light: '浅色',
    dark: '深色',
}

const ACCOUNT_SAVE_DELAY_MS = 300

const darkMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

// Incremented by every local change, so that an account theme fetched before the change does not override it.
let localRevision = 0
let saveTimer: number | undefined

// normalizeThemePreference treats legacy persisted values (e.g. an empty string) as 'auto'.
export function normalizeThemePreference(value: unknown): ThemePreference {
    return value === 'light' || value === 'dark' ? value : 'auto'
}

export function resolveTheme(preference: ThemePreference): ResolvedTheme {
    if (preference === 'auto') return darkMediaQuery.matches ? 'dark' : 'light'
    return preference
}

// currentTheme returns the theme currently applied to the page.
export function currentTheme(): ResolvedTheme {
    return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'
}

function applyTheme(preference: ThemePreference) {
    document.documentElement.dataset.theme = resolveTheme(preference)
}

// setThemePreference applies the preference locally and saves it to the account when signed in.
// Failing to save keeps the local change, so the request is silent.
export function setThemePreference(preference: ThemePreference) {
    const appStore = useAppStore()
    const authStore = useAuthStore()
    localRevision++
    appStore.setTheme(preference)

    if (!authStore.isSignedIn) return
    // Rapid toggles are coalesced into a single request carrying the last choice.
    window.clearTimeout(saveTimer)
    saveTimer = window.setTimeout(() => {
        if (!authStore.isSignedIn) return
        api.mine.updateMineThemeSettings({theme: preference}, {silent: true}).catch(() => {
        })
    }, ACCOUNT_SAVE_DELAY_MS)
}

// syncAccountTheme applies the theme saved in the account, keeping the local preference if the account has none.
async function syncAccountTheme() {
    const appStore = useAppStore()
    const revision = localRevision
    try {
        const {data} = await api.mine.getMineThemeSettings({silent: true})
        if (data.theme && revision === localRevision) appStore.setTheme(data.theme)
    } catch {
        // The local preference is kept if the account theme cannot be fetched.
    }
}

// setupTheme syncs the stored preference to <html data-theme>, reacts to system color scheme changes,
// and loads the account theme whenever a session is signed in, including the one restored on startup.
// The initial data-theme is set earlier by the inline script in index.html, which must use the same storage key and values.
export function setupTheme() {
    const appStore = useAppStore()
    const authStore = useAuthStore()
    watch(() => appStore.theme, theme => applyTheme(normalizeThemePreference(theme)), {immediate: true})
    darkMediaQuery.addEventListener('change', () => applyTheme(normalizeThemePreference(appStore.theme)))
    watch(() => authStore.isSignedIn ? authStore.sessionID : '', sessionID => {
        if (sessionID) void syncAccountTheme()
    }, {immediate: true})
}
