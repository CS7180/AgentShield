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

type AttackResultRepository struct {
	pool *pgxpool.Pool
}

func NewAttackResultRepository(pool *pgxpool.Pool) *AttackResultRepository {
	return &AttackResultRepository{pool: pool}
}

func (r *AttackResultRepository) CreateBatch(
	ctx context.Context,
	scanID uuid.UUID,
	userID uuid.UUID,
	inputs []domain.AttackResultInput,
) ([]*domain.AttackResult, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("inputs are empty")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		INSERT INTO attack_results (
			id, scan_id, user_id, attack_type, attack_prompt, target_response, attack_success,
			severity, owasp_category, defense_intercepted, judge_confidence, judge_reasoning,
			latency_ms, tokens_used, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, NOW(), NOW()
		)
		RETURNING created_at, updated_at`

	results := make([]*domain.AttackResult, 0, len(inputs))
	for _, in := range inputs {
		ar := &domain.AttackResult{
			ID:                 uuid.New(),
			ScanID:             scanID,
			UserID:             userID,
			AttackType:         in.AttackType,
			AttackPrompt:       in.AttackPrompt,
			TargetResponse:     in.TargetResponse,
			AttackSuccess:      in.AttackSuccess,
			Severity:           in.Severity,
			OWASPCategory:      in.OWASPCategory,
			DefenseIntercepted: in.DefenseIntercepted,
			JudgeConfidence:    in.JudgeConfidence,
			JudgeReasoning:     in.JudgeReasoning,
			LatencyMs:          in.LatencyMs,
			TokensUsed:         in.TokensUsed,
		}

		if err := tx.QueryRow(
			ctx,
			query,
			ar.ID,
			ar.ScanID,
			ar.UserID,
			ar.AttackType,
			ar.AttackPrompt,
			ar.TargetResponse,
			ar.AttackSuccess,
			nullIfEmpty(ar.Severity),
			nullIfEmpty(ar.OWASPCategory),
			ar.DefenseIntercepted,
			ar.JudgeConfidence,
			nullIfEmpty(ar.JudgeReasoning),
			ar.LatencyMs,
			ar.TokensUsed,
		).Scan(&ar.CreatedAt, &ar.UpdatedAt); err != nil {
			return nil, fmt.Errorf("insert attack result: %w", err)
		}
		results = append(results, ar)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return results, nil
}

func (r *AttackResultRepository) ListByScanID(
	ctx context.Context,
	scanID uuid.UUID,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]*domain.AttackResult, int, error) {
	countQuery := `SELECT COUNT(*) FROM attack_results WHERE scan_id = $1 AND user_id = $2`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, scanID, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count attack results: %w", err)
	}

	query := `
		SELECT id, scan_id, user_id, attack_type, attack_prompt, target_response, attack_success,
		       severity, owasp_category, defense_intercepted, judge_confidence, judge_reasoning,
		       latency_ms, tokens_used, created_at, updated_at
		FROM attack_results
		WHERE scan_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, scanID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list attack results: %w", err)
	}
	defer rows.Close()

	results := make([]*domain.AttackResult, 0)
	for rows.Next() {
		ar, err := scanAttackResult(rows)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, ar)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate attack results: %w", err)
	}

	return results, total, nil
}

func (r *AttackResultRepository) ListAllByScanID(
	ctx context.Context,
	scanID uuid.UUID,
	userID uuid.UUID,
) ([]*domain.AttackResult, error) {
	query := `
		SELECT id, scan_id, user_id, attack_type, attack_prompt, target_response, attack_success,
		       severity, owasp_category, defense_intercepted, judge_confidence, judge_reasoning,
		       latency_ms, tokens_used, created_at, updated_at
		FROM attack_results
		WHERE scan_id = $1 AND user_id = $2
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, scanID, userID)
	if err != nil {
		return nil, fmt.Errorf("list all attack results: %w", err)
	}
	defer rows.Close()

	results := make([]*domain.AttackResult, 0)
	for rows.Next() {
		ar, err := scanAttackResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, ar)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all attack results: %w", err)
	}
	return results, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAttackResult(scanner rowScanner) (*domain.AttackResult, error) {
	ar := &domain.AttackResult{}
	var severity sql.NullString
	var owaspCategory sql.NullString
	var defenseIntercepted sql.NullBool
	var judgeConfidence sql.NullFloat64
	var judgeReasoning sql.NullString
	var latencyMs sql.NullInt32
	var tokensUsed sql.NullInt32

	if err := scanner.Scan(
		&ar.ID,
		&ar.ScanID,
		&ar.UserID,
		&ar.AttackType,
		&ar.AttackPrompt,
		&ar.TargetResponse,
		&ar.AttackSuccess,
		&severity,
		&owaspCategory,
		&defenseIntercepted,
		&judgeConfidence,
		&judgeReasoning,
		&latencyMs,
		&tokensUsed,
		&ar.CreatedAt,
		&ar.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan attack result: %w", err)
	}

	if severity.Valid {
		ar.Severity = severity.String
	}
	if owaspCategory.Valid {
		ar.OWASPCategory = owaspCategory.String
	}
	if defenseIntercepted.Valid {
		v := defenseIntercepted.Bool
		ar.DefenseIntercepted = &v
	}
	if judgeConfidence.Valid {
		v := judgeConfidence.Float64
		ar.JudgeConfidence = &v
	}
	if judgeReasoning.Valid {
		ar.JudgeReasoning = judgeReasoning.String
	}
	if latencyMs.Valid {
		v := int(latencyMs.Int32)
		ar.LatencyMs = &v
	}
	if tokensUsed.Valid {
		v := int(tokensUsed.Int32)
		ar.TokensUsed = &v
	}
	return ar, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
