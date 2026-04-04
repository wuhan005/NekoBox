import axios from 'axios';

export interface MineQuestionItem {
    id: number;
    createdAt: Date;
    content: string;
    isAnswered: boolean;
    isPrivate: boolean;
}

export interface MineQuestions {
    total: number;
    cursor: string;
    questions: MineQuestionItem[];
}

export function mineQuestions(cursor: string, pageSize: number) {
    return axios.get<MineQuestions, MineQuestions>('/mine/questions', {
        params: {
            cursor,
            pageSize,
        }
    });
}

export interface AnswerQuestionRequest {
    answer: string;
    images: File[];
}

export function answerQuestion(questionID: number | string, data: AnswerQuestionRequest) {
    return axios.put<string, string>(`/mine/questions/${questionID}/answer`, data, {
        headers: {
            'Content-Type': 'multipart/form-data',
        }
    });
}

export function deleteQuestion(questionID: number | string) {
    return axios.delete<string, string>(`/mine/questions/${questionID}`);
}

export function setQuestionVisible(questionID: number | string, isVisible: boolean) {
    return axios.put<string, string>(`/mine/questions/${questionID}/visible`, {
        visible: isVisible,
    });
}

export interface MineProfile {
    email: string;
    name: string;
}

export function getMineProfile() {
    return axios.get<MineProfile, MineProfile>('/mine/settings/profile');
}

export interface UpdateMineProfileRequest {
    name: string;
    oldPassword?: string;
    newPassword?: string;
}

export function updateMineProfile(data: UpdateMineProfileRequest) {
    return axios.put<string, string>('/mine/settings/profile', data);
}

export interface MineBoxSettings {
    intro: string;
    notifyType: 'none' | 'email';
    avatarURL: string;
    backgroundURL: string;
}

export function getMineBoxSettings() {
    return axios.get<MineBoxSettings, MineBoxSettings>('/mine/settings/box');
}

export interface UpdateMineBoxSettingsRequest {
    intro: string;
    notifyType: 'none' | 'email';
    avatar?: File | null;
    background?: File | null;
}

export function updateMineBoxSettings(data: UpdateMineBoxSettingsRequest) {
    return axios.put<string, string>('/mine/settings/box', data, {
        headers: {
            'Content-Type': 'multipart/form-data',
        }
    });
}

export interface MineHarassmentSettings {
    harassmentSettingType: 'none' | 'register_only';
    blockWords: string;
}

export function getMineHarassmentSettings() {
    return axios.get<MineHarassmentSettings, MineHarassmentSettings>('/mine/settings/harassment');
}

export interface UpdateMineHarassmentSettingsRequest {
    harassmentSettingType: 'none' | 'register_only';
    blockWords: string;
}

export function updateMineHarassmentSettings(data: UpdateMineHarassmentSettingsRequest) {
    return axios.put<string, string>('/mine/settings/harassment', data);
}

export function exportData() {
    return axios.post<Blob, Blob>('/mine/settings/export-data', {}, {
        responseType: 'blob',
    })
}

export function deactivateAccount() {
    return axios.post<string, string>('/mine/settings/deactivate', {})
}