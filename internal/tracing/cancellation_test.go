// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tracing

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/flamego/flamego"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestRequestCancellation(t *testing.T) {
	for _, tt := range []struct {
		name            string
		queryErr        error
		cancelRequest   bool
		otherFailure    bool
		writeResponse   bool
		wantQueryStatus codes.Code
		wantHTTPStatus  codes.Code
	}{
		{"client disconnect", context.Canceled, true, false, true, codes.Unset, codes.Unset},
		{"disconnect without response", context.Canceled, true, false, false, codes.Unset, codes.Unset},
		{"canceled operation with live request", context.Canceled, false, false, true, codes.Error, codes.Error},
		{"deadline after disconnect", context.DeadlineExceeded, true, false, true, codes.Error, codes.Error},
		{"other failure after disconnect", context.Canceled, true, true, true, codes.Unset, codes.Error},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := FilterRequestCancellations(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))

			var logs bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&logs)
			logger.AddHook(LogHook())

			database, err := gorm.Open(postgres.Open("host=localhost user=test"), &gorm.Config{
				DryRun:               true,
				DisableAutomaticPing: true,
				Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
			})
			require.NoError(t, err)

			pool, err := database.DB()
			require.NoError(t, err)
			t.Cleanup(func() { _ = pool.Close() })

			require.NoError(t, database.Use(otelgorm.NewPlugin(otelgorm.WithTracerProvider(provider), otelgorm.WithoutMetrics())))
			require.NoError(t, database.Callback().Query().Before("gorm:query").Register("test:query-error", func(tx *gorm.DB) {
				tx.AddError(fmt.Errorf("timeout: %w", tt.queryErr))
			}))

			request := httptest.NewRequest(http.MethodGet, "/questions/1", nil)
			requestCtx, cancel := context.WithCancel(request.Context())
			defer cancel()
			response := httptest.NewRecorder()

			app := flamego.New()
			app.Use(Middleware("test", WithTracerProvider(provider)))
			app.Get("/questions/1", func(c flamego.Context) {
				if tt.cancelRequest {
					cancel()
				}

				ctx := c.Request().Context()
				var row struct{ ID uint }
				err := database.WithContext(ctx).Table("questions").First(&row, 1).Error
				assert.ErrorIs(t, err, tt.queryErr)

				if !tt.writeResponse {
					return
				}
				logger.WithContext(ctx).WithError(errors.Wrap(err, "get question by ID")).Error("Failed to get question by ID")
				if tt.otherFailure {
					logger.WithContext(ctx).Error("Another failure")
				}

				c.ResponseWriter().WriteHeader(http.StatusInternalServerError)
			})

			app.ServeHTTP(response, request.WithContext(requestCtx))

			if tt.writeResponse {
				assert.Equal(t, http.StatusInternalServerError, response.Code)
				assert.Contains(t, logs.String(), "get question by ID: timeout: "+tt.queryErr.Error())
			}

			spans := recorder.Ended()
			require.Len(t, spans, 2)
			query, requestSpan := spans[0], spans[1]
			assert.Equal(t, "gorm.Query", query.Name())
			assert.Equal(t, requestSpan.SpanContext(), query.Parent())
			assert.Equal(t, tt.wantQueryStatus, query.Status().Code)
			assert.Equal(t, tt.wantHTTPStatus, requestSpan.Status().Code)

			if tt.wantQueryStatus == codes.Unset {
				assert.Empty(t, query.Events())
			} else {
				require.Len(t, query.Events(), 1)
				assert.Equal(t, "exception", query.Events()[0].Name)
			}

			if tt.wantHTTPStatus == codes.Unset {
				assert.Empty(t, requestSpan.Events())
				assert.Contains(t, requestSpan.Attributes(), attribute.Bool("http.request.canceled", true))
				assert.NotContains(t, requestSpan.Attributes(), semconv.HTTPStatusCodeKey.Int(500))
			} else {
				assert.Len(t, requestSpan.Events(), 1)
				assert.Contains(t, requestSpan.Attributes(), semconv.HTTPStatusCodeKey.Int(500))
			}
		})
	}
}
