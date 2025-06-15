package bot

import (
	"ai-calories/ai"
	"ai-calories/database"
	"ai-calories/i18n"
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ExpensesBot struct {
	Bot
	MasterPassword string
	Classifier     ai.ExpensesClassifier
}

func (b *ExpensesBot) HandleBot(bot *tgbotapi.BotAPI, db *database.Database, classifier *ai.Classifier) {
	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)
	b.Classifier = (*classifier).(ai.ExpensesClassifier)
	for update := range updates {
		if update.Message != nil { // If we receive a message
			log.Print(update.Message.Text)
			lang := update.Message.From.LanguageCode

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					sendMessage(update.Message.Chat.ID, i18n.GetString("welcome_expenses", lang), bot)
				case "timezone":
					timezone := b.getTimezone(db, update.Message.From.ID, lang)
					sendMessage(update.Message.Chat.ID, timezone, bot)
				case "delete":
					deletedItem := b.deleteLastItem(db, update.Message.From.ID, lang)
					sendMessage(update.Message.Chat.ID, deletedItem, bot)
				case "stats":
					stats, img := b.getMonthlyStats(db, update.Message.From.ID)
					sendImageMessage(update.Message.Chat.ID, stats, img, bot)
				case "annual":
					stats, img := b.getAnnualStats(db, update.Message.From.ID, lang)
					sendImageMessage(update.Message.Chat.ID, stats, img, bot)
				case "authorize":
					password := strings.Replace(update.Message.Text, "/authorize ", "", 1)
					if password != b.MasterPassword {
						continue
					}
					err := db.AddUser(update.Message.From.ID, update.Message.From.UserName)
					if err != nil {
						log.Print(err)
						continue
					}
					sendMessage(update.Message.Chat.ID, "User authorized", bot)
				default:
					sendMessage(update.Message.Chat.ID, i18n.GetString("unknown_command", lang), bot)
				}
			} else {
				err := checkAuthorization(db, update.Message.From.ID)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, i18n.GetString("unauthorized", lang))
					bot.Send(msg)
					continue
				}
				addedItem := b.addItem(db, update.Message.From.ID, update.Message.Text, update.Message.Date, lang)
				sendMessage(update.Message.Chat.ID, addedItem, bot)
			}
		}
	}
}

func (b *ExpensesBot) addItem(db *database.Database, userID int64, userText string, timestamp int, lang string) string {
	defaults := []string{
		i18n.GetString("house", lang),
		i18n.GetString("food", lang),
		i18n.GetString("transport", lang),
		i18n.GetString("entertainment", lang),
		i18n.GetString("health", lang),
		i18n.GetString("education", lang),
		i18n.GetString("other", lang),
	}

	categories, err := db.GetAllUserCategories(userID, defaults)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_adding", lang)
	}

	categories_str := ""
	for _, category := range categories {
		categories_str += category.Category + ", "
	}
	categories_str = strings.TrimRight(categories_str, ", ")

	product, err := b.Classifier.GetProductData(userText, categories_str)
	if err != nil {
		log.Println(err)
		return fmt.Sprintf(i18n.GetString("error_adding", lang), userText)
	}
	product.UserID = userID
	product.Timestamp = timestamp

	err = db.InsertProduct(product)
	if err != nil {
		log.Print(err)
		return fmt.Sprintf(i18n.GetString("error_adding", lang), userText)
	}

	spendings, _ := db.GetUserStatisticsForCurrentMonth(userID)
	total, _ := spendings.Get("🤑 Total")

	return fmt.Sprintf(i18n.GetString("added_to_category", lang), product.Item, product.TotalCost, product.Category, total)
}

func (b *ExpensesBot) deleteLastItem(db *database.Database, userID int64, lang string) string {
	foodItem, err := db.DeleteLastItem(userID)
	if err != nil {
		log.Println(err)
		return ""
	}
	return fmt.Sprintf(i18n.GetString("deleted", lang), foodItem)
}

func (b *ExpensesBot) addCategory(db *database.Database, userID int64, category string, lang string) string {
	category = b.Classifier.GetCategory(category)
	err := db.AddUserCategory(userID, category)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_category", lang)
	}
	return fmt.Sprintf(i18n.GetString("category_added", lang), category)
}

func (b *ExpensesBot) deleteCategory(db *database.Database, userID int64, category string, lang string) string {
	category = b.Classifier.GetCategory(category)
	err := db.DeleteUserCategory(userID, category)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_category", lang)
	}
	return fmt.Sprintf(i18n.GetString("category_deleted", lang), category)
}

func (b *ExpensesBot) getTimezone(db *database.Database, id int64, lang string) string {
	tz, err := db.GetUserTimezone(id)
	if err != nil {
		log.Print(err)
		return "error"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.FixedZone("UTC", 0)
	}
	return fmt.Sprintf(i18n.GetString("timezone", lang), loc)
}

func (b *ExpensesBot) setLocation(db *database.Database, userID int64, location string, lang string) string {
	tz, err := ai.GetTimezone(b.Classifier.GetAI(), location)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_tz_update", lang)
	}
	err = db.SaveUserTimezone(userID, tz)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_tz_update", lang)
	}
	return fmt.Sprintf(i18n.GetString("tz_updated", lang), tz)
}

func (b *ExpensesBot) getMonthlyStats(db *database.Database, userID int64) (string, bytes.Buffer) {
	stats, err := db.GetUserStatisticsForCurrentMonth(userID)
	if err != nil {
		log.Print(err)
		return "", bytes.Buffer{}
	}

	img, err := DrawPieChart(stats)
	if err != nil {
		log.Println(err)
		return "", bytes.Buffer{}
	}

	message := ""
	for it := stats.Iterator(); it.Next(); {
		message += fmt.Sprintf("%s: $%.2f\n", it.Key(), it.Value())
	}

	return message, img
}

func (b *ExpensesBot) getAnnualStats(db *database.Database, userID int64, lang string) (string, bytes.Buffer) {
	stats, err := db.GetUserAnnualStats(userID)
	if err != nil {
		log.Print(err)
		return "", bytes.Buffer{}
	}

	img, err := DrawBarChart(stats)
	if err != nil {
		log.Println(err)
		return "", bytes.Buffer{}
	}

	total := 0.0
	for it := stats.Iterator(); it.Next(); {
		total += it.Value()
	}

	return fmt.Sprintf(i18n.GetString("total_year", lang), total), img
}
