package bot

import (
	"ai-calories/database"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/matheusoliveira/go-ordered-map/omap"
)

type mockTodayTotalsStore struct {
	getTodayNutritionFn func(int64) (database.FoodResult, error)
	getFoodGroupsFn     func(int64, string) (string, error)

	getTodayNutritionCalls int
	getFoodGroupsCalls     int
}

func (m *mockTodayTotalsStore) GetTodayNutrition(userID int64) (database.FoodResult, error) {
	m.getTodayNutritionCalls++
	if m.getTodayNutritionFn == nil {
		return database.FoodResult{}, nil
	}
	return m.getTodayNutritionFn(userID)
}

func (m *mockTodayTotalsStore) GetFoodGroups(userID int64, lang string) (string, error) {
	m.getFoodGroupsCalls++
	if m.getFoodGroupsFn == nil {
		return "", nil
	}
	return m.getFoodGroupsFn(userID, lang)
}

func setDrawPieChartFn(t *testing.T, fn func(omap.OMap[string, float64]) (bytes.Buffer, error)) {
	t.Helper()
	original := drawPieChartFn
	drawPieChartFn = fn
	t.Cleanup(func() {
		drawPieChartFn = original
	})
}

func TestGetTodayTotal_IncludesWeightInFormattedMessage(t *testing.T) {
	setDrawPieChartFn(t, func(values omap.OMap[string, float64]) (bytes.Buffer, error) {
		var img bytes.Buffer
		_, _ = img.WriteString("png")
		return img, nil
	})

	store := &mockTodayTotalsStore{
		getTodayNutritionFn: func(userID int64) (database.FoodResult, error) {
			return database.FoodResult{
				TotalCalories:      1800,
				TotalFat:           65,
				TotalCarbohydrates: 210,
				TotalProtein:       120,
				TotalWeight:        1450,
			}, nil
		},
		getFoodGroupsFn: func(userID int64, lang string) (string, error) {
			return "*Morning* eggs\n", nil
		},
	}

	bot := &FoodBot{}
	message, image := bot.getTodayTotal(store, 42, "en")

	if !strings.Contains(message, "Weight") {
		t.Fatalf("expected message to include weight label, got %q", message)
	}
	if !strings.Contains(message, "1450.00g.") {
		t.Fatalf("expected message to include weight value, got %q", message)
	}
	if image.Len() == 0 {
		t.Fatalf("expected non-empty chart image")
	}
}

func TestGetTodayTotal_IncludesGroupedFoodsText(t *testing.T) {
	setDrawPieChartFn(t, func(values omap.OMap[string, float64]) (bytes.Buffer, error) {
		var img bytes.Buffer
		_, _ = img.WriteString("png")
		return img, nil
	})
	grouped := "*Morning* eggs, toast\n*Afternoon* rice\n"

	store := &mockTodayTotalsStore{
		getTodayNutritionFn: func(userID int64) (database.FoodResult, error) {
			return database.FoodResult{TotalCalories: 1000, TotalWeight: 800}, nil
		},
		getFoodGroupsFn: func(userID int64, lang string) (string, error) {
			return grouped, nil
		},
	}

	bot := &FoodBot{}
	message, _ := bot.getTodayTotal(store, 42, "en")
	if !strings.Contains(message, grouped) {
		t.Fatalf("expected grouped foods in message, got %q", message)
	}
}

func TestGetTodayTotal_ReturnsEmptyOnGetTodayNutritionError(t *testing.T) {
	setDrawPieChartFn(t, func(values omap.OMap[string, float64]) (bytes.Buffer, error) {
		t.Fatal("drawPieChart should not be called")
		return bytes.Buffer{}, nil
	})

	store := &mockTodayTotalsStore{
		getTodayNutritionFn: func(userID int64) (database.FoodResult, error) {
			return database.FoodResult{}, errors.New("nutrition failed")
		},
		getFoodGroupsFn: func(userID int64, lang string) (string, error) {
			t.Fatal("GetFoodGroups should not be called")
			return "", nil
		},
	}

	bot := &FoodBot{}
	message, image := bot.getTodayTotal(store, 42, "en")
	if message != "" || image.Len() != 0 {
		t.Fatalf("expected empty result, got message=%q imageLen=%d", message, image.Len())
	}
	if store.getTodayNutritionCalls != 1 {
		t.Fatalf("expected one GetTodayNutrition call, got %d", store.getTodayNutritionCalls)
	}
	if store.getFoodGroupsCalls != 0 {
		t.Fatalf("expected zero GetFoodGroups calls, got %d", store.getFoodGroupsCalls)
	}
}

func TestGetTodayTotal_ReturnsEmptyOnGetFoodGroupsError(t *testing.T) {
	setDrawPieChartFn(t, func(values omap.OMap[string, float64]) (bytes.Buffer, error) {
		t.Fatal("drawPieChart should not be called")
		return bytes.Buffer{}, nil
	})

	store := &mockTodayTotalsStore{
		getTodayNutritionFn: func(userID int64) (database.FoodResult, error) {
			return database.FoodResult{TotalCalories: 900, TotalWeight: 700}, nil
		},
		getFoodGroupsFn: func(userID int64, lang string) (string, error) {
			return "", errors.New("groups failed")
		},
	}

	bot := &FoodBot{}
	message, image := bot.getTodayTotal(store, 42, "en")
	if message != "" || image.Len() != 0 {
		t.Fatalf("expected empty result, got message=%q imageLen=%d", message, image.Len())
	}
	if store.getTodayNutritionCalls != 1 {
		t.Fatalf("expected one GetTodayNutrition call, got %d", store.getTodayNutritionCalls)
	}
	if store.getFoodGroupsCalls != 1 {
		t.Fatalf("expected one GetFoodGroups call, got %d", store.getFoodGroupsCalls)
	}
}

func TestGetTodayTotal_ReturnsEmptyOnChartError(t *testing.T) {
	setDrawPieChartFn(t, func(values omap.OMap[string, float64]) (bytes.Buffer, error) {
		return bytes.Buffer{}, errors.New("chart failed")
	})

	store := &mockTodayTotalsStore{
		getTodayNutritionFn: func(userID int64) (database.FoodResult, error) {
			return database.FoodResult{
				TotalCalories:      1500,
				TotalFat:           50,
				TotalCarbohydrates: 180,
				TotalProtein:       100,
				TotalWeight:        1200,
			}, nil
		},
		getFoodGroupsFn: func(userID int64, lang string) (string, error) {
			return "*Evening* soup\n", nil
		},
	}

	bot := &FoodBot{}
	message, image := bot.getTodayTotal(store, 42, "en")
	if message != "" || image.Len() != 0 {
		t.Fatalf("expected empty result, got message=%q imageLen=%d", message, image.Len())
	}
}
