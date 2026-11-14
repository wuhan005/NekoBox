// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func TestSystemSettings(t *testing.T) {
	t.Parallel()

	db, cleanup := newTestDB(t)
	ctx := context.Background()

	systemSettingsStore := NewSystemSettingsStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *systemSettings)
	}{
		{"Set", testSystemSettingsSet},
		{"Get", testSystemSettingsGet},
		{"List", testSystemSettingsList},
		{"Delete", testSystemSettingsDelete},
		{"CreateIfNotExist", testSystemSettingsCreateIfNotExist},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				if err := cleanup("system_settings"); err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, ctx, systemSettingsStore.(*systemSettings))
		})
	}
}

func testSystemSettingsSet(t *testing.T, ctx context.Context, db *systemSettings) {
	t.Run("create", func(t *testing.T) {
		err := db.Set(ctx, SetSystemSettingOptions{
			Key:         "upload.max_image_size",
			Type:        SystemSettingTypeInt,
			Value:       " 5242880 ",
			Description: "Max upload image size in bytes",
		})
		require.Nil(t, err)

		got, err := db.Get(ctx, "upload.max_image_size")
		require.Nil(t, err)
		require.Equal(t, SystemSettingTypeInt, got.Type)
		require.Equal(t, "5242880", got.Value)
		require.Equal(t, "Max upload image size in bytes", got.Description)
	})

	t.Run("update", func(t *testing.T) {
		before, err := db.Get(ctx, "upload.max_image_size")
		require.Nil(t, err)

		time.Sleep(10 * time.Millisecond)
		err = db.Set(ctx, SetSystemSettingOptions{
			Key:   "upload.max_image_size",
			Type:  SystemSettingTypeInt,
			Value: "1024",
		})
		require.Nil(t, err)

		got, err := db.Get(ctx, "upload.max_image_size")
		require.Nil(t, err)
		require.Equal(t, before.ID, got.ID)
		require.Equal(t, "1024", got.Value)
		require.Equal(t, "", got.Description)
		require.True(t, got.UpdatedAt.After(before.UpdatedAt))
	})

	t.Run("type mismatch", func(t *testing.T) {
		err := db.Set(ctx, SetSystemSettingOptions{
			Key:   "upload.max_image_size",
			Type:  SystemSettingTypeString,
			Value: "1024",
		})
		require.True(t, errors.Is(err, ErrSystemSettingTypeMismatch))
	})

	t.Run("invalid key", func(t *testing.T) {
		for _, key := range []string{"", "Upload.Size", "upload..size", ".upload", "upload size", strings.Repeat("a", maxSystemSettingKeyLength+1)} {
			err := db.Set(ctx, SetSystemSettingOptions{
				Key:   key,
				Type:  SystemSettingTypeString,
				Value: "foo",
			})
			require.True(t, errors.Is(err, ErrInvalidSystemSettingKey), "key %q", key)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		err := db.Set(ctx, SetSystemSettingOptions{
			Key:   "site.title",
			Type:  "yaml",
			Value: "foo",
		})
		require.True(t, errors.Is(err, ErrInvalidSystemSettingType))
	})

	t.Run("invalid value", func(t *testing.T) {
		err := db.Set(ctx, SetSystemSettingOptions{
			Key:   "site.register_enabled",
			Type:  SystemSettingTypeBool,
			Value: "yes",
		})
		require.True(t, errors.Is(err, ErrInvalidSystemSettingValue))
	})
}

func testSystemSettingsGet(t *testing.T, ctx context.Context, db *systemSettings) {
	for _, opts := range []SetSystemSettingOptions{
		{Key: "site.announcement", Type: SystemSettingTypeString, Value: "Hello, NekoBox!"},
		{Key: "upload.max_image_size", Type: SystemSettingTypeInt, Value: "5242880"},
		{Key: "censor.threshold", Type: SystemSettingTypeFloat, Value: "0.85"},
		{Key: "site.register_enabled", Type: SystemSettingTypeBool, Value: "TRUE"},
		{Key: "site.links", Type: SystemSettingTypeJSON, Value: `{ "github": "https://github.com/wuhan005/NekoBox" }`},
	} {
		err := db.Set(ctx, opts)
		require.Nil(t, err)
	}

	t.Run("string", func(t *testing.T) {
		got, err := db.Get(ctx, "site.announcement")
		require.Nil(t, err)
		v, err := got.AsString()
		require.Nil(t, err)
		require.Equal(t, "Hello, NekoBox!", v)
	})

	t.Run("int", func(t *testing.T) {
		got, err := db.Get(ctx, "upload.max_image_size")
		require.Nil(t, err)
		v, err := got.AsInt()
		require.Nil(t, err)
		require.Equal(t, int64(5242880), v)
	})

	t.Run("float", func(t *testing.T) {
		got, err := db.Get(ctx, "censor.threshold")
		require.Nil(t, err)
		v, err := got.AsFloat()
		require.Nil(t, err)
		require.Equal(t, 0.85, v)
	})

	t.Run("bool", func(t *testing.T) {
		got, err := db.Get(ctx, "site.register_enabled")
		require.Nil(t, err)
		require.Equal(t, "true", got.Value)
		v, err := got.AsBool()
		require.Nil(t, err)
		require.True(t, v)
	})

	t.Run("json", func(t *testing.T) {
		got, err := db.Get(ctx, "site.links")
		require.Nil(t, err)
		require.Equal(t, `{"github":"https://github.com/wuhan005/NekoBox"}`, got.Value)

		var v map[string]string
		err = got.AsJSON(&v)
		require.Nil(t, err)
		require.Equal(t, map[string]string{"github": "https://github.com/wuhan005/NekoBox"}, v)
	})

	t.Run("type mismatch", func(t *testing.T) {
		got, err := db.Get(ctx, "upload.max_image_size")
		require.Nil(t, err)
		_, err = got.AsBool()
		require.True(t, errors.Is(err, ErrSystemSettingTypeMismatch))
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.Get(ctx, "site.not_exist")
		require.Equal(t, ErrSystemSettingNotExist, err)
	})
}

func testSystemSettingsList(t *testing.T, ctx context.Context, db *systemSettings) {
	t.Run("empty", func(t *testing.T) {
		got, err := db.List(ctx)
		require.Nil(t, err)
		require.Empty(t, got)
	})

	t.Run("normal", func(t *testing.T) {
		for _, key := range []string{"site.title", "censor.enabled", "upload.max_image_size"} {
			err := db.Set(ctx, SetSystemSettingOptions{
				Key:   key,
				Type:  SystemSettingTypeString,
				Value: key,
			})
			require.Nil(t, err)
		}

		got, err := db.List(ctx)
		require.Nil(t, err)

		keys := make([]string, 0, len(got))
		for _, setting := range got {
			keys = append(keys, setting.Key)
		}
		require.Equal(t, []string{"censor.enabled", "site.title", "upload.max_image_size"}, keys)
	})
}

func testSystemSettingsDelete(t *testing.T, ctx context.Context, db *systemSettings) {
	err := db.Set(ctx, SetSystemSettingOptions{
		Key:   "site.announcement",
		Type:  SystemSettingTypeString,
		Value: "Hello, NekoBox!",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		err := db.Delete(ctx, "site.announcement")
		require.Nil(t, err)

		_, err = db.Get(ctx, "site.announcement")
		require.Equal(t, ErrSystemSettingNotExist, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := db.Delete(ctx, "site.announcement")
		require.Equal(t, ErrSystemSettingNotExist, err)
	})

	t.Run("recreate with another type", func(t *testing.T) {
		err := db.Set(ctx, SetSystemSettingOptions{
			Key:   "site.announcement",
			Type:  SystemSettingTypeJSON,
			Value: `{"text":"Hello, NekoBox!"}`,
		})
		require.Nil(t, err)
	})
}

func testSystemSettingsCreateIfNotExist(t *testing.T, ctx context.Context, db *systemSettings) {
	err := db.Set(ctx, SetSystemSettingOptions{
		Key:   "site.register_enabled",
		Type:  SystemSettingTypeBool,
		Value: "false",
	})
	require.Nil(t, err)

	defaults := []SetSystemSettingOptions{
		{Key: "site.register_enabled", Type: SystemSettingTypeBool, Value: "true", Description: "Whether new users can sign up"},
		{Key: "upload.max_image_size", Type: SystemSettingTypeInt, Value: "5242880", Description: "Max upload image size in bytes"},
	}

	t.Run("normal", func(t *testing.T) {
		err := db.CreateIfNotExist(ctx, defaults)
		require.Nil(t, err)

		got, err := db.Get(ctx, "site.register_enabled")
		require.Nil(t, err)
		require.Equal(t, "false", got.Value)
		require.Equal(t, "", got.Description)

		got, err = db.Get(ctx, "upload.max_image_size")
		require.Nil(t, err)
		require.Equal(t, "5242880", got.Value)
		require.Equal(t, "Max upload image size in bytes", got.Description)
	})

	t.Run("idempotent", func(t *testing.T) {
		err := db.CreateIfNotExist(ctx, defaults)
		require.Nil(t, err)

		got, err := db.List(ctx)
		require.Nil(t, err)
		require.Len(t, got, 2)
	})

	t.Run("empty", func(t *testing.T) {
		err := db.CreateIfNotExist(ctx, nil)
		require.Nil(t, err)
	})

	t.Run("duplicate key", func(t *testing.T) {
		err := db.CreateIfNotExist(ctx, []SetSystemSettingOptions{
			{Key: "site.title", Type: SystemSettingTypeString, Value: "NekoBox"},
			{Key: "site.title", Type: SystemSettingTypeString, Value: "NekoBox"},
		})
		require.True(t, errors.Is(err, ErrInvalidSystemSettingKey))
	})

	t.Run("invalid value", func(t *testing.T) {
		err := db.CreateIfNotExist(ctx, []SetSystemSettingOptions{
			{Key: "site.title", Type: SystemSettingTypeString, Value: "NekoBox"},
			{Key: "site.links", Type: SystemSettingTypeJSON, Value: "{"},
		})
		require.True(t, errors.Is(err, ErrInvalidSystemSettingValue))

		_, err = db.Get(ctx, "site.title")
		require.Equal(t, ErrSystemSettingNotExist, err)
	})
}

func TestSystemSettingTypeNormalize(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		typ       SystemSettingType
		value     string
		want      string
		wantError error
	}{
		{name: "string keeps spaces", typ: SystemSettingTypeString, value: " foo ", want: " foo "},
		{name: "string empty", typ: SystemSettingTypeString, value: "", want: ""},
		{name: "int", typ: SystemSettingTypeInt, value: " +042 ", want: "42"},
		{name: "int negative", typ: SystemSettingTypeInt, value: "-1", want: "-1"},
		{name: "int invalid", typ: SystemSettingTypeInt, value: "1.5", wantError: ErrInvalidSystemSettingValue},
		{name: "int overflow", typ: SystemSettingTypeInt, value: "9223372036854775808", wantError: ErrInvalidSystemSettingValue},
		{name: "float", typ: SystemSettingTypeFloat, value: "1.50", want: "1.5"},
		{name: "float exponent", typ: SystemSettingTypeFloat, value: "1e3", want: "1000"},
		{name: "float NaN", typ: SystemSettingTypeFloat, value: "NaN", wantError: ErrInvalidSystemSettingValue},
		{name: "float Inf", typ: SystemSettingTypeFloat, value: "+Inf", wantError: ErrInvalidSystemSettingValue},
		{name: "bool", typ: SystemSettingTypeBool, value: "1", want: "true"},
		{name: "bool false", typ: SystemSettingTypeBool, value: "F", want: "false"},
		{name: "bool invalid", typ: SystemSettingTypeBool, value: "yes", wantError: ErrInvalidSystemSettingValue},
		{name: "json object", typ: SystemSettingTypeJSON, value: "{\n  \"a\": [1, 2]\n}", want: `{"a":[1,2]}`},
		{name: "json scalar", typ: SystemSettingTypeJSON, value: `"foo"`, want: `"foo"`},
		{name: "json empty", typ: SystemSettingTypeJSON, value: "", wantError: ErrInvalidSystemSettingValue},
		{name: "json invalid", typ: SystemSettingTypeJSON, value: `{"a":}`, wantError: ErrInvalidSystemSettingValue},
		{name: "unknown type", typ: "yaml", value: "foo", wantError: ErrInvalidSystemSettingType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.typ.Normalize(tc.value)
			if tc.wantError != nil {
				require.True(t, errors.Is(err, tc.wantError), "got error: %v", err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
