// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

type SignIn struct {
	Email    string `json:"email" valid:"required;email;maxlen:100" label:"电子邮箱"`
	Password string `json:"password" valid:"required" label:"密码"`
	Captcha  string `json:"captcha" valid:"required" label:"验证码" msg:"验证码加载错误，请尝试刷新页面重试。"`
}

type SignUp struct {
	Email          string `json:"email" valid:"required;email;maxlen:100" label:"电子邮箱"`
	Domain         string `json:"domain" valid:"required;alphadash;minlen:3;maxlen:20" label:"个性域名"`
	Name           string `json:"name" valid:"required;maxlen:20" label:"昵称"`
	Password       string `json:"password" valid:"required;minlen:8;maxlen:30" label:"密码"`
	RepeatPassword string `json:"repeatPassword" valid:"required;equal:Password" label:"重复密码"`
	Captcha        string `json:"captcha" valid:"required" label:"验证码" msg:"验证码加载错误，请尝试刷新页面重试。"`
}

type ForgotPassword struct {
	Email   string `json:"email" valid:"required;email;maxlen:100" label:"电子邮箱"`
	Captcha string `json:"captcha" valid:"required" label:"验证码" msg:"验证码加载错误，请尝试刷新页面重试。"`
}

type RecoverPassword struct {
	NewPassword    string `json:"newPassword" valid:"required;minlen:8;maxlen:30" label:"新密码"`
	RepeatPassword string `json:"repeatPassword" valid:"required;equal:NewPassword" label:"重复密码"`
	Code           string `json:"code" valid:"required" label:"恢复码"`
}
