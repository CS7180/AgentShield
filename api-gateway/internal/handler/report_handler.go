package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/agentshield/api-gateway/internal/repository/postgres"
	"github.com/agentshield/api-gateway/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ReportRepo is the subset of persistence operations required by ReportHandler.
type ReportRepo interface {
	GetByScanID(ctx context.Context, scanID uuid.UUID) (*domain.Report, error)
	UpsertByScanID(ctx context.Context, report *domain.Report) error
}

type ReportHandler struct {
	repo         ReportRepo
	uploader     storage.Uploader
	reportBucket string
	logger       *zap.Logger
}

type severityCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

type scoreComparison struct {
	Base  *float64 `json:"base,omitempty"`
	Other *float64 `json:"other,omitempty"`
	Delta *float64 `json:"delta,omitempty"`
	Trend string   `json:"trend"`
}

type countComparison struct {
	Base  severityCounts `json:"base"`
	Other severityCounts `json:"other"`
	Delta severityCounts `json:"delta"`
}

type compareResponse struct {
	BaseScanID     uuid.UUID       `json:"base_scan_id"`
	OtherScanID    uuid.UUID       `json:"other_scan_id"`
	OverallScore   scoreComparison `json:"overall_score"`
	SeverityCounts countComparison `json:"severity_counts"`
	GeneratedAt    time.Time       `json:"generated_at"`
}

type upsertReportRequest struct {
	OverallScore   *float64        `json:"overall_score"`
	CriticalCount  int             `json:"critical_count"`
	HighCount      int             `json:"high_count"`
	MediumCount    int             `json:"medium_count"`
	LowCount       int             `json:"low_count"`
	OWASPScorecard json.RawMessage `json:"owasp_scorecard"`
	ReportJSON     json.RawMessage `json:"report_json"`
	PDFBase64      string          `json:"pdf_base64,omitempty"`
}

func NewReportHandler(repo ReportRepo, uploader storage.Uploader, reportBucket string, logger *zap.Logger) *ReportHandler {
	if reportBucket == "" {
		reportBucket = "agentshield-reports"
	}
	return &ReportHandler{
		repo:         repo,
		uploader:     uploader,
		reportBucket: reportBucket,
		logger:       logger,
	}
}

func (h *ReportHandler) GetJSON(c *gin.Context) {
	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	report, err := h.repo.GetByScanID(c.Request.Context(), scanID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			abortNotFound(c, "report not found")
			return
		}
		h.logger.Error("get report", zap.Error(err))
		abortInternal(c, "failed to get report")
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Upsert(c *gin.Context) {
	if h.uploader == nil {
		rid := c.GetString(middleware.RequestIDKey)
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":       "report storage is not configured",
			"code":        "STORAGE_NOT_CONFIGURED",
			"status_code": http.StatusNotImplemented,
			"timestamp":   time.Now().UTC(),
			"request_id":  rid,
		})
		return
	}

	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	userIDStr := c.GetString(middleware.UserIDKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		abortBadRequest(c, "invalid user id", "INVALID_USER")
		return
	}

	var req upsertReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error(), "INVALID_REQUEST")
		return
	}

	if req.OverallScore != nil && (*req.OverallScore < 0 || *req.OverallScore > 100) {
		abortBadRequest(c, "overall_score must be between 0 and 100", "INVALID_REPORT_PAYLOAD")
		return
	}
	if req.CriticalCount < 0 || req.HighCount < 0 || req.MediumCount < 0 || req.LowCount < 0 {
		abortBadRequest(c, "severity counts must be non-negative", "INVALID_REPORT_PAYLOAD")
		return
	}
	if !isJSONObject(req.ReportJSON) {
		abortBadRequest(c, "report_json must be a valid JSON object", "INVALID_REPORT_PAYLOAD")
		return
	}
	if len(req.OWASPScorecard) > 0 && !isJSONObject(req.OWASPScorecard) {
		abortBadRequest(c, "owasp_scorecard must be a valid JSON object", "INVALID_REPORT_PAYLOAD")
		return
	}

	owaspScorecard := req.OWASPScorecard
	if len(owaspScorecard) == 0 {
		owaspScorecard = json.RawMessage(`{}`)
	}

	pdfPath := ""
	var pdfBytes []byte
	if req.PDFBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.PDFBase64)
		if err != nil {
			abortBadRequest(c, "pdf_base64 is invalid base64", "INVALID_REPORT_PAYLOAD")
			return
		}
		if len(decoded) == 0 {
			abortBadRequest(c, "pdf_base64 decoded payload is empty", "INVALID_REPORT_PAYLOAD")
			return
		}
		pdfBytes = decoded
		pdfPath = fmt.Sprintf("%s/scans/%s/report.pdf", userID.String(), scanID.String())
	}

	ctx := c.Request.Context()
	jsonPath := fmt.Sprintf("%s/scans/%s/report.json", userID.String(), scanID.String())
	if err := h.uploader.Upload(ctx, h.reportBucket, jsonPath, "application/json", req.ReportJSON); err != nil {
		h.logger.Error("upload report json", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to upload report json")
		return
	}

	if len(pdfBytes) > 0 {
		if err := h.uploader.Upload(ctx, h.reportBucket, pdfPath, "application/pdf", pdfBytes); err != nil {
			h.logger.Error("upload report pdf", zap.Error(err), zap.String("scan_id", scanID.String()))
			abortInternal(c, "failed to upload report pdf")
			return
		}
	}

	report := &domain.Report{
		ID:             uuid.New(),
		ScanID:         scanID,
		UserID:         userID,
		OverallScore:   req.OverallScore,
		CriticalCount:  req.CriticalCount,
		HighCount:      req.HighCount,
		MediumCount:    req.MediumCount,
		LowCount:       req.LowCount,
		OWASPScorecard: owaspScorecard,
		ReportJSON:     req.ReportJSON,
		ReportPDFPath:  pdfPath,
		ReportJSONPath: jsonPath,
		ReportBucket:   h.reportBucket,
	}

	if err := h.repo.UpsertByScanID(ctx, report); err != nil {
		h.logger.Error("upsert report", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to persist report")
		return
	}

	c.JSON(http.StatusOK, report)
}

func isJSONObject(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var obj map[string]any
	return json.Unmarshal(raw, &obj) == nil
}

func (h *ReportHandler) GetPDF(c *gin.Context) {
	scanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	report, err := h.repo.GetByScanID(c.Request.Context(), scanID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			abortNotFound(c, "report not found")
			return
		}
		h.logger.Error("get report for pdf", zap.Error(err))
		abortInternal(c, "failed to get report")
		return
	}

	if report.ReportPDFPath == "" {
		rid := c.GetString(middleware.RequestIDKey)
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "report pdf not found",
			"code":        "REPORT_PDF_NOT_FOUND",
			"status_code": http.StatusNotFound,
			"timestamp":   time.Now().UTC(),
			"request_id":  rid,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scan_id":     report.ScanID,
		"bucket":      report.ReportBucket,
		"object_path": report.ReportPDFPath,
	})
}

func (h *ReportHandler) Compare(c *gin.Context) {
	baseScanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		abortBadRequest(c, "invalid scan id", "INVALID_SCAN_ID")
		return
	}

	otherScanID, err := uuid.Parse(c.Param("other_id"))
	if err != nil {
		abortBadRequest(c, "invalid compare scan id", "INVALID_SCAN_ID")
		return
	}

	userIDStr := c.GetString(middleware.UserIDKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		abortBadRequest(c, "invalid user id", "INVALID_USER")
		return
	}

	baseReport, err := h.repo.GetByScanID(c.Request.Context(), baseScanID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			abortNotFound(c, "base report not found")
			return
		}
		h.logger.Error("get base report", zap.Error(err), zap.String("scan_id", baseScanID.String()))
		abortInternal(c, "failed to get base report")
		return
	}

	otherReport, err := h.repo.GetByScanID(c.Request.Context(), otherScanID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			abortNotFound(c, "comparison report not found")
			return
		}
		h.logger.Error("get comparison report", zap.Error(err), zap.String("scan_id", otherScanID.String()))
		abortInternal(c, "failed to get comparison report")
		return
	}

	if baseReport.UserID != userID || otherReport.UserID != userID {
		rid := c.GetString(middleware.RequestIDKey)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":       "access denied",
			"code":        "FORBIDDEN",
			"status_code": http.StatusForbidden,
			"timestamp":   time.Now().UTC(),
			"request_id":  rid,
		})
		return
	}

	c.JSON(http.StatusOK, compareResponse{
		BaseScanID:     baseScanID,
		OtherScanID:    otherScanID,
		OverallScore:   compareScores(baseReport.OverallScore, otherReport.OverallScore),
		SeverityCounts: compareSeverityCounts(baseReport, otherReport),
		GeneratedAt:    time.Now().UTC(),
	})
}

func compareScores(base, other *float64) scoreComparison {
	result := scoreComparison{
		Base:  base,
		Other: other,
		Trend: "unknown",
	}

	if base == nil || other == nil {
		return result
	}

	delta := *other - *base
	result.Delta = &delta

	switch {
	case delta > 0:
		result.Trend = "improved"
	case delta < 0:
		result.Trend = "regressed"
	default:
		result.Trend = "unchanged"
	}

	return result
}

func compareSeverityCounts(base, other *domain.Report) countComparison {
	baseCounts := severityCounts{
		Critical: base.CriticalCount,
		High:     base.HighCount,
		Medium:   base.MediumCount,
		Low:      base.LowCount,
	}
	otherCounts := severityCounts{
		Critical: other.CriticalCount,
		High:     other.HighCount,
		Medium:   other.MediumCount,
		Low:      other.LowCount,
	}

	return countComparison{
		Base:  baseCounts,
		Other: otherCounts,
		Delta: severityCounts{
			Critical: otherCounts.Critical - baseCounts.Critical,
			High:     otherCounts.High - baseCounts.High,
			Medium:   otherCounts.Medium - baseCounts.Medium,
			Low:      otherCounts.Low - baseCounts.Low,
		},
	}
}
