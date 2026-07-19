package service

import (
	"ai-calories/internal/model"
	"ai-calories/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PaymentService struct {
	repo *repository.PaymentRepository
}

func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) GetCurrent(userID uint) (*model.CurrentPaymentResponse, error) {
	payment, err := s.repo.GetCurrent(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // free plan
	}
	if err != nil {
		return nil, err
	}

	return &model.CurrentPaymentResponse{
		SKU:            payment.SKU,
		PaymentDate:    payment.PaymentDate,
		ExpirationDate: payment.ExpirationDate,
		Amount:         payment.Amount,
	}, nil
}

func (s *PaymentService) Record(userID uint, req model.RecordPaymentRequest) error {
	expiration, err := time.Parse(time.RFC3339, req.ExpirationDate)
	if err != nil {
		return errors.New("invalid expiration_date format, use ISO 8601")
	}

	payment := &model.PaymentHistory{
		UserID:         userID,
		SKU:            req.SKU,
		PaymentDate:    time.Now(),
		ExpirationDate: expiration,
		Amount:         req.Amount,
	}
	return s.repo.Record(payment)
}

func (s *PaymentService) GetHistory(userID uint) ([]model.PaymentHistoryItem, error) {
	payments, err := s.repo.GetHistory(userID)
	if err != nil {
		return nil, err
	}

	var items []model.PaymentHistoryItem
	for _, p := range payments {
		items = append(items, model.PaymentHistoryItem{
			ID:             p.ID,
			SKU:            p.SKU,
			PaymentDate:    p.PaymentDate,
			ExpirationDate: p.ExpirationDate,
			Amount:         p.Amount,
		})
	}
	return items, nil
}
