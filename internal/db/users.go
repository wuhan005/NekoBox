// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/wuhan005/gadget"
	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/conf"
)

var Users UsersStore

var _ UsersStore = (*users)(nil)

type UsersStore interface {
	Create(ctx context.Context, opts CreateUserOptions) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByDomain(ctx context.Context, domain string) (*User, error)
	Update(ctx context.Context, id uint, opts UpdateUserOptions) error
	SetName(ctx context.Context, id uint, name string) error
	UpdateHarassmentSetting(ctx context.Context, id uint, options HarassmentSettingOptions) error
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, id uint, oldPassword, newPassword string) error
	UpdatePassword(ctx context.Context, id uint, newPassword string) error
	Deactivate(ctx context.Context, id uint) error
}

func NewUsersStore(db *gorm.DB) UsersStore {
	return &users{db}
}

type users struct {
	*gorm.DB
}

type User struct {
	gorm.Model        `json:"-"`
	UID               string                `json:"-"`
	Name              string                `json:"name"`
	Password          string                `json:"-"`
	Email             string                `json:"email"`
	Avatar            string                `json:"avatar"`
	Domain            string                `json:"domain"`
	Background        string                `json:"background"`
	Intro             string                `json:"intro"`
	Notify            NotifyType            `json:"notify"`
	HarassmentSetting HarassmentSettingType `json:"harassment_setting"`
	BlockWords        string                `json:"-"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	u.UID = xid.New().String()
	return nil
}

type NotifyType string

const (
	NotifyTypeEmail NotifyType = "email"
	NotifyTypeNone  NotifyType = "none"
)

type HarassmentSettingType string

const (
	HarassmentSettingNone             HarassmentSettingType = "none"
	HarassmentSettingTypeRegisterOnly HarassmentSettingType = "register_only"
)

func (u *User) EncodePassword() {
	u.Password = gadget.HmacSha1(u.Password, conf.Server.Salt)
}

func (u *User) Authenticate(password string) bool {
	password = gadget.HmacSha1(password, conf.Server.Salt)
	return u.Password == password
}

type CreateUserOptions struct {
	Name       string
	Password   string
	Email      string
	Avatar     string
	Domain     string
	Background string
	Intro      string
}

var (
	ErrUserNotExists   = errors.New("账号不存在")
	ErrBadCredential   = errors.New("邮箱或密码错误")
	ErrDuplicateEmail  = errors.New("这个邮箱已经注册过账号了！")
	ErrDuplicateDomain = errors.New("个性域名重复了，换一个吧~")
)

func (db *users) Create(ctx context.Context, opts CreateUserOptions) error {
	if err := db.validate(ctx, opts); err != nil {
		return err
	}

	newUser := &User{
		Name:       opts.Name,
		Password:   opts.Password,
		Email:      opts.Email,
		Avatar:     opts.Avatar,
		Domain:     opts.Domain,
		Background: opts.Background,
		Intro:      opts.Intro,
		Notify:     NotifyTypeEmail,
	}
	newUser.EncodePassword()

	if err := db.WithContext(ctx).Create(newUser).Error; err != nil {
		return errors.Wrap(err, "create user")
	}
	return nil
}

func (db *users) getBy(ctx context.Context, where string, args ...interface{}) (*User, error) {
	var user User
	if err := db.WithContext(ctx).Where(where, args...).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotExists
		}
		return nil, errors.Wrap(err, "get user")
	}
	return &user, nil
}

func (db *users) GetByID(ctx context.Context, id uint) (*User, error) {
	return db.getBy(ctx, "id = ?", id)
}

func (db *users) GetByEmail(ctx context.Context, email string) (*User, error) {
	return db.getBy(ctx, "email = ?", email)
}

func (db *users) GetByDomain(ctx context.Context, domain string) (*User, error) {
	return db.getBy(ctx, "domain = ?", domain)
}

type UpdateUserOptions struct {
	Name       string
	Avatar     string
	Background string
	Intro      string
	Notify     NotifyType
}

func (db *users) Update(ctx context.Context, id uint, opts UpdateUserOptions) error {
	_, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get user by id")
	}

	switch opts.Notify {
	case NotifyTypeEmail, NotifyTypeNone:
	default:
		return errors.Errorf("unexpected notify type: %q", opts.Notify)
	}

	if err := db.WithContext(ctx).Where("id = ?", id).Updates(&User{
		Name:       opts.Name,
		Avatar:     opts.Avatar,
		Background: opts.Background,
		Intro:      opts.Intro,
		Notify:     opts.Notify,
	}).Error; err != nil {
		return errors.Wrap(err, "update user")
	}
	return nil
}

func (db *users) SetName(ctx context.Context, id uint, name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name cannot be empty")
	}

	if err := db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("name", name).Error; err != nil {
		return errors.Wrap(err, "update user name")
	}
	return nil
}

type HarassmentSettingOptions struct {
	Type       HarassmentSettingType
	BlockWords string
}

func (db *users) UpdateHarassmentSetting(ctx context.Context, id uint, options HarassmentSettingOptions) error {
	typ := options.Type

	switch typ {
	case HarassmentSettingNone, HarassmentSettingTypeRegisterOnly:
	default:
		return errors.Errorf("unexpected harassment setting type: %q", typ)
	}

	if err := db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"HarassmentSetting": typ,
		"BlockWords":        options.BlockWords,
	}).Error; err != nil {
		return errors.Wrap(err, "update user")
	}
	return nil
}

func (db *users) Authenticate(ctx context.Context, email, password string) (*User, error) {
	u, err := db.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrBadCredential
	}

	if !u.Authenticate(password) {
		return nil, ErrBadCredential
	}

	return u, nil
}

func (db *users) ChangePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	u, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get user by id")
	}

	if !u.Authenticate(oldPassword) {
		return ErrBadCredential
	}

	u.Password = newPassword
	u.EncodePassword()

	if err := db.WithContext(ctx).Model(&User{}).Where("id = ?", u.ID).Update("password", u.Password).Error; err != nil {
		return errors.Wrap(err, "change password")
	}
	return nil
}

func (db *users) UpdatePassword(ctx context.Context, id uint, newPassword string) error {
	u, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get user by id")
	}

	u.Password = newPassword
	u.EncodePassword()

	if err := db.WithContext(ctx).Model(&User{}).Where("id = ?", u.ID).Update("password", u.Password).Error; err != nil {
		return errors.Wrap(err, "change password")
	}
	return nil
}

func (db *users) Deactivate(ctx context.Context, id uint) error {
	u, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get user by id")
	}

	if err := db.WithContext(ctx).Model(&User{}).Delete("id = ?", u.ID).Error; err != nil {
		return errors.Wrap(err, "delete user")
	}
	return nil
}

func (db *users) validate(ctx context.Context, opts CreateUserOptions) error {
	if err := db.WithContext(ctx).Model(&User{}).Where("email = ?", opts.Email).First(&User{}).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return errors.Wrap(err, "validate email")
		}
	} else {
		return ErrDuplicateEmail
	}

	if err := db.WithContext(ctx).Model(&User{}).Where("domain = ?", opts.Domain).First(&User{}).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return errors.Wrap(err, "validate name")
		}
	} else {
		return ErrDuplicateDomain
	}

	return nil
}
