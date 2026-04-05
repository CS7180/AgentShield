package handler_test

import (
	"bytes"
	"context"
	"encoding/base64"
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

// fakeReportRepo satisfies handler.ReportRepo.
type fakeReportRepo struct {
	report     *domain.Report
	getErr     error
	upsertErr  error
	upsertSeen *domain.Report
}

func (f *fakeReportRepo) GetByScanID(_ context.Context, _ uuid.UUID) (*domain.Report, error) {
	return f.report, f.getErr
}

func (f *fakeReportRepo) UpsertByScanID(_ context.Context, report *domain.Report) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.upsertSeen = report
	f.report = report
	return nil
}

type fakeUploader struct {
	err      error
	uploads  int
	lastPath string
}

func (f *fakeUploader) Upload(_ context.Context, _, objectPath, _ string, _ []byte) error {
	if f.err != nil {
		return f.err
	}
	f.uploads++
	f.lastPath = objectPath
	return nil
}

func newReportRouter(h *handler.ReportHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/scans/:id/report", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, uuid.New().String())
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.GetJSON(c)
	})
	r.GET("/scans/:id/report/pdf", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, uuid.New().String())
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.GetPDF(c)
	})
	r.PUT("/scans/:id/report", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, "00000000-0000-0000-0000-000000000001")
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.Upsert(c)
	})
	return r
}

func TestGetReport_Returns200_WhenReportExists(t *testing.T) {
	report := &domain.Report{
		ID:           uuid.New(),
		ScanID:       uuid.New(),
		UserID:       uuid.New(),
		ReportBucket: "agentshield-reports",
	}
	repo := &fakeReportRepo{report: report}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+report.ScanID.String()+"/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestGetReport_Returns404_WhenReportNotFound(t *testing.T) {
	repo := &fakeReportRepo{getErr: postgres.ErrNotFound}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusNotFound)
}

func TestGetReport_Returns500_WhenRepoFails(t *testing.T) {
	repo := &fakeReportRepo{getErr: errors.New("db offline")}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

func TestGetReport_Returns400_WhenInvalidScanID(t *testing.T) {
	repo := &fakeReportRepo{}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/not-a-uuid/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetPDF_Returns404_WhenPathMissing(t *testing.T) {
	report := &domain.Report{
		ID:           uuid.New(),
		ScanID:       uuid.New(),
		UserID:       uuid.New(),
		ReportBucket: "agentshield-reports",
	}
	repo := &fakeReportRepo{report: report}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+report.ScanID.String()+"/report/pdf", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusNotFound)
}

func TestGetPDF_Returns200_WhenPathExists(t *testing.T) {
	report := &domain.Report{
		ID:            uuid.New(),
		ScanID:        uuid.New(),
		UserID:        uuid.New(),
		ReportBucket:  "agentshield-reports",
		ReportPDFPath: "00000000-0000-0000-0000-000000000000/scans/test/report.pdf",
	}
	repo := &fakeReportRepo{report: report}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+report.ScanID.String()+"/report/pdf", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestGetPDF_Returns400_WhenInvalidScanID(t *testing.T) {
	repo := &fakeReportRepo{}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/not-a-uuid/report/pdf", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpsertReport_Returns200_WhenValidPayload(t *testing.T) {
	repo := &fakeReportRepo{}
	uploader := &fakeUploader{}
	h := handler.NewReportHandler(repo, uploader, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	body := []byte(`{
		"overall_score": 82.5,
		"critical_count": 1,
		"high_count": 2,
		"medium_count": 3,
		"low_count": 4,
		"owasp_scorecard": {"llm01":"fail"},
		"report_json": {"summary":"ok"},
		"pdf_base64": "` + base64.StdEncoding.EncodeToString([]byte("%PDF-test")) + `"
	}`)

	req := httptest.NewRequest(http.MethodPut, "/scans/11111111-1111-1111-1111-111111111111/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
	if uploader.uploads != 2 {
		t.Fatalf("expected 2 uploads (json + pdf), got %d", uploader.uploads)
	}
	if repo.upsertSeen == nil {
		t.Fatal("expected report upsert to be called")
	}
}

func TestUpsertReport_Returns501_WhenStorageNotConfigured(t *testing.T) {
	repo := &fakeReportRepo{}
	h := handler.NewReportHandler(repo, nil, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	body := []byte(`{"report_json":{"summary":"ok"}}`)
	req := httptest.NewRequest(http.MethodPut, "/scans/11111111-1111-1111-1111-111111111111/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusNotImplemented)
}

func TestUpsertReport_Returns400_WhenInvalidJSONPayload(t *testing.T) {
	repo := &fakeReportRepo{}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	body := []byte(`{"report_json":"not-an-object"}`)
	req := httptest.NewRequest(http.MethodPut, "/scans/11111111-1111-1111-1111-111111111111/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpsertReport_Returns500_WhenUploaderFails(t *testing.T) {
	repo := &fakeReportRepo{}
	uploader := &fakeUploader{err: errors.New("storage down")}
	h := handler.NewReportHandler(repo, uploader, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	body := []byte(`{"report_json":{"summary":"ok"}}`)
	req := httptest.NewRequest(http.MethodPut, "/scans/11111111-1111-1111-1111-111111111111/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}

func TestUpsertReport_Returns500_WhenRepoFails(t *testing.T) {
	repo := &fakeReportRepo{upsertErr: errors.New("db write failed")}
	h := handler.NewReportHandler(repo, &fakeUploader{}, "agentshield-reports", zap.NewNop())
	r := newReportRouter(h)

	body := []byte(`{"report_json":{"summary":"ok"}}`)
	req := httptest.NewRequest(http.MethodPut, "/scans/11111111-1111-1111-1111-111111111111/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusInternalServerError)
}
