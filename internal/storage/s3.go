// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	"github.com/wuhan005/gadget"

	"github.com/wuhan005/NekoBox/internal/conf"
)

func UploadPictureToS3(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()

	key := fmt.Sprintf("%s%d/%02d/%02d/%s", PictureKeyPrefix, year, month, day, randstr.Hex(15))

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               conf.Upload.ImageEndpoint,
			HostnameImmutable: true,
			Source:            aws.EndpointSourceCustom,
		}, nil
	})

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(conf.Upload.ImageAccessID, conf.Upload.ImageAccessSecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return "", errors.Wrap(err, "load config")
	}

	client := s3.NewFromConfig(cfg)
	if err := gadget.Retry(5, func() error {
		if _, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(conf.Upload.ImageBucket),
			Key:           aws.String(key),
			Body:          file,
			ContentLength: aws.Int64(fileHeader.Size),
		}); err != nil {
			return errors.Wrap(err, "put object")
		}
		return nil
	}); err != nil {
		return "", errors.Wrap(err, "retry 5 times")
	}

	return fmt.Sprintf("https://%s/%s", conf.Upload.ImageBucketCDNHost, key), nil
}
