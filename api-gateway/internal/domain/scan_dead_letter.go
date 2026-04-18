package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ScanDeadLetter stores the terminal failure metadata after retries are exhausted.
type ScanDeadLetter struct {
	ID           uuid.UUID       `json:"id"`
	ScanID       uuid.UUID       `json:"scan_id"`
	UserID       uuid.UUID       `json:"user_id"`
	AttemptCount int             `json:"attempt_count"`
	ErrorStage   string          `json:"error_stage"`
	ErrorMessage string          `json:"error_message"`
	Payload      json.RawMessage `json:"payload"`
	FailedAt     time.Time       `json:"failed_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
