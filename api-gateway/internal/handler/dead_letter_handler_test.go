package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/handler"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type fakeDeadLetterRepo struct {
	items []*domain.ScanDeadLetter
	total int
	err   error
}

func (f *fakeDeadLetterRepo) ListByScanID(_ context.Context, _ uuid.UUID, _ uuid.UUID, _, _ int) ([]*domain.ScanDeadLetter, int, error) {
	return f.items, f.total, f.err
}

func newDeadLetterRouter(h *handler.DeadLetterHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/scans/:id/dead-letters", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		h.List(c)
	})
	return r
}

func TestDeadLetterList_Returns200(t *testing.T) {
	scanID := uuid.New()
	userID := uuid.New()

	repo := &fakeDeadLetterRepo{
		items: []*domain.ScanDeadLetter{
			{
				ID:           uuid.New(),
				ScanID:       scanID,
				UserID:       userID,
				AttemptCount: 3,
				ErrorStage:   "run agents",
				ErrorMessage: "run agents: timeout",
				FailedAt:     time.Now().UTC(),
			},
		},
		total: 1,
	}

	h := handler.NewDeadLetterHandler(repo, zap.NewNop())
	r := newDeadLetterRouter(h, userID.String())

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/dead-letters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got, ok := payload["total"].(float64); !ok || int(got) != 1 {
		t.Fatalf("total = %v, want 1", payload["total"])
	}
}

func TestDeadLetterList_Returns400_WhenInvalidScanID(t *testing.T) {
	h := handler.NewDeadLetterHandler(&fakeDeadLetterRepo{}, zap.NewNop())
	r := newDeadLetterRouter(h, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans/not-a-uuid/dead-letters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestDeadLetterList_Returns400_WhenInvalidUserID(t *testing.T) {
	h := handler.NewDeadLetterHandler(&fakeDeadLetterRepo{}, zap.NewNop())
	r := newDeadLetterRouter(h, "not-a-uuid")

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/dead-letters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestDeadLetterList_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeDeadLetterRepo{err: errors.New("db error")}
	h := handler.NewDeadLetterHandler(repo, zap.NewNop())
	r := newDeadLetterRouter(h, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/dead-letters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (body: %s)", w.Code, w.Body.String())
	}
}

func TestDeadLetterList_Returns200_WithLimitOffsetParams(t *testing.T) {
	repo := &fakeDeadLetterRepo{items: []*domain.ScanDeadLetter{}, total: 0}
	h := handler.NewDeadLetterHandler(repo, zap.NewNop())
	r := newDeadLetterRouter(h, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/dead-letters?limit=5&offset=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
}
