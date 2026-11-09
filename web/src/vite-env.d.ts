/// <reference types="vite/client" />

declare module '@e99p1ant/neko-pixel-umd/neko-pixel-app.es.js' {
    export const NekoPixelApp: import('vue').DefineComponent<{ editable?: boolean }>
    export function setAuthToken(token: string): void
    export function setBaseURL(url: string): void
}
