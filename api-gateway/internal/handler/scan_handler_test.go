package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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

// ── Fakes ────────────────────────────────────────────────────────────────────

// fakeScanRepo satisfies handler.ScanRepo.
// Only GetByID, UpdateStatus, MarkStarted, and MarkStopped are exercised
// by Start and Stop; the rest are no-ops.
type fakeScanRepo struct {
	scan *domain.Scan
}

func (f *fakeScanRepo) Create(_ context.Context, _ *domain.Scan) error { return nil }
func (f *fakeScanRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Scan, error) {
	return f.scan, nil
}
func (f *fakeScanRepo) ListByUser(_ context.Context, _ uuid.UUID, _, _ int) ([]*domain.Scan, int, error) {
	return nil, 0, nil
}
func (f *fakeScanRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.ScanStatus) error {
	return nil
}
func (f *fakeScanRepo) MarkStarted(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeScanRepo) MarkStopped(_ context.Context, _ uuid.UUID) error { return nil }

// fakeOrchestrator satisfies handler.OrchestratorClient.
// accepted=true by default so Start tests that reach the orchestrator succeed.
type fakeOrchestrator struct{ accepted bool }

func (f *fakeOrchestrator) StartScan(_ context.Context, _, _, _ string, _ []string) (bool, string, error) {
	return f.accepted, "ok", nil
}
func (f *fakeOrchestrator) StopScan(_ context.Context, _ string) (bool, string, error) {
	return true, "ok", nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newRouter(h *handler.ScanHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	scanID := uuid.New()

	// Inject a fixed user ID and scan ID so ownership middleware is bypassed
	// and handler logic is the only thing under test.
	r.POST("/scans/:id/start", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, scanID.String())
		h.Start(c)
	})
	r.POST("/scans/:id/stop", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, scanID.String())
		h.Stop(c)
	})
	return r
}

func scanWithStatus(status domain.ScanStatus) *domain.Scan {
	return &domain.Scan{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		TargetEndpoint: "https://example.com",
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{"prompt_injection"},
		Status:         status,
	}
}

func doRequest(t *testing.T, r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, wantHTTP int) {
	t.Helper()
	if w.Code != wantHTTP {
		t.Errorf("HTTP status = %d, want %d (body: %s)", w.Code, wantHTTP, w.Body.String())
	}
}

func assertErrorCode(t *testing.T, w *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	got, _ := body["code"].(string)
	if got != wantCode {
		t.Errorf("error code = %q, want %q (body: %s)", got, wantCode, w.Body.String())
	}
}

func newHandler(scan *domain.Scan) (*handler.ScanHandler, *gin.Engine) {
	repo := &fakeScanRepo{scan: scan}
	orch := &fakeOrchestrator{accepted: true}
	h := handler.NewScanHandler(repo, orch, zap.NewNop())
	return h, newRouter(h)
}

// ── Start tests ───────────────────────────────────────────────────────────────

// pending → running is a valid transition; orchestrator accepts → 202.
func TestStart_Returns202_WhenPending(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusPending))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/start")
	assertStatus(t, w, http.StatusAccepted)
}

// queued → running is a valid transition; orchestrator accepts → 202.
func TestStart_Returns202_WhenQueued(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusQueued))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/start")
	assertStatus(t, w, http.StatusAccepted)
}

// running → running is not a valid transition → 400 INVALID_SCAN_STATE.
func TestStart_Returns400_WhenRunning(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusRunning))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/start")
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_STATE")
}

// completed is terminal; cannot be started → 400 INVALID_SCAN_STATE.
func TestStart_Returns400_WhenCompleted(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusCompleted))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/start")
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_STATE")
}

// failed is terminal; cannot be restarted → 400 INVALID_SCAN_STATE.
func TestStart_Returns400_WhenFailed(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusFailed))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/start")
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_STATE")
}

// stopped is terminal; cannot be resumed → 400 INVALID_SCAN_STATE.
func TestStart_Returns400_WhenStopped(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusStopped))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/start")
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_STATE")
}

// ── Stop tests ────────────────────────────────────────────────────────────────

// running → stopped is a valid transition → 200.
func TestStop_Returns200_WhenRunning(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusRunning))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/stop")
	assertStatus(t, w, http.StatusOK)
}

// queued → stopped is a valid transition (cancel before pickup) → 200.
func TestStop_Returns200_WhenQueued(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusQueued))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/stop")
	assertStatus(t, w, http.StatusOK)
}

// pending → stopped is not a valid transition → 400 INVALID_SCAN_STATE.
func TestStop_Returns400_WhenPending(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusPending))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/stop")
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_STATE")
}

// completed is terminal; stop is meaningless → 400 INVALID_SCAN_STATE.
func TestStop_Returns400_WhenCompleted(t *testing.T) {
	_, r := newHandler(scanWithStatus(domain.StatusCompleted))
	w := doRequest(t, r, http.MethodPost, "/scans/"+uuid.New().String()+"/stop")
	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_SCAN_STATE")
}
