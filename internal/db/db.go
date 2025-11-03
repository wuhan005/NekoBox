// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/conf"
)

var AllTables = []interface{}{
	&User{}, &Question{}, &CensorLog{}, &UploadImage{}, &UploadImageQuestion{},
}

func Init(typ, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch typ {
	case "mysql", "":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return nil, errors.Errorf("unknown database type: %q", typ)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "connect to database")
	}

	if err := db.AutoMigrate(AllTables...); err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	Users = NewUsersStore(db)
	Questions = NewQuestionsStore(db)
	CensorLogs = NewCensorLogsStore(db)
	UploadImages = NewUploadImagesStore(db)

	if err := db.Use(otelgorm.NewPlugin(
		otelgorm.WithDBName(conf.Database.Name),
	)); err != nil {
		return nil, errors.Wrap(err, "register otelgorm plugin")
	}

	return db, nil
}
