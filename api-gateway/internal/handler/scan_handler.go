package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/metrics"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/agentshield/api-gateway/internal/repository/postgres"
	"github.com/agentshield/api-gateway/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrchestratorClient interface {
	StartScan(ctx context.Context, scanID, targetEndpoint, mode string, attackTypes []string) (accepted bool, message string, err error)
	StopScan(ctx context.Context, scanID string) (stopped bool, message string, err error)
}

type ScanRepo interface {
	Create(ctx context.Context, scan *domain.Scan) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Scan, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ScanStatus) error
	MarkStarted(ctx context.Context, id uuid.UUID) error
	MarkStopped(ctx context.Context, id uuid.UUID) error
}

type ScanHandler struct {
	repo         ScanRepo
	orchestrator OrchestratorClient
	logger       *zap.Logger
}

func NewScanHandler(repo ScanRepo, orchestrator OrchestratorClient, logger *zap.Logger) *ScanHandler {
	return &ScanHandler{repo: repo, orchestrator: orchestrator, logger: logger}
}

func (h *ScanHandler) Create(c *gin.Context) {
	var req domain.CreateScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error(), "INVALID_REQUEST")
		return
	}
	if err := validation.Validate.Struct(req); err != nil {
		abortBadRequest(c, err.Error(), "INVALID_TARGET_ENDPOINT")
		return
	}

	userIDStr := c.GetString(middleware.UserIDKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		abortBadRequest(c, "invalid user id", "INVALID_USER")
		return
	}

	scan := &domain.Scan{
		ID:             uuid.New(),
		UserID:         userID,
		TargetEndpoint: req.TargetEndpoint,
		Mode:           req.Mode,
		AttackTypes:    req.AttackTypes,
		Status:         domain.StatusPending,
	}

	if err := h.repo.Create(c.Request.Context(), scan); err != nil {
		h.logger.Error("create scan", zap.Error(err))
		abortInternal(c, "failed to create scan")
		return
	}

	metrics.ScanCreationTotal.Inc()
	c.JSON(http.StatusCreated, scan)
}

func (h *ScanHandler) List(c *gin.Context) {
	userIDStr := c.GetString(middleware.UserIDKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		abortBadRequest(c, "invalid user id", "INVALID_USER")
		return
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	scans, total, err := h.repo.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("list scans", zap.Error(err))
		abortInternal(c, "failed to list scans")
		return
	}

	if scans == nil {
		scans = []*domain.Scan{}
	}

	c.JSON(http.StatusOK, domain.ScanListResponse{
		Scans:  scans,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	})
}

func (h *ScanHandler) Get(c *gin.Context) {
	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	scan, err := h.repo.GetByID(c.Request.Context(), scanID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			abortNotFound(c, "scan not found")
			return
		}
		h.logger.Error("get scan", zap.Error(err))
		abortInternal(c, "failed to get scan")
		return
	}

	c.JSON(http.StatusOK, scan)
}

func (h *ScanHandler) Start(c *gin.Context) {
	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	scan, err := h.repo.GetByID(c.Request.Context(), scanID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			abortNotFound(c, "scan not found")
			return
		}
		abortInternal(c, "failed to get scan")
		return
	}

	if scan.Status != domain.StatusPending && scan.Status != domain.StatusQueued {
		abortBadRequest(c, "scan is not in a startable state", "INVALID_SCAN_STATE")
		return
	}

	accepted, message, err := h.orchestrator.StartScan(
		c.Request.Context(),
		scanID.String(),
		scan.TargetEndpoint,
		string(scan.Mode),
		scan.AttackTypes,
	)

	if err != nil || !accepted {
		// Orchestrator unavailable — queue the scan
		_ = h.repo.UpdateStatus(c.Request.Context(), scanID, domain.StatusQueued)
		c.JSON(http.StatusAccepted, gin.H{
			"scan_id": scanID,
			"status":  domain.StatusQueued,
			"message": "scan queued; orchestrator unavailable",
		})
		return
	}

	_ = h.repo.MarkStarted(c.Request.Context(), scanID)
	c.JSON(http.StatusAccepted, gin.H{
		"scan_id": scanID,
		"status":  domain.StatusRunning,
		"message": message,
	})
}

func (h *ScanHandler) Stop(c *gin.Context) {
	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	scan, err := h.repo.GetByID(c.Request.Context(), scanID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			abortNotFound(c, "scan not found")
			return
		}
		abortInternal(c, "failed to get scan")
		return
	}

	if scan.Status != domain.StatusRunning && scan.Status != domain.StatusQueued {
		abortBadRequest(c, "scan is not running", "INVALID_SCAN_STATE")
		return
	}

	_, _, _ = h.orchestrator.StopScan(c.Request.Context(), scanID.String())
	_ = h.repo.MarkStopped(c.Request.Context(), scanID)

	c.JSON(http.StatusOK, gin.H{
		"scan_id": scanID,
		"status":  domain.StatusStopped,
		"message": "scan stopped",
	})
}

// Shared error helpers

func abortBadRequest(c *gin.Context, msg, code string) {
	rid := c.GetString(middleware.RequestIDKey)
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error":       msg,
		"code":        code,
		"status_code": http.StatusBadRequest,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}

func abortNotFound(c *gin.Context, msg string) {
	rid := c.GetString(middleware.RequestIDKey)
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
		"error":       msg,
		"code":        "NOT_FOUND",
		"status_code": http.StatusNotFound,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}

func abortInternal(c *gin.Context, msg string) {
	rid := c.GetString(middleware.RequestIDKey)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error":       msg,
		"code":        "INTERNAL_ERROR",
		"status_code": http.StatusInternalServerError,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}
