// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package censor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/samber/lo"
)

type QiniuTextCensor struct {
	accessKey, accessSecret string
}

func NewQiniuTextCensor(accessKey, accessSecret string) *QiniuTextCensor {
	return &QiniuTextCensor{
		accessKey:    accessKey,
		accessSecret: accessSecret,
	}
}

type QiniuTextCensorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  struct {
		Suggestion string `json:"suggestion"`
		Scenes     struct {
			Antispam struct {
				Suggestion string `json:"suggestion"`
				Details    []struct {
					Label    string  `json:"label"`
					Score    float64 `json:"score"`
					Contexts []struct {
						Context   string `json:"context"`
						Positions []struct {
							StartPos int `json:"startPos"`
							EndPos   int `json:"endPos"`
						} `json:"positions"`
					} `json:"contexts"`
				} `json:"details"`
			} `json:"antispam"`
		} `json:"scenes"`
	} `json:"result"`
}

func (r *QiniuTextCensorResponse) IsPass() bool {
	for _, detail := range r.Result.Scenes.Antispam.Details {
		label := detail.Label

		// These are absolutely not allowed to display to users.
		blockedLabels := []string{"politics", "terrorism", "porn"}
		if lo.Contains(blockedLabels, label) {
			return false
		}
	}

	// ⚠️ Right now, we allow `review` and `pass` to pass the censor.
	return r.Result.Scenes.Antispam.Suggestion != "block"
}

// Censor censors text with Qiniu API.
// https://developer.qiniu.com/censor/7260/api-text-censor
func (c *QiniuTextCensor) Censor(ctx context.Context, text string) (*TextCensorResponse, error) {
	var bodyBuffer bytes.Buffer
	if err := json.NewEncoder(&bodyBuffer).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"text": text,
		},
		"params": map[string]interface{}{
			"scenes": []string{"antispam"},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "encode request body")
	}

	request, err := http.NewRequest(http.MethodPost, "https://ai.qiniuapi.com/v3/text/censor", &bodyBuffer)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}
	request = request.WithContext(ctx)

	credentials := auth.New(c.accessKey, c.accessSecret)
	token, err := credentials.SignRequestV2(request)
	if err != nil {
		return nil, errors.Wrap(err, "sign request")
	}
	request.Header.Set("Authorization", "Qiniu "+token)

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}

	if resp.StatusCode/100 != 2 {
		return nil, errors.Errorf("unexpected error code: %d, response body: %s", resp.StatusCode, string(bodyBytes))
	}

	return QiniuTextCensorParser(bodyBytes)
}

func (*QiniuTextCensor) String() string {
	return "qiniu"
}

func QiniuTextCensorParser(raw []byte) (*TextCensorResponse, error) {
	var responseJSON QiniuTextCensorResponse
	if err := json.Unmarshal(raw, &responseJSON); err != nil {
		return nil, errors.Wrap(err, "decode response body")
	}

	var hint string
	var detailKey string
	var confidence float64
	for _, detail := range responseJSON.Result.Scenes.Antispam.Details {
		if detail.Label == "normal" {
			continue
		}

		// Get the first context as the hint, forbidden type, confidence.
		for _, context := range detail.Contexts {
			hint = context.Context
		}
		detailKey = detail.Label
		confidence = detail.Score
		break
	}

	return &TextCensorResponse{
		SourceName:    "qiniu",
		Pass:          responseJSON.IsPass(),
		ForbiddenType: formatQiniuForbiddenType(detailKey),
		Hint:          hint,
		Confidence:    confidence,
		RawResponse:   raw,
	}, nil
}

func formatQiniuForbiddenType(typ string) ForbiddenType {
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
	default:
		return ""
	}
}
