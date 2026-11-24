package route

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/flamego/cache"
	"github.com/flamego/session"
	"github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"

	"github.com/wuhan005/NekoBox/internal/captcha"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/form"
	"github.com/wuhan005/NekoBox/internal/mail"
	"github.com/wuhan005/NekoBox/internal/response"
	"github.com/wuhan005/NekoBox/internal/settings"
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

// SignUp
// @Summary Sign up
// @Tags Authentication
// @Accept json
// @Produce json
// @Param data body form.SignUp true "Account to create"
// @Success 200 {string} string "Welcome message"
// @Failure 400 {string} string "Invalid request body, captcha failed, or the email or domain has been used"
// @Failure 403 {string} string "Already signed in"
// @Failure 500 {string} string "Internal server error"
// @ID signUp
// @Router /auth/sign-up [post]
func (*AuthHandler) SignUp(ctx context.Context, v captcha.Verifier, c cache.Cache, f form.SignUp) error {
	verifyCaptchaToken(ctx, v, c, f.Captcha)
	if ctx.ResponseWriter().Written() {
		return nil
	}

	avatar, err := settings.GetString(ctx.Request().Context(), settings.KeyDefaultAvatar)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get default avatar setting")
		return ctx.ServerError()
	}
	background, err := settings.GetString(ctx.Request().Context(), settings.KeyDefaultBackground)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get default background setting")
		return ctx.ServerError()
	}

	if err := db.Users.Create(ctx.Request().Context(), db.CreateUserOptions{
		Name:       f.Name,
		Password:   f.Password,
		Email:      f.Email,
		Avatar:     avatar,
		Domain:     f.Domain,
		Background: background,
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

// SignIn
// @Summary Sign in with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param data body form.SignIn true "Credential"
// @Success 200 {object} response.SignIn
// @Failure 400 {string} string "Invalid request body, captcha failed, or wrong email or password"
// @Failure 403 {string} string "Already signed in"
// @Failure 500 {string} string "Internal server error"
// @ID signIn
// @Router /auth/sign-in [post]
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

	context.SignInSession(sess, user)

	return ctx.Success(response.SignIn{
		Profile: &response.SignInUserProfile{
			UID:    user.UID,
			Name:   user.Name,
			Domain: user.Domain,
		},
		SessionID: sess.ID(),
	})
}

// SignOut
// @Summary Sign out the current session
// @Tags Authentication
// @Description The session ID can no longer be used. It succeeds even if the session has expired or is not signed in.
// @Produce json
// @Success 200 {string} string "Message of the signed out"
// @Security SessionToken
// @ID signOut
// @Router /auth/sign-out [post]
func (*AuthHandler) SignOut(ctx context.Context, sess session.Session) error {
	sess.Flush()
	return ctx.Success("已退出登录")
}

// ForgotPassword
// @Summary Send the password recovery email
// @Tags Authentication
// @Description The recovery code in the email expires in 24 hours, and the email can be sent once every 2 minutes.
// @Accept json
// @Produce json
// @Param data body form.ForgotPassword true "Email of the account"
// @Success 200 {string} string "Recovery request acknowledged"
// @Failure 400 {string} string "Invalid request body or captcha failed"
// @Failure 403 {string} string "Already signed in"
// @Failure 500 {string} string "Captcha verification service failed"
// @ID forgotPassword
// @Router /auth/forgot-password [post]
func (*AuthHandler) ForgotPassword(ctx context.Context, v captcha.Verifier, cache cache.Cache, f form.ForgotPassword) error {
	verifyCaptchaToken(ctx, v, cache, f.Captcha)
	if ctx.ResponseWriter().Written() {
		return nil
	}

	const message = "若该邮箱已注册，我们已向其发送了重置密码邮件，请前往查收。"
	requestCtx := ctx.Request().Context()

	user, err := db.Users.GetByEmail(requestCtx, strings.TrimSpace(f.Email))
	if err != nil {
		if !errors.Is(err, db.ErrUserNotExists) {
			logrus.WithContext(requestCtx).WithError(err).Error("Failed to get user by email")
		}
		return ctx.Success(message)
	}

	emailSentCacheKey := "forgot-password-email-sent:" + user.Email
	_, err = cache.Get(requestCtx, emailSentCacheKey)
	if err == nil {
		return ctx.Success(message)
	}
	if !errors.Is(err, os.ErrNotExist) {
		logrus.WithContext(requestCtx).WithError(err).Error("Failed to read password recovery email sent cache")
	}

	code := randstr.String(64)
	recoveryCodeCacheKey := "forgot-password-recovery-code:" + code
	if err := cache.Set(requestCtx, recoveryCodeCacheKey, user.ID, 24*time.Hour); err != nil {
		logrus.WithContext(requestCtx).WithError(err).Error("Failed to set password recovery code cache")
		return ctx.Success(message)
	}

	if err := mail.SendPasswordRecoveryMail(user.Email, code); err != nil {
		logrus.WithContext(requestCtx).WithError(err).Error("Failed to send password recovery email")
		return ctx.Success(message)
	}

	if err := cache.Set(requestCtx, emailSentCacheKey, true, 2*time.Minute); err != nil {
		logrus.WithContext(requestCtx).WithError(err).Error("Failed to set password recovery email cache")
	}

	return ctx.Success(message)
}

// GetRecoverPasswordCode
// @Summary Check a password recovery code
// @Tags Authentication
// @Produce json
// @Param code query string true "Recovery code in the email"
// @Success 200 {object} response.RecoverPassword
// @Failure 400 {string} string "The code is invalid or expired"
// @Failure 403 {string} string "Already signed in"
// @Failure 404 {string} string "The user does not exist"
// @Failure 500 {string} string "Internal server error"
// @ID getRecoverPasswordCode
// @Router /auth/recover-password [get]
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

// RecoverPassword
// @Summary Reset the password with a recovery code
// @Tags Authentication
// @Description All the existing sessions of the user are signed out.
// @Accept json
// @Produce json
// @Param data body form.RecoverPassword true "New password and recovery code"
// @Success 200 {string} string "Message of the password reset"
// @Failure 400 {string} string "Invalid request body, or the code is invalid or expired"
// @Failure 403 {string} string "Already signed in"
// @Failure 404 {string} string "The user does not exist"
// @Failure 500 {string} string "Internal server error"
// @ID recoverPassword
// @Router /auth/recover-password [post]
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
