package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/handler"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/agentshield/api-gateway/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ── Repo variant that can return errors ──────────────────────────────────────

type fakeScanRepoWithErr struct {
	scan      *domain.Scan
	createErr error
	getErr    error
	listScans []*domain.Scan
	listTotal int
	listErr   error
}

func (f *fakeScanRepoWithErr) Create(_ context.Context, _ *domain.Scan) error {
	return f.createErr
}
func (f *fakeScanRepoWithErr) GetByID(_ context.Context, _ uuid.UUID) (*domain.Scan, error) {
	return f.scan, f.getErr
}
func (f *fakeScanRepoWithErr) ListByUser(_ context.Context, _ uuid.UUID, _, _ int) ([]*domain.Scan, int, error) {
	return f.listScans, f.listTotal, f.listErr
}
func (f *fakeScanRepoWithErr) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.ScanStatus) error {
	return nil
}
func (f *fakeScanRepoWithErr) MarkStarted(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeScanRepoWithErr) MarkStopped(_ context.Context, _ uuid.UUID) error { return nil }

// ── Router helpers ────────────────────────────────────────────────────────────

func crudRouter(repo handler.ScanRepo, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handler.NewScanHandler(repo, &fakeOrchestrator{accepted: true}, zap.NewNop())

	r := gin.New()
	r.POST("/scans", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		h.Create(c)
	})
	r.GET("/scans", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		h.List(c)
	})
	r.GET("/scans/:id", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		h.Get(c)
	})
	return r
}

func postJSON(t *testing.T, r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_Returns201_WithValidRequest(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, uuid.New().String())

	body := domain.CreateScanRequest{
		TargetEndpoint: "https://api.example.com/v1/chat",
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{"prompt_injection"},
	}
	w := postJSON(t, r, "/scans", body)
	assertStatus(t, w, http.StatusCreated)
}

func TestCreate_Returns400_WithInvalidJSON(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodPost, "/scans", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestCreate_Returns400_WithHTTPTarget(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, uuid.New().String())

	body := domain.CreateScanRequest{
		TargetEndpoint: "http://insecure.example.com",
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{"prompt_injection"},
	}
	w := postJSON(t, r, "/scans", body)
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_TARGET_ENDPOINT")
}

func TestCreate_Returns400_WithInvalidMode(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, uuid.New().String())

	body := map[string]interface{}{
		"target_endpoint": "https://api.example.com",
		"mode":            "invalid_mode",
		"attack_types":    []string{"prompt_injection"},
	}
	w := postJSON(t, r, "/scans", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestCreate_Returns400_WithEmptyAttackTypes(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, uuid.New().String())

	body := domain.CreateScanRequest{
		TargetEndpoint: "https://api.example.com",
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{},
	}
	w := postJSON(t, r, "/scans", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestCreate_Returns400_WithInvalidUserID(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, "not-a-uuid")

	body := domain.CreateScanRequest{
		TargetEndpoint: "https://api.example.com",
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{"prompt_injection"},
	}
	w := postJSON(t, r, "/scans", body)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestCreate_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeScanRepoWithErr{createErr: errors.New("db down")}
	r := crudRouter(repo, uuid.New().String())

	body := domain.CreateScanRequest{
		TargetEndpoint: "https://api.example.com",
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{"prompt_injection"},
	}
	w := postJSON(t, r, "/scans", body)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestList_Returns200_EmptyList(t *testing.T) {
	repo := &fakeScanRepoWithErr{listScans: []*domain.Scan{}, listTotal: 0}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)

	var body domain.ScanListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Scans) != 0 {
		t.Errorf("want empty scans, got %d", len(body.Scans))
	}
}

func TestList_Returns200_WithScans(t *testing.T) {
	scans := []*domain.Scan{scanWithStatus(domain.StatusPending), scanWithStatus(domain.StatusRunning)}
	repo := &fakeScanRepoWithErr{listScans: scans, listTotal: 2}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)

	var body domain.ScanListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Total != 2 {
		t.Errorf("total = %d, want 2", body.Total)
	}
}

func TestList_Returns400_WithInvalidUserID(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, "not-a-uuid")

	req := httptest.NewRequest(http.MethodGet, "/scans", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestList_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeScanRepoWithErr{listErr: errors.New("db error")}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── Get ───────────────────────────────────────────────────────────────────────

func TestGet_Returns200_ForExistingScan(t *testing.T) {
	scan := scanWithStatus(domain.StatusRunning)
	repo := &fakeScanRepoWithErr{scan: scan}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scan.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)
}

func TestGet_Returns404_WhenNotFound(t *testing.T) {
	repo := &fakeScanRepoWithErr{getErr: postgres.ErrNotFound}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGet_Returns400_WithInvalidScanID(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGet_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeScanRepoWithErr{getErr: errors.New("db error")}
	r := crudRouter(repo, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── Start edge cases ──────────────────────────────────────────────────────────

func startStopRouter(repo handler.ScanRepo, orch handler.OrchestratorClient, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handler.NewScanHandler(repo, orch, zap.NewNop())
	r := gin.New()
	r.POST("/scans/:id/start", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		h.Start(c)
	})
	r.POST("/scans/:id/stop", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		h.Stop(c)
	})
	return r
}

func TestStart_Returns400_WithInvalidScanID(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := startStopRouter(repo, &fakeOrchestrator{accepted: true}, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/scans/not-a-uuid/start", nil))
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_ID")
}

func TestStart_Returns404_WhenScanNotFound(t *testing.T) {
	repo := &fakeScanRepoWithErr{getErr: postgres.ErrNotFound}
	r := startStopRouter(repo, &fakeOrchestrator{accepted: true}, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/start", nil))
	assertStatus(t, w, http.StatusNotFound)
}

func TestStart_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeScanRepoWithErr{getErr: errors.New("db down")}
	r := startStopRouter(repo, &fakeOrchestrator{accepted: true}, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/start", nil))
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestStart_Returns202_WhenOrchestratorRejects(t *testing.T) {
	repo := &fakeScanRepoWithErr{scan: scanWithStatus(domain.StatusPending)}
	r := startStopRouter(repo, &fakeOrchestrator{accepted: false}, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/start", nil))
	assertStatus(t, w, http.StatusAccepted)
}

// ── Stop edge cases ───────────────────────────────────────────────────────────

func TestStop_Returns400_WithInvalidScanID(t *testing.T) {
	repo := &fakeScanRepoWithErr{}
	r := startStopRouter(repo, &fakeOrchestrator{accepted: true}, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/scans/not-a-uuid/stop", nil))
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_ID")
}

func TestStop_Returns404_WhenScanNotFound(t *testing.T) {
	repo := &fakeScanRepoWithErr{getErr: postgres.ErrNotFound}
	r := startStopRouter(repo, &fakeOrchestrator{accepted: true}, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/stop", nil))
	assertStatus(t, w, http.StatusNotFound)
}

func TestStop_Returns500_OnRepoError(t *testing.T) {
	repo := &fakeScanRepoWithErr{getErr: errors.New("db down")}
	r := startStopRouter(repo, &fakeOrchestrator{accepted: true}, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/scans/"+uuid.New().String()+"/stop", nil))
	assertStatus(t, w, http.StatusInternalServerError)
}
