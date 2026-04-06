package main

import (
	"ai-calories/ai"
	bot "ai-calories/bot"
	data "ai-calories/database"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	db := data.NewDatabase(connStr)

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgbot, err := tgbotapi.NewBotAPI(telegramToken)

	aiProvider := os.Getenv("AI_PROVIDER")
	botType := os.Getenv("BOT_TYPE")
	masterPassword := os.Getenv("MASTER_PASSWORD")
	currencyAPIKey := os.Getenv("CURRENCY_API_KEY")
	bot.InitCurrencyAPI(currencyAPIKey)

	if uyuRate := os.Getenv("UYU_RATE"); uyuRate != "" {
		if rate, err := strconv.ParseFloat(uyuRate, 64); err == nil {
			bot.SetManualRate("UYU", rate)
		}
	}

	classifier := ai.NewClassifier(aiProvider, botType)
	chatBot := bot.NewBot(botType, masterPassword)

	if err != nil {
		log.Fatalln(err)
	}
	chatBot.HandleBot(tgbot, db, &classifier)
}
