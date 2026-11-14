package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/dbutil"
	"github.com/wuhan005/NekoBox/internal/security/censor"
	"github.com/wuhan005/NekoBox/internal/settings"
)

func TestCensorQuestionsKeepsInitialSetting(t *testing.T) {
	ctx := context.Background()
	database, _ := dbutil.NewTestDB(t, &db.Question{}, &db.CensorLog{}, &db.SystemSetting{})
	systemSettings, censorLogs := db.SystemSettings, db.CensorLogs
	t.Cleanup(func() { db.SystemSettings, db.CensorLogs = systemSettings, censorLogs })
	db.SystemSettings = db.NewSystemSettingsStore(database)
	db.CensorLogs = db.NewCensorLogsStore(database)

	switchSetting := db.SetSystemSettingOptions{Key: settings.KeyEnableTextCensor, Type: db.SystemSettingTypeBool, Value: "true"}
	require.NoError(t, db.SystemSettings.Set(ctx, switchSetting))
	for _, content := range []string{"first question", "second question"} {
		require.NoError(t, database.Create(&db.Question{Content: content, Answer: "test answer"}).Error)
		for _, text := range []string{content, "test answer"} {
			require.NoError(t, db.CensorLogs.Create(ctx, db.CreateCensorLogOptions{
				SourceName:  "aliyun",
				Input:       text,
				RawResponse: json.RawMessage(`{"data":[{"results":[{"suggestion":"block","label":"abuse"}]}]}`),
			}))
		}
	}

	switchSetting.Value = "false"
	disabled := false
	require.NoError(t, database.Callback().Raw().After("gorm:raw").Register("test:disable_censor", func(tx *gorm.DB) {
		if !disabled && tx.RowsAffected == 1 {
			disabled = true
			require.NoError(t, db.SystemSettings.Set(ctx, switchSetting))
		}
	}))
	require.NoError(t, censorQuestions(ctx, database))
	enabled, err := settings.GetBool(ctx, settings.KeyEnableTextCensor)
	require.NoError(t, err)
	require.False(t, enabled)

	var questions []db.Question
	require.NoError(t, database.Find(&questions).Error)
	require.Len(t, questions, 2)
	for _, question := range questions {
		for _, metadata := range [][]byte{question.ContentCensorMetadata, question.AnswerCensorMetadata} {
			var result censor.TextCensorResponse
			require.NoError(t, json.Unmarshal(metadata, &result))
			require.Equal(t, "aliyun", result.SourceName)
			require.False(t, result.Pass)
		}
	}
}
