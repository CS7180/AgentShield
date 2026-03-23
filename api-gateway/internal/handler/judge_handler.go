package handler

import (
	"net/http"
	"time"

	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type JudgeHandler struct{}

func NewJudgeHandler() *JudgeHandler { return &JudgeHandler{} }

func (h *JudgeHandler) Calibrate(c *gin.Context) {
	rid := c.GetString(middleware.RequestIDKey)
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":       "judge calibration not yet implemented",
		"code":        "NOT_IMPLEMENTED",
		"status_code": http.StatusNotImplemented,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}

func (h *JudgeHandler) CalibrationReport(c *gin.Context) {
	rid := c.GetString(middleware.RequestIDKey)
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":       "calibration report not yet implemented",
		"code":        "NOT_IMPLEMENTED",
		"status_code": http.StatusNotImplemented,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}
