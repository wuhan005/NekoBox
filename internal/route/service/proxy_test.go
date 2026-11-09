package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flamego/flamego"
	"github.com/stretchr/testify/require"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/db"
)

func TestProxyRewritesForwardingHeaders(t *testing.T) {
	forwarded := make(chan *http.Request, 1)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded <- r
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(backend.Close)

	previousBackends := conf.Service.Backends
	t.Cleanup(func() { conf.Service.Backends = previousBackends })
	conf.Service.Backends = []struct {
		Prefix     string `ini:"prefix"`
		ForwardURL string `ini:"forward_url"`
	}{{Prefix: "pixel", ForwardURL: backend.URL + "/api/?fixed=1"}}

	user := &db.User{}
	user.ID = 42
	gateway := flamego.New()
	gateway.Any("/service/{**}", func(ctx flamego.Context) error {
		return Proxy(context.Context{Context: ctx, IsSignedIn: true, User: user})
	})

	req := httptest.NewRequest(http.MethodGet, "https://gateway.example/service/pixel/boards?client=1", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	req.Header.Set("Forwarded", "for=203.0.113.5")
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	req.Header.Set("X-Forwarded-Host", "spoofed.example")
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("Connection", "X-NekoBox-From, X-NekoBox-User-ID")
	req.Header.Set("X-NekoBox-User-ID", "99")
	resp := httptest.NewRecorder()
	gateway.ServeHTTP(resp, req)
	require.Equal(t, http.StatusNoContent, resp.Code)

	out := <-forwarded
	require.Equal(t, "/api/boards?fixed=1", out.URL.RequestURI())
	require.Equal(t, backend.Listener.Addr().String(), out.Host)
	require.Empty(t, out.Header.Get("Forwarded"))
	require.Equal(t, "192.0.2.10", out.Header.Get("X-Forwarded-For"))
	require.Equal(t, "gateway.example", out.Header.Get("X-Forwarded-Host"))
	require.Equal(t, "https", out.Header.Get("X-Forwarded-Proto"))
	require.Equal(t, "nekobox-gateway", out.Header.Get("X-NekoBox-From"))
	require.Equal(t, "42", out.Header.Get("X-NekoBox-User-ID"))
}
