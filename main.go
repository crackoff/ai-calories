package main

import (
	"ai-calories/ai"
	data "ai-calories/database"
	"ai-calories/i18n"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	err = data.CreateTableIfNotExists(db)
	if err != nil {
		log.Fatalln(err)
	}

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatalln(err)
	}
	handleBot(bot, db)
}

func handleBot(bot *tgbotapi.BotAPI, db *sql.DB) {
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
				case "help":
					msg.Text = "Send /start to get started."
				case "total":
					totals, err := data.GetTodayNutrition(db, update.Message.From.ID, lang)
					if err != nil {
						log.Print(err)
						continue
					}
					msg.Text = escapeMarkdownV2(totals)
					msg.ParseMode = "MarkdownV2"
				case "dry":
					dry := strings.Replace(update.Message.Text, "/dry", "", 1)
					if strings.TrimSpace(dry) == "" {
						msg.Text = "Please provide a food item"
					}
					food, err := ai.GetNutritionData(dry)
					if err != nil {
						log.Print(err)
						_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
						continue
					}
					f := i18n.FormatNutrition(food.Calories, food.TotalFat, food.TotalCarbohydrates, food.Protein, lang)
					msg.Text = escapeMarkdownV2(fmt.Sprintf("*%s* (%dг.)\n%s", food.FoodItem, food.TotalWeight, f))
					msg.ParseMode = "MarkdownV2"
				case "set":
					msg.Text = "This feature is in development"
				case "delete":
					foodItem, err := data.DeleteLastFood(db, update.Message.From.ID)
					if err != nil {
						log.Print(err)
						_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
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
				}
			} else {
				log.Print(update.Message.Text)
				food, err := ai.GetNutritionData(update.Message.Text)
				if err != nil {
					log.Print(err)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}

				food.UserID = update.Message.From.ID
				food.Timestamp = update.Message.Time().Unix()
				err = data.InsertFood(db, food)
				if err != nil {
					log.Print(err)
					continue
				}
				f := i18n.FormatNutrition(food.Calories, food.TotalFat, food.TotalCarbohydrates, food.Protein, lang)
				s := fmt.Sprintf(i18n.GetString("added", lang), food.FoodItem, food.TotalWeight, f)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, escapeMarkdownV2(s))
				msg.ParseMode = "MarkdownV2"
				_, err = bot.Send(msg)
				if err != nil {
					log.Print(err)
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
