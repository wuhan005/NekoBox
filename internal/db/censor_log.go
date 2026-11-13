// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/wuhan005/gadget"
	"gorm.io/gorm"
)

var CensorLogs CensorLogsStore

type CensorLogsStore interface {
	GetByText(ctx context.Context, sourceName, text string, noLongerThan ...time.Time) (*CensorLog, error)
	Create(ctx context.Context, options CreateCensorLogOptions) error
}

func NewCensorLogsStore(db *gorm.DB) CensorLogsStore {
	return &censorLogs{db}
}

type censorLogs struct {
	*gorm.DB
}

type CensorLog struct {
	gorm.Model
	SourceName  string
	Input       string
	InputHash   string `gorm:"index"`
	Pass        bool
	RawResponse json.RawMessage
}

var ErrCensorLogsNotFound = errors.New("censor logs dose not exist")

func (db *censorLogs) GetByText(ctx context.Context, sourceName, text string, noLongerThan ...time.Time) (*CensorLog, error) {
	hash := hashText(text)

	var censorLog CensorLog
	q := db.WithContext(ctx).Where("source_name = ? AND input_hash = ?", sourceName, hash)
	if len(noLongerThan) > 0 && !noLongerThan[0].IsZero() {
		q = q.Where("created_at > ?", noLongerThan[0])
	}

	if err := q.First(&censorLog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCensorLogsNotFound
		}
		return nil, err
	}
	return &censorLog, nil
}

type CreateCensorLogOptions struct {
	SourceName  string
	Input       string
	Pass        bool
	RawResponse json.RawMessage
}

func (db *censorLogs) Create(ctx context.Context, options CreateCensorLogOptions) error {
	censorLog := CensorLog{
		SourceName:  options.SourceName,
		Input:       options.Input,
		InputHash:   hashText(options.Input),
		Pass:        options.Pass,
		RawResponse: options.RawResponse,
	}
	return db.WithContext(ctx).Create(&censorLog).Error
}

func hashText(text string) string {
	return gadget.Md5(text)
}
