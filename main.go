package main

import (
	"ai-calories/ai"
	data "ai-calories/database"
	"ai-calories/i18n"
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/imroc/req"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	db := data.NewDatabase(connStr)

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(telegramToken)

	aiProvider := os.Getenv("AI_PROVIDER")
	classifier := ai.NewClassifier(aiProvider)

	if err != nil {
		log.Fatalln(err)
	}
	handleBot(bot, db, classifier)
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

func handleBot(bot *tgbotapi.BotAPI, db *data.Database, classifier *ai.Classifier) {
	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message != nil { // If we receive a message
			lang := update.Message.From.LanguageCode

			if update.Message.IsCommand() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				switch update.Message.Command() {
				case "start":
					msg.Text = i18n.GetString("welcome", lang)
				case "timezone":
					tz, err := db.GetUserTimezone(update.Message.From.ID)
					loc, err := time.LoadLocation(tz)
					if err != nil {
						loc = time.FixedZone("UTC", 0)
					}
					msg.Text = fmt.Sprintf(i18n.GetString("timezone", lang), loc)
				case "total":
					totals, err := db.GetTodayNutrition(update.Message.From.ID, lang)
					if err != nil {
						log.Print(err)
						continue
					}
					msg.Text = escapeMarkdownV2(totals)
					msg.ParseMode = "MarkdownV2"
				case "dry":
					dry := strings.Replace(update.Message.Text, "/dry", "", 1)
					if c, _ := classifier.Classify(dry); c != "food" {
						continue
					}
					if strings.TrimSpace(dry) == "" {
						msg.Text = "Please provide a food item"
					}
					food, err := classifier.GetNutritionData(dry)
					if err != nil {
						log.Print(err)
						continue
					}
					f := i18n.FormatNutrition(food.Calories, food.Fat, food.Carbohydrates, food.Protein, lang)
					msg.Text = escapeMarkdownV2(fmt.Sprintf("*%s* (%.0fg.)\n%s", food.FoodItem, food.Weight, f))
					msg.ParseMode = "MarkdownV2"
				case "set":
					msg.Text = "This feature is in development"
				case "delete":
					foodItem, err := db.DeleteLastFood(update.Message.From.ID)
					if err != nil {
						log.Print(err)
						continue
					}
					msg.Text = fmt.Sprintf(i18n.GetString("deleted", lang), foodItem)
					msg.ParseMode = "MarkdownV2"
				default:
					msg.Text = "Sorry, I don't know that command."
				}
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
					continue
				}
			} else if len(update.Message.Photo) > 0 {
				firstPhoto := update.Message.Photo[0]
				fileID := firstPhoto.FileID
				fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
				if err != nil {
					log.Println("Error getting file:", err)
					continue
				}

				fileURL := "https://api.telegram.org/file/bot" + bot.Token + "/" + fileInfo.FilePath
				img, err := downloadFile(fileURL)
				if err != nil {
					log.Println("Error downloading file:", err)
					continue
				}

				food, err := classifier.GetGetNutritionDataByImage(img, update.Message.Text)
				if err != nil {
					log.Print(err)
					continue
				}
				food.UserID = update.Message.From.ID
				food.Timestamp = time.Unix(int64(update.Message.Date), 0)
				err = db.InsertFood(food)
				if err != nil {
					log.Print(err)
					continue
				}
				f := i18n.FormatNutrition(food.Calories, food.Fat, food.Carbohydrates, food.Protein, lang)
				s := fmt.Sprintf(i18n.GetString("added", lang), food.FoodItem, food.Weight, f)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, escapeMarkdownV2(s))
				msg.ParseMode = "MarkdownV2"
				_, err = bot.Send(msg)
				if err != nil {
					log.Print(err)
					continue
				}
			} else {
				log.Print(update.Message.Text)
				class, _ := classifier.Classify(update.Message.Text)
				if class == "food" {
					food, err := classifier.GetNutritionData(update.Message.Text)
					if err != nil {
						log.Print(err)
						continue
					}
					food.UserID = update.Message.From.ID
					food.Timestamp = time.Unix(int64(update.Message.Date), 0)
					err = db.InsertFood(food)
					if err != nil {
						log.Print(err)
						continue
					}
					f := i18n.FormatNutrition(food.Calories, food.Fat, food.Carbohydrates, food.Protein, lang)
					s := fmt.Sprintf(i18n.GetString("added", lang), food.FoodItem, food.Weight, f)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, escapeMarkdownV2(s))
					msg.ParseMode = "MarkdownV2"
					_, err = bot.Send(msg)
					if err != nil {
						log.Print(err)
						continue
					}
				} else if class == "location" {
					tz, err := classifier.GetTimezone(update.Message.Text)
					if err != nil {
						log.Print(err)
						continue
					}
					err = db.SaveUserTimezone(update.Message.From.ID, tz)
					if err != nil {
						log.Print(err)
						continue
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(i18n.GetString("tz_updated", lang), tz))
					_, err = bot.Send(msg)
					if err != nil {
						log.Print(err)
						continue
					}
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, i18n.GetString("unknown", lang))
					_, err := bot.Send(msg)
					if err != nil {
						log.Print(err)
						continue
					}
				}
			}
		}
	}
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
