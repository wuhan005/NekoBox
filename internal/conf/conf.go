// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package conf

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

// File is the configuration object.
var File *ini.File

func Init() error {
	configFile := os.Getenv("NEKOBOX_CONFIG_PATH")
	if configFile == "" {
		configFile = "conf/app.ini"
	}

	var err error
	File, err = ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, configFile)
	if err != nil {
		return errors.Wrapf(err, "parse %q", configFile)
	}

	if err := File.Section("app").MapTo(&App); err != nil {
		return errors.Wrap(err, "map 'server'")
	}

	if App.ExternalURL == "" {
		return errors.New("app.external_url is required")
	}
	App.ExternalURL = strings.TrimRight(App.ExternalURL, "/")

	if err := File.Section("security").MapTo(&Security); err != nil {
		return errors.Wrap(err, "map 'security'")
	}

	if err := File.Section("server").MapTo(&Server); err != nil {
		return errors.Wrap(err, "map 'server'")
	}

	if err := File.Section("database").MapTo(&Database); err != nil {
		return errors.Wrap(err, "map 'database'")
	}

	if err := File.Section("redis").MapTo(&Redis); err != nil {
		return errors.Wrap(err, "map 'redis'")
	}

	if err := File.Section("recaptcha").MapTo(&Recaptcha); err != nil {
		return errors.Wrap(err, "map 'recaptcha'")
	}

	if err := File.Section("pixel").MapTo(&Pixel); err != nil {
		return errors.Wrap(err, "map 'pixel'")
	}

	if err := File.Section("upload").MapTo(&Upload); err != nil {
		return errors.Wrap(err, "map 'upload'")
	}

	if err := File.Section("mail").MapTo(&Mail); err != nil {
		return errors.Wrap(err, "map 'mail'")
	}

	serviceSections := File.Section("service").ChildSections()
	for _, serviceSection := range serviceSections {
		serviceSection := serviceSection
		var backend struct {
			Prefix     string `ini:"prefix"`
			ForwardURL string `ini:"forward_url"`
		}
		if err := serviceSection.MapTo(&backend); err != nil {
			return errors.Wrapf(err, "map 'service.%s'", serviceSection.Name())
		}
		Service.Backends = append(Service.Backends, backend)
	}

	return nil
}
