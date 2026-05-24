// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/flamego/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMemoryCache(t *testing.T) cache.Cache {
	t.Helper()
	c, err := cache.MemoryIniter()(context.Background())
	require.NoError(t, err)
	return c
}

func TestNew_Defaults(t *testing.T) {
	svc, err := New("")
	require.NoError(t, err)
	assert.Equal(t, TypeRecaptcha, svc.Type())

	svc, err = New(TypeRecaptcha)
	require.NoError(t, err)
	assert.Equal(t, TypeRecaptcha, svc.Type())

	_, err = New("unknown")
	assert.Error(t, err)
}

func TestRecaptcha_UnsupportedSlide(t *testing.T) {
	svc := NewRecaptchaService()

	_, err := svc.Generate(context.Background(), newMemoryCache(t))
	assert.ErrorIs(t, err, ErrUnsupported)

	_, err = svc.VerifyChallenge(context.Background(), newMemoryCache(t), "k", 0, 0)
	assert.ErrorIs(t, err, ErrUnsupported)
}

func TestGoCaptcha_VerifyEmptyOrUnknownToken(t *testing.T) {
	svc, err := NewGoCaptchaService()
	require.NoError(t, err)
	assert.Equal(t, TypeGoCaptcha, svc.Type())

	c := newMemoryCache(t)

	assert.ErrorIs(t, svc.Verify(context.Background(), c, "", ""), ErrVerifyFailed)
	assert.ErrorIs(t, svc.Verify(context.Background(), c, "not-exists", ""), ErrVerifyFailed)
}

func TestGoCaptcha_Generate(t *testing.T) {
	svc, err := NewGoCaptchaService()
	require.NoError(t, err)

	c := newMemoryCache(t)

	data, err := svc.Generate(context.Background(), c)
	require.NoError(t, err)
	assert.NotEmpty(t, data.Key)
	assert.NotEmpty(t, data.Image)
	assert.NotEmpty(t, data.Thumb)
	assert.Greater(t, data.ThumbWidth, 0)
	assert.Greater(t, data.ThumbHeight, 0)
}

func TestGoCaptcha_VerifyChallenge_WrongAnswer(t *testing.T) {
	svc, err := NewGoCaptchaService()
	require.NoError(t, err)

	c := newMemoryCache(t)

	data, err := svc.Generate(context.Background(), c)
	require.NoError(t, err)

	// Pass coordinates that cannot possibly satisfy the slide validator.
	_, err = svc.VerifyChallenge(context.Background(), c, data.Key, -10000, -10000)
	assert.ErrorIs(t, err, ErrVerifyFailed)

	// The same key has been consumed; a second submission must also fail.
	_, err = svc.VerifyChallenge(context.Background(), c, data.Key, -10000, -10000)
	assert.ErrorIs(t, err, ErrVerifyFailed)
}

func TestGoCaptcha_VerifyChallenge_OneShotToken(t *testing.T) {
	svc, err := NewGoCaptchaService()
	require.NoError(t, err)

	c := newMemoryCache(t)

	data, err := svc.Generate(context.Background(), c)
	require.NoError(t, err)

	// Read the target coordinates straight from the cache to simulate a perfect slide.
	raw, err := c.Get(context.Background(), challengeCacheKeyPrefix+data.Key)
	require.NoError(t, err)
	bytes, ok := captchaCacheValueBytes(raw)
	require.True(t, ok)
	var v challengeCacheValue
	require.NoError(t, json.Unmarshal(bytes, &v))

	token, err := svc.VerifyChallenge(context.Background(), c, data.Key, v.X, v.Y)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// The first Verify should succeed; the second one (already consumed) must fail.
	require.NoError(t, svc.Verify(context.Background(), c, token, ""))
	err = svc.Verify(context.Background(), c, token, "")
	assert.ErrorIs(t, err, ErrVerifyFailed)
}

func TestGoCaptcha_VerifyToken_ConcurrentConsume(t *testing.T) {
	svc, err := NewGoCaptchaService()
	require.NoError(t, err)

	c := newMemoryCache(t)

	data, err := svc.Generate(context.Background(), c)
	require.NoError(t, err)

	raw, err := c.Get(context.Background(), challengeCacheKeyPrefix+data.Key)
	require.NoError(t, err)
	bytes, ok := captchaCacheValueBytes(raw)
	require.True(t, ok)
	var v challengeCacheValue
	require.NoError(t, json.Unmarshal(bytes, &v))

	token, err := svc.VerifyChallenge(context.Background(), c, data.Key, v.X, v.Y)
	require.NoError(t, err)

	const workers = 8
	var wg sync.WaitGroup
	successes := make(chan struct{}, workers)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if err := svc.Verify(context.Background(), c, token, ""); err == nil {
				successes <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(successes)

	count := 0
	for range successes {
		count++
	}
	assert.Equal(t, 1, count)
}

func TestCheckChallengeRateLimit(t *testing.T) {
	c := newMemoryCache(t)
	ctx := context.Background()
	ip := "127.0.0.1"

	for i := 0; i < challengeRateLimitMax; i++ {
		require.NoError(t, CheckChallengeRateLimit(ctx, c, ip))
	}
	assert.ErrorIs(t, CheckChallengeRateLimit(ctx, c, ip), ErrRateLimited)
}
