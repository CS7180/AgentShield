package handler

import (
	"net/http"
	"time"

	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler { return &ReportHandler{} }

func (h *ReportHandler) GetJSON(c *gin.Context) {
	// Sprint 2: aggregate results from attack_results table
	rid := c.GetString(middleware.RequestIDKey)
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":       "report generation not yet implemented",
		"code":        "NOT_IMPLEMENTED",
		"status_code": http.StatusNotImplemented,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
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
