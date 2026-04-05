package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

func (r *ReportRepository) GetByScanID(ctx context.Context, scanID uuid.UUID) (*domain.Report, error) {
	query := `
		SELECT id, scan_id, user_id, overall_score::double precision,
		       critical_count, high_count, medium_count, low_count,
		       owasp_scorecard, report_json, report_pdf_path, report_json_path,
		       report_bucket, created_at, updated_at
		FROM reports
		WHERE scan_id = $1
		LIMIT 1`

	report := &domain.Report{}
	var overallScore sql.NullFloat64
	var pdfPath sql.NullString
	var jsonPath sql.NullString

	err := r.pool.QueryRow(ctx, query, scanID).Scan(
		&report.ID,
		&report.ScanID,
		&report.UserID,
		&overallScore,
		&report.CriticalCount,
		&report.HighCount,
		&report.MediumCount,
		&report.LowCount,
		&report.OWASPScorecard,
		&report.ReportJSON,
		&pdfPath,
		&jsonPath,
		&report.ReportBucket,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get report by scan id: %w", err)
	}

	if overallScore.Valid {
		report.OverallScore = &overallScore.Float64
	}
	if pdfPath.Valid {
		report.ReportPDFPath = pdfPath.String
	}
	if jsonPath.Valid {
		report.ReportJSONPath = jsonPath.String
	}

	return report, nil
}

func (r *ReportRepository) UpsertByScanID(ctx context.Context, report *domain.Report) error {
	if report == nil {
		return fmt.Errorf("report is nil")
	}

	query := `
		INSERT INTO reports (
			id, scan_id, user_id, overall_score,
			critical_count, high_count, medium_count, low_count,
			owasp_scorecard, report_json, report_pdf_path, report_json_path, report_bucket,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12, $13,
			NOW(), NOW()
		)
		ON CONFLICT (scan_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			overall_score = EXCLUDED.overall_score,
			critical_count = EXCLUDED.critical_count,
			high_count = EXCLUDED.high_count,
			medium_count = EXCLUDED.medium_count,
			low_count = EXCLUDED.low_count,
			owasp_scorecard = EXCLUDED.owasp_scorecard,
			report_json = EXCLUDED.report_json,
			report_pdf_path = EXCLUDED.report_pdf_path,
			report_json_path = EXCLUDED.report_json_path,
			report_bucket = EXCLUDED.report_bucket,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	var overallScore any
	if report.OverallScore != nil {
		overallScore = *report.OverallScore
	}

	owasp := report.OWASPScorecard
	if len(owasp) == 0 {
		owasp = []byte(`{}`)
	}
	reportJSON := report.ReportJSON
	if len(reportJSON) == 0 {
		reportJSON = []byte(`{}`)
	}

	var pdfPath any
	if report.ReportPDFPath != "" {
		pdfPath = report.ReportPDFPath
	}

	var jsonPath any
	if report.ReportJSONPath != "" {
		jsonPath = report.ReportJSONPath
	}

	if report.ReportBucket == "" {
		report.ReportBucket = "agentshield-reports"
	}

	if err := r.pool.QueryRow(
		ctx,
		query,
		report.ID,
		report.ScanID,
		report.UserID,
		overallScore,
		report.CriticalCount,
		report.HighCount,
		report.MediumCount,
		report.LowCount,
		owasp,
		reportJSON,
		pdfPath,
		jsonPath,
		report.ReportBucket,
	).Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt); err != nil {
		return fmt.Errorf("upsert report by scan id: %w", err)
	}

	return nil
}
