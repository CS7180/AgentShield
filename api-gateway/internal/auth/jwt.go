package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// SupabaseClaims represents the JWT claims issued by Supabase Auth.
// Supabase signs tokens with HS256 using the project's JWT_SECRET.
type SupabaseClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"` // "authenticated" for normal users
}

// ParseSupabaseTokenWithKeyFunc validates a Supabase JWT using the provided
// keyfunc. Use this for ES256 tokens via JWKSKeyFunc.
func ParseSupabaseTokenWithKeyFunc(tokenString string, keyFunc jwt.Keyfunc) (*SupabaseClaims, error) {
	claims := &SupabaseClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		keyFunc,
		jwt.WithAudience("authenticated"),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("missing sub claim")
	}
	return claims, nil
}

// ParseSupabaseToken validates a Supabase-issued JWT and returns its claims.
// The token must be HS256-signed and have audience "authenticated".
func ParseSupabaseToken(tokenString string, secret []byte) (*SupabaseClaims, error) {
	claims := &SupabaseClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secret, nil
		},
		jwt.WithAudience("authenticated"),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("missing sub claim")
	}

	return claims, nil
}
