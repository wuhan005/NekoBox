// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flamego/cache"
	"github.com/pkg/errors"

	"github.com/wuhan005/NekoBox/internal/conf"
)

// Verify URLs accepted by Google reCAPTCHA and Cloudflare Turnstile.
const (
	recaptchaVerifyURLGlobal     = "https://www.google.com/recaptcha/api/siteverify"
	turnstileVerifyURLCloudFlare = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

// recaptchaService implements the invisible captcha flow backed by Google reCAPTCHA v3 or
// Cloudflare Turnstile.
type recaptchaService struct {
	secret     string
	verifyURL  string
	httpClient *http.Client
}

// NewRecaptchaService builds a reCAPTCHA verifier from the global config. An empty server_key
// is allowed at construction time; the upstream call is only made when Verify runs.
func NewRecaptchaService() Service {
	verifyURL := recaptchaVerifyURLGlobal
	switch {
	case conf.Recaptcha.VerifyURL != "":
		verifyURL = conf.Recaptcha.VerifyURL
	case conf.Recaptcha.TurnstileStyle:
		// FYI: https://developers.cloudflare.com/turnstile/migration/migrating-from-recaptcha/
		verifyURL = turnstileVerifyURLCloudFlare
	}

	return &recaptchaService{
		secret:    conf.Recaptcha.ServerKey,
		verifyURL: verifyURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (*recaptchaService) Type() Type { return TypeRecaptcha }

// Verify forwards the token (and optional client IP) to the upstream verify endpoint.
func (s *recaptchaService) Verify(ctx context.Context, _ cache.Cache, token, ip string) error {
	if token == "" {
		return ErrVerifyFailed
	}

	form := url.Values{}
	form.Set("secret", s.secret)
	form.Set("response", token)
	if ip != "" {
		form.Set("remoteip", ip)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return errors.Wrap(ErrInternal, err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(ErrInternal, err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return errors.Wrap(ErrInternal, err.Error())
	}
	if !result.Success {
		return ErrVerifyFailed
	}
	return nil
}

func (*recaptchaService) Generate(_ context.Context, _ cache.Cache) (*ChallengeData, error) {
	return nil, ErrUnsupported
}

func (*recaptchaService) VerifyChallenge(_ context.Context, _ cache.Cache, _ string, _, _ int) (string, error) {
	return "", ErrUnsupported
}
