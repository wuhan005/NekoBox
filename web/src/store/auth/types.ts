import type {SignInUserProfile} from "@/api/api.ts";

export interface AuthState {
    isSignedIn: boolean;

    profile: SignInUserProfile;
    sessionID: string;
}