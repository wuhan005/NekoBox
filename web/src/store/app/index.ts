import {defineStore} from "pinia";
import {type AppState} from "./types";

const useAppStore = defineStore('app', {
    persist: {pick: ['theme']},

    state: (): AppState => ({
        theme: '',
        questionsRevision: 0,
        publicQuestionsRevision: 0,
    }),

    actions: {
        setTheme(theme: string) {
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
