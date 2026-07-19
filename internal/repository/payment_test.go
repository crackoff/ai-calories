package repository

import (
	"ai-calories/internal/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func TestPaymentRepository_GetCurrent_ActivePayment(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewPaymentRepository(db)

	future := time.Now().Add(30 * 24 * time.Hour)
	cols := []string{"id", "user_id", "sku", "payment_date", "expiration_date", "amount"}
	mock.ExpectQuery("SELECT .* FROM `payment_histories`").
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, 10, "premium_monthly", time.Now(), future, 9.99))

	payment, err := repo.GetCurrent(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.SKU != "premium_monthly" {
		t.Fatalf("got sku=%q, want %q", payment.SKU, "premium_monthly")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPaymentRepository_GetCurrent_NoPlan(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectQuery("SELECT .* FROM `payment_histories`").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	payment, err := repo.GetCurrent(99)
	if err == nil {
		t.Fatal("expected ErrRecordNotFound, got nil")
	}
	if payment != nil {
		t.Fatal("expected nil payment for no active plan")
	}
	if !isNotFound(err) {
		t.Fatalf("expected record-not-found error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPaymentRepository_Record_Inserts(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `payment_histories`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	p := &model.PaymentHistory{
		UserID:         5,
		SKU:            "premium_yearly",
		PaymentDate:    time.Now(),
		ExpirationDate: time.Now().Add(365 * 24 * time.Hour),
		Amount:         79.99,
	}
	if err := repo.Record(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPaymentRepository_GetHistory_ReturnsAll(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewPaymentRepository(db)

	now := time.Now()
	cols := []string{"id", "user_id", "sku", "payment_date", "expiration_date", "amount"}
	mock.ExpectQuery("SELECT .* FROM `payment_histories`").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(2, 7, "premium_monthly", now, now.Add(30*24*time.Hour), 9.99).
			AddRow(1, 7, "premium_monthly", now.Add(-31*24*time.Hour), now.Add(-24*time.Hour), 9.99))

	history, err := repo.GetHistory(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("got %d entries, want 2", len(history))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

// isNotFound checks if the error is a GORM record-not-found.
func isNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
