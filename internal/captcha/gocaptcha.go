// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/flamego/cache"
	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	gocaptchaImagesV2 "github.com/wenlng/go-captcha-assets/resources/imagesv2"
	gocaptchaTiles "github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"

	"github.com/wuhan005/NekoBox/internal/conf"
)

// Cache key prefixes; namespaced to avoid colliding with other application caches.
const (
	challengeCacheKeyPrefix = "captcha:go-captcha:challenge:"
	tokenCacheKeyPrefix     = "captcha:go-captcha:token:"
)

// goCaptchaService implements the slide captcha flow backed by go-captcha. The challenge answer
// and the issued one-shot token are persisted via cache.Cache.
type goCaptchaService struct {
	captcha       slide.Captcha
	tokens        *tokenStore
	verifyPadding int
	challengeTTL  time.Duration
	tokenTTL      time.Duration
}

// Defaults applied when conf has not been initialized yet, so the service does not silently use
// a zero TTL that would expire entries immediately.
const (
	defaultVerifyPadding = 5
	defaultChallengeTTL  = 5 * time.Minute
	defaultTokenTTL      = 5 * time.Minute
)

// NewGoCaptchaService loads the embedded assets and builds a slide captcha service. Asset
// loading errors are surfaced to the caller.
func NewGoCaptchaService() (Service, error) {
	bgImages, err := gocaptchaImagesV2.GetImages()
	if err != nil {
		return nil, errors.Wrap(err, "load go-captcha background images")
	}

	tiles, err := gocaptchaTiles.GetTiles()
	if err != nil {
		return nil, errors.Wrap(err, "load go-captcha tile images")
	}

	graphImages := make([]*slide.GraphImage, 0, len(tiles))
	for _, tile := range tiles {
		graphImages = append(graphImages, &slide.GraphImage{
			OverlayImage: tile.OverlayImage,
			MaskImage:    tile.MaskImage,
			ShadowImage:  tile.ShadowImage,
		})
	}

	builder := slide.NewBuilder()
	builder.SetResources(
		slide.WithBackgrounds(bgImages),
		slide.WithGraphImages(graphImages),
	)

	verifyPadding := conf.GoCaptcha.VerifyPadding
	if verifyPadding <= 0 {
		verifyPadding = defaultVerifyPadding
	}
	challengeTTL := time.Duration(conf.GoCaptcha.ChallengeTTL) * time.Second
	if challengeTTL <= 0 {
		challengeTTL = defaultChallengeTTL
	}
	tokenTTL := time.Duration(conf.GoCaptcha.TokenTTL) * time.Second
	if tokenTTL <= 0 {
		tokenTTL = defaultTokenTTL
	}

	return &goCaptchaService{
		captcha:       builder.Make(),
		tokens:        newTokenStore(),
		verifyPadding: verifyPadding,
		challengeTTL:  challengeTTL,
		tokenTTL:      tokenTTL,
	}, nil
}

func (*goCaptchaService) Type() Type { return TypeGoCaptcha }

// challengeCacheValue stores only the coordinates needed to validate the slide answer; we never cache the rendered images.
type challengeCacheValue struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Generate produces a fresh slide challenge: the answer coordinates are persisted in the cache and the encoded images are returned to the client.
func (s *goCaptchaService) Generate(ctx context.Context, c cache.Cache) (*ChallengeData, error) {
	captData, err := s.captcha.Generate()
	if err != nil {
		return nil, errors.Wrap(err, "generate slide captcha")
	}

	block := captData.GetData()
	if block == nil {
		return nil, errors.New("empty slide captcha block")
	}

	masterB64, err := captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, errors.Wrap(err, "encode master image")
	}
	tileB64, err := captData.GetTileImage().ToBase64()
	if err != nil {
		return nil, errors.Wrap(err, "encode tile image")
	}

	key := randstr.String(32)
	value, err := json.Marshal(challengeCacheValue{X: block.X, Y: block.Y})
	if err != nil {
		return nil, errors.Wrap(err, "marshal challenge")
	}
	if err := c.Set(ctx, challengeCacheKeyPrefix+key, value, s.challengeTTL); err != nil {
		return nil, errors.Wrap(err, "cache challenge")
	}

	return &ChallengeData{
		Key:         key,
		Image:       masterB64,
		Thumb:       tileB64,
		ThumbX:      block.DX,
		ThumbY:      block.DY,
		ThumbWidth:  block.Width,
		ThumbHeight: block.Height,
	}, nil
}

// VerifyChallenge checks the user-submitted slide answer; on success it issues and stores a one-shot business token.
func (s *goCaptchaService) VerifyChallenge(ctx context.Context, c cache.Cache, key string, x, y int) (string, error) {
	if key == "" {
		return "", ErrVerifyFailed
	}

	cacheKey := challengeCacheKeyPrefix + key
	raw, err := c.Get(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrVerifyFailed
		}
		return "", errors.Wrap(ErrInternal, err.Error())
	}

	rawBytes, ok := captchaCacheValueBytes(raw)
	if !ok {
		return "", ErrVerifyFailed
	}

	var v challengeCacheValue
	if err := json.Unmarshal(rawBytes, &v); err != nil {
		return "", errors.Wrap(ErrInternal, err.Error())
	}

	// The challenge is single-use: drop it before the bound check so a wrong answer can't be
	// brute-forced against the same cached coordinates.
	_ = c.Delete(ctx, cacheKey)

	if !slide.Validate(x, y, v.X, v.Y, s.verifyPadding) {
		return "", ErrVerifyFailed
	}

	token := randstr.String(64)
	if err := s.tokens.put(ctx, c, token, s.tokenTTL); err != nil {
		return "", errors.Wrap(ErrInternal, err.Error())
	}
	return token, nil
}

// Verify consumes a one-shot business token: missing or already-used tokens are rejected as ErrVerifyFailed.
func (s *goCaptchaService) Verify(ctx context.Context, c cache.Cache, token, _ string) error {
	return s.tokens.consume(ctx, c, token)
}

// captchaCacheValueBytes normalizes the cache value to []byte regardless of whether the backend returned []byte or string.
func captchaCacheValueBytes(v interface{}) ([]byte, bool) {
	switch t := v.(type) {
	case []byte:
		return t, true
	case string:
		return []byte(t), true
	default:
		return nil, false
	}
}
