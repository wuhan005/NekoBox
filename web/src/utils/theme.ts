import {watch} from "vue";
import {useAppStore} from "@/store";

// ThemePreference is the user's choice; 'auto' follows the system color scheme.
export type ThemePreference = 'auto' | 'light' | 'dark'
export type ResolvedTheme = 'light' | 'dark'

export const THEME_PREFERENCES: ThemePreference[] = ['auto', 'light', 'dark']

export const THEME_LABELS: Record<ThemePreference, string> = {
    auto: '跟随系统',
    light: '浅色',
    dark: '深色',
}

const darkMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

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

// setupTheme syncs the stored preference to <html data-theme> and reacts to system color scheme changes.
// The initial data-theme is set earlier by the inline script in index.html, which must use the same storage key and values.
export function setupTheme() {
    const appStore = useAppStore()
    watch(() => appStore.theme, theme => applyTheme(normalizeThemePreference(theme)), {immediate: true})
    darkMediaQuery.addEventListener('change', () => applyTheme(normalizeThemePreference(appStore.theme)))
}
