package bot

import (
	"ai-calories/database"
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"
)

type mockAuthStore struct {
	getUserFn       func(int64) (database.User, error)
	addUserFn       func(int64, string) error
	getFoodsCountFn func(int64) (int, error)

	getUserCalls       int
	addUserCalls       int
	getFoodsCountCalls int
}

func (m *mockAuthStore) GetUser(userID int64) (database.User, error) {
	m.getUserCalls++
	if m.getUserFn == nil {
		return database.User{}, nil
	}
	return m.getUserFn(userID)
}

func (m *mockAuthStore) AddUser(userID int64, username string) error {
	m.addUserCalls++
	if m.addUserFn == nil {
		return nil
	}
	return m.addUserFn(userID, username)
}

func (m *mockAuthStore) GetFoodsCount(userID int64) (int, error) {
	m.getFoodsCountCalls++
	if m.getFoodsCountFn == nil {
		return 0, nil
	}
	return m.getFoodsCountFn(userID)
}

func TestCheckAuthorization_AllowsMembersWithoutDBCalls(t *testing.T) {
	store := &mockAuthStore{}
	err := checkAuthorization(store, 1001, "alice", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.getUserCalls != 0 || store.addUserCalls != 0 || store.getFoodsCountCalls != 0 {
		t.Fatalf("unexpected db calls: %+v", store)
	}
}

func TestCheckAuthorization_CreatesMissingUserAndAllowsUnderLimit(t *testing.T) {
	var addedUserID int64
	var addedUsername string
	store := &mockAuthStore{
		getUserFn: func(userID int64) (database.User, error) {
			return database.User{}, gorm.ErrRecordNotFound
		},
		addUserFn: func(userID int64, username string) error {
			addedUserID = userID
			addedUsername = username
			return nil
		},
		getFoodsCountFn: func(userID int64) (int, error) {
			return 10, nil
		},
	}

	err := checkAuthorization(store, 1002, "bob", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addedUserID != 1002 || addedUsername != "bob" {
		t.Fatalf("unexpected AddUser args: userID=%d username=%q", addedUserID, addedUsername)
	}
	if store.getUserCalls != 1 || store.addUserCalls != 1 || store.getFoodsCountCalls != 1 {
		t.Fatalf("unexpected db calls: %+v", store)
	}
}

func TestCheckAuthorization_DeniesWhenFoodCountExceedsLimit(t *testing.T) {
	store := &mockAuthStore{
		getUserFn: func(userID int64) (database.User, error) {
			return database.User{}, nil
		},
		getFoodsCountFn: func(userID int64) (int, error) {
			return 11, nil
		},
	}

	err := checkAuthorization(store, 1003, "carol", false)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "too many requests") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCheckAuthorization_PropagatesGetUserError(t *testing.T) {
	expectedErr := errors.New("get user failed")
	store := &mockAuthStore{
		getUserFn: func(userID int64) (database.User, error) {
			return database.User{}, expectedErr
		},
	}

	err := checkAuthorization(store, 1004, "dave", false)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("got err=%v, want %v", err, expectedErr)
	}
	if store.getFoodsCountCalls != 0 {
		t.Fatalf("expected GetFoodsCount to not be called")
	}
}

func TestCheckAuthorization_PropagatesAddUserError(t *testing.T) {
	expectedErr := errors.New("add user failed")
	store := &mockAuthStore{
		getUserFn: func(userID int64) (database.User, error) {
			return database.User{}, gorm.ErrRecordNotFound
		},
		addUserFn: func(userID int64, username string) error {
			return expectedErr
		},
	}

	err := checkAuthorization(store, 1005, "erin", false)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("got err=%v, want %v", err, expectedErr)
	}
	if store.getFoodsCountCalls != 0 {
		t.Fatalf("expected GetFoodsCount to not be called")
	}
}

func TestCheckAuthorization_PropagatesGetFoodsCountError(t *testing.T) {
	expectedErr := errors.New("count failed")
	store := &mockAuthStore{
		getUserFn: func(userID int64) (database.User, error) {
			return database.User{}, nil
		},
		getFoodsCountFn: func(userID int64) (int, error) {
			return 0, expectedErr
		},
	}

	err := checkAuthorization(store, 1006, "frank", false)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("got err=%v, want %v", err, expectedErr)
	}
}

func TestEscapeMarkdownV2_EscapesAllTelegramMarkdownV2SpecialChars(t *testing.T) {
	input := "_[]()~`>#+-=|{}.!"
	var expectedBuilder strings.Builder
	for _, ch := range input {
		expectedBuilder.WriteString("\\")
		expectedBuilder.WriteRune(ch)
	}
	expected := expectedBuilder.String()

	got := escapeMarkdownV2(input)
	if got != expected {
		t.Fatalf("unexpected escaped output:\n got: %q\nwant: %q", got, expected)
	}
}
