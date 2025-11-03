// Copyright 2023 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"testing"

	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/dbutil"
)

// newTestDB returns a test database instance with the cleanup function.
func newTestDB(t *testing.T) (*gorm.DB, func(...string) error) {
	return dbutil.NewTestDB(t, AllTables...)
}
