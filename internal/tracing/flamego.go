// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tracing

import (
	"fmt"
	"net/http"

	"github.com/flamego/flamego"

	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// instrumentationName is the name of this instrumentation package.
const instrumentationName = "github.com/wuhan005/NekoBox/internal/tracing"

// Middleware returns a flamego Handler to trace requests to the server.
func Middleware(service string, opts ...Option) flamego.Handler {
	cfg := newConfig(opts)
	tracer := cfg.TracerProvider.Tracer(
		instrumentationName,
		oteltrace.WithInstrumentationVersion("1.0.0"),
	)
	return func(res http.ResponseWriter, req *http.Request, c flamego.Context) {
		savedCtx := c.Request().Context()
		defer func() {
			c.Request().Request = c.Request().WithContext(savedCtx)
		}()

		ctx := cfg.Propagators.Extract(savedCtx, propagation.HeaderCarrier(c.Request().Header))
		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", c.Request().Request)...),
			oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(c.Request().Request)...),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(service, "", c.Request().Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}
		// TODO: span name should be router template not the actual request path, eg /user/:id vs /user/123
		spanName := c.Request().RequestURI
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request().Method)
		}
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		c.Request().Request = c.Request().WithContext(ctx)

		// serve the request to the next middleware
		c.Next()

		status := c.ResponseWriter().Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(status, oteltrace.SpanKindServer)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
	}
}
