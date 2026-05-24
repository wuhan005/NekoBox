// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package captcha abstracts different captcha drivers (reCAPTCHA, go-captcha) behind a unified Verifier and Service.
package captcha

import (
	"context"

	"github.com/flamego/cache"
	"github.com/flamego/flamego"
	"github.com/pkg/errors"
)

// Type identifies the captcha driver currently in use; the frontend uses it to pick the matching UI.
type Type string

const (
	TypeRecaptcha Type = "recaptcha"
	TypeGoCaptcha Type = "go_captcha"
)

// Common driver-agnostic errors. Handlers use them to differentiate client and server failures.
var (
	ErrVerifyFailed = errors.New("captcha: verify failed")
	ErrUnsupported  = errors.New("captcha: operation not supported by current driver")
	ErrInternal     = errors.New("captcha: internal error")
	ErrRateLimited  = errors.New("captcha: rate limited")
)

// ChallengeData carries the data the client needs to render a slide challenge. Unused for reCAPTCHA.
type ChallengeData struct {
	Key         string `json:"key"`
	Image       string `json:"image"`
	Thumb       string `json:"thumb"`
	ThumbX      int    `json:"thumbX"`
	ThumbY      int    `json:"thumbY"`
	ThumbWidth  int    `json:"thumbWidth"`
	ThumbHeight int    `json:"thumbHeight"`
}

// Verifier validates the token submitted by business endpoints. reCAPTCHA hits the Google API,
// go-captcha looks up a one-shot token in the cache.
type Verifier interface {
	Type() Type
	Verify(ctx context.Context, c cache.Cache, token, ip string) error
}

// Service extends Verifier with slide challenge generation and verification, used only by the
// dedicated captcha endpoints.
type Service interface {
	Verifier
	Generate(ctx context.Context, c cache.Cache) (*ChallengeData, error)
	VerifyChallenge(ctx context.Context, c cache.Cache, key string, x, y int) (string, error)
}

// Provider injects the Service into the request scope so handlers can pull captcha.Verifier or
// captcha.Service straight from the DI container.
func Provider(svc Service) flamego.Handler {
	return func(c flamego.Context) {
		c.MapTo(svc, (*Verifier)(nil))
		c.MapTo(svc, (*Service)(nil))
	}
}

// New builds the Service for the given driver type. An empty or unrecognized type falls back to
// reCAPTCHA to preserve backward compatibility.
func New(t Type) (Service, error) {
	switch t {
	case TypeGoCaptcha:
		return NewGoCaptchaService()
	case TypeRecaptcha, "":
		return NewRecaptchaService(), nil
	default:
		return nil, errors.Errorf("unknown captcha type: %q", t)
	}
}
