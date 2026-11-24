import type {ThemePreference} from "@/utils/theme.ts";

export interface AppState {
    theme: ThemePreference;
    questionsRevision: number;
    publicQuestionsRevision: number;
}
