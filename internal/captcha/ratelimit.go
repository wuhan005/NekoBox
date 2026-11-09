// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"context"
	"strconv"
	"time"

	"github.com/flamego/cache"
	"github.com/pkg/errors"
)

const (
	challengeRateLimitKeyPrefix = "captcha:challenge:rate:"
	challengeRateLimitMax       = 30
	challengeRateLimitWindow    = time.Minute
)

// CheckChallengeRateLimit rejects clients that exceed the per-IP challenge generation quota.
func CheckChallengeRateLimit(ctx context.Context, c cache.Cache, ip string) error {
	if ip == "" {
		ip = "unknown"
	}

	key := challengeRateLimitKeyPrefix + ip
	count := 0
	if raw, err := c.Get(ctx, key); err == nil {
		count = rateLimitCount(raw)
	}
	if count >= challengeRateLimitMax {
		return ErrRateLimited
	}

	count++
	if err := c.Set(ctx, key, strconv.Itoa(count), challengeRateLimitWindow); err != nil {
		return errors.Wrap(ErrInternal, err.Error())
	}
	return nil
}

func rateLimitCount(raw interface{}) int {
	switch v := raw.(type) {
	case string:
		n, _ := strconv.Atoi(v)
		return n
	case []byte:
		n, _ := strconv.Atoi(string(v))
		return n
	case int:
		return v
	default:
		return 0
	}
}
