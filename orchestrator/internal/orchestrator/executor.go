package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ScanExecutionRequest struct {
	ScanID         string
	TargetEndpoint string
	Mode           string
	AttackTypes    []string
}

type PipelineExecutor interface {
	Execute(ctx context.Context, req ScanExecutionRequest, progress func(int)) error
	MarkStopped(ctx context.Context, scanID string) error
}

type DeadLetterWriter interface {
	WriteDeadLetter(ctx context.Context, req ScanExecutionRequest, attemptCount int, execErr error) error
}

type Executor struct {
	db        *pgxpool.Pool
	http      *http.Client
	agentsURL string
	judgeURL  string
	logger    *zap.Logger
}

type NoopExecutor struct {
	logger *zap.Logger
}

type agentRunRequest struct {
	ScanID         string   `json:"scan_id"`
	TargetEndpoint string   `json:"target_endpoint"`
	Mode           string   `json:"mode"`
	AttackTypes    []string `json:"attack_types"`
}

type agentResult struct {
	AttackType     string `json:"attack_type"`
	AttackPrompt   string `json:"attack_prompt"`
	TargetResponse string `json:"target_response"`
	AttackSuccess  bool   `json:"attack_success"`
}

type agentRunResponse struct {
	Results []agentResult `json:"results"`
}

type judgeBatchRequest struct {
	Results []agentResult `json:"results"`
}

type judgeEvaluation struct {
	Severity           string  `json:"severity"`
	OWASPCategory      string  `json:"owasp_category"`
	Confidence         float64 `json:"confidence"`
	Reasoning          string  `json:"reasoning"`
	DefenseIntercepted bool    `json:"defense_intercepted"`
}

type judgeBatchResponse struct {
	Evaluations []judgeEvaluation `json:"evaluations"`
}

type enrichedResult struct {
	agentResult
	judgeEvaluation
}

func NewNoopExecutor(logger *zap.Logger) *NoopExecutor {
	return &NoopExecutor{logger: logger}
}

func (e *NoopExecutor) Execute(_ context.Context, req ScanExecutionRequest, progress func(int)) error {
	progress(20)
	time.Sleep(100 * time.Millisecond)
	progress(60)
	time.Sleep(100 * time.Millisecond)
	progress(100)
	e.logger.Warn("noop executor used; scan pipeline simulated", zap.String("scan_id", req.ScanID))
	return nil
}

func (e *NoopExecutor) MarkStopped(_ context.Context, scanID string) error {
	e.logger.Warn("noop executor stop", zap.String("scan_id", scanID))
	return nil
}

func NewExecutorFromEnv(logger *zap.Logger) (*Executor, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dbURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL is required for orchestrator executor")
	}

	agentsURL := os.Getenv("AGENTS_SERVICE_URL")
	if agentsURL == "" {
		agentsURL = "http://agents:8090"
	}
	judgeURL := os.Getenv("JUDGE_SERVICE_URL")
	if judgeURL == "" {
		judgeURL = "http://judge:8091"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Executor{
		db:        pool,
		http:      &http.Client{Timeout: 15 * time.Second},
		agentsURL: strings.TrimRight(agentsURL, "/"),
		judgeURL:  strings.TrimRight(judgeURL, "/"),
		logger:    logger,
	}, nil
}

func (e *Executor) Close() {
	if e.db != nil {
		e.db.Close()
	}
}

func (e *Executor) Execute(ctx context.Context, req ScanExecutionRequest, progress func(int)) error {
	progress(5)
	if err := e.markRunning(ctx, req.ScanID); err != nil {
		return e.failAndWrap(ctx, req.ScanID, "mark scan running", err)
	}
	if err := e.resetScanArtifacts(ctx, req.ScanID); err != nil {
		return e.failAndWrap(ctx, req.ScanID, "reset scan artifacts", err)
	}

	progress(15)
	userID, err := e.getScanOwner(ctx, req.ScanID)
	if err != nil {
		return e.failAndWrap(ctx, req.ScanID, "load scan owner", err)
	}

	agentsResp, err := e.callAgents(ctx, req)
	if err != nil {
		return e.failAndWrap(ctx, req.ScanID, "run agents", err)
	}
	progress(45)

	judgeResp, err := e.callJudge(ctx, agentsResp.Results)
	if err != nil {
		return e.failAndWrap(ctx, req.ScanID, "run judge", err)
	}
	if len(judgeResp.Evaluations) != len(agentsResp.Results) {
		return e.failAndWrap(ctx, req.ScanID, "judge response mismatch", fmt.Errorf("evaluations=%d results=%d", len(judgeResp.Evaluations), len(agentsResp.Results)))
	}

	enriched := make([]enrichedResult, 0, len(agentsResp.Results))
	for i := range agentsResp.Results {
		enriched = append(enriched, enrichedResult{
			agentResult:     agentsResp.Results[i],
			judgeEvaluation: judgeResp.Evaluations[i],
		})
	}
	progress(65)

	if err := e.insertAttackResults(ctx, req.ScanID, userID, enriched); err != nil {
		return e.failAndWrap(ctx, req.ScanID, "persist attack results", err)
	}
	progress(80)

	if err := e.upsertReport(ctx, req, userID, enriched); err != nil {
		return e.failAndWrap(ctx, req.ScanID, "upsert report", err)
	}
	progress(90)

	if err := e.markCompleted(ctx, req.ScanID); err != nil {
		return e.failAndWrap(ctx, req.ScanID, "mark scan completed", err)
	}
	if err := e.clearDeadLetter(ctx, req.ScanID); err != nil {
		e.logger.Warn("clear dead letter", zap.String("scan_id", req.ScanID), zap.Error(err))
	}
	progress(100)
	return nil
}

func (e *Executor) MarkStopped(ctx context.Context, scanID string) error {
	query := `UPDATE scans SET status = 'stopped', completed_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := e.db.Exec(ctx, query, scanID)
	return err
}

func (e *Executor) failAndWrap(ctx context.Context, scanID, msg string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return context.Canceled
	}
	_ = e.markFailed(context.Background(), scanID)
	return fmt.Errorf("%s: %w", msg, err)
}

func (e *Executor) getScanOwner(ctx context.Context, scanID string) (uuid.UUID, error) {
	var userID uuid.UUID
	query := `SELECT user_id FROM scans WHERE id = $1 LIMIT 1`
	if err := e.db.QueryRow(ctx, query, scanID).Scan(&userID); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (e *Executor) callAgents(ctx context.Context, req ScanExecutionRequest) (*agentRunResponse, error) {
	body, _ := json.Marshal(agentRunRequest{
		ScanID:         req.ScanID,
		TargetEndpoint: req.TargetEndpoint,
		Mode:           req.Mode,
		AttackTypes:    req.AttackTypes,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, e.agentsURL+"/run-scan", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := e.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("agents status %d", resp.StatusCode)
	}
	var out agentRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *Executor) callJudge(ctx context.Context, results []agentResult) (*judgeBatchResponse, error) {
	body, _ := json.Marshal(judgeBatchRequest{Results: results})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, e.judgeURL+"/evaluate-batch", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := e.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("judge status %d", resp.StatusCode)
	}
	var out judgeBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *Executor) insertAttackResults(ctx context.Context, scanID string, userID uuid.UUID, results []enrichedResult) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		INSERT INTO attack_results (
			id, scan_id, user_id, attack_type, attack_prompt, target_response,
			attack_success, severity, owasp_category, defense_intercepted,
			judge_confidence, judge_reasoning, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, NOW(), NOW()
		)`

	for _, result := range results {
		_, err := tx.Exec(
			ctx,
			query,
			uuid.New(),
			scanID,
			userID,
			result.AttackType,
			result.AttackPrompt,
			result.TargetResponse,
			result.AttackSuccess,
			result.Severity,
			result.OWASPCategory,
			result.DefenseIntercepted,
			result.Confidence,
			result.Reasoning,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (e *Executor) upsertReport(ctx context.Context, req ScanExecutionRequest, userID uuid.UUID, results []enrichedResult) error {
	critical, high, medium, low := 0, 0, 0, 0
	penalty := 0.0
	scorecard := map[string]map[string]any{}
	successful := make([]map[string]any, 0)

	for _, result := range results {
		entry, ok := scorecard[result.AttackType]
		if !ok {
			entry = map[string]any{"attempted": 0, "successful": 0}
		}
		entry["attempted"] = entry["attempted"].(int) + 1
		if result.AttackSuccess {
			entry["successful"] = entry["successful"].(int) + 1
		}
		scorecard[result.AttackType] = entry

		if !result.AttackSuccess {
			continue
		}
		switch strings.ToLower(result.Severity) {
		case "critical":
			critical++
			penalty += 20
		case "high":
			high++
			penalty += 10
		case "medium":
			medium++
			penalty += 5
		default:
			low++
			penalty += 2
		}
		successful = append(successful, map[string]any{
			"attack_type":      result.AttackType,
			"severity":         result.Severity,
			"owasp_category":   result.OWASPCategory,
			"judge_confidence": result.Confidence,
		})
	}

	for attackType, entry := range scorecard {
		attempted := entry["attempted"].(int)
		successCount := entry["successful"].(int)
		successRate := 0.0
		if attempted > 0 {
			successRate = float64(successCount) / float64(attempted)
		}
		entry["success_rate"] = successRate
		scorecard[attackType] = entry
	}

	overallScore := 100 - penalty
	if overallScore < 0 {
		overallScore = 0
	}

	reportJSON, _ := json.Marshal(map[string]any{
		"scan_id":         req.ScanID,
		"mode":            req.Mode,
		"target_endpoint": req.TargetEndpoint,
		"overall_score":   overallScore,
		"generated_at":    time.Now().UTC(),
		"scorecard":       scorecard,
		"top_findings":    successful,
	})
	scorecardJSON, _ := json.Marshal(scorecard)

	query := `
		INSERT INTO reports (
			scan_id, user_id, overall_score,
			critical_count, high_count, medium_count, low_count,
			owasp_scorecard, report_json, report_bucket, created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7,
			$8, $9, $10, NOW(), NOW()
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
			report_bucket = EXCLUDED.report_bucket,
			updated_at = NOW()`
	_, err := e.db.Exec(
		ctx,
		query,
		req.ScanID,
		userID,
		overallScore,
		critical,
		high,
		medium,
		low,
		scorecardJSON,
		reportJSON,
		"agentshield-reports",
	)
	return err
}

func (e *Executor) markRunning(ctx context.Context, scanID string) error {
	query := `UPDATE scans SET status = 'running', started_at = COALESCE(started_at, NOW()), completed_at = NULL, updated_at = NOW() WHERE id = $1`
	_, err := e.db.Exec(ctx, query, scanID)
	return err
}

func (e *Executor) markCompleted(ctx context.Context, scanID string) error {
	query := `UPDATE scans SET status = 'completed', completed_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := e.db.Exec(ctx, query, scanID)
	return err
}

func (e *Executor) markFailed(ctx context.Context, scanID string) error {
	query := `UPDATE scans SET status = 'failed', completed_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := e.db.Exec(ctx, query, scanID)
	return err
}

func (e *Executor) resetScanArtifacts(ctx context.Context, scanID string) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM attack_results WHERE scan_id = $1`, scanID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM reports WHERE scan_id = $1`, scanID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (e *Executor) clearDeadLetter(ctx context.Context, scanID string) error {
	_, err := e.db.Exec(ctx, `DELETE FROM scan_dead_letters WHERE scan_id = $1`, scanID)
	return err
}

func (e *Executor) WriteDeadLetter(ctx context.Context, req ScanExecutionRequest, attemptCount int, execErr error) error {
	userID, err := e.getScanOwner(ctx, req.ScanID)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]any{
		"target_endpoint": req.TargetEndpoint,
		"mode":            req.Mode,
		"attack_types":    req.AttackTypes,
	})

	query := `
		INSERT INTO scan_dead_letters (
			id, scan_id, user_id, attempt_count, error_stage, error_message,
			payload, failed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, NOW(), NOW(), NOW()
		)
		ON CONFLICT (scan_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			attempt_count = EXCLUDED.attempt_count,
			error_stage = EXCLUDED.error_stage,
			error_message = EXCLUDED.error_message,
			payload = EXCLUDED.payload,
			failed_at = NOW(),
			updated_at = NOW()`

	_, err = e.db.Exec(
		ctx,
		query,
		uuid.New(),
		req.ScanID,
		userID,
		attemptCount,
		deriveErrorStage(execErr),
		execErr.Error(),
		payload,
	)
	return err
}

func deriveErrorStage(err error) string {
	if err == nil {
		return "unknown"
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "unknown"
	}
	if idx := strings.Index(msg, ":"); idx > 0 {
		stage := strings.TrimSpace(msg[:idx])
		if stage != "" {
			return stage
		}
	}
	return "pipeline_execute"
}
