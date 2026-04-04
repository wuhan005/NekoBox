package route

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flamego/session"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"github.com/wuhan005/NekoBox/internal/conf"
	"github.com/wuhan005/NekoBox/internal/context"
	"github.com/wuhan005/NekoBox/internal/db"
	"github.com/wuhan005/NekoBox/internal/dbutil"
	"github.com/wuhan005/NekoBox/internal/form"
	"github.com/wuhan005/NekoBox/internal/mail"
	"github.com/wuhan005/NekoBox/internal/response"
	"github.com/wuhan005/NekoBox/internal/security/censor"
)

type MineHandler struct{}

func NewMineHandler() *MineHandler {
	return &MineHandler{}
}

func (*MineHandler) ListQuestions(ctx context.Context) error {
	pageSize := ctx.QueryInt("pageSize")
	cursorValue := ctx.Query("cursor")

	total, err := db.Questions.Count(ctx.Request().Context(), ctx.User.ID, db.GetQuestionsCountOptions{
		FilterAnswered: false,
		ShowPrivate:    true,
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get questions count")
		return ctx.ServerError()
	}

	questions, err := db.Questions.GetByUserID(ctx.Request().Context(), ctx.User.ID, db.GetQuestionsByUserIDOptions{
		Cursor: &dbutil.Cursor{
			Value:    cursorValue,
			PageSize: pageSize,
		},
		FilterAnswered: false,
		ShowPrivate:    true,
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get questions by user ID")
		return ctx.ServerError()
	}

	respQuestions := lo.Map(questions, func(question *db.Question, _ int) *response.MineQuestionsItem {
		return &response.MineQuestionsItem{
			ID:         question.ID,
			CreatedAt:  question.CreatedAt,
			Content:    question.Content,
			IsAnswered: question.Answer != "",
			IsPrivate:  question.IsPrivate,
		}
	})

	var cursor string
	if len(questions) > 0 {
		cursor = strconv.Itoa(int(questions[len(questions)-1].ID))
	}

	return ctx.Success(&response.MineQuestions{
		Total:     total,
		Cursor:    cursor,
		Questions: respQuestions,
	})
}

func (*MineHandler) Questioner(ctx context.Context) error {
	questionID := uint(ctx.ParamInt("questionID"))
	question, err := db.Questions.GetByID(ctx.Request().Context(), questionID)
	if err != nil {
		if errors.Is(err, db.ErrQuestionNotExist) {
			return ctx.Error(http.StatusNotFound, "提问不存在")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get question")
		return ctx.ServerError()
	}

	if question.UserID != ctx.User.ID {
		return ctx.Error(http.StatusNotFound, "提问不存在")
	}

	ctx.Map(question)
	return nil
}

func (*MineHandler) AnswerQuestion(ctx context.Context, question *db.Question, tx dbutil.Transactor, f form.AnswerQuestion) error {
	answer := f.Answer

	// 🚨 Content security check.
	censorResponse, err := censor.Text(ctx.Request().Context(), answer)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to censor text")
	}
	if err == nil && !censorResponse.Pass {
		errorMessage := censorResponse.ErrorMessage()
		return ctx.Error(http.StatusBadRequest, "%s", errorMessage)
	}

	// Upload image if exists.
	var uploadImage *db.UploadImage
	if len(f.Images) > 0 {
		image := f.Images[0]
		uploadImage, err = uploadImageFile(ctx, uploadImageFileOptions{
			Image:          image,
			UploaderUserID: ctx.User.ID,
		})
		if err != nil {
			if errors.Is(err, ErrUploadImageSizeTooLarge) {
				return ctx.Error(http.StatusBadRequest, "图片文件大小不能大于 5Mb")
			} else {
				logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to upload image")
				return ctx.Error(http.StatusInternalServerError, "上传图片失败，请重试")
			}
		}
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		questionsStore := db.NewQuestionsStore(tx)
		if err := questionsStore.AnswerByID(ctx.Request().Context(), question.ID, f.Answer); err != nil {
			return errors.Wrap(err, "answer by ID")
		}

		// Update censor result.
		if censorResponse != nil {
			if err := questionsStore.UpdateCensor(ctx.Request().Context(), question.ID, db.UpdateQuestionCensorOptions{
				AnswerCensorMetadata: censorResponse.ToJSON(),
			}); err != nil {
				logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update answer censor result")
				return errors.Wrap(err, "update question censor")
			}
		}

		if uploadImage != nil {
			// Bind the uploaded image with the question.
			if err := db.NewUploadImagesStore(tx).BindUploadImageWithQuestion(ctx.Request().Context(), uploadImage.ID, db.UploadImageQuestionTypeAnswer, question.ID); err != nil {
				return errors.Wrap(err, "bind upload image with question")
			}
		}
		return nil
	}); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to answer question")
		return ctx.ServerError()
	}

	go func() {
		if question.ReceiveReplyEmail != "" && question.Answer == "" { // We only send the email when the question has not been answered.
			// Send notification to questioner.
			if err := mail.SendNewAnswerMail(question.ReceiveReplyEmail, ctx.User.Domain, question.ID, question.Content, f.Answer); err != nil {
				logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to send receive reply mail to questioner")
			}
		}
	}()
	return ctx.Success("提问回复成功")
}

func (*MineHandler) DeleteQuestion(ctx context.Context, question *db.Question) error {
	if err := db.Questions.DeleteByID(ctx.Request().Context(), question.ID); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to delete question")
		return ctx.ServerError()
	}
	return ctx.Success("提问删除成功")
}

func (*MineHandler) SetQuestionVisible(ctx context.Context, question *db.Question, f form.QuestionVisible) error {
	if f.Visible {
		if err := db.Questions.SetPublic(ctx.Request().Context(), question.ID); err != nil {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to set question public")
			return ctx.ServerError()
		}
		return ctx.Success("提问已设为公开")

	} else {
		if err := db.Questions.SetPrivate(ctx.Request().Context(), question.ID); err != nil {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to set question private")
			return ctx.ServerError()
		}
		return ctx.Success("提问已设为私密")
	}
}

func (*MineHandler) Profile(ctx context.Context) error {
	user := ctx.User
	return ctx.Success(&response.MineProfile{
		Email: user.Email,
		Name:  user.Name,
	})
}

func (*MineHandler) UpdateProfileSettings(ctx context.Context, tx dbutil.Transactor, f form.UpdateProfile) error {
	if err := tx.Transaction(func(tx *gorm.DB) error {
		usersStore := db.NewUsersStore(tx)
		if err := usersStore.SetName(ctx.Request().Context(), ctx.User.ID, f.Name); err != nil {
			return errors.Wrap(err, "update user profile")
		}

		if f.NewPassword != "" {
			if err := usersStore.ChangePassword(ctx.Request().Context(), ctx.User.ID, f.OldPassword, f.NewPassword); err != nil {
				return errors.Wrap(err, "change password")
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, db.ErrBadCredential) {
			return ctx.Error(http.StatusBadRequest, "旧密码输入错误")
		} else {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update profile")
			return ctx.ServerError()
		}
	}

	return ctx.Success("个人信息更新成功")
}

func (*MineHandler) BoxSettings(ctx context.Context) error {
	user := ctx.User

	return ctx.Success(&response.MineBoxSettings{
		Intro:         user.Intro,
		NotifyType:    string(user.Notify),
		AvatarURL:     user.Avatar,
		BackgroundURL: user.Background,
	})
}

func (*MineHandler) UpdateBoxSettings(ctx context.Context, f form.UpdateBoxSettings) error {
	user := ctx.User

	notifyType := db.NotifyType(f.NotifyType)
	switch notifyType {
	case db.NotifyTypeEmail, db.NotifyTypeNone:
	default:
		return ctx.Error(http.StatusBadRequest, "未知的通知类型")
	}

	var avatarURL string
	if f.Avatar != nil {
		uploadAvatar, err := uploadImageFile(ctx, uploadImageFileOptions{
			Image:          f.Avatar,
			UploaderUserID: user.ID,
		})
		if err != nil {
			if errors.Is(err, ErrUploadImageSizeTooLarge) {
				return ctx.Error(http.StatusBadRequest, "头像文件大小不能大于 5Mb")
			} else {
				logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to upload avatar")
				return ctx.Error(http.StatusInternalServerError, "上传头像失败，请重试")
			}
		}
		avatarURL = "https://" + conf.Upload.ImageBucketCDNHost + "/" + uploadAvatar.Key
	}

	var backgroundURL string
	if f.Background != nil {
		uploadBackground, err := uploadImageFile(ctx, uploadImageFileOptions{
			Image:          f.Background,
			UploaderUserID: ctx.User.ID,
		})
		if err != nil {
			if errors.Is(err, ErrUploadImageSizeTooLarge) {
				return ctx.Error(http.StatusBadRequest, "背景图文件大小不能大于 5Mb")
			} else {
				logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to upload background")
				return ctx.Error(http.StatusInternalServerError, "上传背景图失败，请重试")
			}
		}
		backgroundURL = "https://" + conf.Upload.ImageBucketCDNHost + "/" + uploadBackground.Key
	}

	if err := db.Users.Update(ctx.Request().Context(), user.ID, db.UpdateUserOptions{
		Avatar:     avatarURL,
		Background: backgroundURL,
		Intro:      f.Intro,
		Notify:     notifyType,
	}); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update box settings")
		return ctx.ServerError()
	}
	return ctx.Success("提问箱设置更新成功")
}

func (*MineHandler) HarassmentSettings(ctx context.Context) error {
	user := ctx.User
	return ctx.Success(&response.HarassmentSettings{
		HarassmentSettingType: user.HarassmentSetting,
		BlockWords:            user.BlockWords,
	})
}

func (*MineHandler) UpdateHarassmentSettings(ctx context.Context, f form.UpdateHarassmentSettings) error {
	user := ctx.User

	harassmentSettingType := db.HarassmentSettingType(f.HarassmentSettingType)
	switch harassmentSettingType {
	case db.HarassmentSettingTypeRegisterOnly:
	default:
		harassmentSettingType = db.HarassmentSettingNone
	}

	blockWords := f.BlockWords
	blockWords = strings.ReplaceAll(blockWords, "，", ",")
	blockWords = strings.TrimSpace(blockWords)

	words := make([]string, 0)
	wordSet := make(map[string]struct{})
	for _, word := range strings.Split(blockWords, ",") {
		word := strings.TrimSpace(word)
		if word == "" {
			continue
		}
		if _, ok := wordSet[word]; ok {
			continue
		}
		wordSet[word] = struct{}{}

		if len(word) > 10 {
			return ctx.Error(http.StatusBadRequest, "屏蔽词长度不能超过 10 个字符：%s", word)
		}
		words = append(words, word)
	}
	if len(words) > 10 {
		return ctx.Error(http.StatusBadRequest, "屏蔽词不能超过 10 个")
	}

	if err := db.Users.UpdateHarassmentSetting(ctx.Request().Context(), user.ID, db.HarassmentSettingOptions{
		Type:       harassmentSettingType,
		BlockWords: strings.Join(words, ","),
	}); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update harassment setting")
		return ctx.ServerError()
	}
	return ctx.Success("防骚扰设置更新成功")
}

func (*MineHandler) ExportData(ctx context.Context) error {
	user := ctx.User

	questions, err := db.Questions.GetByUserID(ctx.Request().Context(), user.ID, db.GetQuestionsByUserIDOptions{
		FilterAnswered: false,
		ShowPrivate:    true,
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get questions")
		return ctx.Error(http.StatusInternalServerError, "导出失败：获取问题信息失败")
	}

	f, err := createExportExcelFile(user, questions)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create excel file")
		return ctx.Error(http.StatusInternalServerError, "导出失败：创建Excel文件失败")
	}

	fileName := fmt.Sprintf("NekoBox账号信息导出-%s-%s.xlsx", user.Domain, time.Now().Format("20060102150405"))
	ctx.ResponseWriter().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.ResponseWriter().Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))

	if err := f.Write(ctx.ResponseWriter()); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to write excel file")
		return ctx.Error(http.StatusInternalServerError, "导出失败：写入Excel文件失败")
	}
	return nil
}

func (*MineHandler) Deactivate(ctx context.Context, session session.Session) error {
	if err := db.Users.Deactivate(ctx.Request().Context(), ctx.User.ID); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to deactivate user")
		return ctx.ServerError()
	}

	session.Flush()

	return ctx.Success("您的账号已停用，感谢您使用 NekoBox。期待未来还能再见 👋🏻")
}

func createXLSXStreamWriter(xlsx *excelize.File, sheet string, headers []string) (*excelize.StreamWriter, error) {
	xlsx.NewSheet(sheet)
	sw, err := xlsx.NewStreamWriter(sheet)
	if err != nil {
		return nil, errors.Wrap(err, "new stream writer")
	}

	cols := make([]interface{}, 0, len(headers))
	for _, c := range headers {
		cols = append(cols, c)
	}
	err = sw.SetRow("A1", cols)
	if err != nil {
		return nil, errors.Wrap(err, "set header row")
	}
	return sw, nil
}

func createExportExcelFile(user *db.User, questions []*db.Question) (*excelize.File, error) {
	f := excelize.NewFile()

	sw, err := createXLSXStreamWriter(f, "账号信息", nil)
	if err != nil {
		return nil, errors.Wrap(err, "create xlsx stream writer: 提问")
	}
	// Set personal information sheet.
	personalData := [][]interface{}{
		{"NekoBox 账号信息导出", fmt.Sprintf("导出时间 %s", time.Now().Format("2006-01-02 15:04:05"))},
		{"电子邮箱", user.Email},
		{"昵称", user.Name},
		{"个性域名", user.Domain},
		{"介绍", user.Intro},
		{"头像 URL", user.Avatar},
		{"背景图 URL", user.Background},
		{"注册时间", user.CreatedAt},
	}
	currentRow := 1
	for _, row := range personalData {
		cell, _ := excelize.CoordinatesToCellName(1, currentRow)
		_ = sw.SetRow(cell, row)
		currentRow++
	}
	if err := sw.Flush(); err != nil {
		return nil, errors.Wrap(err, "flush personal data")
	}

	// Set questions sheet.
	sw, err = createXLSXStreamWriter(f, "提问", []string{"提问时间", "问题", "回答"})
	if err != nil {
		return nil, errors.Wrap(err, "create xlsx stream writer: 提问")
	}

	currentRow = 2 // Include header row.
	for _, question := range questions {
		question := question
		vals := []interface{}{question.CreatedAt, question.Content, question.Answer}
		cell, _ := excelize.CoordinatesToCellName(1, currentRow)
		_ = sw.SetRow(cell, vals)
		currentRow++
	}
	if err := sw.Flush(); err != nil {
		return nil, errors.Wrap(err, "flush personal data")
	}

	f.SetActiveSheet(f.GetSheetIndex("提问"))
	f.DeleteSheet("Sheet1") // Delete default sheet.

	return f, nil
}
