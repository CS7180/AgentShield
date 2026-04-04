package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScanStatus string
type ScanMode string

const (
	StatusPending   ScanStatus = "pending"
	StatusQueued    ScanStatus = "queued"
	StatusRunning   ScanStatus = "running"
	StatusCompleted ScanStatus = "completed"
	StatusFailed    ScanStatus = "failed"
	StatusStopped   ScanStatus = "stopped"
)

const (
	ModeRedTeam     ScanMode = "red_team"
	ModeBlueTeam    ScanMode = "blue_team"
	ModeAdversarial ScanMode = "adversarial"
)

type Scan struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	TargetEndpoint string     `json:"target_endpoint"`
	Mode           ScanMode   `json:"mode"`
	AttackTypes    []string   `json:"attack_types"`
	Status         ScanStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateScanRequest struct {
	TargetEndpoint string   `json:"target_endpoint" validate:"required,https_endpoint"`
	Mode           ScanMode `json:"mode" validate:"required,oneof=red_team blue_team adversarial"`
	AttackTypes    []string `json:"attack_types" validate:"required,min=1,dive,oneof=prompt_injection jailbreak data_leakage constraint_drift"`
}

type ScanListResponse struct {
	Scans  []*Scan `json:"scans"`
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
}
