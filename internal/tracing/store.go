// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tracing

import (
	"context"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/flamego/cache"
	"github.com/flamego/session"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// storeTracer creates spans for the flamego cache and session stores, which own their clients and thus cannot be instrumented with driver hooks.
type storeTracer struct {
	tracer oteltrace.Tracer
	attrs  []attribute.KeyValue
}

func newStoreTracer(attrs []attribute.KeyValue) storeTracer {
	return storeTracer{
		tracer: otel.GetTracerProvider().Tracer(instrumentationName, oteltrace.WithInstrumentationVersion("1.0.0")),
		attrs:  attrs,
	}
}

// start only creates a span under a recording parent so that background jobs do not produce orphan traces.
// Keys are never recorded, since session IDs are credentials.
func (t storeTracer) start(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	if !oteltrace.SpanFromContext(ctx).IsRecording() {
		return ctx, oteltrace.SpanFromContext(context.Background())
	}
	return t.tracer.Start(ctx, name,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(t.attrs...),
		oteltrace.WithAttributes(semconv.DBOperationKey.String(name)),
	)
}

func end(span oteltrace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// CacheIniter wraps the flamego cache initer to trace every cache operation, with attrs describing the backend, e.g. semconv.DBSystemRedis.
func CacheIniter(initer cache.Initer, attrs ...attribute.KeyValue) cache.Initer {
	return func(ctx context.Context, args ...interface{}) (cache.Cache, error) {
		c, err := initer(ctx, args...)
		if err != nil {
			return nil, err
		}
		return &tracedCache{Cache: c, tracer: newStoreTracer(attrs)}, nil
	}
}

type tracedCache struct {
	cache.Cache
	tracer storeTracer
}

func (c *tracedCache) Get(ctx context.Context, key string) (interface{}, error) {
	ctx, span := c.tracer.start(ctx, "cache.Get")
	v, err := c.Cache.Get(ctx, key)
	span.SetAttributes(attribute.Bool("cache.hit", err == nil))
	if errors.Is(err, os.ErrNotExist) {
		end(span, nil)
	} else {
		end(span, err)
	}
	return v, err
}

func (c *tracedCache) Set(ctx context.Context, key string, value interface{}, lifetime time.Duration) error {
	ctx, span := c.tracer.start(ctx, "cache.Set")
	err := c.Cache.Set(ctx, key, value, lifetime)
	end(span, err)
	return err
}

func (c *tracedCache) Delete(ctx context.Context, key string) error {
	ctx, span := c.tracer.start(ctx, "cache.Delete")
	err := c.Cache.Delete(ctx, key)
	end(span, err)
	return err
}

// SessionIniter wraps the flamego session initer to trace every session store operation, with attrs describing the backend, e.g. semconv.DBSystemRedis.
func SessionIniter(initer session.Initer, attrs ...attribute.KeyValue) session.Initer {
	return func(ctx context.Context, args ...interface{}) (session.Store, error) {
		s, err := initer(ctx, args...)
		if err != nil {
			return nil, err
		}
		return &tracedSessionStore{Store: s, tracer: newStoreTracer(attrs)}, nil
	}
}

type tracedSessionStore struct {
	session.Store
	tracer storeTracer
}

func (s *tracedSessionStore) Exist(ctx context.Context, sid string) bool {
	ctx, span := s.tracer.start(ctx, "session.Exist")
	ok := s.Store.Exist(ctx, sid)
	end(span, nil)
	return ok
}

func (s *tracedSessionStore) Read(ctx context.Context, sid string) (session.Session, error) {
	ctx, span := s.tracer.start(ctx, "session.Read")
	sess, err := s.Store.Read(ctx, sid)
	end(span, err)
	return sess, err
}

func (s *tracedSessionStore) Destroy(ctx context.Context, sid string) error {
	ctx, span := s.tracer.start(ctx, "session.Destroy")
	err := s.Store.Destroy(ctx, sid)
	end(span, err)
	return err
}

func (s *tracedSessionStore) Save(ctx context.Context, sess session.Session) error {
	ctx, span := s.tracer.start(ctx, "session.Save")
	err := s.Store.Save(ctx, sess)
	end(span, err)
	return err
}
