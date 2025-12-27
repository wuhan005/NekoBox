// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package conf

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit = "dev"
)

var (
	App struct {
		Production            bool   `ini:"production"`
		ExternalURL           string `ini:"external_url"`
		ICP                   string `ini:"icp"`
		UptraceDSN            string `ini:"uptrace_dsn"`
		AliyunAccessKey       string `ini:"aliyun_access_key"`
		AliyunAccessKeySecret string `ini:"aliyun_access_key_secret"`
		SentryDSN             string `ini:"sentry_dsn"`
		MaintenanceMode       bool   `ini:"maintenance_mode"`
	}

	Security struct {
		EnableTextCensor bool `ini:"enable_text_censor"`
	}

	Server struct {
		Port    int    `ini:"port"`
		Salt    string `ini:"salt"`
		XSRFKey string `ini:"xsrf_key"`
	}

	Database struct {
		DSN string

		Type     string `ini:"type"`
		User     string `ini:"user"`
		Password string `ini:"password"`
		Host     string `ini:"host"`
		Port     uint   `ini:"port"`
		Name     string `ini:"name"`
		Schema   string `ini:"schema"` // for postgres
	}

	Redis struct {
		Addr     string `ini:"addr"`
		Password string `ini:"password"`
	}

	Recaptcha struct {
		Domain         string `ini:"domain"`
		SiteKey        string `ini:"site_key"`
		ServerKey      string `ini:"server_key"`
		TurnstileStyle bool   `ini:"turnstile_style"`
	}

	Pixel struct {
		Host string `ini:"host"`
	}

	Upload struct {
		DefaultAvatarURL    string `ini:"default_avatar"`
		DefaultBackground   string `ini:"default_background"`
		AliyunEndpoint      string `ini:"aliyun_endpoint"`
		AliyunAccessID      string `ini:"aliyun_access_id"`
		AliyunAccessSecret  string `ini:"aliyun_access_secret"`
		AliyunBucket        string `ini:"aliyun_bucket"`
		AliyunBucketCDNHost string `ini:"aliyun_bucket_cdn_host"`
		ImageEndpoint       string `ini:"image_endpoint"`
		ImageAccessID       string `ini:"image_access_id"`
		ImageAccessSecret   string `ini:"image_access_secret"`
		ImageBucket         string `ini:"image_bucket"`
		ImageBucketCDNHost  string `ini:"image_bucket_cdn_host"`
	}

	Mail struct {
		Account  string `ini:"account"`
		Password string `ini:"password"`
		Port     int    `ini:"port"`
		SMTP     string `ini:"smtp"`
	}

	Service struct {
		Backends []struct {
			Prefix     string `ini:"prefix"`
			ForwardURL string `ini:"forward_url"`
		}
	}
)
