package auth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentshield/api-gateway/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

// ── ParseSupabaseTokenWithKeyFunc ─────────────────────────────────────────────

func TestParseSupabaseTokenWithKeyFunc_Valid(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-ec-123",
			Audience:  jwt.ClaimStrings{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email: "ec@example.com",
		Role:  "authenticated",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenStr, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected method")
		}
		return &key.PublicKey, nil
	}

	got, err := auth.ParseSupabaseTokenWithKeyFunc(tokenStr, keyFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != "user-ec-123" {
		t.Errorf("subject = %q, want user-ec-123", got.Subject)
	}
	if got.Email != "ec@example.com" {
		t.Errorf("email = %q, want ec@example.com", got.Email)
	}
}

func TestParseSupabaseTokenWithKeyFunc_Expired(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-ec-123",
			Audience:  jwt.ClaimStrings{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenStr, _ := tok.SignedString(key)

	keyFunc := func(t *jwt.Token) (interface{}, error) { return &key.PublicKey, nil }
	_, err := auth.ParseSupabaseTokenWithKeyFunc(tokenStr, keyFunc)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestParseSupabaseTokenWithKeyFunc_WrongAudience(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-ec-123",
			Audience:  jwt.ClaimStrings{"service_role"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenStr, _ := tok.SignedString(key)

	keyFunc := func(t *jwt.Token) (interface{}, error) { return &key.PublicKey, nil }
	_, err := auth.ParseSupabaseTokenWithKeyFunc(tokenStr, keyFunc)
	if err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestParseSupabaseTokenWithKeyFunc_MissingSub(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "",
			Audience:  jwt.ClaimStrings{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenStr, _ := tok.SignedString(key)

	keyFunc := func(t *jwt.Token) (interface{}, error) { return &key.PublicKey, nil }
	_, err := auth.ParseSupabaseTokenWithKeyFunc(tokenStr, keyFunc)
	if err == nil {
		t.Fatal("expected error for missing sub")
	}
}

// ── JWKSKeyFunc (mock HTTP server) ────────────────────────────────────────────

// jwkFromKey converts an ECDSA public key to a JWK map for test server responses.
func jwkFromKey(kid string, pub *ecdsa.PublicKey) map[string]string {
	return map[string]string{
		"kty": "EC",
		"alg": "ES256",
		"kid": kid,
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(pub.X.Bytes()),
		"y":   base64.RawURLEncoding.EncodeToString(pub.Y.Bytes()),
	}
}

// padTo32 ensures the coordinate bytes are exactly 32 bytes (P-256 requirement).
func coordBytes(n *big.Int) []byte {
	b := n.Bytes()
	if len(b) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(b):], b)
		return padded
	}
	return b
}

func jwkFromKeyPadded(kid string, pub *ecdsa.PublicKey) map[string]string {
	return map[string]string{
		"kty": "EC",
		"alg": "ES256",
		"kid": kid,
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(coordBytes(pub.X)),
		"y":   base64.RawURLEncoding.EncodeToString(coordBytes(pub.Y)),
	}
}

func newJWKSServer(keys ...map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"keys": keys})
	}))
}

func signES256(t *testing.T, key *ecdsa.PrivateKey, kid, subject string) string {
	t.Helper()
	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Audience:  jwt.ClaimStrings{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email: subject + "@example.com",
		Role:  "authenticated",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign ES256 token: %v", err)
	}
	return s
}

func TestJWKSKeyFunc_ValidToken(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	kid := "test-key-1"

	srv := newJWKSServer(jwkFromKeyPadded(kid, &key.PublicKey))
	defer srv.Close()

	keyFunc := auth.JWKSKeyFunc(srv.URL)
	tokenStr := signES256(t, key, kid, "user-jwks-1")

	got, err := auth.ParseSupabaseTokenWithKeyFunc(tokenStr, keyFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != "user-jwks-1" {
		t.Errorf("subject = %q, want user-jwks-1", got.Subject)
	}
}

func TestJWKSKeyFunc_UnknownKid(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	kid := "known-key"

	srv := newJWKSServer(jwkFromKeyPadded(kid, &key.PublicKey))
	defer srv.Close()

	keyFunc := auth.JWKSKeyFunc(srv.URL)

	// Sign with a different kid that the server doesn't know
	otherKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tokenStr := signES256(t, otherKey, "unknown-kid", "user-404")

	_, err := auth.ParseSupabaseTokenWithKeyFunc(tokenStr, keyFunc)
	if err == nil {
		t.Fatal("expected error for unknown kid")
	}
}

func TestJWKSKeyFunc_WrongSigningMethod(t *testing.T) {
	srv := newJWKSServer()
	defer srv.Close()

	keyFunc := auth.JWKSKeyFunc(srv.URL)

	// HS256 token passed to an ES256 keyFunc
	hs256Claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Audience:  jwt.ClaimStrings{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email: "user@example.com",
		Role:  "authenticated",
	}
	tokenStr := makeToken(t, hs256Claims, testSecret)
	_, err := auth.ParseSupabaseTokenWithKeyFunc(tokenStr, keyFunc)
	if err == nil {
		t.Fatal("expected error for wrong signing method (HS256 vs ES256)")
	}
}
