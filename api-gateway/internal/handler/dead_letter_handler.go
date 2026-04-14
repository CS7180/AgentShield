package handler

import (
	"context"
	"strconv"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeadLetterRepo interface {
	ListByScanID(ctx context.Context, scanID, userID uuid.UUID, limit, offset int) ([]*domain.ScanDeadLetter, int, error)
}

type DeadLetterHandler struct {
	repo   DeadLetterRepo
	logger *zap.Logger
}

func NewDeadLetterHandler(repo DeadLetterRepo, logger *zap.Logger) *DeadLetterHandler {
	return &DeadLetterHandler{repo: repo, logger: logger}
}

func (h *DeadLetterHandler) List(c *gin.Context) {
	scanID, userID, ok := parseScanAndUser(c)
	if !ok {
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

	items, total, err := h.repo.ListByScanID(c.Request.Context(), scanID, userID, limit, offset)
	if err != nil {
		h.logger.Error("list dead letters", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to list dead letters")
		return
	}
	if items == nil {
		items = []*domain.ScanDeadLetter{}
	}

	c.JSON(200, gin.H{
		"scan_id": scanID,
		"entries": items,
		"total":   total,
		"offset":  offset,
		"limit":   limit,
	})
}
