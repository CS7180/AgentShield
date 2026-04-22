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

type fakeGenerateScanRepo struct {
	scan *domain.Scan
	err  error
}

func (f *fakeGenerateScanRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Scan, error) {
	return f.scan, f.err
}

func newReportGenerateRouter(h *handler.ReportGenerationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := "00000000-0000-0000-0000-000000000001"
	r.POST("/scans/:id/report/generate", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.Generate(c)
	})
	return r
}

func TestReportGenerate_Returns200_AndPersistsReport(t *testing.T) {
	scanID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	reportRepo := &fakeReportRepo{}
	attackRepo := &fakeAttackResultRepo{
		listResults: []*domain.AttackResult{
			{
				ID:             uuid.New(),
				ScanID:         scanID,
				UserID:         userID,
				AttackType:     "prompt_injection",
				AttackPrompt:   "ignore policy",
				TargetResponse: "ok",
				AttackSuccess:  true,
				Severity:       "critical",
			},
			{
				ID:             uuid.New(),
				ScanID:         scanID,
				UserID:         userID,
				AttackType:     "jailbreak",
				AttackPrompt:   "bypass",
				TargetResponse: "blocked",
				AttackSuccess:  false,
				Severity:       "high",
			},
		},
	}
	scanRepo := &fakeGenerateScanRepo{
		scan: &domain.Scan{
			ID:             scanID,
			UserID:         userID,
			TargetEndpoint: "https://example.com/chat",
			Mode:           domain.ModeRedTeam,
		},
	}
	uploader := &fakeUploader{}

	h := handler.NewReportGenerationHandler(
		reportRepo,
		attackRepo,
		scanRepo,
		uploader,
		"agentshield-reports",
		zap.NewNop(),
	)
	r := newReportGenerateRouter(h)

	body := []byte(`{"include_pdf": false}`)
	req := httptest.NewRequest(http.MethodPost, "/scans/"+scanID.String()+"/report/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
	if reportRepo.upsertSeen == nil {
		t.Fatal("expected report upsert")
	}
	if reportRepo.upsertSeen.OverallScore == nil {
		t.Fatal("expected overall score set")
	}
	if uploader.uploads != 1 {
		t.Fatalf("uploads = %d, want 1 (json only)", uploader.uploads)
	}
}

func TestReportGenerate_Returns100Score_WhenNoResults(t *testing.T) {
	scanID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	reportRepo := &fakeReportRepo{}
	attackRepo := &fakeAttackResultRepo{listResults: []*domain.AttackResult{}}
	scanRepo := &fakeGenerateScanRepo{
		scan: &domain.Scan{
			ID:             scanID,
			UserID:         userID,
			TargetEndpoint: "https://example.com/chat",
			Mode:           domain.ModeAdversarial,
		},
	}

	h := handler.NewReportGenerationHandler(
		reportRepo,
		attackRepo,
		scanRepo,
		nil,
		"agentshield-reports",
		zap.NewNop(),
	)
	r := newReportGenerateRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/scans/"+scanID.String()+"/report/generate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	score, _ := payload["overall_score"].(float64)
	if score != 100 {
		t.Fatalf("overall_score = %v, want 100", score)
	}
}

func TestReportGenerate_WithPDF_Returns200_AndUploadsTwice(t *testing.T) {
	scanID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	reportRepo := &fakeReportRepo{}
	// Two *successful* results with different severities so sort.Slice in
	// buildGeneratedReportPayload actually invokes the severityRank comparator.
	attackRepo := &fakeAttackResultRepo{
		listResults: []*domain.AttackResult{
			{
				ID:             uuid.New(),
				ScanID:         scanID,
				UserID:         userID,
				AttackType:     "prompt_injection",
				AttackPrompt:   "ignore policy (test)",
				TargetResponse: "sure",
				AttackSuccess:  true,
				Severity:       "high",
			},
			{
				ID:             uuid.New(),
				ScanID:         scanID,
				UserID:         userID,
				AttackType:     "jailbreak",
				AttackPrompt:   "bypass (test)",
				TargetResponse: "yes",
				AttackSuccess:  true,
				Severity:       "critical",
			},
			{
				ID:             uuid.New(),
				ScanID:         scanID,
				UserID:         userID,
				AttackType:     "data_leakage",
				AttackPrompt:   "dump secrets",
				TargetResponse: "blocked",
				AttackSuccess:  false,
				Severity:       "medium",
			},
		},
	}
	scanRepo := &fakeGenerateScanRepo{
		scan: &domain.Scan{
			ID:             scanID,
			UserID:         userID,
			TargetEndpoint: "https://example.com/chat",
			Mode:           domain.ModeRedTeam,
		},
	}
	uploader := &fakeUploader{}

	h := handler.NewReportGenerationHandler(
		reportRepo,
		attackRepo,
		scanRepo,
		uploader,
		"agentshield-reports",
		zap.NewNop(),
	)
	r := newReportGenerateRouter(h)

	// include_pdf: true triggers severityRank, renderSimpleReportPDF,
	// buildSinglePagePDF and escapePDFText
	body := []byte(`{"include_pdf": true}`)
	req := httptest.NewRequest(http.MethodPost, "/scans/"+scanID.String()+"/report/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
	// JSON upload + PDF upload = 2
	if uploader.uploads != 2 {
		t.Errorf("uploads = %d, want 2 (json + pdf)", uploader.uploads)
	}
}
