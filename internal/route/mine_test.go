package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/dbutil"
	"github.com/wuhan005/NekoBox/internal/response"
)

func TestExportAccountData(t *testing.T) {
	database, _ := dbutil.NewTestDB(t, &db.User{}, &db.Question{})
	users, questions := db.Users, db.Questions
	captchaConfig, recaptchaConfig, redisConfig := conf.Captcha, conf.Recaptcha, conf.Redis
	t.Cleanup(func() {
		db.Users, db.Questions = users, questions
		conf.Captcha, conf.Recaptcha, conf.Redis = captchaConfig, recaptchaConfig, redisConfig
	})
	db.Users = db.NewUsersStore(database)
	db.Questions = db.NewQuestionsStore(database)
	verifier := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	t.Cleanup(verifier.Close)
	conf.Captcha.Type, conf.Recaptcha.VerifyURL, conf.Redis.Addr = "recaptcha", verifier.URL, ""

	user := db.User{Domain: "export-user", Name: "Export Owner", Email: "export@example.com", Password: "test-password", Intro: "Ask away"}
	user.EncodePassword()
	require.NoError(t, database.Create(&user).Error)
	require.NoError(t, database.Create(&[]db.Question{
		{UserID: user.ID, Content: "Public question", Answer: "Public answer"},
		{UserID: user.ID, Content: "Private question", Answer: "Private answer", IsPrivate: true},
		{UserID: user.ID, Content: "Unanswered question"},
		{UserID: user.ID + 1, Content: "Another user's question", Answer: "Another user's answer"},
	}).Error)

	router, err := New(database, nil)
	require.NoError(t, err)
	const exportPath = "/api/mine/settings/export-data"
	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, exportPath, nil))
	require.Equal(t, http.StatusUnauthorized, unauthorized.Code)

	signIn := httptest.NewRecorder()
	router.ServeHTTP(signIn, httptest.NewRequest(http.MethodPost, "/api/auth/sign-in", strings.NewReader(
		`{"email":"export@example.com","password":"test-password","captcha":"test"}`,
	)))
	require.Equal(t, http.StatusOK, signIn.Code, signIn.Body.String())
	var auth struct{ Data response.SignIn }
	require.NoError(t, json.Unmarshal(signIn.Body.Bytes(), &auth))
	require.NotEmpty(t, auth.Data.SessionID)

	export := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, exportPath, nil)
	request.Header.Set("Authorization", "Token "+auth.Data.SessionID)
	router.ServeHTTP(export, request)
	require.Equal(t, http.StatusOK, export.Code)
	require.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", export.Header().Get("Content-Type"))
	require.Regexp(t, `^attachment; filename\*=UTF-8''.*-export-user-\d{14}\.xlsx$`, export.Header().Get("Content-Disposition"))

	workbook, err := excelize.OpenReader(bytes.NewReader(export.Body.Bytes()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = workbook.Close() })
	require.ElementsMatch(t, []string{"账号信息", "提问"}, workbook.GetSheetList())
	for cell, expected := range map[string]string{"B2": user.Email, "B3": user.Name, "B4": user.Domain, "B5": user.Intro} {
		value, err := workbook.GetCellValue("账号信息", cell)
		require.NoError(t, err)
		require.Equal(t, expected, value, cell)
	}
	rows, err := workbook.GetRows("提问")
	require.NoError(t, err)
	require.Len(t, rows, 4)
	require.Equal(t, []string{"提问时间", "问题", "回答"}, rows[0])
	entries := make(map[string]string, len(rows)-1)
	for _, row := range rows[1:] {
		require.GreaterOrEqual(t, len(row), 2)
		entries[row[1]] = ""
		if len(row) == 3 {
			entries[row[1]] = row[2]
		}
	}
	require.Equal(t, map[string]string{
		"Public question": "Public answer", "Private question": "Private answer", "Unanswered question": "",
	}, entries)
}
