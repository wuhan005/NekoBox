// Copyright 2025 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/wuhan005/NekoBox/internal/cmd"
)

func main() {
	app := cli.NewApp()
	app.Name = "NekoBox"
	app.Description = "Anonymous question box"

	app.Commands = []*cli.Command{
		cmd.Web,
		cmd.Censor,
		cmd.Uid,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("Failed to start application")
	}
}
