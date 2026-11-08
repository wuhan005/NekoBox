// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --parseDependency --parseInternal --parseDepth 1 --requiredByDefault --outputTypes json,yaml -g internal/route/route.go --output docs --exclude web,e2e

package nekobox
