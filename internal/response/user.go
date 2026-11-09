package response

type UserProfile struct {
	UID               string `json:"uid"`
	Name              string `json:"name"`
	Avatar            string `json:"avatar"`
	Domain            string `json:"domain"`
	Background        string `json:"background"`
	Intro             string `json:"intro"`
	HarassmentSetting string `json:"harassmentSetting" enums:"none,register_only"`
} // @name UserProfile

type PageQuestionsItem struct {
	ID        uint   `json:"id"`
	CreatedAt Time   `json:"createdAt" format:"date-time"`
	Content   string `json:"content"`
	Answer    string `json:"answer"`
} // @name PageQuestionsItem

type PageQuestions struct {
	Total int64 `json:"total"`
	// Cursor is passed to get the next page, empty if there are no questions.
	Cursor    string               `json:"cursor"`
	Questions []*PageQuestionsItem `json:"questions"`
} // @name PageQuestions

type PageQuestion struct {
	ID                uint     `json:"id"`
	IsOwner           bool     `json:"isOwner"`
	CreatedAt         Time     `json:"createdAt" format:"date-time"`
	AnsweredAt        Time     `json:"answeredAt" format:"date-time"`
	Content           string   `json:"content"`
	Answer            string   `json:"answer"`
	QuestionImageURLs []string `json:"questionImageURLs"`
	AnswerImageURLs   []string `json:"answerImageURLs"`
	HasReplyEmail     bool     `json:"hasReplyEmail"`
	IsPrivate         bool     `json:"isPrivate"`
} // @name PageQuestion
