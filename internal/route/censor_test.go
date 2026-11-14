package route

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/dbutil"
	"github.com/wuhan005/NekoBox/internal/settings"
)

func TestQuestionWritesWithCensorUnavailable(t *testing.T) {
	ctx := context.Background()
	database, _ := dbutil.NewTestDB(t, &db.User{}, &db.Question{}, &db.CensorLog{})
	users, questions, systemSettings, censorLogs := db.Users, db.Questions, db.SystemSettings, db.CensorLogs
	captcha, recaptcha, redis := conf.Captcha, conf.Recaptcha, conf.Redis
	t.Cleanup(func() {
		db.Users, db.Questions, db.SystemSettings, db.CensorLogs = users, questions, systemSettings, censorLogs
		conf.Captcha, conf.Recaptcha, conf.Redis = captcha, recaptcha, redis
	})
	db.Users = db.NewUsersStore(database)
	db.Questions = db.NewQuestionsStore(database)
	db.CensorLogs = db.NewCensorLogsStore(database)
	store := &memorySystemSettings{settings: map[string]*db.SystemSetting{
		settings.KeyEnableTextCensor: {Key: settings.KeyEnableTextCensor, Type: db.SystemSettingTypeBool, Value: "true"},
	}}
	db.SystemSettings = store

	verifier := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	t.Cleanup(verifier.Close)
	conf.Captcha.Type, conf.Recaptcha.VerifyURL, conf.Redis.Addr = "recaptcha", verifier.URL, ""
	owner := db.User{Domain: "test-owner", Email: "owner@example.com", Password: "test-password"}
	owner.EncodePassword()
	require.NoError(t, database.Create(&owner).Error)
	question, err := db.Questions.Create(ctx, db.CreateQuestionOptions{UserID: owner.ID, Content: "existing question"})
	require.NoError(t, err)

	router, err := New(database, nil)
	require.NoError(t, err)
	signIn := httptest.NewRecorder()
	router.ServeHTTP(signIn, httptest.NewRequest(http.MethodPost, "/api/auth/sign-in", strings.NewReader(
		`{"email":"owner@example.com","password":"test-password","captcha":"test"}`,
	)))
	require.Equal(t, http.StatusOK, signIn.Code, signIn.Body.String())
	var auth struct {
		Data struct{ SessionID string }
	}
	require.NoError(t, json.Unmarshal(signIn.Body.Bytes(), &auth))
	require.NotEmpty(t, auth.Data.SessionID)

	send := func(t *testing.T, method, path, token string, fields map[string]string) int {
		t.Helper()
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		for name, value := range fields {
			require.NoError(t, form.WriteField(name, value))
		}
		require.NoError(t, form.Close())
		req := httptest.NewRequest(method, path, &body)
		req.Header.Set("Content-Type", form.FormDataContentType())
		if token != "" {
			req.Header.Set("Authorization", "Token "+token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	answerPath := fmt.Sprintf("/api/mine/questions/%d/answer", question.ID)
	lastAnswer := ""
	for _, test := range []struct {
		text     string
		readErr  error
		response json.RawMessage
		status   int
	}{
		{text: "setting read fails", readErr: context.DeadlineExceeded, status: http.StatusOK},
		{text: "censor reply has no result", response: json.RawMessage(`{"data":[{"results":[]}]}`), status: http.StatusOK},
		{text: "censor rejects content", response: json.RawMessage(`{"data":[{"results":[{"suggestion":"block","label":"abuse"}]}]}`), status: http.StatusBadRequest},
	} {
		t.Run(test.text, func(t *testing.T) {
			store.readErr = test.readErr
			if test.response != nil {
				require.NoError(t, db.CensorLogs.Create(ctx, db.CreateCensorLogOptions{
					SourceName: "aliyun", Input: test.text, RawResponse: test.response,
				}))
			}
			require.Equal(t, test.status, send(t, http.MethodPost, "/api/users/test-owner/questions", "", map[string]string{"content": test.text, "captcha": "test"}))
			require.Equal(t, test.status, send(t, http.MethodPut, answerPath, auth.Data.SessionID, map[string]string{"answer": test.text}))

			var posted []db.Question
			require.NoError(t, database.Where("content = ?", test.text).Find(&posted).Error)
			if test.status == http.StatusOK {
				require.Len(t, posted, 1)
				require.Empty(t, posted[0].ContentCensorMetadata)
				lastAnswer = test.text
			} else {
				require.Empty(t, posted)
			}
			saved, err := db.Questions.GetByID(ctx, question.ID)
			require.NoError(t, err)
			require.Equal(t, lastAnswer, saved.Answer)
			require.Empty(t, saved.AnswerCensorMetadata)
		})
	}
}
