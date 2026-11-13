// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package censor

import (
	"cmp"
	"context"
	"encoding/json"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/green"
	"github.com/cockroachdb/errors"
)

type AliyunTextCensor struct {
	region                     string
	accessKey, accessKeySecret string
}

func NewAliyunTextCensor(region, accessKey, accessKeySecret string) *AliyunTextCensor {
	return &AliyunTextCensor{
		region:          cmp.Or(region, "cn-shanghai"),
		accessKey:       accessKey,
		accessKeySecret: accessKeySecret,
	}
}

type AliyunTextCensorResponse struct {
	Code int `json:"code"`
	Data []struct {
		Code            int    `json:"code"`
		Content         string `json:"content"`
		FilteredContent string `json:"filteredContent"`
		Msg             string `json:"msg"`
		Results         []struct {
			Details []struct {
				Contexts []struct {
					Context   string `json:"context"`
					Positions []struct {
						EndPos   int `json:"endPos"`
						StartPos int `json:"startPos"`
					} `json:"positions"`
				} `json:"contexts"`
				Label string `json:"label"`
			} `json:"details"`
			Label      string  `json:"label"`
			Rate       float64 `json:"rate"`
			Scene      string  `json:"scene"`
			Suggestion string  `json:"suggestion"`
		} `json:"results"`
		TaskId string `json:"taskId"`
	} `json:"data"`
	Msg       string `json:"msg"`
	RequestId string `json:"requestId"`
}

func (r *AliyunTextCensorResponse) IsPass() bool {
	if len(r.Data) == 0 {
		return false
	}

	// ⚠️ Right now, we allow `review` and `pass` to pass the censor.
	for _, result := range r.Data[0].Results {
		return result.Suggestion != "block"
	}
	return false
}

// Censor censors text with Aliyun API.
// https://developer.qiniu.com/censor/7260/api-text-censor
func (c *AliyunTextCensor) Censor(_ context.Context, text string) (*TextCensorResponse, error) {
	client, err := green.NewClientWithAccessKey(c.region, c.accessKey, c.accessKeySecret)
	if err != nil {
		return nil, errors.Wrap(err, "new client with access key")
	}

	content, err := json.Marshal(
		map[string]interface{}{
			"scenes": []string{"antispam"},
			"tasks": []map[string]interface{}{
				{"content": text},
			},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "marshal content")
	}

	req := green.CreateTextScanRequest()
	req.SetContent(content)
	resp, err := client.TextScan(req)
	if err != nil {
		return nil, errors.Wrap(err, "text scan")
	}

	if !resp.IsSuccess() {
		return nil, errors.New("response is not success")
	}

	return AliyunTextCensorParser(resp.GetHttpContentBytes())
}

func (*AliyunTextCensor) String() string {
	return "aliyun"
}

func AliyunTextCensorParser(raw []byte) (*TextCensorResponse, error) {
	var responseJSON AliyunTextCensorResponse
	if err := json.Unmarshal(raw, &responseJSON); err != nil {
		return nil, errors.Wrap(err, "unmarshal response")
	}

	if len(responseJSON.Data) == 0 {
		return nil, errors.New("response data is empty")
	}

	var hint string
	var label string
	var confidence float64
	for _, result := range responseJSON.Data[0].Results {
		if result.Label == "normal" {
			continue
		}

		// Get the first context as the hint, forbidden type, confidence.
		for _, detail := range result.Details {
			for _, context := range detail.Contexts {
				hint = context.Context
			}
		}

		label = result.Label
		confidence = result.Rate
		break
	}

	return &TextCensorResponse{
		SourceName:    "aliyun",
		Pass:          responseJSON.IsPass(),
		ForbiddenType: formatAliyunForbiddenType(label),
		Hint:          hint,
		Confidence:    confidence,
		RawResponse:   raw,
	}, nil
}

func formatAliyunForbiddenType(typ string) ForbiddenType {
	switch typ {
	case "spam":
		return ForbiddenTypeSpam
	case "ad":
		return ForbiddenTypeAd
	case "politics":
		return ForbiddenTypePolitics
	case "terrorism":
		return ForbiddenTypeTerrorism
	case "abuse":
		return ForbiddenTypeAbuse
	case "porn":
		return ForbiddenTypePorn
	case "flood":
		return ForbiddenTypeFlood
	case "contraband":
		return ForbiddenTypeContraband
	case "meaningless":
		return ForbiddenTypeMeaningless
	case "harmful":
		return ForbiddenTypeHarmful
	default:
		return ""
	}
}
