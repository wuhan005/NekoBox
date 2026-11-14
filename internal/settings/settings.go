// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package settings defines the keys and default values of the dynamic system settings stored in the database.
package settings

import (
	"context"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/go-redis/redis/v8"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
)

// Keys of the predefined system settings, named after the static configuration they are migrated from.
const (
	KeyICP               = "app.icp"
	KeyEnableTextCensor  = "security.enable_text_censor"
	KeyDefaultAvatar     = "upload.default_avatar"
	KeyDefaultBackground = "upload.default_background"
)

// defaults returns the predefined system settings, whose values are imported from the static configuration so that existing deployments keep their behavior.
func defaults() []db.SetSystemSettingOptions {
	return []db.SetSystemSettingOptions{
		{
			Key:         KeyICP,
			Type:        db.SystemSettingTypeString,
			Value:       conf.App.ICP,
			Description: "ICP license number shown in the site footer",
		},
		{
			Key:         KeyEnableTextCensor,
			Type:        db.SystemSettingTypeBool,
			Value:       strconv.FormatBool(conf.Security.EnableTextCensor),
			Description: "Whether to check user content with the text censor service",
		},
		{
			Key:         KeyDefaultAvatar,
			Type:        db.SystemSettingTypeString,
			Value:       conf.Upload.DefaultAvatarURL,
			Description: "Avatar URL of newly signed up users",
		},
		{
			Key:         KeyDefaultBackground,
			Type:        db.SystemSettingTypeString,
			Value:       conf.Upload.DefaultBackground,
			Description: "Background URL of newly signed up users",
		},
	}
}

// store is set by Init; it stays nil in CLI commands that skip Init, which then read the database directly.
var store db.SystemSettingsStore

// Store returns the system settings store; writes must go through it so that the cache stays consistent across instances.
func Store() db.SystemSettingsStore {
	if store != nil {
		return store
	}
	return db.SystemSettings
}

// Init enables the Redis cache unless client is nil, and creates the predefined system settings whose keys do not exist in the database yet.
// It must be called after conf.Init and db.Init.
func Init(ctx context.Context, client *redis.Client) error {
	store = db.SystemSettings
	if client != nil {
		store = newCachedStore(db.SystemSettings, client, cacheTTL)
	}
	return initDefaults(ctx, store, defaults())
}

func initDefaults(ctx context.Context, store db.SystemSettingsStore, opts []db.SetSystemSettingOptions) error {
	if err := store.CreateIfNotExist(ctx, opts); err != nil {
		return errors.Wrap(err, "create default system settings")
	}
	return nil
}

// GetString returns the value of the string setting with the given key.
func GetString(ctx context.Context, key string) (string, error) {
	setting, err := get(ctx, Store(), key)
	if err != nil {
		return "", err
	}
	return setting.AsString()
}

// GetBool returns the value of the bool setting with the given key.
func GetBool(ctx context.Context, key string) (bool, error) {
	setting, err := get(ctx, Store(), key)
	if err != nil {
		return false, err
	}
	return setting.AsBool()
}

// get falls back to the predefined default when the key is missing in the database, e.g. in CLI commands that skip Init.
func get(ctx context.Context, store db.SystemSettingsStore, key string) (*db.SystemSetting, error) {
	setting, err := store.Get(ctx, key)
	if err == nil {
		return setting, nil
	}
	if !errors.Is(err, db.ErrSystemSettingNotExist) {
		return nil, errors.Wrapf(err, "get system setting %q", key)
	}

	for _, opt := range defaults() {
		if opt.Key == key {
			return &db.SystemSetting{
				Key:         opt.Key,
				Type:        opt.Type,
				Value:       opt.Value,
				Description: opt.Description,
			}, nil
		}
	}
	return nil, errors.Wrapf(err, "%q", key)
}
