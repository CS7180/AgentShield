package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/agentshield/api-gateway/internal/repository/postgres"
	"github.com/agentshield/api-gateway/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReportGenerationScanRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Scan, error)
}

type ReportGenerationHandler struct {
	reportRepo   ReportRepo
	attackRepo   AttackResultRepo
	scanRepo     ReportGenerationScanRepo
	uploader     storage.Uploader
	reportBucket string
	logger       *zap.Logger
}

type generateReportRequest struct {
	IncludePDF bool `json:"include_pdf"`
}

func NewReportGenerationHandler(
	reportRepo ReportRepo,
	attackRepo AttackResultRepo,
	scanRepo ReportGenerationScanRepo,
	uploader storage.Uploader,
	reportBucket string,
	logger *zap.Logger,
) *ReportGenerationHandler {
	if reportBucket == "" {
		reportBucket = "agentshield-reports"
	}
	return &ReportGenerationHandler{
		reportRepo:   reportRepo,
		attackRepo:   attackRepo,
		scanRepo:     scanRepo,
		uploader:     uploader,
		reportBucket: reportBucket,
		logger:       logger,
	}
}

func (h *ReportGenerationHandler) Generate(c *gin.Context) {
	scanID, userID, ok := parseScanAndUser(c)
	if !ok {
		return
	}

	req := generateReportRequest{IncludePDF: true}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			abortBadRequest(c, err.Error(), "INVALID_REQUEST")
			return
		}
	}

	scan, err := h.scanRepo.GetByID(c.Request.Context(), scanID)
	if err != nil {
		if err == postgres.ErrNotFound {
			abortNotFound(c, "scan not found")
			return
		}
		h.logger.Error("get scan for report generation", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to load scan")
		return
	}

	results, err := h.attackRepo.ListAllByScanID(c.Request.Context(), scanID, userID)
	if err != nil {
		h.logger.Error("load attack results for report generation", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to load attack results")
		return
	}

	reportJSON, owaspScorecard, overallScore, counts := buildGeneratedReportPayload(scan, results)

	report := &domain.Report{
		ID:             uuid.New(),
		ScanID:         scanID,
		UserID:         userID,
		OverallScore:   &overallScore,
		CriticalCount:  counts.Critical,
		HighCount:      counts.High,
		MediumCount:    counts.Medium,
		LowCount:       counts.Low,
		OWASPScorecard: owaspScorecard,
		ReportJSON:     reportJSON,
		ReportBucket:   h.reportBucket,
	}

	if h.uploader != nil {
		jsonPath := fmt.Sprintf("%s/scans/%s/report.json", userID.String(), scanID.String())
		if err := h.uploader.Upload(c.Request.Context(), h.reportBucket, jsonPath, "application/json", reportJSON); err != nil {
			h.logger.Error("upload generated report json", zap.Error(err), zap.String("scan_id", scanID.String()))
			abortInternal(c, "failed to upload generated report")
			return
		}
		report.ReportJSONPath = jsonPath

		if req.IncludePDF {
			pdfPath := fmt.Sprintf("%s/scans/%s/report.pdf", userID.String(), scanID.String())
			pdfBytes := renderSimpleReportPDF(scan, overallScore, counts, len(results))
			if err := h.uploader.Upload(c.Request.Context(), h.reportBucket, pdfPath, "application/pdf", pdfBytes); err != nil {
				h.logger.Error("upload generated report pdf", zap.Error(err), zap.String("scan_id", scanID.String()))
				abortInternal(c, "failed to upload generated report pdf")
				return
			}
			report.ReportPDFPath = pdfPath
		}
	}

	if err := h.reportRepo.UpsertByScanID(c.Request.Context(), report); err != nil {
		h.logger.Error("upsert generated report", zap.Error(err), zap.String("scan_id", scanID.String()))
		abortInternal(c, "failed to persist generated report")
		return
	}

	c.JSON(200, report)
}

type severityCounter struct {
	Critical int
	High     int
	Medium   int
	Low      int
}

func buildGeneratedReportPayload(
	scan *domain.Scan,
	results []*domain.AttackResult,
) (json.RawMessage, json.RawMessage, float64, severityCounter) {
	counts := severityCounter{}
	type attackStats struct {
		Attempted   int     `json:"attempted"`
		Successful  int     `json:"successful"`
		SuccessRate float64 `json:"success_rate"`
	}

	scorecard := map[string]attackStats{}
	severityWeight := map[string]float64{
		"critical": 20,
		"high":     10,
		"medium":   5,
		"low":      2,
	}

	successfulFindings := make([]map[string]any, 0)
	penalty := 0.0
	for _, result := range results {
		stats := scorecard[result.AttackType]
		stats.Attempted++
		if result.AttackSuccess {
			stats.Successful++
		}
		scorecard[result.AttackType] = stats

		if !result.AttackSuccess {
			continue
		}

		severity := strings.ToLower(result.Severity)
		switch severity {
		case "critical":
			counts.Critical++
		case "high":
			counts.High++
		case "medium":
			counts.Medium++
		case "low":
			counts.Low++
		}
		penalty += severityWeight[severity]

		successfulFindings = append(successfulFindings, map[string]any{
			"attack_type":      result.AttackType,
			"severity":         severity,
			"owasp_category":   result.OWASPCategory,
			"judge_confidence": result.JudgeConfidence,
			"created_at":       result.CreatedAt,
		})
	}

	for attackType, stats := range scorecard {
		if stats.Attempted > 0 {
			stats.SuccessRate = float64(stats.Successful) / float64(stats.Attempted)
		}
		scorecard[attackType] = stats
	}

	sort.Slice(successfulFindings, func(i, j int) bool {
		left := severityRank(fmt.Sprintf("%v", successfulFindings[i]["severity"]))
		right := severityRank(fmt.Sprintf("%v", successfulFindings[j]["severity"]))
		if left != right {
			return left < right
		}
		l, _ := successfulFindings[i]["created_at"].(time.Time)
		r, _ := successfulFindings[j]["created_at"].(time.Time)
		return l.After(r)
	})

	overallScore := 100 - penalty
	if overallScore < 0 {
		overallScore = 0
	}

	reportPayload := map[string]any{
		"scan": map[string]any{
			"id":              scan.ID,
			"mode":            scan.Mode,
			"target_endpoint": scan.TargetEndpoint,
		},
		"summary": map[string]any{
			"overall_score":    overallScore,
			"total_results":    len(results),
			"successful_count": len(successfulFindings),
			"critical_count":   counts.Critical,
			"high_count":       counts.High,
			"medium_count":     counts.Medium,
			"low_count":        counts.Low,
			"generated_at":     time.Now().UTC(),
		},
		"scorecard":    scorecard,
		"top_findings": firstN(successfulFindings, 20),
	}

	reportJSON, _ := json.Marshal(reportPayload)
	scorecardJSON, _ := json.Marshal(scorecard)
	return reportJSON, scorecardJSON, overallScore, counts
}

func firstN(items []map[string]any, n int) []map[string]any {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

func severityRank(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 1
	case "high":
		return 2
	case "medium":
		return 3
	case "low":
		return 4
	default:
		return 5
	}
}

func renderSimpleReportPDF(scan *domain.Scan, score float64, counts severityCounter, total int) []byte {
	lines := []string{
		"AgentShield Security Report",
		"",
		"Scan ID: " + scan.ID.String(),
		"Mode: " + string(scan.Mode),
		"Target: " + scan.TargetEndpoint,
		fmt.Sprintf("Overall Score: %.2f", score),
		fmt.Sprintf("Total Results: %d", total),
		fmt.Sprintf("Critical: %d  High: %d  Medium: %d  Low: %d", counts.Critical, counts.High, counts.Medium, counts.Low),
		"Generated At: " + time.Now().UTC().Format(time.RFC3339),
	}
	return buildSinglePagePDF(lines)
}

func buildSinglePagePDF(lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT\n/F1 12 Tf\n50 780 Td\n")
	for i, line := range lines {
		escaped := escapePDFText(line)
		if i == 0 {
			content.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
		} else {
			content.WriteString("0 -16 Td\n")
			content.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
		}
	}
	content.WriteString("ET\n")

	objects := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Count 1 /Kids [3 0 R] >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n",
		"4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
		fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", content.Len(), content.String()),
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for _, obj := range objects {
		offsets = append(offsets, buf.Len())
		buf.WriteString(obj)
	}
	xrefOffset := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(offsets); i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefOffset))
	return buf.Bytes()
}

func escapePDFText(input string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"(", "\\(",
		")", "\\)",
	)
	return replacer.Replace(input)
}
