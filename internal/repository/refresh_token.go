package repository

import (
	"ai-calories/internal/model"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Save(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) FindByToken(token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	result := r.db.Where("token = ?", token).First(&rt)
	if result.Error != nil {
		return nil, result.Error
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}

func (r *RefreshTokenRepository) DeleteAllForUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}
