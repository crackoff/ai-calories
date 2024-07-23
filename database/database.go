package database

import (
	"ai-calories/i18n"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

type Database struct {
	*gorm.DB
}

func NewDatabase(dsn string) *Database {
	conn, err := gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", dsn),
	}))
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	db := &Database{DB: conn}
	err = db.AutoMigrate(&Food{}, &UserTimezone{})
	if err != nil {
		log.Fatalln("Failed to migrate:", err)
	}
	return db
}

func (db *Database) GetTodayNutrition(userID int64, lang string) (string, error) {
	// Prepare the query
	startOfDay, err := db.GetStartOfDay(userID)
	if err != nil {
		return "", err
	}

	var result struct {
		TotalCalories      float64
		TotalFat           float64
		TotalCarbohydrates float64
		TotalProtein       float64
	}
	db.Model(&Food{}).
		Select("SUM(calories) AS total_calories, SUM(fat) AS total_fat, SUM(carbohydrates) AS total_carbohydrates, SUM(protein) AS total_protein").
		Where("user_id = ? AND timestamp >= ?", userID, startOfDay).
		Scan(&result)

	// Format the results
	total := i18n.FormatNutrition(result.TotalCalories, result.TotalFat, result.TotalCarbohydrates, result.TotalProtein, lang)
	groups, err := db.getFoodGroups(userID, lang)
	if err != nil {
		log.Print(err)
		return "", err
	}
	s := fmt.Sprintf(i18n.GetString("total_today", lang), total, groups)

	return s, nil
}

func (db *Database) GetStartOfDay(userID int64) (int64, error) {
	var loc *time.Location
	tz, err := db.GetUserTimezone(userID)
	if err != nil {
		loc = time.FixedZone("UTC", 0)
	}

	loc, err = time.LoadLocation(tz)
	if err != nil {
		loc = time.FixedZone("UTC", 0)
	}
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	return startOfDay, err
}

func (db *Database) GetUserTimezone(userID int64) (string, error) {
	var userTimezone UserTimezone
	result := db.Where("user_id = ?", userID).First(&userTimezone)
	if result.Error != nil {
		return "", result.Error
	}

	return userTimezone.Timezone, nil
}

func (db *Database) getFoodGroups(userID int64, lang string) (string, error) {
	startOfDay, err := db.GetStartOfDay(userID)
	if err != nil {
		return "", err
	}

	var foodGroups []struct {
		TimeOfDay string
		FoodItems string
	}
	query := `
		SELECT
			CASE
				WHEN HOUR(timestamp) BETWEEN 0 AND 11 THEN ?
				WHEN HOUR(timestamp) BETWEEN 12 AND 17 THEN ?
				ELSE ?
			END as time_of_day,
			GROUP_CONCAT(food_item SEPARATOR ', ') as food_items
		FROM foods
		WHERE user_id = ?
		  AND timestamp >= ?
		GROUP BY time_of_day
		ORDER BY CASE WHEN time_of_day = ? THEN 1 WHEN time_of_day = ? THEN 2 ELSE 3 END`

	err = db.Raw(query,
		i18n.GetString("morning", lang),
		i18n.GetString("afternoon", lang),
		i18n.GetString("evening", lang),
		userID, startOfDay,
		i18n.GetString("morning", lang),
		i18n.GetString("afternoon", lang),
	).Scan(&foodGroups).Error

	if err != nil {
		return "", err
	}

	var result string
	for _, group := range foodGroups {
		result += fmt.Sprintf("*%s* %s\n", group.TimeOfDay, group.FoodItems)
	}

	return result, nil
}

func (db *Database) InsertFood(f Food) error {
	return db.Create(&f).Error
}

func (db *Database) DeleteLastFood(userId int64) (string, error) {
	var lastFood Food
	result := db.Order("timestamp DESC").Limit(1).Find(&lastFood, "user_id = ?", userId)
	if result.Error != nil {
		return "", result.Error
	}

	if result.RowsAffected > 0 {
		err := db.Delete(&lastFood).Error
		if err != nil {
			return "", err
		}
		return lastFood.FoodItem, nil
	}

	return "", errors.New("no food item found for the user")
}

func (db *Database) SaveUserTimezone(userId int64, timezone string) error {
	var userTimezone UserTimezone
	// First, try to find the record
	result := db.Where("user_id = ?", userId).First(&userTimezone)

	var err error = nil
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Record not found, create a new one
		userTimezone = UserTimezone{UserID: userId, Timezone: timezone}
		err = db.Create(&userTimezone).Error
	} else {
		// Record found, update it
		err = db.Model(&userTimezone).Update("timezone", timezone).Error
	}

	return err
}
