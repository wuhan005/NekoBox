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
	"github.com/wuhan005/NekoBox/internal/response"
)

type CaptchaHandler struct{}

func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{}
}

// Challenge generates a slide challenge image. Only available for the go-captcha driver.
// @Summary Generate a slide captcha challenge
// @Tags Captcha
// @Description Only available for the go-captcha driver.
// @Produce json
// @Success 200 {object} captcha.ChallengeData
// @Failure 400 {string} string "The captcha driver does not support slide challenges"
// @Failure 429 {string} string "Too many requests"
// @Failure 500 {string} string "Internal server error"
// @ID getCaptchaChallenge
// @Router /captcha/challenge [get]
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
// @Summary Verify a slide captcha challenge
// @Tags Captcha
// @Description Only available for the go-captcha driver. The returned token is passed as the captcha of the business endpoints.
// @Accept json
// @Produce json
// @Param data body form.VerifyCaptcha true "Slide answer"
// @Success 200 {object} response.VerifyCaptcha
// @Failure 400 {string} string "Invalid request body, wrong answer, or the captcha driver does not support slide challenges"
// @Failure 500 {string} string "Internal server error"
// @ID verifyCaptcha
// @Router /captcha/verify [post]
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
	return ctx.Success(response.VerifyCaptcha{Token: token})
}
