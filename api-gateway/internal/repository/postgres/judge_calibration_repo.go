package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JudgeCalibrationRecord struct {
	UserID      uuid.UUID
	SampleCount int
	ReportJSON  json.RawMessage
	GeneratedAt time.Time
	UpdatedAt   time.Time
}

type JudgeCalibrationRepository struct {
	pool *pgxpool.Pool
}

func NewJudgeCalibrationRepository(pool *pgxpool.Pool) *JudgeCalibrationRepository {
	return &JudgeCalibrationRepository{pool: pool}
}

func (r *JudgeCalibrationRepository) UpsertLatestByUser(
	ctx context.Context,
	userID uuid.UUID,
	sampleCount int,
	reportJSON json.RawMessage,
	generatedAt time.Time,
) error {
	if len(reportJSON) == 0 {
		reportJSON = []byte(`{}`)
	}
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO judge_calibrations (user_id, sample_count, report_json, generated_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			sample_count = EXCLUDED.sample_count,
			report_json = EXCLUDED.report_json,
			generated_at = EXCLUDED.generated_at,
			updated_at = NOW()`
	_, err := r.pool.Exec(ctx, query, userID, sampleCount, reportJSON, generatedAt)
	if err != nil {
		return fmt.Errorf("upsert judge calibration: %w", err)
	}
	return nil
}

func (r *JudgeCalibrationRepository) GetLatestByUser(ctx context.Context, userID uuid.UUID) (*JudgeCalibrationRecord, error) {
	query := `
		SELECT user_id, sample_count, report_json, generated_at, updated_at
		FROM judge_calibrations
		WHERE user_id = $1
		LIMIT 1`
	record := &JudgeCalibrationRecord{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&record.UserID,
		&record.SampleCount,
		&record.ReportJSON,
		&record.GeneratedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get judge calibration: %w", err)
	}
	return record, nil
}

func (r *JudgeCalibrationRepository) GetLatestJSONByUser(ctx context.Context, userID uuid.UUID) (json.RawMessage, error) {
	record, err := r.GetLatestByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return record.ReportJSON, nil
}
