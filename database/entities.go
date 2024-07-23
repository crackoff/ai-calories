package database

import (
	"gorm.io/gorm"
	"time"
)

type Food struct {
	gorm.Model
	UserID        int64     `json:"user_id"`
	Timestamp     time.Time `json:"timestamp"`
	FoodItem      string    `json:"food_item"`
	Weight        float64   `json:"weight"`
	Calories      float64   `json:"calories"`
	Fat           float64   `json:"fat"`
	Carbohydrates float64   `json:"carbohydrates"`
	Protein       float64   `json:"protein"`
}

type UserTimezone struct {
	gorm.Model
	UserID   int64  `json:"user_id"`
	Timezone string `json:"timezone"`
}
