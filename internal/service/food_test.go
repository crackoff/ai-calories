package service

import (
	"ai-calories/internal/model"
	"math"
	"testing"
)

func requireClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %.8f, want %.8f", got, want)
	}
}

func newCache(calories, protein, fat, carbs float64) *model.FoodCache {
	return &model.FoodCache{
		Calories100g: calories,
		Protein100g:  protein,
		Fat100g:      fat,
		Carbs100g:    carbs,
	}
}

// ---- grams mode ----

func TestCalculateNutrition_GramsMode_FullPortion(t *testing.T) {
	// 200g of food with 100kcal/100g
	cache := newCache(100, 20, 5, 10)
	grams, calories, protein, fat, carbs := calculateNutrition(cache, "grams", 200)

	requireClose(t, grams, 200)
	requireClose(t, calories, 200)
	requireClose(t, protein, 40)
	requireClose(t, fat, 10)
	requireClose(t, carbs, 20)
}

func TestCalculateNutrition_GramsMode_HalfPortion(t *testing.T) {
	cache := newCache(400, 30, 10, 50)
	grams, calories, protein, fat, carbs := calculateNutrition(cache, "grams", 50)

	requireClose(t, grams, 50)
	requireClose(t, calories, 200)
	requireClose(t, protein, 15)
	requireClose(t, fat, 5)
	requireClose(t, carbs, 25)
}

// ---- kcal mode ----

func TestCalculateNutrition_KcalMode_DerivedGrams(t *testing.T) {
	// 330 kcal of food that has 200 kcal/100g
	cache := newCache(200, 10, 8, 30)
	grams, calories, protein, fat, carbs := calculateNutrition(cache, "kcal", 330)

	requireClose(t, grams, 165)  // (330/200)*100
	requireClose(t, calories, 330)
	requireClose(t, protein, 16.5)  // (165/100)*10
	requireClose(t, fat, 13.2)      // (165/100)*8
	requireClose(t, carbs, 49.5)    // (165/100)*30
}

func TestCalculateNutrition_KcalMode_ZeroCalories100g(t *testing.T) {
	// Edge case: zero calories_100g → grams stays 0
	cache := newCache(0, 0, 0, 0)
	grams, calories, protein, fat, carbs := calculateNutrition(cache, "kcal", 100)

	requireClose(t, grams, 0)
	requireClose(t, calories, 100)
	requireClose(t, protein, 0)
	requireClose(t, fat, 0)
	requireClose(t, carbs, 0)
}

// ---- default (unknown mode falls back to grams) ----

func TestCalculateNutrition_UnknownModeFallsBackToGrams(t *testing.T) {
	cache := newCache(100, 5, 2, 15)
	grams, calories, protein, fat, carbs := calculateNutrition(cache, "unknown", 100)

	requireClose(t, grams, 100)
	requireClose(t, calories, 100)
	requireClose(t, protein, 5)
	requireClose(t, fat, 2)
	requireClose(t, carbs, 15)
}
