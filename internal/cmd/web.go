// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"github.com/urfave/cli/v2"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/route"
)

var Web = &cli.Command{
	Name:   "web",
	Usage:  "Start web server",
	Action: runWeb,
}

func runWeb(ctx *cli.Context) error {
	if err := conf.Init(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if conf.App.UptraceDSN != "" {
		uptrace.ConfigureOpentelemetry(
			uptrace.WithDSN(conf.App.UptraceDSN),
			uptrace.WithServiceName("nekobox"),
			uptrace.WithServiceVersion(conf.BuildCommit),
		)
		defer func() {
			flush, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := uptrace.Shutdown(flush); err != nil {
				logrus.WithError(err).Warn("Failed to shut down Uptrace")
			}
		}()
		logrus.WithContext(ctx.Context).Debug("Tracing enabled.")
	}

	logrus.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	)))

	dbType := conf.Database.Type

	var dsn string
	switch dbType {
	case "mysql", "":
		dsn = conf.MySQLDsn()
	case "postgres":
		dsn = conf.PostgresDsn()
	default:
		return errors.Errorf("unknown database type: %q", dbType)
	}
	conf.Database.DSN = dsn

	database, err := db.Init(dbType, dsn)
	if err != nil {
		return errors.Wrap(err, "connect to database")
	}
	sqlDB, err := database.DB()
	if err != nil {
		return errors.Wrap(err, "get database connection pool")
	}
	defer func() { _ = sqlDB.Close() }()

	var notifications sync.WaitGroup
	r, err := route.New(database, &notifications)
	if err != nil {
		return errors.Wrap(err, "init route")
	}

	// Flamego's Redis stores own their clients, so readiness uses a separate pool.
	redisProbe := redis.NewClient(&redis.Options{
		Addr:       conf.Redis.Addr,
		Password:   conf.Redis.Password,
		MaxRetries: -1,
	})
	defer func() { _ = redisProbe.Close() }()

	signals, stopSignals := signal.NotifyContext(ctx.Context, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	host := "0.0.0.0"
	if addr := os.Getenv("FLAMEGO_ADDR"); addr != "" {
		host, _, _ = strings.Cut(addr, ":")
	}
	address := net.JoinHostPort(host, strconv.Itoa(conf.Server.Port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrapf(err, "listen on %s", address)
	}
	logrus.WithContext(ctx.Context).WithFields(logrus.Fields{
		"addr": listener.Addr(), "external_url": conf.App.ExternalURL,
	}).Info("Listening for HTTP requests")

	requests, cancelRequests := context.WithCancel(context.WithoutCancel(ctx.Context))
	defer cancelRequests()
	server := &http.Server{
		Handler: webProbes(r, signals, func(ctx context.Context) error {
			if err := sqlDB.PingContext(ctx); err != nil {
				return err
			}
			return redisProbe.Ping(ctx).Err()
		}),
		BaseContext:       func(net.Listener) context.Context { return requests },
		ReadHeaderTimeout: 3 * time.Second,
	}
	return serveWeb(signals, listener, server, &notifications, conf.Server.DrainDelay, conf.Server.ShutdownTimeout)
}

// Probes bypass session and other application middleware.
func webProbes(next http.Handler, signals context.Context, ready func(context.Context) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/-/healthz":
			w.WriteHeader(http.StatusOK)
		case "/-/readyz":
			probe, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if signals.Err() != nil || ready(probe) != nil || signals.Err() != nil {
				http.Error(w, "Not ready", http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func serveWeb(signals context.Context, listener net.Listener, server *http.Server, notifications *sync.WaitGroup, drainDelay, timeout time.Duration) error {
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()

	select {
	case err := <-served:
		_ = server.Close()
		return errors.Wrap(err, "serve HTTP")
	case <-signals.Done():
	}

	logrus.WithField("drain_delay", drainDelay).Info("Draining web server")
	shutdown, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	time.Sleep(drainDelay)

	if err := server.Shutdown(shutdown); err != nil {
		_ = server.Close()
		return errors.Wrap(err, "drain HTTP requests")
	}
	if err := <-served; !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "serve HTTP")
	}

	// After HTTP shutdown, no handlers can enqueue another notification.
	sent := make(chan struct{})
	go func() {
		notifications.Wait()
		close(sent)
	}()
	select {
	case <-sent:
	case <-shutdown.Done():
		return errors.Wrap(shutdown.Err(), "drain email notifications")
	}

	logrus.Info("Web server stopped")
	return nil
}
