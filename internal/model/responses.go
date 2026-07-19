package model

import "time"

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type FoodEntryResponse struct {
	ID            uint      `json:"id"`
	FoodItem      string    `json:"food_item"`
	Weight        float64   `json:"weight"`
	Calories      float64   `json:"calories"`
	Protein       float64   `json:"protein"`
	Fat           float64   `json:"fat"`
	Carbohydrates float64   `json:"carbohydrates"`
	FromCache     bool      `json:"from_cache"`
	Timestamp     time.Time `json:"timestamp"`
}

type FoodCacheSearchResult struct {
	ID           uint    `json:"id"`
	FoodName     string  `json:"food_name"`
	Calories100g float64 `json:"calories_100g"`
	ImageURL     *string `json:"image_url"`
}

type NutritionSummary struct {
	Date               string          `json:"date"`
	TotalCalories      float64         `json:"total_calories"`
	TotalProtein       float64         `json:"total_protein"`
	TotalFat           float64         `json:"total_fat"`
	TotalCarbohydrates float64         `json:"total_carbohydrates"`
	Meals              []MealGroup     `json:"meals"`
	MacrosBreakdown    MacrosBreakdown `json:"macros_breakdown"`
}

type MealGroup struct {
	Period  string              `json:"period"`
	Entries []FoodEntryResponse `json:"entries"`
}

type MacrosBreakdown struct {
	ProteinPct float64 `json:"protein_pct"`
	FatPct     float64 `json:"fat_pct"`
	CarbsPct   float64 `json:"carbs_pct"`
}

type HistoryDataPoint struct {
	Date     string  `json:"date"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type FoodHistoryResponse struct {
	Period string             `json:"period"`
	Data   []HistoryDataPoint `json:"data"`
}

type UserProfileResponse struct {
	ID           uint    `json:"id"`
	Email        *string `json:"email"`
	AuthProvider *string `json:"auth_provider"`
	Language     string  `json:"language"`
	Timezone     string  `json:"timezone"`
}

type CurrentPaymentResponse struct {
	SKU            string    `json:"sku"`
	PaymentDate    time.Time `json:"payment_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	Amount         float64   `json:"amount"`
}

type PaymentHistoryItem struct {
	ID             uint      `json:"id"`
	SKU            string    `json:"sku"`
	PaymentDate    time.Time `json:"payment_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	Amount         float64   `json:"amount"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
