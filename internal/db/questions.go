// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/dbutil"
)

var Questions QuestionsStore

type QuestionsStore interface {
	Create(ctx context.Context, opts CreateQuestionOptions) (*Question, error)
	GetByID(ctx context.Context, id uint) (*Question, error)
	GetByUserID(ctx context.Context, userID uint, opts GetQuestionsByUserIDOptions) ([]*Question, error)
	GetByAskUserID(ctx context.Context, userID uint, opts GetQuestionsByAskUserIDOptions) ([]*Question, error)
	AnswerByID(ctx context.Context, id uint, answer string) error
	DeleteByID(ctx context.Context, id uint) error
	UpdateCensor(ctx context.Context, id uint, opts UpdateQuestionCensorOptions) error
	Count(ctx context.Context, userID uint, opts GetQuestionsCountOptions) (int64, error)
	SetPrivate(ctx context.Context, id uint) error
	SetPublic(ctx context.Context, id uint) error
}

func NewQuestionsStore(db *gorm.DB) QuestionsStore {
	return &questions{db}
}

type questions struct {
	*gorm.DB
}

type Question struct {
	dbutil.Model
	FromIP                string         `json:"-"`
	UserID                uint           `gorm:"index:idx_question_user_id" json:"-"`
	Content               string         `json:"content"`
	ContentCensorMetadata datatypes.JSON `json:"-"`
	Token                 string         `json:"-"`
	Answer                string         `json:"answer"`
	AnswerCensorMetadata  datatypes.JSON `json:"-"`
	ReceiveReplyEmail     string         `json:"-"`
	AskerUserID           uint           `json:"-"`
	IsPrivate             bool           `gorm:"default: FALSE; NOT NULL" json:"-"`
}

type CreateQuestionOptions struct {
	FromIP            string
	UserID            uint
	Content           string
	ReceiveReplyEmail string
	AskerUserID       uint
	IsPrivate         bool
}

func (db *questions) Create(ctx context.Context, opts CreateQuestionOptions) (*Question, error) {
	question := Question{
		FromIP:            opts.FromIP,
		UserID:            opts.UserID,
		Token:             randstr.String(6),
		Content:           opts.Content,
		ReceiveReplyEmail: opts.ReceiveReplyEmail,
		AskerUserID:       opts.AskerUserID,
		IsPrivate:         opts.IsPrivate,
	}
	return &question, db.WithContext(ctx).Create(&question).Error
}

type UpdateQuestionCensorOptions struct {
	ContentCensorMetadata json.RawMessage
	AnswerCensorMetadata  json.RawMessage
}

func (db *questions) UpdateCensor(ctx context.Context, id uint, opts UpdateQuestionCensorOptions) error {
	question, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get by ID")
	}

	contentCensorMetadata := question.ContentCensorMetadata
	if checkTextCensorResponseValid(opts.ContentCensorMetadata) {
		contentCensorMetadata = datatypes.JSON(opts.ContentCensorMetadata)
	}
	answerCensorMetadata := question.AnswerCensorMetadata
	if checkTextCensorResponseValid(opts.AnswerCensorMetadata) {
		answerCensorMetadata = datatypes.JSON(opts.AnswerCensorMetadata)
	}

	return db.WithContext(ctx).Model(&Question{}).Where("id = ?", id).Updates(&Question{
		ContentCensorMetadata: contentCensorMetadata,
		AnswerCensorMetadata:  answerCensorMetadata,
	}).Error
}

func checkTextCensorResponseValid(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}

	if bytes.EqualFold(raw, []byte("null")) {
		return false
	}

	var response struct {
		SourceName string `json:"source_name"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return false
	}
	return response.SourceName != ""
}

var ErrQuestionNotExist = errors.New("提问不存在")

func (db *questions) GetByID(ctx context.Context, id uint) (*Question, error) {
	var question Question
	if err := db.WithContext(ctx).First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotExist
		}
		return nil, errors.Wrap(err, "get question by ID")
	}
	return &question, nil
}

func (db *questions) getBy(ctx context.Context, cursor *dbutil.Cursor, whereQuery string, args ...interface{}) ([]*Question, error) {
	var questions []*Question
	q := db.WithContext(ctx).Where(whereQuery, args...)

	if cursor != nil {
		cursorID := cursor.Value
		if cursorID != nil && fmt.Sprintf("%v", cursorID) != "" {
			// For we ordered by ID DESC, so we need to use `>` instead of `<`.
			q = q.Where(`id < ?`, cursorID)
		}

		limit := cursor.Limit()
		q = q.Limit(limit)
	}

	q = q.Order("created_at DESC")
	if err := q.Find(&questions).Error; err != nil {
		return nil, errors.Wrap(err, "get questions by page ID")
	}
	return questions, nil
}

type GetQuestionsByUserIDOptions struct {
	*dbutil.Cursor
	FilterAnswered bool
	ShowPrivate    bool
}

func (db *questions) GetByUserID(ctx context.Context, userID uint, opts GetQuestionsByUserIDOptions) ([]*Question, error) {
	where := `user_id = ?`
	args := userID

	if opts.FilterAnswered {
		where = `user_id = ? AND answer != ''`
	}
	if !opts.ShowPrivate {
		where += ` AND is_private = false`
	}

	questions, err := db.getBy(ctx, opts.Cursor, where, args)
	if err != nil {
		return nil, errors.Wrap(err, "get by")
	}
	return questions, nil
}

type GetQuestionsByAskUserIDOptions struct {
	*dbutil.Cursor
	FilterAnswered bool
	ShowPrivate    bool
}

func (db *questions) GetByAskUserID(ctx context.Context, userID uint, opts GetQuestionsByAskUserIDOptions) ([]*Question, error) {
	where := `asker_user_id = ?`
	args := userID

	if opts.FilterAnswered {
		where = `asker_user_id = ? AND answer != ''`
	}
	if !opts.ShowPrivate {
		where += ` AND is_private = false`
	}

	questions, err := db.getBy(ctx, opts.Cursor, where, args)
	if err != nil {
		return nil, errors.Wrap(err, "get by")
	}
	return questions, nil
}

func (db *questions) AnswerByID(ctx context.Context, id uint, answer string) error {
	var question Question
	if err := db.WithContext(ctx).First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrQuestionNotExist
		}
		return errors.Wrap(err, "get question by ID")
	}

	if err := db.WithContext(ctx).Model(&question).Where("id = ?", id).Update("answer", answer).Error; err != nil {
		return errors.Wrap(err, "update question answer")
	}
	return nil
}

func (db *questions) DeleteByID(ctx context.Context, id uint) error {
	var question Question
	if err := db.WithContext(ctx).First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrQuestionNotExist
		}
		return errors.Wrap(err, "get question by ID")
	}

	if err := db.WithContext(ctx).Delete(&Question{}, id).Error; err != nil {
		return errors.Wrap(err, "delete question")
	}
	return nil
}

type GetQuestionsCountOptions struct {
	FilterAnswered bool
	ShowPrivate    bool
}

func (db *questions) Count(ctx context.Context, userID uint, opts GetQuestionsCountOptions) (int64, error) {
	q := db.WithContext(ctx).Model(&Question{})
	if opts.FilterAnswered {
		q = q.Where(`user_id = ? AND answer != ''`, userID)
	} else {
		q = q.Where(`user_id = ?`, userID)
	}
	if !opts.ShowPrivate {
		q = q.Where(`is_private = ?`, false)
	}

	var count int64
	return count, q.Count(&count).Error
}

func (db *questions) SetPrivate(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Model(&Question{}).Where("id = ?", id).Update("is_private", true).Error
}

func (db *questions) SetPublic(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Model(&Question{}).Where("id = ?", id).Update("is_private", false).Error
}
