package orchestrator

import "context"

// Stub is an in-process no-op orchestrator used when ORCHESTRATOR_ENABLED=false.
// StartScan always returns accepted=false so the scan is queued in the DB.
type Stub struct{}

func NewStub() *Stub { return &Stub{} }

func (s *Stub) StartScan(_ context.Context, scanID, _, _ string, _ []string) (bool, string, error) {
	return false, "orchestrator disabled; scan queued for processing", nil
}

func (s *Stub) StopScan(_ context.Context, scanID string) (bool, string, error) {
	return true, "scan marked stopped (stub)", nil
}
