import axios from 'axios'

export interface Profile {
    uid: string;
    name: string;
    avatar: string;
    domain: string;
    background: string;
    intro: string;
    harassmentSetting: string;
}

export function getUserProfile(domain: string) {
    return axios.get<Profile, Profile>(`/users/${domain}/profile`)
}

export interface PageQuestionItem {
    id: number;
    createdAt: Date;
    content: string;
    answer: string;
}

export interface PageQuestions {
    total: number;
    cursor: string;
    questions: PageQuestionItem[];
}

export function getUserQuestions(domain: string, cursor: string, pageSize: number) {
    return axios.get<PageQuestions, PageQuestions>(`/users/${domain}/questions`, {
        params: {
            cursor,
            pageSize,
        }
    })
}

export interface PostQuestionRequest {
    content: string;
    receiveReplyEmail: string;
    images: File[];
    isPrivate: boolean;
    captcha: string;
}

export function postQuestion(domain: string, data: PostQuestionRequest) {
    return axios.post<string, string>(`/users/${domain}/questions`, data, {
        headers: {
            'Content-Type': 'multipart/form-data',
        }
    })
}

export interface PageQuestion {
    id: number;
    isOwner: boolean;
    createdAt: Date | null;
    answeredAt: Date | null;
    content: string;
    answer: string;
    questionImageURLs: string[];
    answerImageURLs: string[];
    isPrivate: boolean;
    hasReplyEmail: boolean;
}

export function getUserQuestion(domain: string, questionID: number | string, token?: string) {
    return axios.get<PageQuestion, PageQuestion>(`/users/${domain}/questions/${questionID}`, {
        params: {
            t: token,
        }
    })
}