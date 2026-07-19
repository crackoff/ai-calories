package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtpkg "ai-calories/pkg/jwt"
)

const testJWTSecret = "middleware-test-secret"

func makeToken(t *testing.T, userID uint, ttl time.Duration) string {
	t.Helper()
	token, err := jwtpkg.GenerateAccessToken(userID, testJWTSecret, ttl)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

func TestAuthMiddleware_ValidToken_PassesThrough(t *testing.T) {
	token := makeToken(t, 42, 15*time.Minute)

	var capturedID uint
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = getUserID(r)
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(testJWTSecret)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", w.Code)
	}
	if capturedID != 42 {
		t.Fatalf("got userID=%d, want 42", capturedID)
	}
}

func TestAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})

	handler := AuthMiddleware(testJWTSecret)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})

	handler := AuthMiddleware(testJWTSecret)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken_Returns401(t *testing.T) {
	token := makeToken(t, 1, -time.Second)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})

	handler := AuthMiddleware(testJWTSecret)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_WrongScheme_Returns401(t *testing.T) {
	token := makeToken(t, 1, 15*time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})

	handler := AuthMiddleware(testJWTSecret)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token "+token) // wrong scheme
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", w.Code)
	}
}
