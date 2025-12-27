package service

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/context"
)

func Proxy(ctx context.Context) error {
	span := trace.SpanFromContext(ctx.Request().Context())

	var userID uint
	if ctx.IsLogged {
		userID = ctx.User.ID
	}

	if span.SpanContext().IsValid() {
		span.SetAttributes(
			attribute.Int("nekobox.service.user-id", int(userID)),
		)
	}

	uri := ctx.Param("**")
	basePath := strings.Split(uri, "/")[0]
	forwardPath := strings.TrimPrefix(uri, basePath)

	var forwardURLStr string
	for _, backend := range conf.Service.Backends {
		if backend.Prefix == basePath {
			forwardURLStr = backend.ForwardURL
			break
		}
	}
	if len(forwardURLStr) == 0 {
		return ctx.JSONError(http.StatusNotFound, "页面不存在")
	}

	forwardURL, err := url.Parse(forwardURLStr)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to parse forward URL")
		return ctx.JSONError(http.StatusInternalServerError, "服务网关内部错误")
	}

	reverseProxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = forwardURL
			req.URL.Path = strings.TrimRight(req.URL.Path, "/") + forwardPath
			req.Host = forwardURL.Host

			traceHeaders := http.Header{}
			otel.GetTextMapPropagator().Inject(ctx.Request().Context(), propagation.HeaderCarrier(traceHeaders))
			for key := range traceHeaders {
				req.Header.Set(key, traceHeaders.Get(key))
			}

			req.Header.Set("X-NekoBox-From", "nekobox-gateway")
			req.Header.Set("X-NekoBox-User-ID", strconv.Itoa(int(userID)))
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to handle reverse proxy request")
			_ = ctx.JSONError(http.StatusInternalServerError, "服务网关内部错误")
		},
	}

	reverseProxy.ServeHTTP(ctx.ResponseWriter(), ctx.Request().Request)
	return nil
}
