// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/session"

	"github.com/wuhan005/NekoBox/internal/db"
)

const SessionKeyUserID = "nekobox:auth:user-id"

// authenticatedUser returns the user object of the authenticated user.
func authenticatedUser(ctx flamego.Context, sess session.Session) *db.User {
	uid, ok := sess.Get(SessionKeyUserID).(uint)
	if !ok {
		return nil
	}

	user, _ := db.Users.GetByID(ctx.Request().Context(), uid)
	return user
}

type ToggleOptions struct {
	UserSignInRequired  bool
	UserSignOutRequired bool
}

func Toggle(options *ToggleOptions) flamego.Handler {
	return func(ctx Context) error {
		if options.UserSignOutRequired && ctx.IsSignedIn {
			return ctx.Error(http.StatusForbidden, "请先登出账号")
		}

		if options.UserSignInRequired && !ctx.IsSignedIn {
			return ctx.Error(http.StatusUnauthorized, "请先登录账号")
		}
		return nil
	}
}
