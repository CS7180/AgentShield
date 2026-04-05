package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Report stores aggregated scan output and artifact paths.
type Report struct {
	ID             uuid.UUID       `json:"id"`
	ScanID         uuid.UUID       `json:"scan_id"`
	UserID         uuid.UUID       `json:"user_id"`
	OverallScore   *float64        `json:"overall_score,omitempty"`
	CriticalCount  int             `json:"critical_count"`
	HighCount      int             `json:"high_count"`
	MediumCount    int             `json:"medium_count"`
	LowCount       int             `json:"low_count"`
	OWASPScorecard json.RawMessage `json:"owasp_scorecard"`
	ReportJSON     json.RawMessage `json:"report_json"`
	ReportPDFPath  string          `json:"report_pdf_path,omitempty"`
	ReportJSONPath string          `json:"report_json_path,omitempty"`
	ReportBucket   string          `json:"report_bucket"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
