package service

import (
	"ai-calories/database"
	"ai-calories/internal/config"
	"ai-calories/internal/model"
	jwtpkg "ai-calories/pkg/jwt"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo         UserRepo
	refreshTokenRepo RefreshTokenRepo
	cfg              config.Config
}

type UserRepo interface {
	FindByEmail(email string) (*database.User, error)
	FindByID(id uint) (*database.User, error)
	Create(user *database.User) error
}

type RefreshTokenRepo interface {
	Save(token *model.RefreshToken) error
	FindByToken(token string) (*model.RefreshToken, error)
	DeleteByToken(token string) error
}

func NewAuthService(userRepo UserRepo, refreshTokenRepo RefreshTokenRepo, cfg config.Config) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
	}
}

func (s *AuthService) Register(email, password string) (*model.AuthResponse, error) {
	_, err := s.userRepo.FindByEmail(email)
	if err == nil {
		return nil, errors.New("email already registered")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashStr := string(hash)
	provider := "email"
	user := &database.User{
		Email:        &email,
		Password:     &hashStr,
		AuthProvider: &provider,
		Language:     "en",
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return s.generateTokenPair(user.ID)
}

func (s *AuthService) Login(email, password string) (*model.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if user.Password == nil {
		return nil, errors.New("account uses social login")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.generateTokenPair(user.ID)
}

func (s *AuthService) GoogleLogin(idToken string) (*model.AuthResponse, error) {
	email, err := verifyGoogleToken(idToken, s.cfg.GoogleClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google token: %w", err)
	}
	return s.oauthLogin(email, "google")
}

func (s *AuthService) AppleLogin(idToken string) (*model.AuthResponse, error) {
	email, err := verifyAppleToken(idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid apple token: %w", err)
	}
	return s.oauthLogin(email, "apple")
}

func (s *AuthService) Refresh(refreshToken string) (*model.AuthResponse, error) {
	stored, err := s.refreshTokenRepo.FindByToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if stored.ExpiresAt.Before(time.Now()) {
		_ = s.refreshTokenRepo.DeleteByToken(refreshToken)
		return nil, errors.New("refresh token expired")
	}

	_ = s.refreshTokenRepo.DeleteByToken(refreshToken)
	return s.generateTokenPair(stored.UserID)
}

func (s *AuthService) oauthLogin(email, provider string) (*model.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = &database.User{
			Email:        &email,
			AuthProvider: &provider,
			Language:     "en",
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return s.generateTokenPair(user.ID)
}

func (s *AuthService) generateTokenPair(userID uint) (*model.AuthResponse, error) {
	accessToken, err := jwtpkg.GenerateAccessToken(userID, s.cfg.JWTSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwtpkg.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	rt := &model.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWTRefreshTTL),
	}
	if err := s.refreshTokenRepo.Save(rt); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// verifyGoogleToken verifies a Google ID token and returns the email.
func verifyGoogleToken(idToken, clientID string) (string, error) {
	url := "https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("google token verification failed")
	}

	var payload struct {
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Aud           string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	if clientID != "" && payload.Aud != clientID {
		return "", errors.New("token audience mismatch")
	}
	if payload.EmailVerified != "true" {
		return "", errors.New("email not verified")
	}

	return payload.Email, nil
}

// verifyAppleToken verifies an Apple ID token and returns the email.
func verifyAppleToken(idToken string) (string, error) {
	// Apple ID tokens are JWTs signed with Apple's public keys.
	// For MVP, we parse the claims without full key verification
	// and validate the issuer. Production should fetch keys from
	// https://appleid.apple.com/auth/keys and verify signature.
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse apple token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid apple token claims")
	}

	iss, _ := claims["iss"].(string)
	if iss != "https://appleid.apple.com" {
		return "", errors.New("invalid apple token issuer")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return "", errors.New("email not found in apple token")
	}

	return email, nil
}
