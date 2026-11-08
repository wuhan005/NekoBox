/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface CaptchaChallenge {
  image: string;
  key: string;
  thumb: string;
  thumbHeight: number;
  thumbWidth: number;
  thumbX: number;
  thumbY: number;
}

export interface CaptchaConfig {
  /** Type is the captcha driver in use, the frontend renders the matching UI. */
  type: "recaptcha" | "go_captcha";
}

export interface ForgotPassword {
  captcha: string;
  email: string;
}

export interface HarassmentSettings {
  blockWords: string;
  harassmentSettingType: "none" | "register_only";
}

export interface MineBoxSettings {
  avatarURL: string;
  backgroundURL: string;
  intro: string;
  notifyType: "none" | "email";
}

export interface MineProfile {
  email: string;
  name: string;
}

export interface MineQuestions {
  /** Cursor is passed to get the next page, empty if there are no questions. */
  cursor: string;
  questions: MineQuestionsItem[];
  total: number;
}

export interface MineQuestionsItem {
  content: string;
  /** @format date-time */
  createdAt: string;
  id: number;
  isAnswered: boolean;
  isPrivate: boolean;
}

export interface MineSentQuestions {
  /** Cursor is passed to get the next page, empty if there are no questions. */
  cursor: string;
  questions: MineSentQuestionsItem[];
  total: number;
}

export interface MineSentQuestionsItem {
  content: string;
  /** @format date-time */
  createdAt: string;
  id: number;
  isAnswered: boolean;
  isPrivate: boolean;
  /** TargetDomain and TargetName are empty if the receiver no longer exists. */
  targetDomain: string;
  targetName: string;
}

export interface PageQuestion {
  answer: string;
  answerImageURLs: string[];
  /** @format date-time */
  answeredAt: string;
  content: string;
  /** @format date-time */
  createdAt: string;
  hasReplyEmail: boolean;
  id: number;
  isOwner: boolean;
  isPrivate: boolean;
  questionImageURLs: string[];
}

export interface PageQuestions {
  /** Cursor is passed to get the next page, empty if there are no questions. */
  cursor: string;
  questions: PageQuestionsItem[];
  total: number;
}

export interface PageQuestionsItem {
  answer: string;
  content: string;
  /** @format date-time */
  createdAt: string;
  id: number;
}

export interface QuestionVisible {
  visible: boolean;
}

export interface RecoverPassword {
  code: string;
  newPassword: string;
  repeatPassword: string;
}

export interface RecoverPasswordResp {
  name: string;
}

export interface SignIn {
  captcha: string;
  email: string;
  password: string;
}

export interface SignInResp {
  profile: SignInUserProfile;
  /** SessionID is sent back as the `Authorization: Token <sessionID>` header to authenticate the later requests. */
  sessionID: string;
}

export interface SignInUserProfile {
  domain: string;
  name: string;
  uid: string;
}

export interface SignUp {
  captcha: string;
  domain: string;
  email: string;
  name: string;
  password: string;
  repeatPassword: string;
}

export interface UpdateHarassmentSettings {
  /** BlockWords are at most 10 words separated by commas, each of them is at most 10 bytes. */
  blockWords: string;
  /** HarassmentSettingType falls back to none if it is unknown. */
  harassmentSettingType: "none" | "register_only";
}

export interface UpdateProfile {
  name: string;
  /** NewPassword changes the password if it is not empty. */
  newPassword?: string;
  /** OldPassword is required only when changing the password. */
  oldPassword?: string;
}

export interface UserProfile {
  avatar: string;
  background: string;
  domain: string;
  harassmentSetting: "none" | "register_only";
  intro: string;
  name: string;
  uid: string;
}

export interface VerifyCaptcha {
  key: string;
  x: number;
  y: number;
}

export interface VerifyCaptchaResp {
  /** Token is the one-shot token passed as the captcha of the business endpoints. */
  token: string;
}

export type ForgotPasswordData = string;

export type GetRecoverPasswordCodeData = RecoverPasswordResp;

export type RecoverPasswordData = string;

export type SignInData = SignInResp;

export type SignUpData = string;

export type GetCaptchaChallengeData = CaptchaChallenge;

export type GetCaptchaConfigData = CaptchaConfig;

export type VerifyCaptchaData = VerifyCaptchaResp;

export type ListMineQuestionsData = MineQuestions;

export type ListMineSentQuestionsData = MineSentQuestions;

export type DeleteQuestionData = string;

export interface AnswerQuestionPayload {
  /** Answer, at most 1000 characters */
  answer: string;
  /** Image of the answer, at most 5 MiB */
  "images[]"?: File;
}

export type AnswerQuestionData = string;

export type SetQuestionVisibleData = string;

export type GetMineBoxSettingsData = MineBoxSettings;

export interface UpdateMineBoxSettingsPayload {
  /** Introduction, at most 100 characters */
  intro: string;
  /** Notification of the new questions */
  notifyType: "none" | "email";
  /**
   * Avatar, at most 5 MiB
   * @format binary
   */
  avatar?: File;
  /**
   * Background, at most 5 MiB
   * @format binary
   */
  background?: File;
}

export type UpdateMineBoxSettingsData = string;

export type DeactivateAccountData = string;

/** @format binary */
export type ExportMineDataData = File;

export type GetMineHarassmentSettingsData = HarassmentSettings;

export type UpdateMineHarassmentSettingsData = string;

export type GetMineProfileData = MineProfile;

export type UpdateMineProfileData = string;

export type GetUserProfileData = UserProfile;

export type ListUserQuestionsData = PageQuestions;

export interface PostQuestionPayload {
  /** Content of the question, at most 1000 characters */
  content: string;
  /** Email to notify when the question is answered */
  receiveReplyEmail?: string;
  /** Whether to keep the question private after it is answered */
  isPrivate?: boolean;
  /** Image of the question, at most 5 MiB */
  "images[]"?: File;
  /** Captcha token */
  captcha: string;
}

export type PostQuestionData = string;

export type GetUserQuestionData = PageQuestion;

import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  HeadersDefaults,
  ResponseType,
} from "axios";
import axios from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  JsonApi = "application/vnd.api+json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || "/api",
    });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[
            method.toLowerCase() as keyof HeadersDefaults
          ]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input;
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] =
        property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(
          key,
          isFileType ? formItem : this.stringifyFormItem(formItem),
        );
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (
      type === ContentType.FormData &&
      body &&
      body !== null &&
      typeof body === "object"
    ) {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (
      type === ContentType.Text &&
      body &&
      body !== null &&
      typeof body !== "string"
    ) {
      body = JSON.stringify(body);
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type ? { "Content-Type": type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: body,
      url: path,
    });
  };
}

/**
 * @title NekoBox API
 * @version 1.0
 * @baseUrl /api
 * @contact
 *
 * The successful responses wrap the documented schema as `{"data": ...}`, and the errors respond `{"error": <status code>, "msg": "<message>"}`.
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  auth = {
    /**
     * @description The recovery code in the email expires in 24 hours, and the email can be sent once every 2 minutes.
     *
     * @tags Authentication
     * @name ForgotPassword
     * @summary Send the password recovery email
     * @request POST:/auth/forgot-password
     */
    forgotPassword: (data: ForgotPassword, params: RequestParams = {}) =>
      this.request<ForgotPasswordData, string>({
        path: `/auth/forgot-password`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Authentication
     * @name GetRecoverPasswordCode
     * @summary Check a password recovery code
     * @request GET:/auth/recover-password
     */
    getRecoverPasswordCode: (
      query: {
        /** Recovery code in the email */
        code: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<GetRecoverPasswordCodeData, string>({
        path: `/auth/recover-password`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Authentication
     * @name RecoverPassword
     * @summary Reset the password with a recovery code
     * @request POST:/auth/recover-password
     */
    recoverPassword: (data: RecoverPassword, params: RequestParams = {}) =>
      this.request<RecoverPasswordData, string>({
        path: `/auth/recover-password`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Authentication
     * @name SignIn
     * @summary Sign in with email and password
     * @request POST:/auth/sign-in
     */
    signIn: (data: SignIn, params: RequestParams = {}) =>
      this.request<SignInData, string>({
        path: `/auth/sign-in`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Authentication
     * @name SignUp
     * @summary Sign up
     * @request POST:/auth/sign-up
     */
    signUp: (data: SignUp, params: RequestParams = {}) =>
      this.request<SignUpData, string>({
        path: `/auth/sign-up`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  captcha = {
    /**
     * @description Only available for the go-captcha driver.
     *
     * @tags Captcha
     * @name GetCaptchaChallenge
     * @summary Generate a slide captcha challenge
     * @request GET:/captcha/challenge
     */
    getCaptchaChallenge: (params: RequestParams = {}) =>
      this.request<GetCaptchaChallengeData, string>({
        path: `/captcha/challenge`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Captcha
     * @name GetCaptchaConfig
     * @summary Get the captcha config
     * @request GET:/captcha/config
     */
    getCaptchaConfig: (params: RequestParams = {}) =>
      this.request<GetCaptchaConfigData, any>({
        path: `/captcha/config`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * @description Only available for the go-captcha driver. The returned token is passed as the captcha of the business endpoints.
     *
     * @tags Captcha
     * @name VerifyCaptcha
     * @summary Verify a slide captcha challenge
     * @request POST:/captcha/verify
     */
    verifyCaptcha: (data: VerifyCaptcha, params: RequestParams = {}) =>
      this.request<VerifyCaptchaData, string>({
        path: `/captcha/verify`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  mine = {
    /**
     * No description
     *
     * @tags My Questions
     * @name ListMineQuestions
     * @summary List the questions received by the signed-in user
     * @request GET:/mine/questions
     * @secure
     */
    listMineQuestions: (
      query?: {
        /** Cursor returned by the previous page, empty for the first page */
        cursor?: string;
        /** Page size, defaults to 20 and at most 100 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<ListMineQuestionsData, string>({
        path: `/mine/questions`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags My Questions
     * @name ListMineSentQuestions
     * @summary List the questions asked by the signed-in user
     * @request GET:/mine/questions/sent
     * @secure
     */
    listMineSentQuestions: (
      query?: {
        /** Cursor returned by the previous page, empty for the first page */
        cursor?: string;
        /** Page size, defaults to 20 and at most 100 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<ListMineSentQuestionsData, string>({
        path: `/mine/questions/sent`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags My Questions
     * @name DeleteQuestion
     * @summary Delete a received question
     * @request DELETE:/mine/questions/{questionID}
     * @secure
     */
    deleteQuestion: (questionId: number, params: RequestParams = {}) =>
      this.request<DeleteQuestionData, string>({
        path: `/mine/questions/${questionId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description The asker is notified by email on the first answer if they left an email.
     *
     * @tags My Questions
     * @name AnswerQuestion
     * @summary Answer a received question
     * @request PUT:/mine/questions/{questionID}/answer
     * @secure
     */
    answerQuestion: (
      questionId: number,
      data: AnswerQuestionPayload,
      params: RequestParams = {},
    ) =>
      this.request<AnswerQuestionData, string>({
        path: `/mine/questions/${questionId}/answer`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.FormData,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags My Questions
     * @name SetQuestionVisible
     * @summary Set whether a received question is public
     * @request PUT:/mine/questions/{questionID}/visible
     * @secure
     */
    setQuestionVisible: (
      questionId: number,
      data: QuestionVisible,
      params: RequestParams = {},
    ) =>
      this.request<SetQuestionVisibleData, string>({
        path: `/mine/questions/${questionId}/visible`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Settings
     * @name GetMineBoxSettings
     * @summary Get the question box settings of the signed-in user
     * @request GET:/mine/settings/box
     * @secure
     */
    getMineBoxSettings: (params: RequestParams = {}) =>
      this.request<GetMineBoxSettingsData, string>({
        path: `/mine/settings/box`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description The avatar and the background are kept if they are not uploaded.
     *
     * @tags Settings
     * @name UpdateMineBoxSettings
     * @summary Update the question box settings of the signed-in user
     * @request PUT:/mine/settings/box
     * @secure
     */
    updateMineBoxSettings: (
      data: UpdateMineBoxSettingsPayload,
      params: RequestParams = {},
    ) =>
      this.request<UpdateMineBoxSettingsData, string>({
        path: `/mine/settings/box`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.FormData,
        format: "json",
        ...params,
      }),

    /**
     * @description The account can no longer sign in, its question box is no longer accessible, and the session is signed out.
     *
     * @tags Settings
     * @name DeactivateAccount
     * @summary Deactivate the signed-in user
     * @request POST:/mine/settings/deactivate
     * @secure
     */
    deactivateAccount: (params: RequestParams = {}) =>
      this.request<DeactivateAccountData, string>({
        path: `/mine/settings/deactivate`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Export the profile and the received questions as an Excel workbook.
     *
     * @tags Settings
     * @name ExportMineData
     * @summary Export the data of the signed-in user
     * @request POST:/mine/settings/export-data
     * @secure
     */
    exportMineData: (params: RequestParams = {}) =>
      this.request<ExportMineDataData, string>({
        path: `/mine/settings/export-data`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Settings
     * @name GetMineHarassmentSettings
     * @summary Get the harassment settings of the signed-in user
     * @request GET:/mine/settings/harassment
     * @secure
     */
    getMineHarassmentSettings: (params: RequestParams = {}) =>
      this.request<GetMineHarassmentSettingsData, string>({
        path: `/mine/settings/harassment`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Settings
     * @name UpdateMineHarassmentSettings
     * @summary Update the harassment settings of the signed-in user
     * @request PUT:/mine/settings/harassment
     * @secure
     */
    updateMineHarassmentSettings: (
      data: UpdateHarassmentSettings,
      params: RequestParams = {},
    ) =>
      this.request<UpdateMineHarassmentSettingsData, string>({
        path: `/mine/settings/harassment`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Settings
     * @name GetMineProfile
     * @summary Get the profile settings of the signed-in user
     * @request GET:/mine/settings/profile
     * @secure
     */
    getMineProfile: (params: RequestParams = {}) =>
      this.request<GetMineProfileData, string>({
        path: `/mine/settings/profile`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description The password is changed only if the new password is not empty.
     *
     * @tags Settings
     * @name UpdateMineProfile
     * @summary Update the profile settings of the signed-in user
     * @request PUT:/mine/settings/profile
     * @secure
     */
    updateMineProfile: (data: UpdateProfile, params: RequestParams = {}) =>
      this.request<UpdateMineProfileData, string>({
        path: `/mine/settings/profile`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  users = {
    /**
     * No description
     *
     * @tags Question Box
     * @name GetUserProfile
     * @summary Get the profile of a question box
     * @request GET:/users/{domain}/profile
     */
    getUserProfile: (domain: string, params: RequestParams = {}) =>
      this.request<GetUserProfileData, string>({
        path: `/users/${domain}/profile`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Question Box
     * @name ListUserQuestions
     * @summary List the answered public questions of a question box
     * @request GET:/users/{domain}/questions
     */
    listUserQuestions: (
      domain: string,
      query?: {
        /** Cursor returned by the previous page, empty for the first page */
        cursor?: string;
        /** Page size, defaults to 20 and at most 100 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<ListUserQuestionsData, string>({
        path: `/users/${domain}/questions`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Signing in is required if the owner only allows registered users to ask.
     *
     * @tags Question Box
     * @name PostQuestion
     * @summary Ask a question
     * @request POST:/users/{domain}/questions
     */
    postQuestion: (
      domain: string,
      data: PostQuestionPayload,
      params: RequestParams = {},
    ) =>
      this.request<PostQuestionData, string>({
        path: `/users/${domain}/questions`,
        method: "POST",
        body: data,
        type: ContentType.FormData,
        format: "json",
        ...params,
      }),

    /**
     * @description Unanswered or private questions are only visible to the owner, the asker, or with the token of the private link.
     *
     * @tags Question Box
     * @name GetUserQuestion
     * @summary Get a question of a question box
     * @request GET:/users/{domain}/questions/{questionID}
     */
    getUserQuestion: (
      domain: string,
      questionId: number,
      query?: {
        /** Token of the private link */
        t?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<GetUserQuestionData, string>({
        path: `/users/${domain}/questions/${questionId}`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),
  };
}
