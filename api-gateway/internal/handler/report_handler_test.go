package handler_test

import (
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

// fakeReportRepo satisfies handler.ReportRepo.
type fakeReportRepo struct {
	scan *domain.Scan
	err  error
}

func (f *fakeReportRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Scan, error) {
	return f.scan, f.err
}

func newReportRouter(h *handler.ReportHandler, scanID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/scans/:id/report", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, uuid.New().String())
		h.GetJSON(c)
	})
	_ = scanID
	return r
}

func TestGetReport_Returns200_WhenScanExists(t *testing.T) {
	scan := &domain.Scan{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		TargetEndpoint: "https://example.com",
		Mode:           domain.ModeRedTeam,
		AttackTypes:    []string{"prompt_injection"},
		Status:         domain.StatusCompleted,
	}
	repo := &fakeReportRepo{scan: scan}
	h := handler.NewReportHandler(repo, zap.NewNop())
	r := newReportRouter(h, scan.ID)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scan.ID.String()+"/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
}

func TestGetReport_Returns404_WhenScanNotFound(t *testing.T) {
	repo := &fakeReportRepo{scan: nil, err: errors.New("not found")}
	h := handler.NewReportHandler(repo, zap.NewNop())
	r := newReportRouter(h, uuid.New())

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.New().String()+"/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusNotFound)
}

func TestGetReport_Returns400_WhenInvalidScanID(t *testing.T) {
	repo := &fakeReportRepo{}
	h := handler.NewReportHandler(repo, zap.NewNop())
	r := newReportRouter(h, uuid.New())

	req := httptest.NewRequest(http.MethodGet, "/scans/not-a-uuid/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
}
