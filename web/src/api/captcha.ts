import axios from 'axios'

export type CaptchaType = 'recaptcha' | 'go_captcha'

export interface CaptchaConfig {
    type: CaptchaType;
}

export interface CaptchaChallenge {
    key: string;
    image: string;
    thumb: string;
    thumbX: number;
    thumbY: number;
    thumbWidth: number;
    thumbHeight: number;
}

export interface CaptchaVerifyRequest {
    key: string;
    x: number;
    y: number;
}

export interface CaptchaVerifyResponse {
    token: string;
}

// getCaptchaConfig fetches the captcha driver enabled by the backend so the UI can adapt.
// Called once per session.
export function getCaptchaConfig() {
    return axios.get<CaptchaConfig, CaptchaConfig>('/captcha/config')
}

// getCaptchaChallenge fetches a fresh slide challenge. Only available when the backend
// runs in go_captcha mode.
export function getCaptchaChallenge() {
    return axios.get<CaptchaChallenge, CaptchaChallenge>('/captcha/challenge')
}

// verifyCaptchaSlide submits the user's slide answer; on success returns a one-shot token
// that the caller forwards to business endpoints.
export function verifyCaptchaSlide(data: CaptchaVerifyRequest) {
    return axios.post<CaptchaVerifyResponse, CaptchaVerifyResponse>('/captcha/verify', data)
}
