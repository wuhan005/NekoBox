// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package censor

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/db"
)

var (
	censorCacheNoMoreThan = 31 * 24 * time.Hour // 1 month
)

// Text checks the text for sensitive content.
// It will save the censor log to the database and invoke the callback function.
func Text(ctx context.Context, text string) (*TextCensorResponse, error) {
	if !conf.Security.EnableTextCensor {
		return &TextCensorResponse{Pass: true}, nil
	}

	var responses []*TextCensorResponse

	for _, censor := range []TextCensor{
		// For we are using aliyun's API, we don't need to use qiniu's API.
		//NewQiniuTextCensor(conf.App.QiniuAccessKey, conf.App.QiniuAccessSecret),
		NewAliyunTextCensor(conf.App.AliyunAccessKey, conf.App.AliyunAccessKeySecret),
	} {
		sourceName := censor.String()

		// Try to get the censor from the database log.
		censorLog, err := db.CensorLogs.GetByText(ctx, sourceName, text, time.Now().Add(-censorCacheNoMoreThan))
		if err != nil {
			if !errors.Is(err, db.ErrCensorLogsNotFound) {
				logrus.WithContext(ctx).WithError(err).Error("Failed to get censor log")
			}
		} else {
			raw := censorLog.RawResponse
			// We got the previous censor log cache, seems like we saved money for the API call.
			switch censorLog.SourceName {
			case "qiniu":
				// HACK: Qiniu's API response is not accurate, so we need to check the text again with aliyun API.
				response, err := QiniuTextCensorParser(raw)
				if err == nil {
					if response.Pass {
						return response, nil
					}

					responses = append(responses, response) // If the text is not passed, we need to check it with aliyun API.
					continue
				}

			case "aliyun":
				return AliyunTextCensorParser(raw)

			default:
				logrus.WithContext(ctx).WithField("censor_source", censorLog.SourceName).Error("Unknown censor source")
			}
		}

		censorText := RemoveTrustedURL(text)
		response, err := censor.Censor(ctx, censorText)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).WithField("censor_source", sourceName).Error("Failed to censor text")
			continue
		}
		responses = append(responses, response)

		// Save the censor log to the database.
		if err := db.CensorLogs.Create(ctx, db.CreateCensorLogOptions{
			SourceName:  sourceName,
			Input:       text,
			Pass:        response.Pass,
			RawResponse: response.RawResponse,
		}); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("Failed to save censor log")
		}

		if response.Pass {
			return response, nil
		}
	}

	if len(responses) == 0 {
		return nil, errors.New("no censor source available")
	}
	return responses[len(responses)-1], nil
}
