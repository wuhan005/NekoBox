// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package settings

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
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

		icp, err := get(ctx, store, KeyICP)
		require.Nil(t, err)
		require.Equal(t, "Example ICP", icp.Value)

		enableTextCensor, err := get(ctx, store, KeyEnableTextCensor)
		require.Nil(t, err)
		require.Equal(t, "false", enableTextCensor.Value)
	})
}

func testSettingsGet(t *testing.T, ctx context.Context, store db.SystemSettingsStore) {
	setStaticConf(t, "Static ICP", true)

	t.Run("fall back to default", func(t *testing.T) {
		got, err := get(ctx, store, KeyICP)
		require.Nil(t, err)
		v, err := got.AsString()
		require.Nil(t, err)
		require.Equal(t, "Static ICP", v)

		got, err = get(ctx, store, KeyEnableTextCensor)
		require.Nil(t, err)
		enabled, err := got.AsBool()
		require.Nil(t, err)
		require.True(t, enabled)
	})

	t.Run("database takes precedence", func(t *testing.T) {
		err := store.Set(ctx, db.SetSystemSettingOptions{
			Key:   KeyICP,
			Type:  db.SystemSettingTypeString,
			Value: "Dynamic ICP",
		})
		require.Nil(t, err)

		got, err := get(ctx, store, KeyICP)
		require.Nil(t, err)
		require.Equal(t, "Dynamic ICP", got.Value)
	})

	t.Run("unknown key", func(t *testing.T) {
		_, err := get(ctx, store, "app.not_exist")
		require.True(t, errors.Is(err, db.ErrSystemSettingNotExist))
	})
}
