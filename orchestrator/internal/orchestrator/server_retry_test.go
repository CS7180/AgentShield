package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	pb "github.com/agentshield/api-gateway/proto/orchestrator"
	"go.uber.org/zap"
)

type flakyExecutor struct {
	mu           sync.Mutex
	failTimes    int
	callCount    int
	deadLettered int
	lastAttempts int
}

func (f *flakyExecutor) Execute(_ context.Context, _ ScanExecutionRequest, progress func(int)) error {
	f.mu.Lock()
	f.callCount++
	attempt := f.callCount
	shouldFail := attempt <= f.failTimes
	f.mu.Unlock()

	if shouldFail {
		return fmt.Errorf("run agents: transient failure")
	}

	progress(100)
	return nil
}

func (f *flakyExecutor) MarkStopped(_ context.Context, _ string) error {
	return nil
}

func (f *flakyExecutor) WriteDeadLetter(_ context.Context, _ ScanExecutionRequest, attemptCount int, _ error) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deadLettered++
	f.lastAttempts = attemptCount
	return nil
}

func TestServerRunPipeline_RetryThenSuccess(t *testing.T) {
	t.Setenv("ORCHESTRATOR_EXEC_MAX_ATTEMPTS", "3")
	t.Setenv("ORCHESTRATOR_EXEC_RETRY_BASE_MS", "1")
	t.Setenv("ORCHESTRATOR_EXEC_RETRY_MAX_MS", "2")

	exec := &flakyExecutor{failTimes: 1}
	s := NewServer(exec, zap.NewNop())

	scanID := "33333333-3333-3333-3333-333333333333"
	resp, err := s.StartScan(context.Background(), &pb.StartScanRequest{
		ScanId:         scanID,
		TargetEndpoint: "https://example.com",
		Mode:           "red_team",
		AttackTypes:    []string{"prompt_injection"},
	})
	if err != nil {
		t.Fatalf("start scan error: %v", err)
	}
	if !resp.Accepted {
		t.Fatalf("expected accepted, got %+v", resp)
	}

	waitStatus(t, s, scanID, ScanStatusCompleted, 2*time.Second)

	exec.mu.Lock()
	callCount := exec.callCount
	deadLettered := exec.deadLettered
	exec.mu.Unlock()

	if callCount != 2 {
		t.Fatalf("execute call count = %d, want 2", callCount)
	}
	if deadLettered != 0 {
		t.Fatalf("dead letters = %d, want 0", deadLettered)
	}
}

func TestServerRunPipeline_DeadLetterOnExhaustedRetries(t *testing.T) {
	t.Setenv("ORCHESTRATOR_EXEC_MAX_ATTEMPTS", "2")
	t.Setenv("ORCHESTRATOR_EXEC_RETRY_BASE_MS", "1")
	t.Setenv("ORCHESTRATOR_EXEC_RETRY_MAX_MS", "2")

	exec := &flakyExecutor{failTimes: 10}
	s := NewServer(exec, zap.NewNop())

	scanID := "44444444-4444-4444-4444-444444444444"
	resp, err := s.StartScan(context.Background(), &pb.StartScanRequest{
		ScanId:         scanID,
		TargetEndpoint: "https://example.com",
		Mode:           "red_team",
		AttackTypes:    []string{"prompt_injection"},
	})
	if err != nil {
		t.Fatalf("start scan error: %v", err)
	}
	if !resp.Accepted {
		t.Fatalf("expected accepted, got %+v", resp)
	}

	waitStatus(t, s, scanID, ScanStatusFailed, 2*time.Second)

	exec.mu.Lock()
	callCount := exec.callCount
	deadLettered := exec.deadLettered
	lastAttempts := exec.lastAttempts
	exec.mu.Unlock()

	if callCount != 2 {
		t.Fatalf("execute call count = %d, want 2", callCount)
	}
	if deadLettered != 1 {
		t.Fatalf("dead letters = %d, want 1", deadLettered)
	}
	if lastAttempts != 2 {
		t.Fatalf("dead letter attempt count = %d, want 2", lastAttempts)
	}
}

func waitStatus(t *testing.T, s *Server, scanID, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, _ := s.manager.GetStatus(scanID)
		if status == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	status, _ := s.manager.GetStatus(scanID)
	t.Fatalf("status = %s, want %s", status, want)
}
