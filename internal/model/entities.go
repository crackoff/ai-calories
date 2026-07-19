package model

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"size:512;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
}

type FoodCache struct {
	gorm.Model
	FoodName     string  `json:"food_name" gorm:"size:255;not null;uniqueIndex"`
	Calories100g float64 `json:"calories_100g" gorm:"column:calories_100g;not null"`
	Protein100g  float64 `json:"protein_100g" gorm:"column:protein_100g;not null"`
	Fat100g      float64 `json:"fat_100g" gorm:"column:fat_100g;not null"`
	Carbs100g    float64 `json:"carbs_100g" gorm:"column:carbs_100g;not null"`
	ImageURL     *string `json:"image_url" gorm:"size:512"`
	Source       string  `json:"source" gorm:"size:10;default:ai"`
}

type PaymentHistory struct {
	gorm.Model
	UserID         uint      `json:"user_id" gorm:"not null;index"`
	SKU            string    `json:"sku" gorm:"size:12;not null"`
	PaymentDate    time.Time `json:"payment_date" gorm:"not null"`
	ExpirationDate time.Time `json:"expiration_date" gorm:"not null"`
	Amount         float64   `json:"amount" gorm:"type:decimal(10,2);not null"`
}
