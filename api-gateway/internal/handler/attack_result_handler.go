package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/agentshield/api-gateway/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AttackResultRepo interface {
	CreateBatch(ctx context.Context, scanID uuid.UUID, userID uuid.UUID, inputs []domain.AttackResultInput) ([]*domain.AttackResult, error)
	ListByScanID(ctx context.Context, scanID uuid.UUID, userID uuid.UUID, limit int, offset int) ([]*domain.AttackResult, int, error)
	ListAllByScanID(ctx context.Context, scanID uuid.UUID, userID uuid.UUID) ([]*domain.AttackResult, error)
}

type AttackResultHandler struct {
	repo   AttackResultRepo
	logger *zap.Logger
}

func NewAttackResultHandler(repo AttackResultRepo, logger *zap.Logger) *AttackResultHandler {
	return &AttackResultHandler{repo: repo, logger: logger}
}

func (h *AttackResultHandler) CreateBatch(c *gin.Context) {
	scanID, userID, ok := parseScanAndUser(c)
	if !ok {
		return
	}

	var req domain.CreateAttackResultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error(), "INVALID_REQUEST")
		return
	}
	if err := validation.Validate.Struct(req); err != nil {
		abortBadRequest(c, err.Error(), "INVALID_REQUEST")
		return
	}

	results, err := h.repo.CreateBatch(c.Request.Context(), scanID, userID, req.Results)
	if err != nil {
		h.logger.Error("create attack results", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to create attack results")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"scan_id": scanID,
		"count":   len(results),
		"results": results,
	})
}

func (h *AttackResultHandler) List(c *gin.Context) {
	scanID, userID, ok := parseScanAndUser(c)
	if !ok {
		return
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	results, total, err := h.repo.ListByScanID(c.Request.Context(), scanID, userID, limit, offset)
	if err != nil {
		h.logger.Error("list attack results", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to list attack results")
		return
	}
	if results == nil {
		results = []*domain.AttackResult{}
	}

	c.JSON(http.StatusOK, domain.AttackResultListResponse{
		Results: results,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
	})
}

func parseScanAndUser(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return uuid.Nil, uuid.Nil, false
	}
	userID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		abortBadRequest(c, "invalid user id", "INVALID_USER")
		return uuid.Nil, uuid.Nil, false
	}
	return scanID, userID, true
}
