package bot

import (
	"ai-calories/ai"
	"ai-calories/database"
	"bytes"
	"errors"
	"io"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/imroc/req"
	"gorm.io/gorm"
)

type Bot interface {
	HandleBot(bot *tgbotapi.BotAPI, db *database.Database, classifier *ai.Classifier)
}

type authStore interface {
	GetUser(int64) (database.User, error)
	AddUser(int64, string) error
	GetFoodsCount(int64) (int, error)
}

func NewBot(botType string, masterPassword string) Bot {
	switch botType {
	case "food":
		return &FoodBot{MasterPassword: masterPassword}
	case "expenses":
		return &ExpensesBot{MasterPassword: masterPassword}
	}
	return nil
}

func checkAuthorization(db authStore, userID int64, username string, is_member bool) error {
	if is_member {
		return nil
	}

	_, err := db.GetUser(userID)
	if err == gorm.ErrRecordNotFound {
		if err := db.AddUser(userID, username); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	count, err := db.GetFoodsCount(userID)
	if err != nil {
		return err
	}
	if count > 10 {
		return errors.New("too many requests, please join the Tribute channel https://t.me/tribute/app?startapp=sw6x to continue using this bot")
	}

	return nil
}

func isGroupChat(msg *tgbotapi.Message) bool {
	return msg.Chat.Type == "group" || msg.Chat.Type == "supergroup"
}

func escapeMarkdownV2(text string) string {
	// Characters to be escaped in MarkdownV2
	escapeChars := []string{"_", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	// Replace each character in the text with its escaped version
	for _, ch := range escapeChars {
		text = strings.ReplaceAll(text, ch, "\\"+ch)
	}
	return text
}

func downloadFile(fileURL string) (*bytes.Buffer, error) {
	resp, err := req.Get(fileURL)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Response().Body)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Response().Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func sendMessage(chatID int64, message string, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2(message))
	msg.ParseMode = "MarkdownV2"
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func sendImageMessage(chatID int64, message string, img bytes.Buffer, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileReader{Name: "image.png", Reader: &img})
	msg.Caption = escapeMarkdownV2(message)
	msg.ParseMode = "MarkdownV2"
	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}
