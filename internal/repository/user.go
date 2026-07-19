package repository

import (
	"ai-calories/database"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*database.User, error) {
	var user database.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*database.User, error) {
	var user database.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) Create(user *database.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) UpdateTimezone(userID int64, timezone string) error {
	var tz database.UserTimezone
	result := r.db.Where("user_id = ?", userID).FirstOrCreate(&tz, database.UserTimezone{UserID: userID, Timezone: "UTC"})
	if result.Error != nil {
		return result.Error
	}
	return r.db.Model(&tz).Update("timezone", timezone).Error
}

func (r *UserRepository) UpdateLanguage(id uint, language string) error {
	return r.db.Model(&database.User{}).Where("id = ?", id).Update("language", language).Error
}

func (r *UserRepository) GetTimezone(userID int64) (string, error) {
	var tz database.UserTimezone
	result := r.db.Where("user_id = ?", userID).FirstOrCreate(&tz, database.UserTimezone{UserID: userID, Timezone: "UTC"})
	if result.Error != nil {
		return "UTC", result.Error
	}
	return tz.Timezone, nil
}
