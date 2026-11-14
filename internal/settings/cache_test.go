// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package settings

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"

	"github.com/wuhan005/NekoBox/internal/db"
)

// memoryStore is an in-memory db.SystemSettingsStore that counts the reads reaching it.
type memoryStore struct {
	mu       sync.Mutex
	settings map[string]*db.SystemSetting
	gets     int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{settings: make(map[string]*db.SystemSetting)}
}

func (s *memoryStore) getCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gets
}

func (s *memoryStore) Get(_ context.Context, key string) (*db.SystemSetting, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gets++
	setting, ok := s.settings[key]
	if !ok {
		return nil, db.ErrSystemSettingNotExist
	}
	copied := *setting
	return &copied, nil
}

func (s *memoryStore) List(context.Context) ([]*db.SystemSetting, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings := make([]*db.SystemSetting, 0, len(s.settings))
	for _, setting := range s.settings {
		copied := *setting
		settings = append(settings, &copied)
	}
	return settings, nil
}

func (s *memoryStore) Set(_ context.Context, opts db.SetSystemSettingOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings[opts.Key] = &db.SystemSetting{Key: opts.Key, Type: opts.Type, Value: opts.Value, Description: opts.Description}
	return nil
}

func (s *memoryStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.settings[key]; !ok {
		return db.ErrSystemSettingNotExist
	}
	delete(s.settings, key)
	return nil
}

func (s *memoryStore) CreateIfNotExist(_ context.Context, opts []db.SetSystemSettingOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, opt := range opts {
		if _, ok := s.settings[opt.Key]; !ok {
			s.settings[opt.Key] = &db.SystemSetting{Key: opt.Key, Type: opt.Type, Value: opt.Value, Description: opt.Description}
		}
	}
	return nil
}

func newTestRedisClient(t *testing.T) *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// newTestKey returns a random key so that tests sharing a Redis instance do not interfere with each other.
func newTestKey(t *testing.T, client *redis.Client) string {
	key := "test." + strings.ToLower(randstr.Hex(8))
	t.Cleanup(func() { _ = client.Del(context.Background(), cacheKey(key)).Err() })
	return key
}

func TestCachedStore(t *testing.T) {
	client := newTestRedisClient(t)
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, client *redis.Client)
	}{
		{"Get", testCachedStoreGet},
		{"Set", testCachedStoreSet},
		{"Delete", testCachedStoreDelete},
		{"CreateIfNotExist", testCachedStoreCreateIfNotExist},
		{"AcrossInstances", testCachedStoreAcrossInstances},
		{"TTL", testCachedStoreTTL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, ctx, client)
		})
	}
}

func testCachedStoreGet(t *testing.T, ctx context.Context, client *redis.Client) {
	t.Run("hit", func(t *testing.T) {
		memory := newMemoryStore()
		store := newCachedStore(memory, client, time.Minute)
		key := newTestKey(t, client)
		require.Nil(t, memory.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeString, Value: "foo"}))

		for i := 0; i < 3; i++ {
			got, err := store.Get(ctx, key)
			require.Nil(t, err)
			require.Equal(t, "foo", got.Value)
		}
		require.Equal(t, 1, memory.getCount())
	})

	t.Run("not exist", func(t *testing.T) {
		memory := newMemoryStore()
		store := newCachedStore(memory, client, time.Minute)
		key := newTestKey(t, client)

		for i := 0; i < 3; i++ {
			_, err := store.Get(ctx, key)
			require.Equal(t, db.ErrSystemSettingNotExist, err)
		}
		require.Equal(t, 1, memory.getCount())
	})

	t.Run("corrupted cache entry", func(t *testing.T) {
		memory := newMemoryStore()
		store := newCachedStore(memory, client, time.Minute)
		key := newTestKey(t, client)
		require.Nil(t, memory.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeString, Value: "foo"}))
		require.Nil(t, client.Set(ctx, cacheKey(key), "{", time.Minute).Err())

		got, err := store.Get(ctx, key)
		require.Nil(t, err)
		require.Equal(t, "foo", got.Value)

		got, err = store.Get(ctx, key)
		require.Nil(t, err)
		require.Equal(t, "foo", got.Value)
		require.Equal(t, 1, memory.getCount())
	})

	t.Run("redis unavailable", func(t *testing.T) {
		memory := newMemoryStore()
		unavailable := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 100 * time.Millisecond})
		t.Cleanup(func() { _ = unavailable.Close() })
		store := newCachedStore(memory, unavailable, time.Minute)
		require.Nil(t, memory.Set(ctx, db.SetSystemSettingOptions{Key: "test.unavailable", Type: db.SystemSettingTypeBool, Value: "true"}))

		got, err := store.Get(ctx, "test.unavailable")
		require.Nil(t, err)
		require.Equal(t, "true", got.Value)

		require.Nil(t, store.Set(ctx, db.SetSystemSettingOptions{Key: "test.unavailable", Type: db.SystemSettingTypeBool, Value: "false"}))
		got, err = store.Get(ctx, "test.unavailable")
		require.Nil(t, err)
		require.Equal(t, "false", got.Value)
	})
}

func testCachedStoreSet(t *testing.T, ctx context.Context, client *redis.Client) {
	memory := newMemoryStore()
	store := newCachedStore(memory, client, time.Minute)
	key := newTestKey(t, client)

	_, err := store.Get(ctx, key)
	require.Equal(t, db.ErrSystemSettingNotExist, err)

	require.Nil(t, store.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeString, Value: "foo"}))
	got, err := store.Get(ctx, key)
	require.Nil(t, err)
	require.Equal(t, "foo", got.Value)

	require.Nil(t, store.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeString, Value: "bar"}))
	got, err = store.Get(ctx, key)
	require.Nil(t, err)
	require.Equal(t, "bar", got.Value)
}

func testCachedStoreDelete(t *testing.T, ctx context.Context, client *redis.Client) {
	memory := newMemoryStore()
	store := newCachedStore(memory, client, time.Minute)
	key := newTestKey(t, client)
	require.Nil(t, store.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeString, Value: "foo"}))

	_, err := store.Get(ctx, key)
	require.Nil(t, err)

	require.Nil(t, store.Delete(ctx, key))
	_, err = store.Get(ctx, key)
	require.Equal(t, db.ErrSystemSettingNotExist, err)

	err = store.Delete(ctx, key)
	require.True(t, errors.Is(err, db.ErrSystemSettingNotExist))
}

func testCachedStoreCreateIfNotExist(t *testing.T, ctx context.Context, client *redis.Client) {
	memory := newMemoryStore()
	store := newCachedStore(memory, client, time.Minute)
	key := newTestKey(t, client)

	_, err := store.Get(ctx, key)
	require.Equal(t, db.ErrSystemSettingNotExist, err)

	require.Nil(t, store.CreateIfNotExist(ctx, []db.SetSystemSettingOptions{{Key: key, Type: db.SystemSettingTypeInt, Value: "42"}}))
	got, err := store.Get(ctx, key)
	require.Nil(t, err)
	require.Equal(t, "42", got.Value)
}

func testCachedStoreAcrossInstances(t *testing.T, ctx context.Context, client *redis.Client) {
	memory := newMemoryStore()
	instance1 := newCachedStore(memory, client, time.Minute)
	instance2 := newCachedStore(memory, newTestRedisClient(t), time.Minute)
	key := newTestKey(t, client)
	require.Nil(t, instance1.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeBool, Value: "true"}))

	got, err := instance2.Get(ctx, key)
	require.Nil(t, err)
	require.Equal(t, "true", got.Value)

	require.Nil(t, instance1.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeBool, Value: "false"}))
	got, err = instance2.Get(ctx, key)
	require.Nil(t, err)
	require.Equal(t, "false", got.Value)
}

func testCachedStoreTTL(t *testing.T, ctx context.Context, client *redis.Client) {
	memory := newMemoryStore()
	store := newCachedStore(memory, client, time.Minute)
	key := newTestKey(t, client)
	require.Nil(t, memory.Set(ctx, db.SetSystemSettingOptions{Key: key, Type: db.SystemSettingTypeString, Value: "foo"}))

	_, err := store.Get(ctx, key)
	require.Nil(t, err)

	ttl, err := client.TTL(ctx, cacheKey(key)).Result()
	require.Nil(t, err)
	require.True(t, ttl > 0 && ttl <= time.Minute, "ttl: %v", ttl)
}
