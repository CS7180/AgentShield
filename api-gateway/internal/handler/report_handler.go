package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ReportRepo is the subset of persistence operations required by ReportHandler.
type ReportRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Scan, error)
}

type ReportHandler struct {
	repo   ReportRepo
	logger *zap.Logger
}

func NewReportHandler(repo ReportRepo, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{repo: repo, logger: logger}
}

func (h *ReportHandler) GetJSON(c *gin.Context) {
	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	scan, err := h.repo.GetByID(c.Request.Context(), scanID)
	if err != nil {
		h.logger.Error("get scan for report", zap.Error(err))
		abortNotFound(c, "scan not found")
		return
	}

	c.JSON(http.StatusOK, scan)
}

func (h *ReportHandler) GetPDF(c *gin.Context) {
	rid := c.GetString(middleware.RequestIDKey)
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":       "PDF report generation not yet implemented",
		"code":        "NOT_IMPLEMENTED",
		"status_code": http.StatusNotImplemented,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}

func (h *ReportHandler) Compare(c *gin.Context) {
	rid := c.GetString(middleware.RequestIDKey)
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":       "scan comparison not yet implemented",
		"code":        "NOT_IMPLEMENTED",
		"status_code": http.StatusNotImplemented,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}
