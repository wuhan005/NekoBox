// Copyright 2023 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUsers(t *testing.T) {
	t.Parallel()

	db, cleanup := newTestDB(t)
	ctx := context.Background()

	usersStore := NewUsersStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *users)
	}{
		{"Create", testUsersCreate},
		{"GetByID", testUsersGetByID},
		{"GetByEmail", testUsersGetByEmail},
		{"GetByDomain", testUsersGetByDomain},
		{"Update", testUsersUpdate},
		{"SetName", testUsersSetName},
		{"UpdateHarassmentSetting", testUsersUpdateHarassmentSetting},
		{"Authenticate", testUsersAuthenticate},
		{"ChangePassword", testUsersChangePassword},
		{"UpdatePassword", testUsersUpdatePassword},
		{"Deactivate", testUsersDeactivate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				if err := cleanup("users"); err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, ctx, usersStore.(*users))
		})
	}
}

func testUsersCreate(t *testing.T, ctx context.Context, db *users) {
	t.Run("normal", func(t *testing.T) {
		err := db.Create(ctx, CreateUserOptions{
			Name:       "E99p1ant",
			Password:   "super_secret",
			Email:      "i@github.red",
			Avatar:     "avater.png",
			Domain:     "e99",
			Background: "background.png",
			Intro:      "Be cool, but also be warm.",
		})
		require.Nil(t, err)
	})

	t.Run("repeat email", func(t *testing.T) {
		err := db.Create(ctx, CreateUserOptions{
			Name:       "E99p1ant",
			Password:   "super_secret",
			Email:      "i@github.red",
			Avatar:     "avater.png",
			Domain:     "e99p1ant",
			Background: "background.png",
			Intro:      "Be cool, but also be warm.",
		})
		require.Equal(t, ErrDuplicateEmail, err)
	})

	t.Run("repeat domain", func(t *testing.T) {
		err := db.Create(ctx, CreateUserOptions{
			Name:       "E99p1ant",
			Password:   "super_secret",
			Email:      "e99@github.red",
			Avatar:     "avater.png",
			Domain:     "e99",
			Background: "background.png",
			Intro:      "Be cool, but also be warm.",
		})
		require.Equal(t, ErrDuplicateDomain, err)
	})
}

func testUsersGetByID(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		got, err := db.GetByID(ctx, 1)
		require.Nil(t, err)

		got.CreatedAt = time.Time{}
		got.UpdatedAt = time.Time{}

		want := &User{
			Model: gorm.Model{
				ID: 1,
			},
			UID:        got.UID,
			Name:       "E99p1ant",
			Password:   "super_secret",
			Email:      "i@github.red",
			Avatar:     "avater.png",
			Domain:     "e99",
			Background: "background.png",
			Intro:      "Be cool, but also be warm.",
			Notify:     NotifyTypeEmail,
		}
		want.EncodePassword()
		require.Equal(t, want, got)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.GetByID(ctx, 404)
		require.Equal(t, ErrUserNotExists, err)
	})
}

func testUsersGetByEmail(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		got, err := db.GetByEmail(ctx, "i@github.red")
		require.Nil(t, err)

		got.CreatedAt = time.Time{}
		got.UpdatedAt = time.Time{}

		want := &User{
			Model: gorm.Model{
				ID: 1,
			},
			UID:        got.UID,
			Name:       "E99p1ant",
			Password:   "super_secret",
			Email:      "i@github.red",
			Avatar:     "avater.png",
			Domain:     "e99",
			Background: "background.png",
			Intro:      "Be cool, but also be warm.",
			Notify:     NotifyTypeEmail,
		}
		want.EncodePassword()
		require.Equal(t, want, got)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.GetByEmail(ctx, "404")
		require.Equal(t, ErrUserNotExists, err)
	})
}

func testUsersGetByDomain(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		got, err := db.GetByDomain(ctx, "e99")
		require.Nil(t, err)

		got.CreatedAt = time.Time{}
		got.UpdatedAt = time.Time{}

		want := &User{
			Model: gorm.Model{
				ID: 1,
			},
			UID:        got.UID,
			Name:       "E99p1ant",
			Password:   "super_secret",
			Email:      "i@github.red",
			Avatar:     "avater.png",
			Domain:     "e99",
			Background: "background.png",
			Intro:      "Be cool, but also be warm.",
			Notify:     NotifyTypeEmail,
		}
		want.EncodePassword()
		require.Equal(t, want, got)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.GetByDomain(ctx, "404")
		require.Equal(t, ErrUserNotExists, err)
	})
}

func testUsersUpdate(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		err := db.Update(ctx, 1, UpdateUserOptions{
			Name:       "e99",
			Avatar:     "new_avatar.png",
			Background: "new_background.png",
			Intro:      "Be cool, but also be warm!!",
			Notify:     NotifyTypeNone,
		})
		require.Nil(t, err)

		got, err := db.GetByID(ctx, 1)
		require.Nil(t, err)

		got.CreatedAt = time.Time{}
		got.UpdatedAt = time.Time{}

		want := &User{
			Model: gorm.Model{
				ID: 1,
			},
			UID:        got.UID,
			Name:       "e99",
			Password:   "super_secret",
			Email:      "i@github.red",
			Avatar:     "new_avatar.png",
			Domain:     "e99",
			Background: "new_background.png",
			Intro:      "Be cool, but also be warm!!",
			Notify:     NotifyTypeNone,
		}
		want.EncodePassword()
		require.Equal(t, want, got)
	})
}

func testUsersSetName(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		err := db.SetName(ctx, 1, "new-name")
		require.Nil(t, err)

		got, err := db.GetByID(ctx, 1)
		require.Nil(t, err)
		require.Equal(t, "new-name", got.Name)
	})

	t.Run("empty name", func(t *testing.T) {
		err := db.SetName(ctx, 1, "   ")
		require.EqualError(t, err, "name cannot be empty")
	})
}

func testUsersUpdateHarassmentSetting(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		err := db.UpdateHarassmentSetting(ctx, 1, HarassmentSettingOptions{
			Type: HarassmentSettingNone,
		})
		require.Nil(t, err)
	})

	t.Run("unexpected harassment setting", func(t *testing.T) {
		err := db.UpdateHarassmentSetting(ctx, 1, HarassmentSettingOptions{
			Type: "not found",
		})
		require.NotNil(t, err)
	})
}

func testUsersAuthenticate(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	got, err := db.Authenticate(ctx, "i@github.red", "super_secret")
	require.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &User{
		Model: gorm.Model{
			ID: 1,
		},
		UID:        got.UID,
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
		Notify:     NotifyTypeEmail,
	}
	want.EncodePassword()
	require.Equal(t, want, got)
}

func testUsersChangePassword(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		err := db.ChangePassword(ctx, 1, "super_secret", "new_password")
		require.Nil(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		err := db.ChangePassword(ctx, 1, "wrong_password", "new_password")
		require.Equal(t, ErrBadCredential, err)
	})
}

func testUsersUpdatePassword(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		err := db.UpdatePassword(ctx, 1, "new_password")
		require.Nil(t, err)
	})
}

func testUsersDeactivate(t *testing.T, ctx context.Context, db *users) {
	err := db.Create(ctx, CreateUserOptions{
		Name:       "E99p1ant",
		Password:   "super_secret",
		Email:      "i@github.red",
		Avatar:     "avater.png",
		Domain:     "e99",
		Background: "background.png",
		Intro:      "Be cool, but also be warm.",
	})
	require.Nil(t, err)

	t.Run("normal", func(t *testing.T) {
		err := db.Deactivate(ctx, 1)
		require.Nil(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := db.Deactivate(ctx, 404)
		require.NotNil(t, err)
	})
}
