import {defineStore} from "pinia";
import {type AppState} from "./types";
import {type ThemePreference} from "@/utils/theme.ts";

const useAppStore = defineStore('app', {
    persist: {pick: ['theme']},

    state: (): AppState => ({
        theme: 'auto',
        questionsRevision: 0,
        publicQuestionsRevision: 0,
    }),

    actions: {
        setTheme(theme: ThemePreference) {
            this.theme = theme;
        },
        invalidateQuestions() {
            this.questionsRevision++;
        },
        invalidateAnsweredQuestions() {
            this.invalidateQuestions();
            this.publicQuestionsRevision++;
        },
    }
})

export default useAppStore;
