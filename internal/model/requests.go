package model

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OAuthRequest struct {
	IDToken string `json:"id_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogFoodRequest struct {
	FoodCacheID *uint   `json:"food_cache_id,omitempty"`
	FreeText    *string `json:"free_text,omitempty"`
	InputMode   string  `json:"input_mode"` // "grams" or "kcal"
	Value       float64 `json:"value"`
}

type UpdateTimezoneRequest struct {
	Timezone string `json:"timezone"`
}

type UpdateLanguageRequest struct {
	Language string `json:"language"`
}

type RecordPaymentRequest struct {
	SKU            string  `json:"sku"`
	Amount         float64 `json:"amount"`
	ExpirationDate string  `json:"expiration_date"` // ISO 8601
}
