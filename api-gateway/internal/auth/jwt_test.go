package auth_test

import (
	"testing"
	"time"

	"github.com/agentshield/api-gateway/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

var testSecret = []byte("test-secret-key-for-unit-tests-only")

func makeToken(t *testing.T, claims jwt.Claims, secret []byte) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func validClaims() *auth.SupabaseClaims {
	return &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-uuid-123",
			Audience:  jwt.ClaimStrings{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email: "test@example.com",
		Role:  "authenticated",
	}
}

func TestParseSupabaseToken_Valid(t *testing.T) {
	tokenStr := makeToken(t, validClaims(), testSecret)

	got, err := auth.ParseSupabaseToken(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != "user-uuid-123" {
		t.Errorf("want sub=user-uuid-123, got %s", got.Subject)
	}
	if got.Email != "test@example.com" {
		t.Errorf("want email=test@example.com, got %s", got.Email)
	}
}

func TestParseSupabaseToken_Expired(t *testing.T) {
	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	tokenStr := makeToken(t, claims, testSecret)

	_, err := auth.ParseSupabaseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestParseSupabaseToken_WrongSecret(t *testing.T) {
	tokenStr := makeToken(t, validClaims(), []byte("other-secret"))

	_, err := auth.ParseSupabaseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestParseSupabaseToken_WrongAudience(t *testing.T) {
	claims := validClaims()
	claims.Audience = jwt.ClaimStrings{"service_role"}
	tokenStr := makeToken(t, claims, testSecret)

	_, err := auth.ParseSupabaseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for wrong audience, got nil")
	}
}

func TestParseSupabaseToken_MissingSub(t *testing.T) {
	claims := validClaims()
	claims.Subject = ""
	tokenStr := makeToken(t, claims, testSecret)

	_, err := auth.ParseSupabaseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for missing sub, got nil")
	}
}

func TestParseSupabaseToken_Malformed(t *testing.T) {
	_, err := auth.ParseSupabaseToken("not.a.jwt", testSecret)
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}
