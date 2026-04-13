package domain

import (
	"time"

	"github.com/google/uuid"
)

type AttackResult struct {
	ID                 uuid.UUID `json:"id"`
	ScanID             uuid.UUID `json:"scan_id"`
	UserID             uuid.UUID `json:"user_id"`
	AttackType         string    `json:"attack_type"`
	AttackPrompt       string    `json:"attack_prompt"`
	TargetResponse     string    `json:"target_response"`
	AttackSuccess      bool      `json:"attack_success"`
	Severity           string    `json:"severity,omitempty"`
	OWASPCategory      string    `json:"owasp_category,omitempty"`
	DefenseIntercepted *bool     `json:"defense_intercepted,omitempty"`
	JudgeConfidence    *float64  `json:"judge_confidence,omitempty"`
	JudgeReasoning     string    `json:"judge_reasoning,omitempty"`
	LatencyMs          *int      `json:"latency_ms,omitempty"`
	TokensUsed         *int      `json:"tokens_used,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AttackResultInput struct {
	AttackType         string   `json:"attack_type" validate:"required,oneof=prompt_injection jailbreak data_leakage constraint_drift"`
	AttackPrompt       string   `json:"attack_prompt" validate:"required,min=1"`
	TargetResponse     string   `json:"target_response" validate:"required,min=1"`
	AttackSuccess      bool     `json:"attack_success"`
	Severity           string   `json:"severity,omitempty" validate:"omitempty,oneof=critical high medium low"`
	OWASPCategory      string   `json:"owasp_category,omitempty"`
	DefenseIntercepted *bool    `json:"defense_intercepted,omitempty"`
	JudgeConfidence    *float64 `json:"judge_confidence,omitempty" validate:"omitempty,gte=0,lte=1"`
	JudgeReasoning     string   `json:"judge_reasoning,omitempty"`
	LatencyMs          *int     `json:"latency_ms,omitempty" validate:"omitempty,gte=0"`
	TokensUsed         *int     `json:"tokens_used,omitempty" validate:"omitempty,gte=0"`
}

type CreateAttackResultsRequest struct {
	Results []AttackResultInput `json:"results" validate:"required,min=1,dive"`
}

type AttackResultListResponse struct {
	Results []*AttackResult `json:"results"`
	Total   int             `json:"total"`
	Offset  int             `json:"offset"`
	Limit   int             `json:"limit"`
}
