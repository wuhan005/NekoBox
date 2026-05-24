// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/flamego/cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/NekoBox/internal/captcha"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/form"
)

type CaptchaHandler struct{}

func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{}
}

// captchaConfigResponse is exposed to the frontend so it can decide which captcha UI to render.
type captchaConfigResponse struct {
	Type string `json:"type"`
}

// Config returns the active captcha driver so the frontend can switch UI accordingly.
func (*CaptchaHandler) Config(ctx context.Context, v captcha.Verifier) error {
	return ctx.Success(captchaConfigResponse{Type: string(v.Type())})
}

// Challenge generates a slide challenge image. Only available for the go-captcha driver.
func (*CaptchaHandler) Challenge(ctx context.Context, svc captcha.Service, c cache.Cache) error {
	if err := captcha.CheckChallengeRateLimit(ctx.Request().Context(), c, ctx.IP()); err != nil {
		if errors.Is(err, captcha.ErrRateLimited) {
			return ctx.Error(http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to check captcha challenge rate limit")
		return ctx.ServerError()
	}

	data, err := svc.Generate(ctx.Request().Context(), c)
	if err != nil {
		if errors.Is(err, captcha.ErrUnsupported) {
			return ctx.Error(http.StatusBadRequest, "当前验证码模式不支持滑块挑战")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to generate captcha challenge")
		return ctx.ServerError()
	}
	return ctx.Success(data)
}

// Verify checks the slide answer and, on success, returns a one-shot token for business endpoints.
func (*CaptchaHandler) Verify(ctx context.Context, svc captcha.Service, c cache.Cache, f form.VerifyCaptcha) error {
	token, err := svc.VerifyChallenge(ctx.Request().Context(), c, f.Key, f.X, f.Y)
	if err != nil {
		switch {
		case errors.Is(err, captcha.ErrUnsupported):
			return ctx.Error(http.StatusBadRequest, "当前验证码模式不支持滑块挑战")
		case errors.Is(err, captcha.ErrVerifyFailed):
			return ctx.Error(http.StatusBadRequest, "滑动校验失败，请重试")
		default:
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to verify captcha challenge")
			return ctx.ServerError()
		}
	}
	return ctx.Success(map[string]string{"token": token})
}
