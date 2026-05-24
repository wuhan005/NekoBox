import axios from 'axios'


export interface SignUpRequest {
    email: string;
    domain: string;
    name: string;
    password: string;
    repeatPassword: string;
    captcha: string;
}

export function signUp(data: SignUpRequest) {
    return axios.post<string, string>('/auth/sign-up', data)
}

export interface SignInRequest {
    email: string;
    password: string;
    captcha: string;
}

export interface SignInResponse {
    profile: UserProfile;
    sessionID: string;
}

export interface UserProfile {
    uid: string;
    name: string;
    domain: string;
}

export function signIn(data: SignInRequest) {
    return axios.post<SignInResponse, SignInResponse>('/auth/sign-in', data)
}

export interface ForgotPasswordRequest {
    email: string;
    captcha: string;
}

export function forgotPassword(data: ForgotPasswordRequest) {
    return axios.post<string, string>('/auth/forgot-password', data);
}

export interface GetRecoverPasswordCodeResponse {
    name: string;
}

export function getRecoverPasswordCode(code: string) {
    return axios.get<GetRecoverPasswordCodeResponse, GetRecoverPasswordCodeResponse>('/auth/recover-password', {
        params: {
            code,
        }
    });
}

export interface RecoverPasswordRequest {
    newPassword: string;
    repeatPassword: string;
    code: string;
}

export function recoverPassword(data: RecoverPasswordRequest) {
    return axios.post<string, string>('/auth/recover-password', data);
}