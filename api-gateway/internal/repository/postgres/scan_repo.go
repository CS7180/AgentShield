package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/agentshield/api-gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type ScanRepository struct {
	pool *pgxpool.Pool
}

func NewScanRepository(pool *pgxpool.Pool) *ScanRepository {
	return &ScanRepository{pool: pool}
}

func (r *ScanRepository) Create(ctx context.Context, scan *domain.Scan) error {
	query := `
		INSERT INTO scans (id, user_id, target_endpoint, mode, attack_types, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		scan.ID,
		scan.UserID,
		scan.TargetEndpoint,
		string(scan.Mode),
		scan.AttackTypes,
		string(scan.Status),
	).Scan(&scan.CreatedAt, &scan.UpdatedAt)
}

func (r *ScanRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Scan, error) {
	query := `
		SELECT id, user_id, target_endpoint, mode, attack_types, status,
		       created_at, started_at, completed_at, updated_at
		FROM scans WHERE id = $1`

	scan := &domain.Scan{}
	var mode, status string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&scan.ID,
		&scan.UserID,
		&scan.TargetEndpoint,
		&mode,
		&scan.AttackTypes,
		&status,
		&scan.CreatedAt,
		&scan.StartedAt,
		&scan.CompletedAt,
		&scan.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get scan: %w", err)
	}
	scan.Mode = domain.ScanMode(mode)
	scan.Status = domain.ScanStatus(status)
	return scan, nil
}

func (r *ScanRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
	countQuery := `SELECT COUNT(*) FROM scans WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count scans: %w", err)
	}

	query := `
		SELECT id, user_id, target_endpoint, mode, attack_types, status,
		       created_at, started_at, completed_at, updated_at
		FROM scans WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list scans: %w", err)
	}
	defer rows.Close()

	var scans []*domain.Scan
	for rows.Next() {
		scan := &domain.Scan{}
		var mode, status string
		if err := rows.Scan(
			&scan.ID,
			&scan.UserID,
			&scan.TargetEndpoint,
			&mode,
			&scan.AttackTypes,
			&status,
			&scan.CreatedAt,
			&scan.StartedAt,
			&scan.CompletedAt,
			&scan.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan row: %w", err)
		}
		scan.Mode = domain.ScanMode(mode)
		scan.Status = domain.ScanStatus(status)
		scans = append(scans, scan)
	}

	return scans, total, rows.Err()
}

func (r *ScanRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ScanStatus) error {
	query := `UPDATE scans SET status = $2, updated_at = NOW() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, string(status))
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ScanRepository) MarkStarted(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE scans SET status = 'running', started_at = NOW(), updated_at = NOW() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark started: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ScanRepository) MarkStopped(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE scans SET status = 'stopped', completed_at = NOW(), updated_at = NOW() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark stopped: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
