package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}
	return gormDB, mock
}

func TestFoodCacheRepository_Search_ReturnsResults(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewFoodCacheRepository(db)

	cols := []string{"id", "food_name", "calories_100g", "protein_100g", "fat_100g", "carbs_100g", "source"}
	rows := sqlmock.NewRows(cols).
		AddRow(1, "banana", 89.0, 1.1, 0.3, 23.0, "ai").
		AddRow(2, "banana bread", 265.0, 4.0, 11.0, 36.0, "ai")

	mock.ExpectQuery("SELECT .* FROM `food_caches`").
		WillReturnRows(rows)

	results, err := repo.Search("banana", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].FoodName != "banana" {
		t.Fatalf("got food_name=%q, want %q", results[0].FoodName, "banana")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestFoodCacheRepository_Search_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewFoodCacheRepository(db)

	mock.ExpectQuery("SELECT .* FROM `food_caches`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "food_name"}))

	results, err := repo.Search("zzz", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestFoodCacheRepository_FindByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewFoodCacheRepository(db)

	cols := []string{"id", "food_name", "calories_100g", "protein_100g", "fat_100g", "carbs_100g", "source"}
	mock.ExpectQuery("SELECT .* FROM `food_caches`").
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(5, "apple", 52.0, 0.3, 0.2, 14.0, "ai"))

	item, err := repo.FindByID(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.FoodName != "apple" {
		t.Fatalf("got food_name=%q, want %q", item.FoodName, "apple")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestFoodCacheRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewFoodCacheRepository(db)

	mock.ExpectQuery("SELECT .* FROM `food_caches`").
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(99)
	if err == nil {
		t.Fatal("expected error for missing record, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
