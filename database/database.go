package database

import (
	"ai-calories/i18n"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/matheusoliveira/go-ordered-map/omap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	err = db.AutoMigrate(&User{}, &Food{}, &UserTimezone{}, &Expense{}, &ExpenseCategory{})
	if err != nil {
		log.Fatalln("Failed to migrate:", err)
	}

	return db
}

func (db *Database) GetUser(userID int64) (User, error) {
	var user User
	result := db.Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		return User{}, result.Error
	}
	return user, nil
}

func (db *Database) AddUser(userID int64, username string) error {
	user := User{UserID: userID, Username: username}
	err := db.Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetTodayNutrition(userID int64) (FoodResult, error) {
	// Prepare the query
	startOfDay, err := db.GetStartOfDay(userID)
	if err != nil {
		return FoodResult{}, err
	}

	var result FoodResult
	db.Model(&Food{}).
		Select("SUM(calories) AS total_calories, SUM(fat) AS total_fat, SUM(carbohydrates) AS total_carbohydrates, SUM(protein) AS total_protein").
		Where("user_id = ? AND timestamp >= ?", userID, startOfDay).
		Scan(&result)

	return result, nil
}

func (db *Database) GetStartOfDay(userID int64) (time.Time, error) {
	var loc *time.Location
	tz, err := db.GetUserTimezone(userID)
	if err != nil {
		tz = "UTC"
	}

	loc, err = time.LoadLocation(tz)
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return startOfDay, err
}

func (db *Database) GetUserTimezone(userID int64) (string, error) {
	var userTimezone UserTimezone = UserTimezone{UserID: userID, Timezone: "UTC"}
	result := db.Where("user_id = ?", userID).FirstOrCreate(&userTimezone)
	if result.Error != nil {
		return "", result.Error
	}

	return userTimezone.Timezone, nil
}

func (db *Database) GetFoodGroups(userID int64, lang string) (string, error) {
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
		  AND deleted_at IS NULL
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

func (db *Database) InsertFood(f Food) (float64, error) {
	err := db.Create(&f).Error
	if err != nil {
		return 0, err
	}

	totalCalories := 0.0

	startOfDay, err := db.GetStartOfDay(f.UserID)
	if err != nil {
		return 0, err
	}

	db.Model(&Food{}).
		Select("SUM(calories) AS total_calories").
		Where("user_id = ? AND timestamp >= ?", f.UserID, startOfDay).
		Scan(&totalCalories)

	return totalCalories, err
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

func (db *Database) GetAllUserCategories(userID int64, defaults []string) ([]ExpenseCategory, error) {
	var categories []ExpenseCategory
	result := db.Find(&categories, "user_id = ?", userID)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(categories) == 0 {
		// For new users
		for _, category := range defaults {
			_ = db.AddUserCategory(userID, category)
			categories = append(categories, ExpenseCategory{UserID: userID, Category: category})
		}
	}

	return categories, nil
}

func (db *Database) AddUserCategory(userID int64, category string) error {
	expenseCategory := ExpenseCategory{UserID: userID, Category: category}
	err := db.Create(&expenseCategory).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteUserCategory(userID int64, category string) error {
	var cat ExpenseCategory
	result := db.Find(&cat, "user_id = ? AND category = ?", userID, category)
	if result.Error != nil {
		return result.Error
	}
	err := db.Delete(&cat).Error
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) InsertProduct(product Expense) error {
	err := db.Create(&product).Error
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteLastItem(userId int64) (string, error) {
	var lastItem Expense
	result := db.Order("timestamp DESC").Limit(1).Find(&lastItem, "user_id = ?", userId)
	if result.Error != nil {
		return "", result.Error
	}

	err := db.Delete(&lastItem).Error
	if err != nil {
		return "", err
	}

	return lastItem.Item, nil
}

func (db *Database) GetUserStatisticsForCurrentMonth(userId int64) (omap.OMap[string, float64], error) {
	query := `select ifnull(c.category, '🤑 Total'),
			  	     ifnull(sum(total_cost), 0.0) as total_cost
			from expense_categories c
			left join expenses p on p.user_id = c.user_id
				  and instr(c.category, p.category) > 0
				  and month(from_unixtime(p.timestamp)) = month(now())
				  and year(from_unixtime(p.timestamp)) = year(now())
				  and p.deleted_at IS NULL
			where c.user_id = ?
			group by c.category with rollup;`
	rows, err := db.Raw(query, userId).Rows()
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var statistics = omap.New[string, float64]()
	for rows.Next() {
		var category string
		var totalCost float64
		err = rows.Scan(&category, &totalCost)
		if err != nil {
			return nil, err
		}
		statistics.Put(category, totalCost)
	}

	return statistics, nil
}

func (db *Database) GetUserAnnualStats(userId int64) (omap.OMap[string, float64], error) {
	// Querying spending for the last 1 year
	query := `with recursive nums as (
				select 0 as num union all select num + 1 from nums where num < 11
			), year_behind as (
				select date_format(date_sub(curdate(), interval num month), '%Y-%m-01') as month
				from nums
			)
			select monthname(y.month) as month
				 , coalesce(sum(total_cost), 0) as total
			from year_behind y
			left join expenses p on y.month = date_format(from_unixtime(timestamp), '%Y-%m-01')
				  and timestamp >= unix_timestamp(date_sub(now(), interval 1 year))
				  and p.user_id = ?
				  and p.deleted_at IS NULL
			group by y.month
			order by y.month;`

	rows, err := db.Raw(query, userId).Rows()
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	statistics := omap.New[string, float64]()
	for rows.Next() {
		var month string
		var total float64
		err = rows.Scan(&month, &total)
		if err != nil {
			return nil, err
		}
		statistics.Put(month, total)
	}

	return statistics, nil
}
