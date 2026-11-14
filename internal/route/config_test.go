package route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/settings"
)

// memorySystemSettings is an in-memory db.SystemSettingsStore that only supports Get.
type memorySystemSettings struct {
	db.SystemSettingsStore
	settings map[string]*db.SystemSetting
	readErr  error
}

func (s *memorySystemSettings) Get(_ context.Context, key string) (*db.SystemSetting, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	setting, ok := s.settings[key]
	if !ok {
		return nil, db.ErrSystemSettingNotExist
	}
	return setting, nil
}

func TestGlobalConfig(t *testing.T) {
	app, captcha, recaptcha, redis, systemSettings := conf.App, conf.Captcha, conf.Recaptcha, conf.Redis, db.SystemSettings
	t.Cleanup(func() {
		conf.App, conf.Captcha, conf.Recaptcha, conf.Redis, db.SystemSettings = app, captcha, recaptcha, redis, systemSettings
	})
	conf.App.ExternalURL = "https://box.example.com"
	conf.App.ICP = "Static ICP"
	conf.Recaptcha.SiteKey = "public-key"
	conf.Recaptcha.ServerKey = "private-key"
	conf.Recaptcha.VerifyURL = "https://internal.example/verify"
	conf.Redis.Addr = "" // Use the router's in-memory session store.
	store := &memorySystemSettings{settings: map[string]*db.SystemSetting{
		settings.KeyICP: {Key: settings.KeyICP, Type: db.SystemSettingTypeString, Value: "Example ICP"},
	}}
	db.SystemSettings = store

	for _, tt := range []struct {
		name, driver, domain, icp, wantCaptcha string
		readErr                                error
	}{
		{"default driver and domain", "", "", "Example ICP", `{"type":"recaptcha","siteKey":"public-key","domain":"https://recaptcha.net"}`, nil},
		{"custom script domain", "recaptcha", "https://www.recaptcha.net", "Example ICP", `{"type":"recaptcha","siteKey":"public-key","domain":"https://www.recaptcha.net"}`, nil},
		{"slide captcha", "go_captcha", "https://www.recaptcha.net", "Example ICP", `{"type":"go_captcha"}`, nil},
		{"settings read fails", "recaptcha", "", "", `{"type":"recaptcha","siteKey":"public-key","domain":"https://recaptcha.net"}`, context.DeadlineExceeded},
		{"ICP key is missing", "go_captcha", "", "", `{"type":"go_captcha"}`, db.ErrSystemSettingNotExist},
	} {
		t.Run(tt.name, func(t *testing.T) {
			conf.Captcha.Type, conf.Recaptcha.Domain = tt.driver, tt.domain
			store.readErr = tt.readErr
			router, err := New(nil, nil)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/config", nil))
			require.Equal(t, http.StatusOK, w.Code)
			require.JSONEq(t, `{"data":{"externalURL":"https://box.example.com","icp":"`+tt.icp+`","captcha":`+tt.wantCaptcha+`}}`, w.Body.String())
		})
	}
}
