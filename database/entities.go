package database

import (
	"time"

	"gorm.io/gorm"
)

// -----------------------------
// Common structures
// -----------------------------

type User struct {
	gorm.Model
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

type UserTimezone struct {
	gorm.Model
	UserID   int64  `json:"user_id"`
	Timezone string `json:"timezone"`
}

// -----------------------------
// Food structures
// -----------------------------

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

// -----------------------------
// Expenses structures
// -----------------------------

type Expense struct {
	gorm.Model
	ExpenseID int     `json:"expense_id"`
	UserID    int64   `json:"user_id"`
	Timestamp int     `json:"timestamp"`
	Item      string  `json:"item"`
	TotalCost float64 `json:"total_cost"`
	Currency  string  `json:"currency"`
	Category  string  `json:"category"`
}

type ExpenseCategory struct {
	gorm.Model
	ExpenseCategoryID int    `json:"expense_category_id"`
	UserID            int64  `json:"user_id"`
	Category          string `json:"category"`
}

type FoodResult struct {
	TotalCalories      float64
	TotalFat           float64
	TotalCarbohydrates float64
	TotalProtein       float64
}
