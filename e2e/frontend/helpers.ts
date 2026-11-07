import { type Page, type Route } from '@playwright/test';

export const avatar = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/lp8AAAAASUVORK5CYII=';
export const image = { name: 'first-answer.png', mimeType: 'image/png', buffer: Buffer.from(avatar.split(',')[1], 'base64') };
export const profile = (domain: string) => ({
    uid: domain, domain, name: domain === 'alice' ? 'Alice' : 'Bob',
    intro: 'Existing introduction', avatar, background: avatar, harassmentSetting: 'none',
});
export const question = (id: number) => ({
    id, isOwner: true, content: `Question ${id}`, answer: `Answer ${id}`,
    createdAt: '2026-11-07T12:00:00+08:00', answeredAt: '2026-11-07T13:00:00+08:00',
    questionImageURLs: [], answerImageURLs: [], isPrivate: id === 1, hasReplyEmail: false,
});

export function respond(route: Route, data: unknown, status = 200) {
    return route.fulfill({ status, json: { data, msg: status >= 400 ? 'Request failed' : '' } });
}

export async function holdNextRequest(page: Page, url: string | RegExp) {
    let capture!: (route: Route) => void;
    const request = new Promise<Route>(resolve => { capture = resolve; });
    await page.route(url, capture, { times: 1 });
    return { request };
}

export const recaptchaScript = `window.grecaptcha = {
    ready: callback => callback(),
    execute: () => {
        window.__captchaCalls = (window.__captchaCalls || 0) + 1;
        return window.__holdCaptcha
            ? new Promise(resolve => { window.__resolveCaptcha = resolve; })
            : Promise.resolve('browser-test-token');
    }
};`;

export async function mockFrontend({page, baseURL}: {page: Page; baseURL?: string}) {
    let posted = 3;
    await page.addInitScript(() => localStorage.setItem('auth', JSON.stringify({
        isSignedIn: true, sessionID: 'test-session', profile: { uid: 'alice', domain: 'alice', name: 'Alice' },
    })));
    const origin = new URL(baseURL!).origin;
    await page.route('**/*', async route => {
        const request = route.request();
        const url = new URL(request.url());
        const path = url.pathname;
        if (url.hostname === 'recaptcha.net') return route.fulfill({contentType: 'application/javascript', body: recaptchaScript});
        if (path.startsWith('/api/')) {
            if (path === '/api/auth/sign-in') return respond(route, {profile: {uid: 'alice', domain: 'alice', name: 'Alice'}, sessionID: 'test-session'});
            if (path === '/api/auth/sign-up' || path === '/api/auth/forgot-password') return respond(route, 'Saved');
            if (path === '/api/mine/questions' || path === '/api/mine/questions/sent') return respond(route, {total: 3, cursor: '1', questions: [3, 2, 1].map(id => ({...question(id), isAnswered: true, targetDomain: 'alice', targetName: 'Alice'}))});
            if (path === '/api/captcha/config') return respond(route, { type: 'recaptcha' });
            if (path === '/api/mine/settings/profile') {
                return respond(route, request.method() === 'GET' ? { name: 'Alice', email: 'alice@example.com' } : 'Saved');
            }
            if (path === '/api/mine/settings/box') {
                return respond(route, request.method() === 'GET' ? {
                    intro: 'Existing introduction', notifyType: 'email', avatarURL: avatar, backgroundURL: avatar,
                } : 'Saved');
            }
            if (path === '/api/mine/settings/harassment') {
                return respond(route, request.method() === 'GET' ? {
                    harassmentSettingType: 'register_only', blockWords: 'blocked',
                } : 'Saved');
            }
            if (/^\/api\/users\/\w+\/profile$/.test(path)) return respond(route, profile(path.split('/')[3]));
            if (/^\/api\/users\/\w+\/questions$/.test(path)) {
                return respond(route, request.method() === 'GET' ? {
                    total: 3, cursor: '1', questions: [3, 2, 1].map(question),
                } : `发送问题成功！<a href="/_/${path.split('/')[3]}/${++posted}?t=private-${posted}" target="_blank">私密链接 ${posted}</a>`);
            }
            if (/^\/api\/users\/\w+\/questions\/\d+$/.test(path)) {
                return respond(route, question(Number(path.split('/').at(-1))));
            }
            if (/^\/api\/mine\/questions\/\d+\/(answer|visible)$/.test(path)) return respond(route, 'Saved');
            return respond(route, null, 404);
        }
        if (url.hostname === 'json.n3ko.cc') {
            return route.fulfill({ json: { data: [{ content: 'Test log', date: '2026-11-07' }] } });
        }
        if (url.origin !== origin) return route.abort();
        return route.continue();
    });
}
