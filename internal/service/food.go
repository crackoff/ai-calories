package service

import (
	"ai-calories/ai"
	"ai-calories/database"
	"ai-calories/internal/model"
	"ai-calories/internal/repository"
	"errors"
	"strings"
	"time"
)

type FoodService struct {
	foodRepo      *repository.FoodRepository
	foodCacheRepo *repository.FoodCacheRepository
	userRepo      *repository.UserRepository
	classifier    ai.Classifier
}

func NewFoodService(
	foodRepo *repository.FoodRepository,
	foodCacheRepo *repository.FoodCacheRepository,
	userRepo *repository.UserRepository,
	classifier ai.Classifier,
) *FoodService {
	return &FoodService{
		foodRepo:      foodRepo,
		foodCacheRepo: foodCacheRepo,
		userRepo:      userRepo,
		classifier:    classifier,
	}
}

func (s *FoodService) LogFood(userID int64, req model.LogFoodRequest) (*model.FoodEntryResponse, error) {
	if req.FoodCacheID != nil {
		return s.logFromCache(userID, *req.FoodCacheID, req.InputMode, req.Value)
	}
	if req.FreeText != nil {
		return s.logFromAI(userID, *req.FreeText, req.InputMode, req.Value)
	}
	return nil, errors.New("either food_cache_id or free_text is required")
}

func (s *FoodService) logFromCache(userID int64, cacheID uint, inputMode string, value float64) (*model.FoodEntryResponse, error) {
	cached, err := s.foodCacheRepo.FindByID(cacheID)
	if err != nil {
		return nil, errors.New("food not found in cache")
	}

	grams, calories, protein, fat, carbs := calculateNutrition(cached, inputMode, value)

	food := &database.Food{
		UserID:        userID,
		Timestamp:     time.Now(),
		FoodItem:      cached.FoodName,
		Weight:        grams,
		Calories:      calories,
		Protein:       protein,
		Fat:           fat,
		Carbohydrates: carbs,
	}
	if err := s.foodRepo.InsertFood(food); err != nil {
		return nil, err
	}

	return &model.FoodEntryResponse{
		ID:            food.ID,
		FoodItem:      food.FoodItem,
		Weight:        food.Weight,
		Calories:      food.Calories,
		Protein:       food.Protein,
		Fat:           food.Fat,
		Carbohydrates: food.Carbohydrates,
		FromCache:     true,
		Timestamp:     food.Timestamp,
	}, nil
}

func (s *FoodService) logFromAI(userID int64, freeText string, inputMode string, value float64) (*model.FoodEntryResponse, error) {
	foodClassifier, ok := s.classifier.(ai.FoodClassifier)
	if !ok {
		return nil, errors.New("classifier is not a food classifier")
	}

	// Ask AI for nutrition data for 100g of the food
	aiResult, err := foodClassifier.GetNutritionData("100g of " + freeText)
	if err != nil {
		return nil, err
	}

	// Save to cache for future use
	cached := &model.FoodCache{
		FoodName:     strings.ToLower(strings.TrimSpace(freeText)),
		Calories100g: aiResult.Calories,
		Protein100g:  aiResult.Protein,
		Fat100g:      aiResult.Fat,
		Carbs100g:    aiResult.Carbohydrates,
		Source:       "ai",
	}
	_ = s.foodCacheRepo.Upsert(cached)

	// Calculate actual values
	grams, calories, protein, fat, carbs := calculateNutrition(cached, inputMode, value)

	food := &database.Food{
		UserID:        userID,
		Timestamp:     time.Now(),
		FoodItem:      aiResult.FoodItem,
		Weight:        grams,
		Calories:      calories,
		Protein:       protein,
		Fat:           fat,
		Carbohydrates: carbs,
	}
	if err := s.foodRepo.InsertFood(food); err != nil {
		return nil, err
	}

	return &model.FoodEntryResponse{
		ID:            food.ID,
		FoodItem:      food.FoodItem,
		Weight:        food.Weight,
		Calories:      food.Calories,
		Protein:       food.Protein,
		Fat:           food.Fat,
		Carbohydrates: food.Carbohydrates,
		FromCache:     false,
		Timestamp:     food.Timestamp,
	}, nil
}

func calculateNutrition(cached *model.FoodCache, inputMode string, value float64) (grams, calories, protein, fat, carbs float64) {
	switch inputMode {
	case "kcal":
		if cached.Calories100g > 0 {
			grams = (value / cached.Calories100g) * 100
		}
		calories = value
	default: // "grams"
		grams = value
		calories = (value / 100) * cached.Calories100g
	}

	protein = (grams / 100) * cached.Protein100g
	fat = (grams / 100) * cached.Fat100g
	carbs = (grams / 100) * cached.Carbs100g
	return
}

func (s *FoodService) GetTodaySummary(userID int64) (*model.NutritionSummary, error) {
	tz, _ := s.userRepo.GetTimezone(userID)
	startOfDay := s.foodRepo.GetStartOfDay(userID, tz)
	return s.getSummaryForDay(userID, startOfDay, tz)
}

func (s *FoodService) GetDateSummary(userID int64, date time.Time) (*model.NutritionSummary, error) {
	tz, _ := s.userRepo.GetTimezone(userID)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	return s.getSummaryForDay(userID, startOfDay, tz)
}

func (s *FoodService) getSummaryForDay(userID int64, startOfDay time.Time, timezone string) (*model.NutritionSummary, error) {
	totals, err := s.foodRepo.GetDaySummary(userID, startOfDay)
	if err != nil {
		return nil, err
	}

	foods, err := s.foodRepo.GetFoodsByDate(userID, startOfDay, timezone)
	if err != nil {
		return nil, err
	}

	// Group foods by meal time
	mealGroups := map[string][]model.FoodEntryResponse{
		"Morning":   {},
		"Afternoon": {},
		"Evening":   {},
	}

	loc, _ := time.LoadLocation(timezone)
	for _, f := range foods {
		entry := model.FoodEntryResponse{
			ID:            f.ID,
			FoodItem:      f.FoodItem,
			Weight:        f.Weight,
			Calories:      f.Calories,
			Protein:       f.Protein,
			Fat:           f.Fat,
			Carbohydrates: f.Carbohydrates,
			Timestamp:     f.Timestamp,
		}
		hour := f.Timestamp.In(loc).Hour()
		switch {
		case hour < 12:
			mealGroups["Morning"] = append(mealGroups["Morning"], entry)
		case hour < 18:
			mealGroups["Afternoon"] = append(mealGroups["Afternoon"], entry)
		default:
			mealGroups["Evening"] = append(mealGroups["Evening"], entry)
		}
	}

	meals := []model.MealGroup{
		{Period: "Morning", Entries: mealGroups["Morning"]},
		{Period: "Afternoon", Entries: mealGroups["Afternoon"]},
		{Period: "Evening", Entries: mealGroups["Evening"]},
	}

	// Calculate macro percentages
	totalGramsMacros := totals.TotalProtein + totals.TotalFat + totals.TotalCarbohydrates
	var macros model.MacrosBreakdown
	if totalGramsMacros > 0 {
		macros.ProteinPct = (totals.TotalProtein / totalGramsMacros) * 100
		macros.FatPct = (totals.TotalFat / totalGramsMacros) * 100
		macros.CarbsPct = (totals.TotalCarbohydrates / totalGramsMacros) * 100
	}

	return &model.NutritionSummary{
		Date:               startOfDay.Format("2006-01-02"),
		TotalCalories:      totals.TotalCalories,
		TotalProtein:       totals.TotalProtein,
		TotalFat:           totals.TotalFat,
		TotalCarbohydrates: totals.TotalCarbohydrates,
		Meals:              meals,
		MacrosBreakdown:    macros,
	}, nil
}

func (s *FoodService) GetFoodsByDate(userID int64, date time.Time) ([]model.FoodEntryResponse, error) {
	tz, _ := s.userRepo.GetTimezone(userID)
	foods, err := s.foodRepo.GetFoodsByDate(userID, date, tz)
	if err != nil {
		return nil, err
	}

	var result []model.FoodEntryResponse
	for _, f := range foods {
		result = append(result, model.FoodEntryResponse{
			ID:            f.ID,
			FoodItem:      f.FoodItem,
			Weight:        f.Weight,
			Calories:      f.Calories,
			Protein:       f.Protein,
			Fat:           f.Fat,
			Carbohydrates: f.Carbohydrates,
			Timestamp:     f.Timestamp,
		})
	}
	return result, nil
}

func (s *FoodService) GetFoodHistory(userID int64, period string) (*model.FoodHistoryResponse, error) {
	tz, _ := s.userRepo.GetTimezone(userID)
	rows, err := s.foodRepo.GetFoodHistory(userID, period, tz)
	if err != nil {
		return nil, err
	}

	var data []model.HistoryDataPoint
	for _, r := range rows {
		data = append(data, model.HistoryDataPoint{
			Date:     r.DateLabel,
			Calories: r.Calories,
			Protein:  r.Protein,
			Fat:      r.Fat,
			Carbs:    r.Carbs,
		})
	}

	return &model.FoodHistoryResponse{
		Period: period,
		Data:   data,
	}, nil
}

func (s *FoodService) DeleteFood(userID int64, foodID uint) error {
	return s.foodRepo.DeleteByID(foodID, userID)
}

func (s *FoodService) DeleteLastFood(userID int64) error {
	return s.foodRepo.DeleteLast(userID)
}
