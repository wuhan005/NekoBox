package response

import (
	"github.com/wuhan005/NekoBox/internal/db"
)

type MineQuestionsItem struct {
	ID         uint   `json:"id"`
	CreatedAt  Time   `json:"createdAt"`
	Content    string `json:"content"`
	IsAnswered bool   `json:"isAnswered"`
	IsPrivate  bool   `json:"isPrivate"`
}

type MineQuestions struct {
	Total     int64                `json:"total"`
	Cursor    string               `json:"cursor"`
	Questions []*MineQuestionsItem `json:"questions"`
}

type MineSentQuestionsItem struct {
	ID           uint   `json:"id"`
	CreatedAt    Time   `json:"createdAt"`
	Content      string `json:"content"`
	IsAnswered   bool   `json:"isAnswered"`
	IsPrivate    bool   `json:"isPrivate"`
	TargetDomain string `json:"targetDomain"`
	TargetName   string `json:"targetName"`
}

type MineSentQuestions struct {
	Total     int64                    `json:"total"`
	Cursor    string                   `json:"cursor"`
	Questions []*MineSentQuestionsItem `json:"questions"`
}

type MineProfile struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type MineBoxSettings struct {
	Intro         string `json:"intro"`
	NotifyType    string `json:"notifyType"`
	AvatarURL     string `json:"avatarURL"`
	BackgroundURL string `json:"backgroundURL"`
}

type HarassmentSettings struct {
	HarassmentSettingType db.HarassmentSettingType `json:"harassmentSettingType"`
	BlockWords            string                   `json:"blockWords"`
}
