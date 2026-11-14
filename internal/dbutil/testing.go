// Copyright 2023 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbutil

import (
	"context"
	"crypto/rand"
	"flag"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var flagParseOnce sync.Once

func NewTestDB(t *testing.T, migrationTables ...interface{}) (testDB *gorm.DB, cleanup func(...string) error) {
	dbType := os.Getenv("DB_TYPE")

	var dsn string
	var dialectFunc func(string) gorm.Dialector

	switch dbType {
	case "mysql":
		dsn = os.ExpandEnv("$DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_DATABASE?charset=utf8mb4&parseTime=True&loc=Local")
		dialectFunc = mysql.Open
	case "postgres":
		dsn = os.ExpandEnv("postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT?sslmode=disable")
		dialectFunc = postgres.Open
	default:
		t.Fatalf("Unknown database type: %q", dbType)
	}

	db, err := gorm.Open(dialectFunc(dsn), &gorm.Config{
		NowFunc:                Now,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}

	ctx := context.Background()
	dbname := "nekobox-test-" + rand.Text()

	err = db.WithContext(ctx).Exec(`CREATE DATABASE ` + QuoteIdentifier(dbType, dbname)).Error
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	switch dbType {
	case "mysql":
		dsn = strings.ReplaceAll(dsn, os.Getenv("DB_DATABASE"), dbname)
	case "postgres":
		cfg, err := url.Parse(dsn)
		if err != nil {
			t.Fatalf("Failed to parse DSN")
		}
		cfg.Path = "/" + dbname
		dsn = cfg.String()
	}

	flagParseOnce.Do(flag.Parse)

	testDB, err = gorm.Open(dialectFunc(dsn), &gorm.Config{
		NowFunc:                Now,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("Failed to open test connection: %v", err)
	}

	err = testDB.AutoMigrate(migrationTables...)
	if err != nil {
		t.Fatalf("Failed to auto migrate tables: %v", err)
	}

	t.Cleanup(func() {
		defer func() {
			if database, err := db.DB(); err == nil {
				_ = database.Close()
			}
		}()

		if t.Failed() {
			t.Logf("DATABASE %s left intact for inspection", dbname)
			return
		}

		database, err := testDB.DB()
		if err != nil {
			t.Fatalf("Failed to get currently open database: %v", err)
		}

		err = database.Close()
		if err != nil {
			t.Fatalf("Failed to close currently open database: %v", err)
		}

		err = db.WithContext(ctx).Exec(`DROP DATABASE ` + QuoteIdentifier(dbType, dbname)).Error
		if err != nil {
			t.Fatalf("Failed to drop test database: %v", err)
		}
	})

	return testDB, func(tables ...string) error {
		if t.Failed() {
			return nil
		}

		for _, table := range tables {
			var query string
			switch dbType {
			case "mysql":
				query = `TRUNCATE TABLE ` + QuoteIdentifier(dbType, table)
			case "postgres":
				query = `TRUNCATE TABLE ` + QuoteIdentifier(dbType, table) + ` RESTART IDENTITY CASCADE`
			}

			err := testDB.WithContext(ctx).Exec(query).Error
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// QuoteIdentifier quotes an "identifier" (e.g. a table or a column name) to be
// used as part of an SQL statement.
func QuoteIdentifier(typ, s string) string {
	if typ == "postgres" {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}
