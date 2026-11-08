package form

import (
	"mime/multipart"
)

type AnswerQuestion struct {
	Answer string                  `form:"answer" valid:"required;maxlen:1000" label:"回答内容"`
	Images []*multipart.FileHeader `form:"images[]" label:"图片"`
}

type QuestionVisible struct {
	Visible bool `json:"visible"`
} // @name QuestionVisible

type UpdateProfile struct {
	Name string `json:"name" valid:"required;maxlen:20" label:"昵称"`
	// OldPassword is required only when changing the password.
	OldPassword string `json:"oldPassword,omitempty" label:"旧密码"`
	// NewPassword changes the password if it is not empty.
	NewPassword string `json:"newPassword,omitempty" valid:"maxlen:30" label:"新密码"`
} // @name UpdateProfile

type UpdateBoxSettings struct {
	Intro      string                `form:"intro" valid:"required;maxlen:100" label:"提问箱介绍"`
	NotifyType string                `form:"notifyType" valid:"required" label:"通知类型"`
	Avatar     *multipart.FileHeader `form:"avatar" label:"提问箱头像"`
	Background *multipart.FileHeader `form:"background" label:"提问箱背景"`
}

type UpdateHarassmentSettings struct {
	// HarassmentSettingType falls back to none if it is unknown.
	HarassmentSettingType string `json:"harassmentSettingType" enums:"none,register_only"`
	// BlockWords are at most 10 words separated by commas, each of them is at most 10 bytes.
	BlockWords string `json:"blockWords"`
} // @name UpdateHarassmentSettings
