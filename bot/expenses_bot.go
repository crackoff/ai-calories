package bot

import (
	"ai-calories/ai"
	"ai-calories/database"
	"ai-calories/i18n"
	"bytes"
	"fmt"
	"log"
	"strconv"
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
			chatID := update.Message.Chat.ID
			userID := update.Message.From.ID

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					sendMessage(chatID, i18n.GetString("welcome_expenses", lang), bot)
				case "timezone":
					timezone := b.getTimezone(db, userID, lang)
					sendMessage(chatID, timezone, bot)
				case "delete":
					deletedItem := b.deleteLastItem(db, userID, chatID, lang)
					sendMessage(chatID, deletedItem, bot)
				case "stats":
					stats, img := b.getMonthlyStats(db, chatID)
					sendImageMessage(chatID, stats, img, bot)
				case "annual":
					stats, img := b.getAnnualStats(db, chatID, lang)
					sendImageMessage(chatID, stats, img, bot)
				case "categories":
					result := b.getCategories(db, chatID, lang)
					sendMessage(chatID, result, bot)
				case "add_category":
					arg := strings.TrimSpace(update.Message.CommandArguments())
					if arg == "" {
						sendMessage(chatID, "Usage: /add_category <category name>", bot)
						continue
					}
					result := b.addCategory(db, chatID, arg, lang)
					sendMessage(chatID, result, bot)
				case "migrate":
					args := strings.Fields(update.Message.CommandArguments())
					if len(args) != 2 {
						sendMessage(chatID, "Usage: /migrate <source_user_id> <target_chat_id>", bot)
						continue
					}
					sourceUserID, err1 := strconv.ParseInt(args[0], 10, 64)
					targetChatID, err2 := strconv.ParseInt(args[1], 10, 64)
					if err1 != nil || err2 != nil {
						sendMessage(chatID, "Invalid user ID or chat ID", bot)
						continue
					}
					catCount, expCount, err := db.MigrateUserToChat(sourceUserID, targetChatID)
					if err != nil {
						log.Print(err)
						sendMessage(chatID, fmt.Sprintf("Migration failed: %v", err), bot)
						continue
					}
					sendMessage(chatID, fmt.Sprintf("Migration complete: copied %d categories and %d expenses", catCount, expCount), bot)
				case "authorize":
					if isGroupChat(update.Message) {
						sendMessage(chatID, i18n.GetString("authorize_private", lang), bot)
						continue
					}
					password := strings.Replace(update.Message.Text, "/authorize ", "", 1)
					if password != b.MasterPassword {
						continue
					}
					err := db.AddUser(userID, update.Message.From.UserName)
					if err != nil {
						log.Print(err)
						continue
					}
					sendMessage(chatID, "User authorized", bot)
				default:
					sendMessage(chatID, i18n.GetString("unknown_command", lang), bot)
				}
			} else {
				err := checkAuthorization(db, userID, update.Message.From.UserName, false)
				if err != nil {
					msg := tgbotapi.NewMessage(chatID, i18n.GetString("unauthorized", lang))
					bot.Send(msg)
					continue
				}
				addedItem := b.addItem(db, userID, chatID, update.Message.Text, update.Message.Date, lang)
				sendMessage(chatID, addedItem, bot)
			}
		}
	}
}

func (b *ExpensesBot) addItem(db *database.Database, userID int64, chatID int64, userText string, timestamp int, lang string) string {
	defaults := []string{
		i18n.GetString("house", lang),
		i18n.GetString("food", lang),
		i18n.GetString("transport", lang),
		i18n.GetString("entertainment", lang),
		i18n.GetString("health", lang),
		i18n.GetString("education", lang),
		i18n.GetString("other", lang),
	}

	categories, err := db.GetAllUserCategories(chatID, defaults)
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
	product.ChatID = chatID
	product.Timestamp = timestamp

	originalCost := product.TotalCost
	originalCurrency := product.Currency

	if strings.ToUpper(strings.TrimSpace(product.Currency)) != "USD" && product.Currency != "" {
		converted, err := ConvertToUSD(product.TotalCost, product.Currency)
		if err != nil {
			log.Println(err)
			return fmt.Sprintf(i18n.GetString("error_conversion", lang), product.Currency)
		}
		product.TotalCost = converted
		product.Currency = "USD"
	}

	err = db.InsertProduct(product)
	if err != nil {
		log.Print(err)
		return fmt.Sprintf(i18n.GetString("error_adding", lang), userText)
	}

	spendings, _ := db.GetUserStatisticsForCurrentMonth(chatID)
	total, _ := spendings.Get("🤑 Total")

	if strings.ToUpper(strings.TrimSpace(originalCurrency)) != "USD" && originalCurrency != "" {
		return fmt.Sprintf(i18n.GetString("added_to_category_converted", lang), product.Item, originalCost, originalCurrency, product.TotalCost, product.Category, total)
	}
	return fmt.Sprintf(i18n.GetString("added_to_category", lang), product.Item, product.TotalCost, product.Category, total)
}

func (b *ExpensesBot) deleteLastItem(db *database.Database, userID int64, chatID int64, lang string) string {
	foodItem, err := db.DeleteLastItem(userID, chatID)
	if err != nil {
		log.Println(err)
		return ""
	}
	return fmt.Sprintf(i18n.GetString("deleted", lang), foodItem)
}

func (b *ExpensesBot) addCategory(db *database.Database, chatID int64, category string, lang string) string {
	category = b.Classifier.GetCategory(category)
	err := db.AddUserCategory(chatID, category)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_category", lang)
	}
	return fmt.Sprintf(i18n.GetString("category_added", lang), category)
}

func (b *ExpensesBot) deleteCategory(db *database.Database, chatID int64, category string, lang string) string {
	category = b.Classifier.GetCategory(category)
	err := db.DeleteUserCategory(chatID, category)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_category", lang)
	}
	return fmt.Sprintf(i18n.GetString("category_deleted", lang), category)
}

func (b *ExpensesBot) getCategories(db *database.Database, chatID int64, lang string) string {
	defaults := []string{
		i18n.GetString("house", lang),
		i18n.GetString("food", lang),
		i18n.GetString("transport", lang),
		i18n.GetString("entertainment", lang),
		i18n.GetString("health", lang),
		i18n.GetString("education", lang),
		i18n.GetString("other", lang),
	}
	categories, err := db.GetAllUserCategories(chatID, defaults)
	if err != nil {
		log.Print(err)
		return i18n.GetString("error_category", lang)
	}
	message := ""
	for _, cat := range categories {
		message += cat.Category + "\n"
	}
	return fmt.Sprintf(i18n.GetString("categories_list", lang), message)
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

func (b *ExpensesBot) getMonthlyStats(db *database.Database, chatID int64) (string, bytes.Buffer) {
	stats, err := db.GetUserStatisticsForCurrentMonth(chatID)
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

func (b *ExpensesBot) getAnnualStats(db *database.Database, chatID int64, lang string) (string, bytes.Buffer) {
	stats, err := db.GetUserAnnualStats(chatID)
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
