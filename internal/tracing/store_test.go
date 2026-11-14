// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tracing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func setupRecorder(t *testing.T) (*tracetest.SpanRecorder, context.Context) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	ctx, parent := tp.Tracer("test").Start(context.Background(), "request")
	t.Cleanup(func() { parent.End() })
	return recorder, ctx
}

func spanNames(recorder *tracetest.SpanRecorder) []string {
	var names []string
	for _, s := range recorder.Ended() {
		names = append(names, s.Name())
	}
	return names
}

func TestCacheIniter(t *testing.T) {
	recorder, ctx := setupRecorder(t)
	c, err := CacheIniter(cache.MemoryIniter(), semconv.DBSystemRedis)(context.Background())
	require.NoError(t, err)

	require.NoError(t, c.Set(ctx, "key", "value", time.Minute))
	v, err := c.Get(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, "value", v)
	_, err = c.Get(ctx, "missing")
	assert.ErrorIs(t, err, os.ErrNotExist)
	require.NoError(t, c.Delete(ctx, "key"))

	spans := recorder.Ended()
	assert.Equal(t, []string{"cache.Set", "cache.Get", "cache.Get", "cache.Delete"}, spanNames(recorder))
	for _, s := range spans {
		assert.Contains(t, s.Attributes(), semconv.DBSystemRedis)
		assert.NotEqual(t, codes.Error, s.Status().Code, "a cache miss must not mark the span as failed")
	}
	assert.Contains(t, spans[1].Attributes(), semconv.DBOperationKey.String("cache.Get"))
}

func TestCacheIniter_NoParentSpan(t *testing.T) {
	recorder, _ := setupRecorder(t)
	c, err := CacheIniter(cache.MemoryIniter())(context.Background())
	require.NoError(t, err)

	require.NoError(t, c.Set(context.Background(), "key", "value", time.Minute))
	assert.Empty(t, recorder.Ended())
}

func TestSessionIniter(t *testing.T) {
	recorder, ctx := setupRecorder(t)
	store, err := SessionIniter(session.MemoryIniter(), semconv.DBSystemRedis)(context.Background())
	require.NoError(t, err)

	assert.False(t, store.Exist(ctx, "sid"))
	sess, err := store.Read(ctx, "sid")
	require.NoError(t, err)
	require.NoError(t, store.Save(ctx, sess))
	require.NoError(t, store.Destroy(ctx, "sid"))

	assert.Equal(t, []string{"session.Exist", "session.Read", "session.Save", "session.Destroy"}, spanNames(recorder))
}
