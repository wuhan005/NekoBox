package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wuhan005/NekoBox/internal/conf"
)

func TestGlobalConfig(t *testing.T) {
	app, captcha, recaptcha, redis := conf.App, conf.Captcha, conf.Recaptcha, conf.Redis
	t.Cleanup(func() { conf.App, conf.Captcha, conf.Recaptcha, conf.Redis = app, captcha, recaptcha, redis })
	conf.App.ExternalURL = "https://box.example.com"
	conf.App.ICP = "Example ICP"
	conf.Recaptcha.SiteKey = "public-key"
	conf.Recaptcha.ServerKey = "private-key"
	conf.Recaptcha.VerifyURL = "https://internal.example/verify"
	conf.Redis.Addr = "" // Use the router's in-memory session store.

	for _, tt := range []struct {
		name, driver, domain, want string
	}{
		{"default driver and domain", "", "", `{"type":"recaptcha","siteKey":"public-key","domain":"https://recaptcha.net"}`},
		{"custom script domain", "recaptcha", "https://www.recaptcha.net", `{"type":"recaptcha","siteKey":"public-key","domain":"https://www.recaptcha.net"}`},
		{"slide captcha", "go_captcha", "https://www.recaptcha.net", `{"type":"go_captcha"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			conf.Captcha.Type, conf.Recaptcha.Domain = tt.driver, tt.domain
			router, err := New(nil, nil)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/config", nil))
			require.Equal(t, http.StatusOK, w.Code)
			require.JSONEq(t, `{"data":{"externalURL":"https://box.example.com","icp":"Example ICP","captcha":`+tt.want+`}}`, w.Body.String())
		})
	}
}
