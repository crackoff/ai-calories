package bot

import (
	"ai-calories/ai"
	"ai-calories/database"
	"bytes"
	"io"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/imroc/req"
)

type Bot interface {
	HandleBot(bot *tgbotapi.BotAPI, db *database.Database, classifier *ai.Classifier)
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

func checkAuthorization(db *database.Database, userID int64) error {
	_, err := db.GetUser(userID)
	if err != nil {
		return err
	}
	return nil
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
	msg.Caption = message
	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}
