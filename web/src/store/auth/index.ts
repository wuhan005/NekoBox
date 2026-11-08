import {defineStore} from "pinia";
import {type AuthState} from "./types";
import type {SignInUserProfile} from "@/api/api.ts";

const useAuthStore = defineStore('auth', {
    persist: true,

    state: (): AuthState => ({
        isSignedIn: false,
        profile: {} as SignInUserProfile,
        sessionID: '',
    }),

    actions: {
        signIn(profile: SignInUserProfile, sessionID: string) {
            this.isSignedIn = true;
            this.profile = profile;
            this.sessionID = sessionID;
        },
        setProfileName(name: string) {
            this.profile.name = name;
        },
        signOut() {
            this.isSignedIn = false;
            this.profile = {} as SignInUserProfile;
            this.sessionID = '';
        }
    }
})

export default useAuthStore;