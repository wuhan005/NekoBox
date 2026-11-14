// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package settings

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/NekoBox/internal/db"
)

const (
	cacheKeyPrefix = "settings:system-setting:"
	// cacheTTL bounds how long a stale entry can survive a read racing with a write, or a direct database edit.
	cacheTTL = time.Minute
)

var _ db.SystemSettingsStore = (*cachedStore)(nil)

// cachedStore is a read-through Redis cache in front of the database store, shared by all instances.
// Writes go to the database first and then invalidate the cached keys; Redis errors never fail a read.
type cachedStore struct {
	store  db.SystemSettingsStore
	client *redis.Client
	ttl    time.Duration
}

func newCachedStore(store db.SystemSettingsStore, client *redis.Client, ttl time.Duration) *cachedStore {
	return &cachedStore{
		store:  store,
		client: client,
		ttl:    ttl,
	}
}

// cacheEntry caches a missing key as a nil setting, so that falling back to the default value does not hit the database either.
type cacheEntry struct {
	Setting *db.SystemSetting `json:"setting"`
}

func cacheKey(key string) string {
	return cacheKeyPrefix + key
}

func (s *cachedStore) Get(ctx context.Context, key string) (*db.SystemSetting, error) {
	raw, err := s.client.Get(ctx, cacheKey(key)).Bytes()
	if err == nil {
		var entry cacheEntry
		if decodeErr := json.Unmarshal(raw, &entry); decodeErr != nil {
			logrus.WithContext(ctx).WithError(decodeErr).WithField("key", key).Warn("Failed to decode cached system setting")
		} else if entry.Setting == nil {
			return nil, db.ErrSystemSettingNotExist
		} else {
			return entry.Setting, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		logrus.WithContext(ctx).WithError(err).WithField("key", key).Warn("Failed to get system setting from cache")
		return s.store.Get(ctx, key)
	}

	setting, err := s.store.Get(ctx, key)
	if err != nil && !errors.Is(err, db.ErrSystemSettingNotExist) {
		return nil, err
	}

	if raw, marshalErr := json.Marshal(cacheEntry{Setting: setting}); marshalErr != nil {
		logrus.WithContext(ctx).WithError(marshalErr).WithField("key", key).Warn("Failed to encode system setting for cache")
	} else if setErr := s.client.Set(ctx, cacheKey(key), raw, s.ttl).Err(); setErr != nil {
		logrus.WithContext(ctx).WithError(setErr).WithField("key", key).Warn("Failed to set system setting to cache")
	}
	return setting, err
}

func (s *cachedStore) List(ctx context.Context) ([]*db.SystemSetting, error) {
	return s.store.List(ctx)
}

func (s *cachedStore) Set(ctx context.Context, opts db.SetSystemSettingOptions) error {
	err := s.store.Set(ctx, opts)
	s.invalidate(ctx, opts.Key)
	return err
}

func (s *cachedStore) Delete(ctx context.Context, key string) error {
	err := s.store.Delete(ctx, key)
	s.invalidate(ctx, key)
	return err
}

func (s *cachedStore) CreateIfNotExist(ctx context.Context, opts []db.SetSystemSettingOptions) error {
	err := s.store.CreateIfNotExist(ctx, opts)

	keys := make([]string, 0, len(opts))
	for _, opt := range opts {
		keys = append(keys, opt.Key)
	}
	s.invalidate(ctx, keys...)
	return err
}

// invalidate runs even if the database write failed, since a failed write may still have been committed.
func (s *cachedStore) invalidate(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}

	cacheKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		cacheKeys = append(cacheKeys, cacheKey(key))
	}
	if err := s.client.Del(ctx, cacheKeys...).Err(); err != nil {
		logrus.WithContext(ctx).WithError(err).WithField("keys", keys).Warn("Failed to invalidate system settings cache")
	}
}
