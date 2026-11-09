// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/flamego/cache"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/wuhan005/NekoBox/internal/conf"
)

// tokenStore persists one-shot business tokens. When Redis is configured, tokens are stored as
// plain keys and consumed atomically via GETDEL; otherwise flamego memory cache plus a per-key
// mutex is used (tests and single-process dev).
type tokenStore struct {
	redis *redis.Client
	mu    sync.Map // key -> *sync.Mutex
}

func newTokenStore() *tokenStore {
	ts := &tokenStore{}
	if conf.Redis.Addr != "" {
		ts.redis = redis.NewClient(&redis.Options{
			Addr:     conf.Redis.Addr,
			Password: conf.Redis.Password,
			DB:       0,
		})
	}
	return ts
}

func (s *tokenStore) put(ctx context.Context, c cache.Cache, token string, ttl time.Duration) error {
	key := tokenCacheKeyPrefix + token
	if s.redis != nil {
		return s.redis.SetEX(ctx, key, "1", ttl).Err()
	}
	return c.Set(ctx, key, []byte("1"), ttl)
}

func (s *tokenStore) consume(ctx context.Context, c cache.Cache, token string) error {
	if token == "" {
		return ErrVerifyFailed
	}

	key := tokenCacheKeyPrefix + token
	if s.redis != nil {
		val, err := s.redis.GetDel(ctx, key).Result()
		if errors.Is(err, redis.Nil) {
			return ErrVerifyFailed
		}
		if err != nil {
			return errors.Wrap(ErrInternal, err.Error())
		}
		if val == "" {
			return ErrVerifyFailed
		}
		return nil
	}
	return s.consumeMemory(ctx, c, key)
}

func (s *tokenStore) consumeMemory(ctx context.Context, c cache.Cache, key string) error {
	muIface, _ := s.mu.LoadOrStore(key, &sync.Mutex{})
	mu := muIface.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	defer s.mu.Delete(key)

	if _, err := c.Get(ctx, key); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrVerifyFailed
		}
		return errors.Wrap(ErrInternal, err.Error())
	}
	if err := c.Delete(ctx, key); err != nil {
		return errors.Wrap(ErrInternal, err.Error())
	}
	return nil
}
