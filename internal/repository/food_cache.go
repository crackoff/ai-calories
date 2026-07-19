package repository

import (
	"ai-calories/internal/model"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FoodCacheRepository struct {
	db *gorm.DB
}

func NewFoodCacheRepository(db *gorm.DB) *FoodCacheRepository {
	return &FoodCacheRepository{db: db}
}

func (r *FoodCacheRepository) Search(query string, limit int) ([]model.FoodCache, error) {
	normalized := strings.ToLower(strings.TrimSpace(query))
	var results []model.FoodCache
	err := r.db.Where("food_name LIKE ?", normalized+"%").
		Limit(limit).
		Order("food_name ASC").
		Find(&results).Error
	return results, err
}

func (r *FoodCacheRepository) FindByID(id uint) (*model.FoodCache, error) {
	var fc model.FoodCache
	result := r.db.First(&fc, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &fc, nil
}

func (r *FoodCacheRepository) Upsert(fc *model.FoodCache) error {
	fc.FoodName = strings.ToLower(strings.TrimSpace(fc.FoodName))
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "food_name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"calories_100g", "protein_100g", "fat_100g", "carbs_100g", "source", "updated_at",
		}),
	}).Create(fc).Error
}
