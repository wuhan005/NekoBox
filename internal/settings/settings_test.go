// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/dbutil"
)

// setStaticConf overrides the static configuration that the predefined settings are migrated from.
func setStaticConf(t *testing.T, icp string, enableTextCensor bool) {
	app, security, upload := conf.App, conf.Security, conf.Upload
	t.Cleanup(func() { conf.App, conf.Security, conf.Upload = app, security, upload })

	conf.App.ICP = icp
	conf.Security.EnableTextCensor = enableTextCensor
	conf.Upload.DefaultAvatarURL = "https://example.com/avatar.png"
	conf.Upload.DefaultBackground = "https://example.com/background.png"
}

func TestSettings(t *testing.T) {
	testDB, cleanup := dbutil.NewTestDB(t, db.AllTables...)
	ctx := context.Background()

	store := db.NewSystemSettingsStore(testDB)
	previousStore := db.SystemSettings
	db.SystemSettings = store
	t.Cleanup(func() { db.SystemSettings = previousStore })

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, store db.SystemSettingsStore)
	}{
		{"Init", testSettingsInit},
		{"Get", testSettingsGet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				if err := cleanup("system_settings"); err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, ctx, store)
		})
	}
}

func testSettingsInit(t *testing.T, ctx context.Context, store db.SystemSettingsStore) {
	setStaticConf(t, "Example ICP", true)

	err := initDefaults(ctx, store, defaults())
	require.Nil(t, err)

	got, err := store.List(ctx)
	require.Nil(t, err)
	require.Len(t, got, len(defaults()))

	values := make(map[string]string, len(got))
	for _, setting := range got {
		values[setting.Key] = setting.Value
	}
	require.Equal(t, map[string]string{
		KeyICP:               "Example ICP",
		KeyEnableTextCensor:  "true",
		KeyDefaultAvatar:     "https://example.com/avatar.png",
		KeyDefaultBackground: "https://example.com/background.png",
	}, values)

	t.Run("keep values changed at runtime", func(t *testing.T) {
		err := store.Set(ctx, db.SetSystemSettingOptions{
			Key:   KeyEnableTextCensor,
			Type:  db.SystemSettingTypeBool,
			Value: "false",
		})
		require.Nil(t, err)

		setStaticConf(t, "Changed ICP", true)
		err = initDefaults(ctx, store, defaults())
		require.Nil(t, err)

		icp, err := GetString(ctx, KeyICP)
		require.Nil(t, err)
		require.Equal(t, "Example ICP", icp)

		enableTextCensor, err := GetBool(ctx, KeyEnableTextCensor)
		require.Nil(t, err)
		require.False(t, enableTextCensor)
	})
}

func testSettingsGet(t *testing.T, ctx context.Context, store db.SystemSettingsStore) {
	setStaticConf(t, "Static ICP", true)

	t.Run("missing keys return an error", func(t *testing.T) {
		v, err := GetString(ctx, KeyICP)
		require.ErrorIs(t, err, db.ErrSystemSettingNotExist)
		require.Empty(t, v)

		enabled, err := GetBool(ctx, KeyEnableTextCensor)
		require.ErrorIs(t, err, db.ErrSystemSettingNotExist)
		require.False(t, enabled)
	})

	t.Run("read the database value", func(t *testing.T) {
		err := store.Set(ctx, db.SetSystemSettingOptions{
			Key:   KeyICP,
			Type:  db.SystemSettingTypeString,
			Value: "Dynamic ICP",
		})
		require.Nil(t, err)

		got, err := GetString(ctx, KeyICP)
		require.Nil(t, err)
		require.Equal(t, "Dynamic ICP", got)
	})
}
