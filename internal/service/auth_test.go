package service

import (
	"ai-calories/database"
	"ai-calories/internal/config"
	"ai-calories/internal/model"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

// ---- mock UserRepo ----

type mockUserRepo struct {
	findByEmailFn func(email string) (*database.User, error)
	findByIDFn    func(id uint) (*database.User, error)
	createFn      func(user *database.User) error
}

func (m *mockUserRepo) FindByEmail(email string) (*database.User, error) {
	return m.findByEmailFn(email)
}
func (m *mockUserRepo) FindByID(id uint) (*database.User, error) {
	return m.findByIDFn(id)
}
func (m *mockUserRepo) Create(user *database.User) error {
	return m.createFn(user)
}

// ---- mock RefreshTokenRepo ----

type mockRefreshTokenRepo struct {
	savedTokens   []*model.RefreshToken
	saveErr       error
	findByTokenFn func(token string) (*model.RefreshToken, error)
	deleteErr     error
}

func (m *mockRefreshTokenRepo) Save(token *model.RefreshToken) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedTokens = append(m.savedTokens, token)
	return nil
}
func (m *mockRefreshTokenRepo) FindByToken(token string) (*model.RefreshToken, error) {
	return m.findByTokenFn(token)
}
func (m *mockRefreshTokenRepo) DeleteByToken(token string) error {
	return m.deleteErr
}

// ---- helpers ----

func testCfg() config.Config {
	return config.Config{
		JWTSecret:     "unit-test-secret",
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 720 * time.Hour,
	}
}

func newAuthSvc(u UserRepo, rt RefreshTokenRepo) *AuthService {
	return NewAuthService(u, rt, testCfg())
}

// ---- Register tests ----

func TestAuthService_Register_Success(t *testing.T) {
	email := "test@example.com"
	userRepo := &mockUserRepo{
		findByEmailFn: func(e string) (*database.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createFn: func(u *database.User) error {
			u.Model.ID = 7
			return nil
		},
	}
	rtRepo := &mockRefreshTokenRepo{}

	resp, err := newAuthSvc(userRepo, rtRepo).Register(email, "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Fatal("expected non-empty refresh token")
	}
	if len(rtRepo.savedTokens) != 1 {
		t.Fatalf("expected 1 saved refresh token, got %d", len(rtRepo.savedTokens))
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	email := "existing@example.com"
	userRepo := &mockUserRepo{
		findByEmailFn: func(e string) (*database.User, error) {
			return &database.User{Email: &email}, nil
		},
	}
	rtRepo := &mockRefreshTokenRepo{}

	_, err := newAuthSvc(userRepo, rtRepo).Register(email, "pass")
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if err.Error() != "email already registered" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

// ---- Login tests ----

func TestAuthService_Login_ValidCredentials(t *testing.T) {
	// Pre-hash the password the same way Register does.
	svc := newAuthSvc(nil, nil)
	_ = svc // just to trigger package import; we'll use the service directly

	// Register first to get a real hash, then Login against it.
	email := "user@example.com"
	password := "securepass"

	var createdUser database.User
	userRepo := &mockUserRepo{
		findByEmailFn: func(e string) (*database.User, error) {
			if createdUser.Email == nil {
				return nil, gorm.ErrRecordNotFound
			}
			return &createdUser, nil
		},
		createFn: func(u *database.User) error {
			createdUser = *u
			createdUser.Model.ID = 5
			return nil
		},
	}
	rtRepo := &mockRefreshTokenRepo{}
	authSvc := newAuthSvc(userRepo, rtRepo)

	if _, err := authSvc.Register(email, password); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	// Now login
	resp, err := authSvc.Login(email, password)
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected non-empty access token after login")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	email := "user@example.com"
	password := "correct"

	var createdUser database.User
	userRepo := &mockUserRepo{
		findByEmailFn: func(e string) (*database.User, error) {
			if createdUser.Email == nil {
				return nil, gorm.ErrRecordNotFound
			}
			return &createdUser, nil
		},
		createFn: func(u *database.User) error {
			createdUser = *u
			createdUser.Model.ID = 6
			return nil
		},
	}
	rtRepo := &mockRefreshTokenRepo{}
	authSvc := newAuthSvc(userRepo, rtRepo)

	if _, err := authSvc.Register(email, password); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err := authSvc.Login(email, "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestAuthService_Login_SocialOnlyAccount(t *testing.T) {
	email := "google@example.com"
	provider := "google"
	userRepo := &mockUserRepo{
		findByEmailFn: func(e string) (*database.User, error) {
			return &database.User{
				Email:        &email,
				AuthProvider: &provider,
				Password:     nil, // no password set
			}, nil
		},
	}
	rtRepo := &mockRefreshTokenRepo{}

	_, err := newAuthSvc(userRepo, rtRepo).Login(email, "any")
	if err == nil {
		t.Fatal("expected error for social-only account, got nil")
	}
	if err.Error() != "account uses social login" {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

// ---- Refresh tests ----

func TestAuthService_Refresh_ValidToken(t *testing.T) {
	stored := &model.RefreshToken{
		UserID:    10,
		Token:     "valid-refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	userRepo := &mockUserRepo{}
	rtRepo := &mockRefreshTokenRepo{
		findByTokenFn: func(token string) (*model.RefreshToken, error) {
			if token == stored.Token {
				return stored, nil
			}
			return nil, errors.New("not found")
		},
	}

	resp, err := newAuthSvc(userRepo, rtRepo).Refresh(stored.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
}

func TestAuthService_Refresh_ExpiredToken(t *testing.T) {
	stored := &model.RefreshToken{
		UserID:    11,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-time.Hour), // already expired
	}
	rtRepo := &mockRefreshTokenRepo{
		findByTokenFn: func(token string) (*model.RefreshToken, error) {
			return stored, nil
		},
	}

	_, err := newAuthSvc(nil, rtRepo).Refresh(stored.Token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if err.Error() != "refresh token expired" {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	rtRepo := &mockRefreshTokenRepo{
		findByTokenFn: func(token string) (*model.RefreshToken, error) {
			return nil, errors.New("not found")
		},
	}

	_, err := newAuthSvc(nil, rtRepo).Refresh("bogus")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
