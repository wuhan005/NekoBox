package route

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/dbutil"
	"github.com/wuhan005/NekoBox/internal/response"
)

func TestSessionRevocation(t *testing.T) {
	database, _ := dbutil.NewTestDB(t, &db.User{})
	users := db.Users
	captchaConfig, recaptchaConfig, redisConfig := conf.Captcha, conf.Recaptcha, conf.Redis
	t.Cleanup(func() {
		db.Users = users
		conf.Captcha, conf.Recaptcha, conf.Redis = captchaConfig, recaptchaConfig, redisConfig
	})
	db.Users = db.NewUsersStore(database)
	verifier := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	t.Cleanup(verifier.Close)
	conf.Captcha.Type, conf.Recaptcha.VerifyURL, conf.Redis.Addr = "recaptcha", verifier.URL, ""

	router, err := New(database, nil)
	require.NoError(t, err)

	createUser := func(t *testing.T, email string) *db.User {
		user := db.User{Domain: strings.Split(email, "@")[0], Name: "Session Owner", Email: email, Password: "old-password"}
		user.EncodePassword()
		require.NoError(t, database.Create(&user).Error)
		return &user
	}
	signIn := func(t *testing.T, email, password string) string {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/auth/sign-in", strings.NewReader(
			`{"email":"`+email+`","password":"`+password+`","captcha":"test"}`,
		)))
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		var auth struct{ Data response.SignIn }
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &auth))
		require.NotEmpty(t, auth.Data.SessionID)
		return auth.Data.SessionID
	}
	do := func(method, path, sessionID, body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Authorization", "Token "+sessionID)
		router.ServeHTTP(recorder, request)
		return recorder
	}
	profileStatus := func(sessionID string) int {
		return do(http.MethodGet, "/api/mine/settings/profile", sessionID, "").Code
	}

	t.Run("sign out", func(t *testing.T) {
		createUser(t, "sign-out@example.com")
		current := signIn(t, "sign-out@example.com", "old-password")
		other := signIn(t, "sign-out@example.com", "old-password")

		require.Equal(t, http.StatusOK, do(http.MethodPost, "/api/auth/sign-out", current, "").Code)
		require.Equal(t, http.StatusUnauthorized, profileStatus(current))
		require.Equal(t, http.StatusOK, profileStatus(other))

		require.Equal(t, http.StatusOK, do(http.MethodPost, "/api/auth/sign-out", current, "").Code)
	})

	t.Run("change password", func(t *testing.T) {
		createUser(t, "change-password@example.com")
		current := signIn(t, "change-password@example.com", "old-password")
		other := signIn(t, "change-password@example.com", "old-password")

		wrongPassword := do(http.MethodPut, "/api/mine/settings/profile", current, `{"name":"Session Owner","oldPassword":"wrong-password","newPassword":"new-password"}`)
		require.Equal(t, http.StatusBadRequest, wrongPassword.Code, wrongPassword.Body.String())
		require.Equal(t, http.StatusOK, profileStatus(other))

		changed := do(http.MethodPut, "/api/mine/settings/profile", current, `{"name":"Session Owner","oldPassword":"old-password","newPassword":"new-password"}`)
		require.Equal(t, http.StatusOK, changed.Code, changed.Body.String())
		require.Equal(t, http.StatusOK, profileStatus(current))
		require.Equal(t, http.StatusUnauthorized, profileStatus(other))
		require.Equal(t, http.StatusUnauthorized, profileStatus(other))

		renamed := do(http.MethodPut, "/api/mine/settings/profile", current, `{"name":"Renamed Owner"}`)
		require.Equal(t, http.StatusOK, renamed.Code, renamed.Body.String())
		require.Equal(t, http.StatusOK, profileStatus(current))
		require.Equal(t, http.StatusOK, profileStatus(signIn(t, "change-password@example.com", "new-password")))
	})

	t.Run("recover password", func(t *testing.T) {
		user := createUser(t, "recover-password@example.com")
		first := signIn(t, "recover-password@example.com", "old-password")
		second := signIn(t, "recover-password@example.com", "old-password")

		require.NoError(t, db.Users.UpdatePassword(context.Background(), user.ID, "new-password"))
		require.Equal(t, http.StatusUnauthorized, profileStatus(first))
		require.Equal(t, http.StatusUnauthorized, profileStatus(second))
		require.Equal(t, http.StatusUnauthorized, do(http.MethodPut, "/api/mine/settings/profile", first, `{"name":"Stale Token"}`).Code)
		require.Equal(t, http.StatusOK, profileStatus(signIn(t, "recover-password@example.com", "new-password")))
	})

	t.Run("stale session version", func(t *testing.T) {
		createUser(t, "legacy-session@example.com")
		sessionID := signIn(t, "legacy-session@example.com", "old-password")
		require.NoError(t, database.Model(&db.User{}).Where("email = ?", "legacy-session@example.com").Update("session_version", 1).Error)
		require.Equal(t, http.StatusUnauthorized, profileStatus(sessionID))
	})
}
