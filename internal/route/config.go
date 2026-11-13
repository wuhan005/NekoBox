package route

import (
	"github.com/wuhan005/NekoBox/internal/captcha"
	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/response"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// Get returns the public site configuration.
// @Summary Get the public site configuration
// @Tags Config
// @Produce json
// @Success 200 {object} response.GlobalConfig
// @ID getGlobalConfig
// @Router /config [get]
func (*ConfigHandler) Get(ctx context.Context, v captcha.Verifier) error {
	config := response.GlobalConfig{
		ExternalURL: conf.App.ExternalURL,
		ICP:         conf.App.ICP,
		Captcha:     response.CaptchaConfig{Type: v.Type()},
	}
	if config.Captcha.Type == captcha.TypeRecaptcha {
		config.Captcha.SiteKey = conf.Recaptcha.SiteKey
		config.Captcha.Domain = conf.Recaptcha.Domain
		if config.Captcha.Domain == "" {
			config.Captcha.Domain = "https://recaptcha.net"
		}
	}
	return ctx.Success(config)
}
