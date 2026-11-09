package cmd

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWebProbes(t *testing.T) {
	signals, stop := context.WithCancel(context.Background())
	defer stop()
	var dependencyErr error
	handler := webProbes(http.NotFoundHandler(), signals, func(context.Context) error {
		return dependencyErr
	})
	status := func(path string) int {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w.Code
	}

	require.Equal(t, http.StatusOK, status("/-/readyz"))
	require.Equal(t, http.StatusNotFound, status("/missing"))
	dependencyErr = context.DeadlineExceeded
	require.Equal(t, http.StatusServiceUnavailable, status("/-/readyz"))
	require.Equal(t, http.StatusOK, status("/-/healthz"))

	dependencyErr = nil
	stop()
	require.Equal(t, http.StatusServiceUnavailable, status("/-/readyz"))
	require.Equal(t, http.StatusOK, status("/-/healthz"))
}

func TestServeWebDrainsRequestsAndNotifications(t *testing.T) {
	signals, stop := context.WithCancel(context.Background())
	defer stop()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	started := make(chan struct{})
	requestCanFinish := make(chan struct{})
	notificationCanFinish := make(chan struct{})
	releaseNotification := sync.OnceFunc(func() { close(notificationCanFinish) })
	var notifications sync.WaitGroup
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-requestCanFinish:
		case <-r.Context().Done():
			return
		}
		notifications.Add(1)
		go func() {
			defer notifications.Done()
			<-notificationCanFinish
		}()
		w.WriteHeader(http.StatusNoContent)
	})}
	t.Cleanup(func() {
		_ = server.Close()
		releaseNotification()
	})
	shuttingDown := make(chan struct{})
	server.RegisterOnShutdown(func() { close(shuttingDown) })
	const drainDelay = 20 * time.Millisecond
	stopped := make(chan error, 1)
	go func() { stopped <- serveWeb(signals, listener, server, &notifications, drainDelay, 3*time.Second) }()

	client := &http.Client{Timeout: 3 * time.Second}
	status := make(chan int, 1)
	go func() {
		resp, err := client.Get("http://" + listener.Addr().String())
		if err != nil {
			status <- 0
			return
		}
		defer func() { _ = resp.Body.Close() }()
		status <- resp.StatusCode
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not reach the server")
	}
	drainingSince := time.Now()
	stop()
	select {
	case <-shuttingDown:
		require.GreaterOrEqual(t, time.Since(drainingSince), drainDelay)
	case <-time.After(3 * time.Second):
		t.Fatal("server did not start shutdown")
	}
	close(requestCanFinish)
	require.Equal(t, http.StatusNoContent, <-status)

	select {
	case err := <-stopped:
		t.Fatalf("server stopped before the notification finished: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	releaseNotification()
	require.NoError(t, <-stopped)
}
