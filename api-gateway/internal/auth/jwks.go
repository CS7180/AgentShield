package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

type jwk struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

// JWKSKeyFunc returns a jwt.Keyfunc that validates ES256 tokens using
// Supabase's public JWKS endpoint. Keys are cached and refreshed on kid miss.
func JWKSKeyFunc(supabaseURL string) jwt.Keyfunc {
	var (
		mu   sync.RWMutex
		keys map[string]*ecdsa.PublicKey
	)

	fetch := func() error {
		resp, err := http.Get(supabaseURL + "/auth/v1/.well-known/jwks.json") //nolint:noctx
		if err != nil {
			return fmt.Errorf("fetch jwks: %w", err)
		}
		defer resp.Body.Close()

		var jwksResp jwksResponse
		if err := json.NewDecoder(resp.Body).Decode(&jwksResp); err != nil {
			return fmt.Errorf("decode jwks: %w", err)
		}

		newKeys := make(map[string]*ecdsa.PublicKey)
		for _, k := range jwksResp.Keys {
			if k.Kty != "EC" || k.Alg != "ES256" {
				continue
			}
			pub, err := ecPublicKeyFromJWK(k)
			if err != nil {
				continue
			}
			newKeys[k.Kid] = pub
		}

		mu.Lock()
		keys = newKeys
		mu.Unlock()
		return nil
	}

	// Eagerly populate cache; failure is non-fatal, will retry on first request.
	_ = fetch()

	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, _ := token.Header["kid"].(string)

		mu.RLock()
		pub, ok := keys[kid]
		mu.RUnlock()

		if !ok {
			// Refresh once on kid miss (key rotation).
			if err := fetch(); err != nil {
				return nil, fmt.Errorf("refresh jwks: %w", err)
			}
			mu.RLock()
			pub, ok = keys[kid]
			mu.RUnlock()
			if !ok {
				return nil, fmt.Errorf("unknown kid: %q", kid)
			}
		}

		return pub, nil
	}
}

func ecPublicKeyFromJWK(k jwk) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
	if err != nil {
		return nil, fmt.Errorf("decode x: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y: %w", err)
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}
