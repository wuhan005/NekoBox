// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"github.com/urfave/cli/v2"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/route"
	"github.com/wuhan005/NekoBox/internal/tracing"
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

	db, err := db.Init(dbType, dsn)
	if err != nil {
		return errors.Wrap(err, "connect to database")
	}

	logrus.WithContext(ctx.Context).WithField("external_url", conf.App.ExternalURL).Info("Starting web server")
	r := route.New(db)
	r.Use(tracing.Middleware("NekoBox"))
	r.Run(conf.Server.Port)

	return nil
}
