// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/dbutil"
)

// Context represents context of a request.
type Context struct {
	flamego.Context

	IsSignedIn bool
	User       *db.User
}

// Success sends a successful response with optional data.
func (c *Context) Success(data ...interface{}) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.ResponseWriter().WriteHeader(http.StatusOK)

	var d interface{}
	if len(data) == 1 {
		d = data[0]
	}

	err := json.NewEncoder(c.ResponseWriter()).Encode(
		map[string]interface{}{
			"data": d,
		},
	)
	if err != nil {
		logrus.WithContext(c.Request().Context()).WithError(err).Error("Failed to encode")
		return c.ServerError()
	}
	return nil
}

// ServerError sends a 500 Internal Server Error response.
func (c *Context) ServerError() error {
	return c.Error(http.StatusInternalServerError, "服务器内部错误，请重试")
}

// Error sends an error response with a specific status code and message.
func (c *Context) Error(statusCode int, message string, v ...interface{}) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.ResponseWriter().WriteHeader(statusCode)

	err := json.NewEncoder(c.ResponseWriter()).Encode(
		map[string]interface{}{
			"error": statusCode,
			"msg":   fmt.Sprintf(message, v...),
		},
	)
	if err != nil {
		logrus.WithContext(c.Request().Context()).WithError(err).Error("Failed to encode")
		return c.ServerError()
	}
	return nil
}

// Status sets the HTTP status code for the response.
func (c *Context) Status(statusCode int) {
	c.ResponseWriter().WriteHeader(statusCode)
}

// IP retrieves the client's IP address from the request.
func (c *Context) IP() string {
	ipHeader := conf.App.IPHeader
	if ipHeader != "" {
		return c.Request().Header.Get(ipHeader)
	}
	return c.Request().RemoteAddr
}

// Contexter initializes a classic context for a request.
func Contexter(gormDB *gorm.DB) flamego.Handler {
	return func(ctx flamego.Context, sess session.Session) {
		c := Context{
			Context: ctx,
		}

		// Get user from session or header when possible
		c.User = authenticatedUser(c.Context, sess)

		var userID uint
		if c.User != nil {
			c.IsSignedIn = true
			userID = c.User.ID
		}

		span := trace.SpanFromContext(ctx.Request().Context())
		if span.IsRecording() {
			span.SetAttributes(
				attribute.Bool("nekobox.user.is-login", c.IsSignedIn),
				attribute.Int("nekobox.user.id", int(userID)),
			)
		}
		c.ResponseWriter().Header().Set("Trace-ID", span.SpanContext().TraceID().String())

		c.MapTo(gormDB, (*dbutil.Transactor)(nil))
		c.Map(c)
	}
}
