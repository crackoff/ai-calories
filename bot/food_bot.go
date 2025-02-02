package bot

import (
	"ai-calories/ai"
	"ai-calories/database"
	"ai-calories/i18n"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type FoodBot struct {
	Bot
	MasterPassword string
}

func (b *FoodBot) HandleBot(bot *tgbotapi.BotAPI, db *database.Database, classifier *ai.Classifier) {
	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message != nil {
			lang := update.Message.From.LanguageCode
			fc := (*classifier).(ai.FoodClassifier)
			if update.Message.IsCommand() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				switch update.Message.Command() {
				case "start":
					msg.Text = i18n.GetString("welcome_food", lang)
				case "timezone":
					tz, _ := db.GetUserTimezone(update.Message.From.ID)
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
					if c, _ := fc.Classify(dry); c != "food" {
						continue
					}
					if strings.TrimSpace(dry) == "" {
						msg.Text = "Please provide a food item"
					}
					food, err := fc.GetNutritionData(dry)
					if err != nil {
						log.Print(err)
						continue
					}
					f := i18n.FormatNutrition(food.Calories, 0.0, food.Fat, food.Carbohydrates, food.Protein, lang)
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
					msg.Text = "User authorized"
				default:
					msg.Text = "Sorry, I don't know that command."
				}
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
					continue
				}
			} else if len(update.Message.Photo) > 0 {
				err := checkAuthorization(db, update.Message.From.ID)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, i18n.GetString("unauthorized", lang))
					bot.Send(msg)
					continue
				}

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

				food, err := fc.GetGetNutritionDataByImage(img, update.Message.Text)
				if err != nil {
					log.Print(err)
					continue
				}
				food.UserID = update.Message.From.ID
				food.Timestamp = time.Unix(int64(update.Message.Date), 0)
				totalCalories, err := db.InsertFood(food)
				if err != nil {
					log.Print(err)
					continue
				}
				f := i18n.FormatNutrition(food.Calories, totalCalories, food.Fat, food.Carbohydrates, food.Protein, lang)
				s := fmt.Sprintf(i18n.GetString("added", lang), food.FoodItem, food.Weight, f)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, escapeMarkdownV2(s))
				msg.ParseMode = "MarkdownV2"
				_, err = bot.Send(msg)
				if err != nil {
					log.Print(err)
					continue
				}
			} else {
				err := checkAuthorization(db, update.Message.From.ID)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, i18n.GetString("unauthorized", lang))
					bot.Send(msg)
					continue
				}

				log.Print(update.Message.Text)
				class, err := fc.Classify(update.Message.Text)
				if err != nil {
					log.Print(err)
				}
				if class == "food" {
					food, err := fc.GetNutritionData(update.Message.Text)
					if err != nil {
						log.Print(err)
						continue
					}
					food.UserID = update.Message.From.ID
					food.Timestamp = time.Unix(int64(update.Message.Date), 0)
					totalCalories, err := db.InsertFood(food)
					if err != nil {
						log.Print(err)
						continue
					}
					f := i18n.FormatNutrition(food.Calories, totalCalories, food.Fat, food.Carbohydrates, food.Protein, lang)
					s := fmt.Sprintf(i18n.GetString("added", lang), food.FoodItem, food.Weight, f)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, escapeMarkdownV2(s))
					msg.ParseMode = "MarkdownV2"
					_, err = bot.Send(msg)
					if err != nil {
						log.Print(err)
						continue
					}
				} else if class == "location" {
					tz, err := ai.GetTimezone(fc.GetAI(), update.Message.Text)
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
					log.Printf("[WARNING] classified as %s\n", class)
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
