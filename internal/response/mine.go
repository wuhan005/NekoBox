package response

import (
	"github.com/wuhan005/NekoBox/internal/db"
)

type MineQuestionsItem struct {
	ID         uint   `json:"id"`
	CreatedAt  Time   `json:"createdAt" format:"date-time"`
	Content    string `json:"content"`
	IsAnswered bool   `json:"isAnswered"`
	IsPrivate  bool   `json:"isPrivate"`
} // @name MineQuestionsItem

type MineQuestions struct {
	Total int64 `json:"total"`
	// Cursor is passed to get the next page, empty if there are no questions.
	Cursor    string               `json:"cursor"`
	Questions []*MineQuestionsItem `json:"questions"`
} // @name MineQuestions

type MineSentQuestionsItem struct {
	ID         uint   `json:"id"`
	CreatedAt  Time   `json:"createdAt" format:"date-time"`
	Content    string `json:"content"`
	IsAnswered bool   `json:"isAnswered"`
	IsPrivate  bool   `json:"isPrivate"`
	// TargetDomain and TargetName are empty if the receiver no longer exists.
	TargetDomain string `json:"targetDomain"`
	TargetName   string `json:"targetName"`
} // @name MineSentQuestionsItem

type MineSentQuestions struct {
	Total int64 `json:"total"`
	// Cursor is passed to get the next page, empty if there are no questions.
	Cursor    string                   `json:"cursor"`
	Questions []*MineSentQuestionsItem `json:"questions"`
} // @name MineSentQuestions

type MineProfile struct {
	Email string `json:"email"`
	Name  string `json:"name"`
} // @name MineProfile

type MineBoxSettings struct {
	Intro         string `json:"intro"`
	NotifyType    string `json:"notifyType" enums:"none,email"`
	AvatarURL     string `json:"avatarURL"`
	BackgroundURL string `json:"backgroundURL"`
} // @name MineBoxSettings

type HarassmentSettings struct {
	HarassmentSettingType db.HarassmentSettingType `json:"harassmentSettingType" swaggertype:"string" enums:"none,register_only"`
	BlockWords            string                   `json:"blockWords"`
} // @name HarassmentSettings
