// Copyright 2023 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

var UploadImages UploadImagesStore

var _ UploadImagesStore = (*uploadImages)(nil)

type UploadImagesStore interface {
	Create(ctx context.Context, opts CreateUploadImageOptions) (*UploadImage, error)
	BindUploadImageWithQuestion(ctx context.Context, uploadImageID uint, typ UploadImageQuestionType, questionID uint) error
	GetByQuestionID(ctx context.Context, questionID uint) ([]*UploadImage, error)
	GetByTypeQuestionID(ctx context.Context, typ UploadImageQuestionType, questionID uint) ([]*UploadImage, error)
}

func NewUploadImagesStore(db *gorm.DB) UploadImagesStore {
	return &uploadImages{db}
}

type uploadImages struct {
	*gorm.DB
}

type UploadImage struct {
	gorm.Model
	UploaderUserID uint
	Name           string
	FileSize       int64
	Md5            string
	Key            string
}

type UploadImageQuestionType string

const (
	UploadImageQuestionTypeAsk    UploadImageQuestionType = "ask"
	UploadImageQuestionTypeAnswer UploadImageQuestionType = "answer"
)

type UploadImageQuestion struct {
	Type          UploadImageQuestionType
	UploadImageID uint
	QuestionID    uint
}

type CreateUploadImageOptions struct {
	UploaderUserID uint
	Name           string
	FileSize       int64
	Md5            string
	Key            string
}

func (db *uploadImages) Create(ctx context.Context, opts CreateUploadImageOptions) (*UploadImage, error) {
	image := &UploadImage{
		UploaderUserID: opts.UploaderUserID,
		Name:           opts.Name,
		FileSize:       opts.FileSize,
		Md5:            opts.Md5,
		Key:            opts.Key,
	}
	if err := db.Model(&UploadImage{}).WithContext(ctx).Create(image).Error; err != nil {
		return nil, errors.Wrap(err, "create")
	}
	return image, nil
}

func (db *uploadImages) BindUploadImageWithQuestion(ctx context.Context, uploadImageID uint, typ UploadImageQuestionType, questionID uint) error {
	uploadImageQuestion := &UploadImageQuestion{
		Type:          typ,
		UploadImageID: uploadImageID,
		QuestionID:    questionID,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// Unbind the question's previous image of the same type.
		if err := tx.Where("type = ? AND question_id = ?", typ, questionID).Delete(&UploadImageQuestion{}).Error; err != nil {
			return errors.Wrap(err, "unbind previous image with question")
		}

		if err := tx.Model(&UploadImageQuestion{}).WithContext(ctx).Create(uploadImageQuestion).Error; err != nil {
			return errors.Wrap(err, "bind upload image with question")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction")
	}
	return nil
}

func (db *uploadImages) getBy(ctx context.Context, where string, args ...interface{}) ([]*UploadImage, error) {
	var uploadImageQuestions []*UploadImageQuestion
	if err := db.WithContext(ctx).Model(&UploadImageQuestion{}).Where(where, args...).Find(&uploadImageQuestions).Error; err != nil {
		return nil, errors.Wrap(err, "get image questions")
	}

	uploadImageIDs := lo.Map(uploadImageQuestions, func(item *UploadImageQuestion, _ int) uint {
		return item.UploadImageID
	})

	var uploadImages []*UploadImage
	if err := db.WithContext(ctx).Model(&UploadImage{}).Where("id IN ?", uploadImageIDs).Order("id ASC").Find(&uploadImages).Error; err != nil {
		return nil, errors.Wrap(err, "get images")
	}
	return uploadImages, nil
}

func (db *uploadImages) GetByQuestionID(ctx context.Context, questionID uint) ([]*UploadImage, error) {
	return db.getBy(ctx, "question_id = ?", questionID)
}

func (db *uploadImages) GetByTypeQuestionID(ctx context.Context, typ UploadImageQuestionType, questionID uint) ([]*UploadImage, error) {
	return db.getBy(ctx, "type = ? AND question_id = ?", typ, questionID)
}
