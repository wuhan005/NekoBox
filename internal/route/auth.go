package route

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/session"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"

	"github.com/wuhan005/NekoBox/internal/captcha"
	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/form"
	"github.com/wuhan005/NekoBox/internal/mail"
	"github.com/wuhan005/NekoBox/internal/response"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// verifyCaptchaToken validates the captcha token; on failure it writes the error response
// directly via ctx and returns. Callers MUST check ctx.ResponseWriter().Written() afterwards
// and bail out before running any business logic, because ctx.Error itself returns nil.
func verifyCaptchaToken(ctx context.Context, v captcha.Verifier, c cache.Cache, token string) {
	err := v.Verify(ctx.Request().Context(), c, token, ctx.IP())
	if err == nil {
		return
	}
	if errors.Is(err, captcha.ErrVerifyFailed) {
		_ = ctx.Error(http.StatusBadRequest, "验证码校验失败，请重试")
		return
	}
	logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to verify captcha")
	_ = ctx.Error(http.StatusInternalServerError, "验证码请求失败，请稍后再试")
}

func (*AuthHandler) SignUp(ctx context.Context, v captcha.Verifier, c cache.Cache, f form.SignUp) error {
	verifyCaptchaToken(ctx, v, c, f.Captcha)
	if ctx.ResponseWriter().Written() {
		return nil
	}

	if err := db.Users.Create(ctx.Request().Context(), db.CreateUserOptions{
		Name:       f.Name,
		Password:   f.Password,
		Email:      f.Email,
		Avatar:     conf.Upload.DefaultAvatarURL,
		Domain:     f.Domain,
		Background: conf.Upload.DefaultBackground,
		Intro:      "问你想问的",
	}); err != nil {
		switch {
		case errors.Is(err, db.ErrDuplicateEmail),
			errors.Is(err, db.ErrDuplicateDomain):
			return ctx.Error(http.StatusBadRequest, "%s", errors.Cause(err).Error())

		default:
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create new user")
			return ctx.ServerError()
		}
	}

	return ctx.Success("注册成功，欢迎来到 NekoBox！")
}

func (*AuthHandler) SignIn(ctx context.Context, sess session.Session, v captcha.Verifier, c cache.Cache, f form.SignIn) error {
	verifyCaptchaToken(ctx, v, c, f.Captcha)
	if ctx.ResponseWriter().Written() {
		return nil
	}

	user, err := db.Users.Authenticate(ctx.Request().Context(), f.Email, f.Password)
	if err != nil {
		if errors.Is(err, db.ErrBadCredential) {
			return ctx.Error(http.StatusBadRequest, "电子邮箱或密码错误")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to authenticate user")
		return ctx.ServerError()
	}

	sess.Set(context.SessionKeyUserID, user.ID)

	return ctx.Success(response.SignIn{
		Profile: &response.SignInUserProfile{
			UID:    user.UID,
			Name:   user.Name,
			Domain: user.Domain,
		},
		SessionID: sess.ID(),
	})
}

func (*AuthHandler) ForgotPassword(ctx context.Context, v captcha.Verifier, cache cache.Cache, f form.ForgotPassword) error {
	verifyCaptchaToken(ctx, v, cache, f.Captcha)
	if ctx.ResponseWriter().Written() {
		return nil
	}

	email := strings.TrimSpace(f.Email)

	user, err := db.Users.GetByEmail(ctx.Request().Context(), email)
	if err != nil {
		if errors.Is(err, db.ErrUserNotExists) {
			return ctx.Error(http.StatusNotFound, "用户邮箱不存在")
		} else {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get user by email")
			return ctx.ServerError()
		}
	}

	emailSentCacheKey := "forgot-password-email-sent:" + user.Email
	_, err = cache.Get(ctx.Request().Context(), emailSentCacheKey)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to read password recovery email sent cache")
		}
	} else {
		return ctx.Error(http.StatusTooManyRequests, "邮件发送太频繁，请稍后再试")
	}

	code := randstr.String(64)
	recoveryCodeCacheKey := "forgot-password-recovery-code:" + code
	if err := cache.Set(ctx.Request().Context(), recoveryCodeCacheKey, user.ID, 24*time.Hour); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to set password recovery code cache")
		return ctx.ServerError()
	}

	if err := mail.SendPasswordRecoveryMail(user.Email, code); err != nil {
		return ctx.Error(http.StatusInternalServerError, "邮件发送失败，请稍后再试")
	}

	if err := cache.Set(ctx.Request().Context(), emailSentCacheKey, time.Now(), 2*time.Minute); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to set password recovery email cache")
	}

	return ctx.Success(fmt.Sprintf("邮件已发送至 %s，请查收。", user.Email))
}

func (*AuthHandler) GetRecoverPasswordCode(ctx context.Context, cache cache.Cache) error {
	code := ctx.Query("code")
	user, ok := checkRecoverPasswordCode(ctx, code, cache)
	if !ok {
		return nil
	}

	return ctx.Success(response.RecoverPassword{
		Name: user.Name,
	})
}

func (*AuthHandler) RecoverPassword(ctx context.Context, cache cache.Cache, f form.RecoverPassword) error {
	code := f.Code

	user, ok := checkRecoverPasswordCode(ctx, code, cache)
	if !ok {
		return nil
	}

	recoveryCodeCacheKey := "forgot-password-recovery-code:" + code
	if err := cache.Delete(ctx.Request().Context(), recoveryCodeCacheKey); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to delete password recovery code cache")
		return ctx.ServerError()
	}

	if err := db.Users.UpdatePassword(ctx.Request().Context(), user.ID, f.NewPassword); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update user password")
		return ctx.ServerError()
	}

	return ctx.Success("密码重置成功，请使用新密码登录。")
}

func checkRecoverPasswordCode(ctx context.Context, code string, cache cache.Cache) (*db.User, bool) {
	recoveryCodeCacheKey := "forgot-password-recovery-code:" + code
	userIDItf, err := cache.Get(ctx.Request().Context(), recoveryCodeCacheKey)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = ctx.Error(http.StatusBadRequest, "邮件已过期，请重新发送")
		} else {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to read password recovery code cache")
			_ = ctx.ServerError()
		}
		return nil, false
	}

	userID, ok := userIDItf.(uint)
	if !ok {
		logrus.WithContext(ctx.Request().Context()).WithField("user_id_itf", userIDItf).Error("Failed to convert user id interface to uint")
		_ = ctx.ServerError()
		return nil, false
	}

	user, err := db.Users.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		_ = ctx.Error(http.StatusNotFound, "用户不存在")
		return nil, false
	}

	return user, true
}
