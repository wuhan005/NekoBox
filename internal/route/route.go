// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"strings"
	"time"

	"github.com/flamego/cache"
	cacheRedis "github.com/flamego/cache/redis"
	"github.com/flamego/cors"
	"github.com/flamego/flamego"
	"github.com/flamego/recaptcha"
	"github.com/flamego/session"
	"github.com/flamego/session/mysql"
	"github.com/flamego/session/postgres"
	sessionRedis "github.com/flamego/session/redis"
	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/form"
	"github.com/wuhan005/NekoBox/route/service"
)

func New(db *gorm.DB) *flamego.Flame {
	f := flamego.Classic()
	if conf.App.Production {
		flamego.SetEnv(flamego.EnvTypeProd)
	}

	// We prefer to save session into database,
	// if no database configuration, the session will be saved into memory instead.
	var sessionStorage interface{}
	initer := session.MemoryIniter()
	if conf.Database.DSN != "" {
		switch conf.Database.Type {
		case "mysql":
			initer = mysql.Initer()
			sessionStorage = mysql.Config{
				DSN:      conf.Database.DSN,
				Lifetime: 7 * 24 * time.Hour,
			}
		case "postgres":
			initer = postgres.Initer()
			sessionStorage = postgres.Config{
				DSN:      conf.Database.DSN,
				Lifetime: 7 * 24 * time.Hour,
			}
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
		ReadIDFunc: func(r *http.Request) string {
			authorizationHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			groups := strings.Fields(authorizationHeader)
			if len(groups) != 2 {
				return ""
			}
			return groups[1]
		},
		WriteIDFunc: func(_ http.ResponseWriter, _ *http.Request, _ string, _ bool) {},
	})

	f.Use(
		cors.CORS(cors.Options{
			Scheme:      "*",
			AllowDomain: []string{conf.App.CorsAllowDomain},
			Methods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		}),
		sessioner,
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
		context.Contexter(db),
	)

	reqUserSignOut := context.Toggle(&context.ToggleOptions{UserSignOutRequired: true})
	reqUserSignIn := context.Toggle(&context.ToggleOptions{UserSignInRequired: true})

	f.Group("/api", func() {
		authHandler := NewAuthHandler()
		f.Group("/auth", func() {
			f.Post("/sign-up", form.Bind(form.SignUp{}), authHandler.SignUp)
			f.Post("/sign-in", form.Bind(form.SignIn{}), authHandler.SignIn)
			f.Post("/forgot-password", form.Bind(form.ForgotPassword{}), authHandler.ForgotPassword)
			f.Combo("/recover-password").Get(authHandler.GetRecoverPasswordCode).Post(form.Bind(form.RecoverPassword{}), authHandler.RecoverPassword)
		}, reqUserSignOut)

		userHandler := NewUserHandler()
		f.Group("/users/{domain}", func() {
			f.Get("/profile", userHandler.Profile)
			f.Group("/questions", func() {
				f.Combo("").
					Get(userHandler.ListQuestions).
					Post(form.BindMultipart(form.PostQuestion{}), userHandler.PostQuestion)
				f.Get("/{questionID}", userHandler.GetQuestion)
			})
		}, userHandler.Domainer)

		mineHandler := NewMineHandler()
		f.Group("/mine", func() {
			f.Group("/questions", func() {
				f.Get("", mineHandler.ListQuestions)
				f.Group("/{questionID}", func() {
					f.Put("/answer", form.BindMultipart(form.AnswerQuestion{}), mineHandler.AnswerQuestion)
					f.Delete("", mineHandler.DeleteQuestion)
					f.Put("/visible", form.Bind(form.QuestionVisible{}), mineHandler.SetQuestionVisible)
				}, mineHandler.Questioner)
			})

			f.Group("/settings", func() {
				f.Combo("/profile").Get(mineHandler.Profile).Put(form.Bind(form.UpdateProfile{}), mineHandler.UpdateProfileSettings)
				f.Combo("/box").Get(mineHandler.BoxSettings).Put(form.BindMultipart(form.UpdateBoxSettings{}), mineHandler.UpdateBoxSettings)
				f.Combo("/harassment").Get(mineHandler.HarassmentSettings).Put(form.Bind(form.UpdateHarassmentSettings{}), mineHandler.UpdateHarassmentSettings)
				f.Post("/export-data", mineHandler.ExportData)
				f.Post("/deactivate", mineHandler.Deactivate)
			})
		}, reqUserSignIn)
	})

	f.Any("/service/{**}", service.Proxy)
	f.NotFound(func(ctx context.Context) error {
		return ctx.Error(http.StatusNotFound, "资源不存在")
	})

	return f
}
