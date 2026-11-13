// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package response

type VerifyCaptcha struct {
	// Token is the one-shot token passed as the captcha of the business endpoints.
	Token string `json:"token"`
} // @name VerifyCaptchaResp
