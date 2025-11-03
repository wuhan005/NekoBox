// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/security/censor"
)

var Censor = &cli.Command{
	Name:   "censor",
	Usage:  "Censor the questions",
	Action: runCensor,
}

func runCensor(ctx *cli.Context) error {
	if err := conf.Init(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if !conf.Security.EnableTextCensor {
		return errors.New("text censor is disabled")
	}

	dbType := "mysql"
	conf.Database.DSN = conf.MySQLDsn()

	database, err := db.Init(dbType, conf.Database.DSN)
	if err != nil {
		return errors.Wrap(err, "connect to database")
	}

	// Check all the unprocessed questions.
	var questions []db.Question
	if err := database.Raw(`SELECT * FROM questions WHERE content_censor_metadata IS NULL`).Find(&questions).Error; err != nil {
		return errors.Wrap(err, "query questions")
	}

	logrus.WithContext(ctx.Context).WithField("count", len(questions)).Info("Found un-censor questions")

	for i, question := range questions {
		question := question

		content := question.Content
		answer := question.Answer
		contentCensorResponse, err := censor.Text(ctx.Context, content)
		if err != nil {
			logrus.WithContext(ctx.Context).WithField("question_id", question.ID).WithError(err).Error("Failed to censor content")
		} else {
			// We don't want to update the `updated_at` field, so just execute the raw SQL.
			if err := database.Debug().Exec(`UPDATE questions SET content_censor_metadata = ? WHERE id = ?`, contentCensorResponse.ToJSON(), question.ID).Error; err != nil {
				logrus.WithContext(ctx.Context).WithField("question_id", question.ID).WithError(err).Error("Failed to update content censor metadata")
			}
		}

		if answer != "" && question.AnswerCensorMetadata == nil {
			answerCensorResponse, err := censor.Text(ctx.Context, answer)
			if err != nil {
				logrus.WithContext(ctx.Context).WithField("question_id", question.ID).WithError(err).Error("Failed to censor answer")
			} else {
				// We don't want to update the `updated_at` field, so just execute the raw SQL.
				if err := database.Exec(`UPDATE questions SET answer_censor_metadata = ? WHERE id = ?`, answerCensorResponse.ToJSON(), question.ID).Error; err != nil {
					logrus.WithContext(ctx.Context).WithField("question_id", question.ID).WithError(err).Error("Failed to update answer censor metadata")
				}
			}
		}

		if i%1000 == 0 {
			logrus.WithContext(ctx.Context).WithField("count", i).Trace("Processed questions")
		}
	}

	logrus.WithContext(ctx.Context).WithField("count", len(questions)).Info("Processed all questions")
	return nil
}
