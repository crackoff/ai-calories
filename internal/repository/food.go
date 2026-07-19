package repository

import (
	"ai-calories/database"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type FoodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) *FoodRepository {
	return &FoodRepository{db: db}
}

func (r *FoodRepository) InsertFood(food *database.Food) error {
	return r.db.Create(food).Error
}

func (r *FoodRepository) DeleteByID(id uint, userID int64) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&database.Food{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *FoodRepository) DeleteLast(userID int64) error {
	var food database.Food
	result := r.db.Where("user_id = ?", userID).Order("timestamp DESC").First(&food)
	if result.Error != nil {
		return result.Error
	}
	return r.db.Delete(&food).Error
}

func (r *FoodRepository) GetFoodsByDate(userID int64, date time.Time, timezone string) ([]database.Food, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var foods []database.Food
	result := r.db.Where("user_id = ? AND timestamp >= ? AND timestamp < ?", userID, startOfDay, endOfDay).
		Order("timestamp ASC").Find(&foods)
	return foods, result.Error
}

func (r *FoodRepository) GetDaySummary(userID int64, startOfDay time.Time) (*database.FoodResult, error) {
	endOfDay := startOfDay.Add(24 * time.Hour)
	var result database.FoodResult
	err := r.db.Model(&database.Food{}).
		Select("COALESCE(SUM(calories),0) as total_calories, COALESCE(SUM(fat),0) as total_fat, COALESCE(SUM(carbohydrates),0) as total_carbohydrates, COALESCE(SUM(protein),0) as total_protein, COALESCE(SUM(weight),0) as total_weight").
		Where("user_id = ? AND timestamp >= ? AND timestamp < ?", userID, startOfDay, endOfDay).
		Scan(&result).Error
	return &result, err
}

func (r *FoodRepository) GetFoodHistory(userID int64, period string, timezone string) ([]HistoryRow, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	var dateFormat string
	var startDate time.Time

	switch period {
	case "week":
		startDate = time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, loc)
		dateFormat = "%Y-%m-%d"
	case "month":
		startDate = time.Date(now.Year(), now.Month(), now.Day()-29, 0, 0, 0, 0, loc)
		dateFormat = "%Y-%m-%d"
	case "year":
		startDate = time.Date(now.Year()-1, now.Month()+1, 1, 0, 0, 0, 0, loc)
		dateFormat = "%Y-%m"
	default:
		startDate = time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, loc)
		dateFormat = "%Y-%m-%d"
	}

	var rows []HistoryRow
	query := fmt.Sprintf(`
		SELECT
			DATE_FORMAT(CONVERT_TZ(timestamp, '+00:00', '%s'), '%s') as date_label,
			COALESCE(SUM(calories), 0) as calories,
			COALESCE(SUM(protein), 0) as protein,
			COALESCE(SUM(fat), 0) as fat,
			COALESCE(SUM(carbohydrates), 0) as carbs
		FROM foods
		WHERE user_id = ? AND timestamp >= ? AND deleted_at IS NULL
		GROUP BY date_label
		ORDER BY date_label ASC`,
		timezone, dateFormat)

	err = r.db.Raw(query, userID, startDate).Scan(&rows).Error
	return rows, err
}

type HistoryRow struct {
	DateLabel string  `gorm:"column:date_label"`
	Calories  float64 `gorm:"column:calories"`
	Protein   float64 `gorm:"column:protein"`
	Fat       float64 `gorm:"column:fat"`
	Carbs     float64 `gorm:"column:carbs"`
}

func (r *FoodRepository) GetStartOfDay(userID int64, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}
