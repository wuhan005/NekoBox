// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

// VerifyCaptcha is the slide-captcha answer payload. Key is the challenge ID issued by the server,
// X and Y are the final tile coordinates the user dragged to (X is allowed to be zero).
type VerifyCaptcha struct {
	Key string `json:"key" valid:"required" label:"验证码 ID"`
	X   int    `json:"x" label:"X 坐标"`
	Y   int    `json:"y" label:"Y 坐标"`
}
