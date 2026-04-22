package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/handler"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type fakeAttackResultRepo struct {
	createInputs []domain.AttackResultInput
	createErr    error
	listResults  []*domain.AttackResult
	listTotal    int
	listErr      error
}

func (f *fakeAttackResultRepo) CreateBatch(
	_ context.Context,
	scanID uuid.UUID,
	userID uuid.UUID,
	inputs []domain.AttackResultInput,
) ([]*domain.AttackResult, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.createInputs = inputs
	out := make([]*domain.AttackResult, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, &domain.AttackResult{
			ID:             uuid.New(),
			ScanID:         scanID,
			UserID:         userID,
			AttackType:     in.AttackType,
			AttackPrompt:   in.AttackPrompt,
			TargetResponse: in.TargetResponse,
			AttackSuccess:  in.AttackSuccess,
			Severity:       in.Severity,
		})
	}
	return out, nil
}

func (f *fakeAttackResultRepo) ListByScanID(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int, _ int) ([]*domain.AttackResult, int, error) {
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return f.listResults, f.listTotal, nil
}

func (f *fakeAttackResultRepo) ListAllByScanID(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]*domain.AttackResult, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResults, nil
}

func newAttackResultsRouter(h *handler.AttackResultHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := "00000000-0000-0000-0000-000000000001"
	r.POST("/scans/:id/attack-results", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.CreateBatch(c)
	})
	r.GET("/scans/:id/attack-results", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.List(c)
	})
	return r
}

func TestAttackResultsCreateBatch_Returns201(t *testing.T) {
	repo := &fakeAttackResultRepo{}
	h := handler.NewAttackResultHandler(repo, zap.NewNop())
	r := newAttackResultsRouter(h)

	body := []byte(`{
		"results":[
			{
				"attack_type":"prompt_injection",
				"attack_prompt":"ignore all rules",
				"target_response":"ok",
				"attack_success": true,
				"severity":"high"
			}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/attack-results", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusCreated)
	if len(repo.createInputs) != 1 {
		t.Fatalf("create input length = %d, want 1", len(repo.createInputs))
	}
}

func TestAttackResultsCreateBatch_Returns400_WhenPayloadInvalid(t *testing.T) {
	repo := &fakeAttackResultRepo{}
	h := handler.NewAttackResultHandler(repo, zap.NewNop())
	r := newAttackResultsRouter(h)

	body := []byte(`{"results":[{"attack_type":"invalid"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/attack-results", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_REQUEST")
}

func TestAttackResultsList_Returns200(t *testing.T) {
	repo := &fakeAttackResultRepo{
		listResults: []*domain.AttackResult{
			{
				ID:             uuid.New(),
				ScanID:         uuid.New(),
				UserID:         uuid.New(),
				AttackType:     "jailbreak",
				AttackPrompt:   "prompt",
				TargetResponse: "resp",
				AttackSuccess:  true,
				Severity:       "critical",
			},
		},
		listTotal: 1,
	}
	h := handler.NewAttackResultHandler(repo, zap.NewNop())
	r := newAttackResultsRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/attack-results?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestAttackResultsCreateBatch_Returns400_WithInvalidScanID(t *testing.T) {
	repo := &fakeAttackResultRepo{}
	h := handler.NewAttackResultHandler(repo, zap.NewNop())
	r := newAttackResultsRouter(h)

	body := []byte(`{"results":[{"attack_type":"prompt_injection","attack_prompt":"x","target_response":"y","attack_success":true,"severity":"high"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/scans/not-a-uuid/attack-results", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestAttackResultsCreateBatch_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeAttackResultRepo{createErr: errors.New("db down")}
	h := handler.NewAttackResultHandler(repo, zap.NewNop())
	r := newAttackResultsRouter(h)

	body := []byte(`{"results":[{"attack_type":"prompt_injection","attack_prompt":"x","target_response":"y","attack_success":true,"severity":"high"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/attack-results", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

func TestAttackResultsList_Returns400_WithInvalidScanID(t *testing.T) {
	repo := &fakeAttackResultRepo{}
	h := handler.NewAttackResultHandler(repo, zap.NewNop())
	r := newAttackResultsRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/not-a-uuid/attack-results", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestAttackResultsList_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeAttackResultRepo{listErr: errors.New("db down")}
	h := handler.NewAttackResultHandler(repo, zap.NewNop())
	r := newAttackResultsRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/attack-results", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}
