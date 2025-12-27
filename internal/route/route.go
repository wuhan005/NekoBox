// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"encoding/gob"
	"io"
	"net/http"
	"time"

	"github.com/flamego/cache"
	cacheRedis "github.com/flamego/cache/redis"
	"github.com/flamego/csrf"
	"github.com/flamego/flamego"
	"github.com/flamego/recaptcha"
	"github.com/flamego/session"
	"github.com/flamego/session/mysql"
	sessionRedis "github.com/flamego/session/redis"
	"github.com/flamego/template"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/form"
	templatepkg "github.com/wuhan005/NekoBox/internal/template"
	"github.com/wuhan005/NekoBox/route"
	"github.com/wuhan005/NekoBox/route/auth"
	"github.com/wuhan005/NekoBox/route/pixel"
	"github.com/wuhan005/NekoBox/route/question"
	"github.com/wuhan005/NekoBox/route/service"
	"github.com/wuhan005/NekoBox/route/user"
	"github.com/wuhan005/NekoBox/static"
	"github.com/wuhan005/NekoBox/templates"
)

func New() *flamego.Flame {
	f := flamego.Classic()
	if conf.App.Production {
		flamego.SetEnv(flamego.EnvTypeProd)
	}

	templateFS, err := template.EmbedFS(templates.FS, ".", []string{".html"})
	if err != nil {
		logrus.WithError(err).Fatal("Failed to embed templates file system")
	}

	// We prefer to save session into database,
	// if no database configuration, the session will be saved into memory instead.
	gob.Register(time.Time{})
	gob.Register(context.Flash{})
	var sessionStorage interface{}
	initer := session.MemoryIniter()
	if conf.Database.DSN != "" {
		initer = mysql.Initer()
		sessionStorage = mysql.Config{
			DSN:      conf.Database.DSN,
			Lifetime: 7 * 24 * time.Hour,
		}
	}
	if conf.Redis.Addr != "" {
		initer = sessionRedis.Initer()
		sessionStorage = sessionRedis.Config{
			Options: &cacheRedis.Options{
				Addr:     conf.Redis.Addr,
				Password: conf.Redis.Password,
				DB:       1,
			},
			Lifetime: 30 * 24 * time.Hour,
		}
	}
	sessioner := session.Sessioner(session.Options{
		Initer: initer,
		Config: sessionStorage,
	})

	f.Use(flamego.Static(flamego.StaticOptions{
		FileSystem: http.FS(static.FS),
		Prefix:     "/static",
	}))

	reqUserSignOut := context.Toggle(&context.ToggleOptions{UserSignOutRequired: true})
	reqUserSignIn := context.Toggle(&context.ToggleOptions{UserSignInRequired: true})

	f.Group("", func() {
		f.Get("/", route.Home)
		f.Get("/pixel", reqUserSignIn, pixel.Index)
		f.Get("/sponsor", route.Sponsor)
		f.Get("/change-logs", route.ChangeLogs)
		f.Get("/robots.txt", func(c context.Context) {
			_, _ = c.ResponseWriter().Write([]byte("User-agent: *\nDisallow: /_/"))
		})
		f.Get("/favicon.ico", func(c context.Context) {
			fs, _ := static.FS.Open("favicon.ico")
			defer func() { _ = fs.Close() }()
			c.ResponseWriter().Header().Set("Content-Type", "image/x-icon")
			_, _ = io.Copy(c.ResponseWriter(), fs)
		})

		f.Group("", func() {
			f.Combo("/register").Get(auth.Register).Post(form.Bind(form.Register{}), auth.RegisterAction)
			f.Combo("/login").Get(auth.Login).Post(form.Bind(form.Login{}), auth.LoginAction)
			f.Combo("/forgot-password").Get(auth.ForgotPassword).Post(form.Bind(form.ForgotPassword{}), auth.ForgotPasswordAction)
			f.Combo("/recover-password").Get(auth.RecoverPassword).Post(form.Bind(form.RecoverPassword{}), auth.RecoverPasswordAction)
		}, reqUserSignOut)

		f.Group("/_/{domain}", func() {
			f.Combo("").Get(question.List).Post(form.Bind(form.NewQuestion{}), question.New)
			f.Group("/{questionID}", func() {
				f.Get("", question.Item)
				f.Post("/delete", question.Delete)
				f.Post("/set-private", question.SetPrivate)
				f.Post("/set-public", question.SetPublic)
				f.Post("/answer", reqUserSignIn, form.Bind(form.PublishAnswerQuestion{}), question.PublishAnswer)
			}, question.Questioner)
		}, question.Pager)

		f.Group("/user", func() {
			f.Get("/questions", user.QuestionList)

			f.Group("/profile", func() {
				f.Get("", user.Profile)
				f.Post("/update", form.Bind(form.UpdateProfile{}), user.UpdateProfile)
				f.Post("/export", user.ExportProfile)
				f.Combo("/deactivate").Get(user.DeactivateProfile).Post(user.DeactivateProfileAction)
			})
			f.Post("/harassment/update", form.Bind(form.UpdateHarassment{}), user.UpdateHarassment)

			f.Get("/logout", auth.Logout)
		}, reqUserSignIn)

		f.Group("/api/v1", func() {
			f.Group("/user", func() {
				f.Get("", reqUserSignIn, user.ProfileAPI)

				f.Group("/{domain}", func() {
					f.Group("/questions", func() {
						f.Get("", question.ListAPI)
					})
				})
			})

			f.Any("/pixel/{**}", reqUserSignIn, pixel.Proxy)
			f.Any("/service/{**}", service.Proxy)
		}, context.APIEndpoint)
	},
		cache.Cacher(cache.Options{
			Initer: cacheRedis.Initer(),
			Config: cacheRedis.Config{
				Options: &cacheRedis.Options{
					Addr:     conf.Redis.Addr,
					Password: conf.Redis.Password,
					DB:       0,
				},
			},
		}),
		recaptcha.V3(
			recaptcha.Options{
				Secret: conf.Recaptcha.ServerKey,
				VerifyURL: func() recaptcha.VerifyURL {
					if conf.Recaptcha.TurnstileStyle {
						// FYI: https://developers.cloudflare.com/turnstile/migration/migrating-from-recaptcha/
						return "https://challenges.cloudflare.com/turnstile/v0/siteverify"
					}
					return recaptcha.VerifyURLGlobal
				}(),
			},
		),
		sessioner,
		csrf.Csrfer(csrf.Options{
			Secret: conf.Server.XSRFKey,
			Header: "X-CSRF-Token",
		}),
		template.Templater(template.Options{
			FileSystem: templateFS,
			FuncMaps:   templatepkg.FuncMap(),
		}),
		context.Contexter(),
	)
	f.NotFound(func(ctx flamego.Context) {
		ctx.Redirect("/")
	})

	return f
}
