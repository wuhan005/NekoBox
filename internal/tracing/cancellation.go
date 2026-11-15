// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tracing

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type requestTraceKey struct{}

type requestTrace struct {
	ctx      context.Context
	canceled atomic.Bool
	failed   atomic.Bool
}

func requestTraceFromContext(ctx context.Context) *requestTrace {
	if ctx == nil {
		return nil
	}

	request, _ := ctx.Value(requestTraceKey{}).(*requestTrace)
	return request
}

func (r *requestTrace) isCanceled(err error) bool {
	return r.ctx.Err() == context.Canceled && errors.Is(err, context.Canceled)
}

// LogHook adds logs to the active span and leaves request cancellation in the application logs.
func LogHook() logrus.Hook {
	return logHook{Hook: otellogrus.NewHook()}
}

type logHook struct{ logrus.Hook }

func (h logHook) Fire(entry *logrus.Entry) error {
	if request := requestTraceFromContext(entry.Context); request != nil {
		err, _ := entry.Data[logrus.ErrorKey].(error)
		if request.isCanceled(err) {
			if entry.Level <= logrus.ErrorLevel {
				request.canceled.Store(true)
			}
			return nil
		}

		if entry.Level <= logrus.ErrorLevel {
			request.failed.Store(true)
		}
	}

	return h.Hook.Fire(entry)
}

// FilterRequestCancellations keeps spans from dependency instrumentation while omitting cancellation errors.
func FilterRequestCancellations(provider trace.TracerProvider) trace.TracerProvider {
	return cancellationTracerProvider{TracerProvider: provider}
}

type cancellationTracerProvider struct{ trace.TracerProvider }

func (p cancellationTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return cancellationTracer{Tracer: p.TracerProvider.Tracer(name, opts...)}
}

type cancellationTracer struct{ trace.Tracer }

func (t cancellationTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	request := requestTraceFromContext(ctx)
	ctx, span := t.Tracer.Start(ctx, name, opts...)

	if request == nil || !span.IsRecording() {
		return ctx, span
	}

	filtered := &cancellationSpan{Span: span, request: request}
	return trace.ContextWithSpan(ctx, filtered), filtered
}

type cancellationSpan struct {
	trace.Span
	request     *requestTrace
	canceledMsg atomic.Pointer[string]
}

func (s *cancellationSpan) RecordError(err error, opts ...trace.EventOption) {
	if s.request.isCanceled(err) {
		message := err.Error()
		s.canceledMsg.Store(&message)
		return
	}

	s.canceledMsg.Store(nil)
	s.Span.RecordError(err, opts...)
}

func (s *cancellationSpan) SetStatus(code codes.Code, description string) {
	// GORM and Redis set the status to the same error message after RecordError.
	if message := s.canceledMsg.Load(); code == codes.Error && message != nil && description == *message {
		return
	}

	s.Span.SetStatus(code, description)
}
