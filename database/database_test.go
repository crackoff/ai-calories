package database

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newSQLMockDB(t *testing.T) (*Database, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}

	return &Database{DB: gormDB}, mock
}

func TestGetFoodsCount_ReturnsCount(t *testing.T) {
	db, mock := newSQLMockDB(t)
	userID := int64(11)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `foods`.*user_id = \\?.*`foods`\\.`deleted_at` IS NULL").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	count, err := db.GetFoodsCount(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 7 {
		t.Fatalf("got count=%d, want 7", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetFoodsCount_PropagatesDBError(t *testing.T) {
	db, mock := newSQLMockDB(t)
	userID := int64(12)
	expectedErr := errors.New("count query failed")

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `foods`.*user_id = \\?.*`foods`\\.`deleted_at` IS NULL").
		WithArgs(userID).
		WillReturnError(expectedErr)

	_, err := db.GetFoodsCount(userID)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("got err=%v, want %v", err, expectedErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetTodayNutrition_ReturnsAggregatesIncludingWeight(t *testing.T) {
	db, mock := newSQLMockDB(t)
	userID := int64(21)

	mock.ExpectQuery("SELECT .*FROM `user_timezones`.*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "timezone"}).AddRow(1, userID, "UTC"))

	mock.ExpectQuery("SELECT .*SUM\\(weight\\) AS total_weight.*FROM `foods`.*user_id = \\?.*timestamp >= \\?.*").
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_calories",
			"total_fat",
			"total_carbohydrates",
			"total_protein",
			"total_weight",
		}).AddRow(1500.0, 55.0, 170.0, 120.0, 1250.0))

	result, err := db.GetTodayNutrition(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCalories != 1500 {
		t.Fatalf("got total_calories=%v, want 1500", result.TotalCalories)
	}
	if result.TotalFat != 55 {
		t.Fatalf("got total_fat=%v, want 55", result.TotalFat)
	}
	if result.TotalCarbohydrates != 170 {
		t.Fatalf("got total_carbs=%v, want 170", result.TotalCarbohydrates)
	}
	if result.TotalProtein != 120 {
		t.Fatalf("got total_protein=%v, want 120", result.TotalProtein)
	}
	if result.TotalWeight != 1250 {
		t.Fatalf("got total_weight=%v, want 1250", result.TotalWeight)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetTodayNutrition_PropagatesQueryError(t *testing.T) {
	db, mock := newSQLMockDB(t)
	userID := int64(22)
	expectedErr := errors.New("aggregate query failed")

	mock.ExpectQuery("SELECT .*FROM `user_timezones`.*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "timezone"}).AddRow(1, userID, "UTC"))

	mock.ExpectQuery("SELECT .*SUM\\(weight\\) AS total_weight.*FROM `foods`.*user_id = \\?.*timestamp >= \\?.*").
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnError(expectedErr)

	_, err := db.GetTodayNutrition(userID)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("got err=%v, want %v", err, expectedErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
