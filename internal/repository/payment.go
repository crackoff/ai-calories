package repository

import (
	"ai-calories/internal/model"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) GetCurrent(userID uint) (*model.PaymentHistory, error) {
	var payment model.PaymentHistory
	result := r.db.Where("user_id = ? AND expiration_date > ?", userID, time.Now()).
		Order("expiration_date DESC").
		First(&payment)
	if result.Error != nil {
		return nil, result.Error
	}
	return &payment, nil
}

func (r *PaymentRepository) Record(payment *model.PaymentHistory) error {
	return r.db.Create(payment).Error
}

func (r *PaymentRepository) GetHistory(userID uint) ([]model.PaymentHistory, error) {
	var payments []model.PaymentHistory
	err := r.db.Where("user_id = ?", userID).
		Order("payment_date DESC").
		Find(&payments).Error
	return payments, err
}
