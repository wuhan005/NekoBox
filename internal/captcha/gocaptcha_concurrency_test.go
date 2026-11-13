package captcha

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/flamego/cache"
	cacheRedis "github.com/flamego/cache/redis"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

type delayedChallengeCache struct{ cache.Cache }

func (c delayedChallengeCache) Get(ctx context.Context, key string) (interface{}, error) {
	value, err := c.Cache.Get(ctx, key)
	if err == nil && strings.HasPrefix(key, challengeCacheKeyPrefix) {
		// Keep successful reads in flight so concurrent verifications overlap.
		time.Sleep(5 * time.Millisecond)
	}
	return value, err
}

func TestGoCaptcha_VerifyChallenge_Concurrent(t *testing.T) {
	for _, backend := range []string{"memory", "redis"} {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			c := newMemoryCache(t)
			services := []*goCaptchaService{{tokens: &tokenStore{}, verifyPadding: 5, tokenTTL: time.Minute}}
			if backend == "redis" {
				addr := os.Getenv("NEKOBOX_TEST_REDIS_ADDR")
				if addr == "" {
					t.Skip("NEKOBOX_TEST_REDIS_ADDR is not set")
				}
				var err error
				c, err = cacheRedis.Initer()(ctx, cacheRedis.Config{Options: &cacheRedis.Options{Addr: addr}})
				require.NoError(t, err)
				services = nil
				for i := 0; i < 2; i++ {
					client := redis.NewClient(&redis.Options{Addr: addr})
					t.Cleanup(func() { _ = client.Close() })
					services = append(services, &goCaptchaService{
						tokens: &tokenStore{redis: client}, verifyPadding: 5, tokenTTL: time.Minute,
					})
				}
			}

			key := randstr.String(32)
			require.NoError(t, c.Set(ctx, challengeCacheKeyPrefix+key, []byte(`{"x":100,"y":100}`), time.Minute))
			t.Cleanup(func() { _ = c.Delete(ctx, challengeCacheKeyPrefix+key) })
			c = delayedChallengeCache{c}

			type result struct {
				token string
				err   error
			}
			const workers = 8
			start := make(chan struct{})
			results := make(chan result, workers)
			for i := 0; i < workers; i++ {
				svc := services[i%len(services)]
				go func() {
					<-start
					token, err := svc.VerifyChallenge(ctx, c, key, 100, 100)
					results <- result{token, err}
				}()
			}
			close(start)

			var tokens []string
			for i := 0; i < workers; i++ {
				r := <-results
				if r.err == nil {
					tokens = append(tokens, r.token)
				} else {
					require.ErrorIs(t, r.err, ErrVerifyFailed)
					require.Empty(t, r.token)
				}
			}
			require.Len(t, tokens, 1)
			require.NotEmpty(t, tokens[0])
			require.NoError(t, services[0].Verify(ctx, c, tokens[0], ""))
			require.ErrorIs(t, services[0].Verify(ctx, c, tokens[0], ""), ErrVerifyFailed)
		})
	}
}

type deleteFailureCache struct{ cache.Cache }

func (deleteFailureCache) Delete(context.Context, string) error {
	return errors.New("cache delete failed")
}

func TestGoCaptcha_VerifyChallenge_DeleteFailure(t *testing.T) {
	ctx := context.Background()
	c := deleteFailureCache{newMemoryCache(t)}
	require.NoError(t, c.Set(ctx, challengeCacheKeyPrefix+"delete-failure", []byte(`{"x":100,"y":100}`), time.Minute))
	svc := &goCaptchaService{tokens: &tokenStore{}, verifyPadding: 5, tokenTTL: time.Minute}

	token, err := svc.VerifyChallenge(ctx, c, "delete-failure", 100, 100)
	require.ErrorIs(t, err, ErrInternal)
	require.Empty(t, token)
}
