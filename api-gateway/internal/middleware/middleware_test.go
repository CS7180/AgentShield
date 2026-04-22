package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentshield/api-gateway/internal/auth"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/agentshield/api-gateway/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

var testSecret = []byte("test-secret-key-for-unit-tests-only")

func hs256KeyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}
	return testSecret, nil
}

func makeToken(subject, email string, exp time.Time) string {
	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Audience:  jwt.ClaimStrings{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email: email,
		Role:  "authenticated",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString(testSecret)
	return s
}

func newRouter(mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw...)
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

func get(r *gin.Engine, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── JWTAuth ───────────────────────────────────────────────────────────────────

func TestJWTAuth_MissingHeader(t *testing.T) {
	r := newRouter(middleware.JWTAuth(hs256KeyFunc))
	w := get(r, "/test", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	for _, h := range []string{"Token abc", "Bearer", "justtoken"} {
		r := newRouter(middleware.JWTAuth(hs256KeyFunc))
		w := get(r, "/test", map[string]string{"Authorization": h})
		if w.Code != http.StatusUnauthorized {
			t.Errorf("header %q: want 401, got %d", h, w.Code)
		}
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	r := newRouter(middleware.JWTAuth(hs256KeyFunc))
	w := get(r, "/test", map[string]string{"Authorization": "Bearer not.a.jwt"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	token := makeToken(uuid.New().String(), "x@example.com", time.Now().Add(-time.Hour))
	r := newRouter(middleware.JWTAuth(hs256KeyFunc))
	w := get(r, "/test", map[string]string{"Authorization": "Bearer " + token})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

func TestJWTAuth_ValidToken_SetsContextValues(t *testing.T) {
	userID := uuid.New().String()
	token := makeToken(userID, "user@example.com", time.Now().Add(time.Hour))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.JWTAuth(hs256KeyFunc))
	r.GET("/test", func(c *gin.Context) {
		if got := c.GetString(middleware.UserIDKey); got != userID {
			t.Errorf("user_id = %q, want %q", got, userID)
		}
		if got := c.GetString(middleware.UserEmailKey); got != "user@example.com" {
			t.Errorf("user_email = %q, want user@example.com", got)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := get(r, "/test", map[string]string{"Authorization": "Bearer " + token})
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// ── RequestID ─────────────────────────────────────────────────────────────────

func TestRequestID_GeneratesUUID_WhenAbsent(t *testing.T) {
	r := newRouter(middleware.RequestID())
	w := get(r, "/test", nil)
	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Fatal("X-Request-ID should be set")
	}
	if _, err := uuid.Parse(rid); err != nil {
		t.Errorf("X-Request-ID should be a valid UUID, got %q", rid)
	}
}

func TestRequestID_PreservesClientHeader(t *testing.T) {
	existing := "my-trace-id-abc"
	r := newRouter(middleware.RequestID())
	w := get(r, "/test", map[string]string{"X-Request-ID": existing})
	if got := w.Header().Get("X-Request-ID"); got != existing {
		t.Errorf("X-Request-ID = %q, want %q", got, existing)
	}
}

// ── Recovery ──────────────────────────────────────────────────────────────────

func TestRecovery_Returns500_OnPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(zap.NewNop()))
	r.GET("/panic", func(c *gin.Context) { panic("intentional test panic") })

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", w.Code)
	}
}

// ── Ownership ─────────────────────────────────────────────────────────────────

type mockOwnershipRepo struct {
	ownerID uuid.UUID
	err     error
}

func (m *mockOwnershipRepo) GetByID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return m.ownerID, m.err
}

func ownershipRouter(repo middleware.OwnershipRepo, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/scans/:id", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Next()
	}, middleware.Ownership(repo), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestOwnership_InvalidScanID(t *testing.T) {
	r := ownershipRouter(&mockOwnershipRepo{}, uuid.New().String())
	w := get(r, "/scans/not-a-uuid", nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", w.Code)
	}
}

func TestOwnership_InvalidUserID(t *testing.T) {
	r := ownershipRouter(&mockOwnershipRepo{}, "not-a-uuid")
	w := get(r, "/scans/"+uuid.New().String(), nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", w.Code)
	}
}

func TestOwnership_ScanNotFound(t *testing.T) {
	repo := &mockOwnershipRepo{err: postgres.ErrNotFound}
	r := ownershipRouter(repo, uuid.New().String())
	w := get(r, "/scans/"+uuid.New().String(), nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

func TestOwnership_RepoError(t *testing.T) {
	repo := &mockOwnershipRepo{err: fmt.Errorf("db connection lost")}
	r := ownershipRouter(repo, uuid.New().String())
	w := get(r, "/scans/"+uuid.New().String(), nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", w.Code)
	}
}

func TestOwnership_WrongOwner(t *testing.T) {
	repo := &mockOwnershipRepo{ownerID: uuid.New()} // different owner
	r := ownershipRouter(repo, uuid.New().String()) // different user
	w := get(r, "/scans/"+uuid.New().String(), nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", w.Code)
	}
}

func TestOwnership_CorrectOwner(t *testing.T) {
	ownerID := uuid.New()
	repo := &mockOwnershipRepo{ownerID: ownerID}
	r := ownershipRouter(repo, ownerID.String())
	w := get(r, "/scans/"+uuid.New().String(), nil)
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// ── RateLimit ─────────────────────────────────────────────────────────────────

type mockLimiter struct {
	allow bool
	err   error
}

func (m *mockLimiter) AllowRequest(_ context.Context, _ string, _ int, _ float64) (bool, error) {
	return m.allow, m.err
}

func TestGlobalRateLimit_AllowsWhenNoUserID(t *testing.T) {
	// No user_id set → middleware skips rate limiting and calls Next
	r := newRouter(middleware.GlobalRateLimit(&mockLimiter{allow: false}))
	w := get(r, "/test", nil)
	if w.Code != http.StatusOK {
		t.Errorf("want 200 (skip), got %d", w.Code)
	}
}

func TestGlobalRateLimit_AllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(middleware.UserIDKey, uuid.New().String()) })
	r.Use(middleware.GlobalRateLimit(&mockLimiter{allow: true}))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := get(r, "/test", nil)
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

func TestGlobalRateLimit_BlocksWhenExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(middleware.UserIDKey, uuid.New().String()) })
	r.Use(middleware.GlobalRateLimit(&mockLimiter{allow: false}))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := get(r, "/test", nil)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", w.Code)
	}
}

func TestScanCreateRateLimit_AllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(middleware.UserIDKey, uuid.New().String()) })
	r.Use(middleware.ScanCreateRateLimit(&mockLimiter{allow: true}))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := get(r, "/test", nil)
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

func TestScanCreateRateLimit_BlocksWhenExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(middleware.UserIDKey, uuid.New().String()) })
	r.Use(middleware.ScanCreateRateLimit(&mockLimiter{allow: false}))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := get(r, "/test", nil)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", w.Code)
	}
}
