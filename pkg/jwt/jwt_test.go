package jwt

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key-for-tests"

func TestGenerateAccessToken_ValidTokenRoundtrip(t *testing.T) {
	userID := uint(42)
	token, err := GenerateAccessToken(userID, testSecret, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateAccessToken(token, testSecret)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("got UserID=%d, want %d", claims.UserID, userID)
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	token, err := GenerateAccessToken(1, testSecret, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ValidateAccessToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	token, err := GenerateAccessToken(1, testSecret, -1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error generating expired token: %v", err)
	}

	_, err = ValidateAccessToken(token, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateAccessToken_MalformedToken(t *testing.T) {
	_, err := ValidateAccessToken("not.a.token", testSecret)
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}

func TestGenerateRefreshToken_UniqueEachCall(t *testing.T) {
	a, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == b {
		t.Fatal("expected distinct refresh tokens, got identical values")
	}
	// 32 random bytes → 64 hex chars
	if len(a) != 64 {
		t.Fatalf("got token length %d, want 64", len(a))
	}
}
