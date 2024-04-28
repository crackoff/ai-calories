package database

type Food struct {
	ID                 int     `json:"id"`
	UserID             int64   `json:"user_id"`
	Timestamp          int64   `json:"timestamp"`
	FoodItem           string  `json:"food_item"`
	TotalWeight        int     `json:"total_weight"`
	Calories           int     `json:"calories"`
	TotalFat           float64 `json:"total_fat"`
	SaturatedFat       float64 `json:"saturated_fat"`
	Cholesterol        float64 `json:"cholesterol"`
	Sodium             float64 `json:"sodium"`
	TotalCarbohydrates float64 `json:"total_carbohydrates"`
	DietaryFiber       float64 `json:"dietary_fiber"`
	Sugars             float64 `json:"sugars"`
	Protein            float64 `json:"protein"`
}
