// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var SystemSettings SystemSettingsStore

var _ SystemSettingsStore = (*systemSettings)(nil)

type SystemSettingsStore interface {
	// Get returns the system setting with the given key, or ErrSystemSettingNotExist if it does not exist.
	Get(ctx context.Context, key string) (*SystemSetting, error)
	// List returns all the system settings ordered by key.
	List(ctx context.Context) ([]*SystemSetting, error)
	// Set creates or replaces the system setting with the given key; changing the type of an existing setting returns ErrSystemSettingTypeMismatch.
	Set(ctx context.Context, opts SetSystemSettingOptions) error
	// Delete deletes the system setting with the given key, or returns ErrSystemSettingNotExist if it does not exist.
	Delete(ctx context.Context, key string) error
	// CreateIfNotExist creates the given system settings whose keys do not exist yet, leaving existing ones untouched.
	CreateIfNotExist(ctx context.Context, opts []SetSystemSettingOptions) error
}

func NewSystemSettingsStore(db *gorm.DB) SystemSettingsStore {
	return &systemSettings{db}
}

type systemSettings struct {
	*gorm.DB
}

// SystemSetting is a dynamic system setting entry whose value is stored as text in the canonical form of its type.
// Soft delete is intentionally not used so that a deleted key can be created again under the unique index.
type SystemSetting struct {
	ID          uint              `gorm:"primaryKey" json:"-"`
	Key         string            `gorm:"size:128;not null;uniqueIndex" json:"key"`
	Type        SystemSettingType `gorm:"size:16;not null" json:"type"`
	Value       string            `gorm:"type:text;not null" json:"value"`
	Description string            `gorm:"size:255;not null;default:''" json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type SystemSettingType string

const (
	SystemSettingTypeString SystemSettingType = "string"
	SystemSettingTypeInt    SystemSettingType = "int"
	SystemSettingTypeFloat  SystemSettingType = "float"
	SystemSettingTypeBool   SystemSettingType = "bool"
	SystemSettingTypeJSON   SystemSettingType = "json"
)

const maxSystemSettingKeyLength = 128

// systemSettingKeyRegexp matches lowercase keys namespaced by dots, e.g. "site.announcement" or "upload.max_image_size".
var systemSettingKeyRegexp = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`)

// systemSettingKeyColumn must be used instead of raw SQL strings to reference the key column, since "key" is a reserved word in MySQL.
var systemSettingKeyColumn = clause.Column{Name: "key"}

var (
	ErrSystemSettingNotExist     = errors.New("system setting does not exist")
	ErrSystemSettingTypeMismatch = errors.New("system setting type mismatch")
	ErrInvalidSystemSettingKey   = errors.New("invalid system setting key")
	ErrInvalidSystemSettingType  = errors.New("invalid system setting type")
	ErrInvalidSystemSettingValue = errors.New("invalid system setting value")
)

// IsValid reports whether the type is one of the supported system setting types.
func (t SystemSettingType) IsValid() bool {
	switch t {
	case SystemSettingTypeString, SystemSettingTypeInt, SystemSettingTypeFloat, SystemSettingTypeBool, SystemSettingTypeJSON:
		return true
	default:
		return false
	}
}

// Normalize validates the value against the type and returns its canonical form.
func (t SystemSettingType) Normalize(value string) (string, error) {
	switch t {
	case SystemSettingTypeString:
		return value, nil

	case SystemSettingTypeInt:
		v, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return "", errors.Wrapf(ErrInvalidSystemSettingValue, "parse %q as int", value)
		}
		return strconv.FormatInt(v, 10), nil

	case SystemSettingTypeFloat:
		v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return "", errors.Wrapf(ErrInvalidSystemSettingValue, "parse %q as finite float", value)
		}
		return strconv.FormatFloat(v, 'g', -1, 64), nil

	case SystemSettingTypeBool:
		v, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return "", errors.Wrapf(ErrInvalidSystemSettingValue, "parse %q as bool", value)
		}
		return strconv.FormatBool(v), nil

	case SystemSettingTypeJSON:
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(value)); err != nil {
			return "", errors.Wrapf(ErrInvalidSystemSettingValue, "parse %q as JSON", value)
		}
		return buf.String(), nil

	default:
		return "", errors.Wrapf(ErrInvalidSystemSettingType, "%q", t)
	}
}

func (c *SystemSetting) expectType(want SystemSettingType) error {
	if c.Type != want {
		return errors.Wrapf(ErrSystemSettingTypeMismatch, "key %q has type %q, want %q", c.Key, c.Type, want)
	}
	return nil
}

// AsString returns the value of a string setting.
func (c *SystemSetting) AsString() (string, error) {
	if err := c.expectType(SystemSettingTypeString); err != nil {
		return "", err
	}
	return c.Value, nil
}

// AsInt returns the value of an int setting.
func (c *SystemSetting) AsInt() (int64, error) {
	if err := c.expectType(SystemSettingTypeInt); err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(ErrInvalidSystemSettingValue, "parse value of key %q as int", c.Key)
	}
	return v, nil
}

// AsFloat returns the value of a float setting.
func (c *SystemSetting) AsFloat() (float64, error) {
	if err := c.expectType(SystemSettingTypeFloat); err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(c.Value, 64)
	if err != nil {
		return 0, errors.Wrapf(ErrInvalidSystemSettingValue, "parse value of key %q as float", c.Key)
	}
	return v, nil
}

// AsBool returns the value of a bool setting.
func (c *SystemSetting) AsBool() (bool, error) {
	if err := c.expectType(SystemSettingTypeBool); err != nil {
		return false, err
	}
	v, err := strconv.ParseBool(c.Value)
	if err != nil {
		return false, errors.Wrapf(ErrInvalidSystemSettingValue, "parse value of key %q as bool", c.Key)
	}
	return v, nil
}

// AsJSON unmarshals the value of a JSON setting into v.
func (c *SystemSetting) AsJSON(v interface{}) error {
	if err := c.expectType(SystemSettingTypeJSON); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(c.Value), v); err != nil {
		return errors.Wrapf(err, "unmarshal value of key %q", c.Key)
	}
	return nil
}

func validateSystemSettingKey(key string) error {
	if len(key) == 0 || len(key) > maxSystemSettingKeyLength || !systemSettingKeyRegexp.MatchString(key) {
		return errors.Wrapf(ErrInvalidSystemSettingKey, "%q", key)
	}
	return nil
}

func whereSystemSettingKey(key string) clause.Expression {
	return clause.Eq{Column: systemSettingKeyColumn, Value: key}
}

func (db *systemSettings) Get(ctx context.Context, key string) (*SystemSetting, error) {
	if err := validateSystemSettingKey(key); err != nil {
		return nil, err
	}

	var setting SystemSetting
	if err := db.WithContext(ctx).Where(whereSystemSettingKey(key)).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSystemSettingNotExist
		}
		return nil, errors.Wrap(err, "get system setting")
	}
	return &setting, nil
}

func (db *systemSettings) List(ctx context.Context) ([]*SystemSetting, error) {
	settings := make([]*SystemSetting, 0)
	if err := db.WithContext(ctx).
		Order(clause.OrderByColumn{Column: systemSettingKeyColumn}).
		Find(&settings).Error; err != nil {
		return nil, errors.Wrap(err, "list system settings")
	}
	return settings, nil
}

type SetSystemSettingOptions struct {
	Key         string
	Type        SystemSettingType
	Value       string
	Description string
}

func (db *systemSettings) Set(ctx context.Context, opts SetSystemSettingOptions) error {
	if err := validateSystemSettingKey(opts.Key); err != nil {
		return err
	}
	value, err := opts.Type.Normalize(opts.Value)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var setting SystemSetting
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereSystemSettingKey(opts.Key)).First(&setting).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&SystemSetting{
				Key:         opts.Key,
				Type:        opts.Type,
				Value:       value,
				Description: opts.Description,
			}).Error; err != nil {
				return errors.Wrap(err, "create system setting")
			}
			return nil
		} else if err != nil {
			return errors.Wrap(err, "get system setting")
		}

		if err := setting.expectType(opts.Type); err != nil {
			return err
		}

		if err := tx.Model(&setting).Updates(map[string]interface{}{
			"value":       value,
			"description": opts.Description,
		}).Error; err != nil {
			return errors.Wrap(err, "update system setting")
		}
		return nil
	})
}

func (db *systemSettings) Delete(ctx context.Context, key string) error {
	if err := validateSystemSettingKey(key); err != nil {
		return err
	}

	result := db.WithContext(ctx).Where(whereSystemSettingKey(key)).Delete(&SystemSetting{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "delete system setting")
	}
	if result.RowsAffected == 0 {
		return ErrSystemSettingNotExist
	}
	return nil
}

func (db *systemSettings) CreateIfNotExist(ctx context.Context, opts []SetSystemSettingOptions) error {
	settings, err := buildSystemSettings(opts)
	if err != nil {
		return err
	}
	if len(settings) == 0 {
		return nil
	}

	if err := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{systemSettingKeyColumn},
		DoNothing: true,
	}).Create(&settings).Error; err != nil {
		return errors.Wrap(err, "create system settings")
	}
	return nil
}

// buildSystemSettings validates and normalizes the options, rejecting duplicate keys.
func buildSystemSettings(opts []SetSystemSettingOptions) ([]*SystemSetting, error) {
	settings := make([]*SystemSetting, 0, len(opts))
	keys := make(map[string]struct{}, len(opts))
	for _, opt := range opts {
		if err := validateSystemSettingKey(opt.Key); err != nil {
			return nil, err
		}
		if _, ok := keys[opt.Key]; ok {
			return nil, errors.Wrapf(ErrInvalidSystemSettingKey, "duplicate key %q", opt.Key)
		}
		keys[opt.Key] = struct{}{}

		value, err := opt.Type.Normalize(opt.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "key %q", opt.Key)
		}
		settings = append(settings, &SystemSetting{
			Key:         opt.Key,
			Type:        opt.Type,
			Value:       value,
			Description: opt.Description,
		})
	}
	return settings, nil
}
