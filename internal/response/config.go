package response

import "github.com/wuhan005/NekoBox/internal/captcha"

type GlobalConfig struct {
	ExternalURL string        `json:"externalURL"`
	ICP         string        `json:"icp"`
	Captcha     CaptchaConfig `json:"captcha"`
} // @name GlobalConfig

type CaptchaConfig struct {
	Type captcha.Type `json:"type" swaggertype:"string" enums:"recaptcha,go_captcha"`
	// SiteKey is the public reCAPTCHA site key.
	SiteKey string `json:"siteKey,omitempty"`
	// Domain is the origin used to load the reCAPTCHA script.
	Domain string `json:"domain,omitempty"`
} // @name CaptchaConfig
