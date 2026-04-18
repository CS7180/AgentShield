package postgres

import (
	"context"
	"fmt"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScanDeadLetterRepository struct {
	pool *pgxpool.Pool
}

func NewScanDeadLetterRepository(pool *pgxpool.Pool) *ScanDeadLetterRepository {
	return &ScanDeadLetterRepository{pool: pool}
}

func (r *ScanDeadLetterRepository) ListByScanID(ctx context.Context, scanID, userID uuid.UUID, limit, offset int) ([]*domain.ScanDeadLetter, int, error) {
	countQuery := `SELECT COUNT(*) FROM scan_dead_letters WHERE scan_id = $1 AND user_id = $2`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, scanID, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count dead letters: %w", err)
	}

	query := `
		SELECT id, scan_id, user_id, attempt_count, error_stage, error_message,
		       payload, failed_at, created_at, updated_at
		FROM scan_dead_letters
		WHERE scan_id = $1 AND user_id = $2
		ORDER BY failed_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, scanID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list dead letters: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.ScanDeadLetter, 0)
	for rows.Next() {
		item := &domain.ScanDeadLetter{}
		if err := rows.Scan(
			&item.ID,
			&item.ScanID,
			&item.UserID,
			&item.AttemptCount,
			&item.ErrorStage,
			&item.ErrorMessage,
			&item.Payload,
			&item.FailedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan dead letter row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate dead letters: %w", err)
	}

	return items, total, nil
}
