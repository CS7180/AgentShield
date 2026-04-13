package orchestrator

import (
	"context"
	"sync"
	"testing"
	"time"

	pb "github.com/agentshield/api-gateway/proto/orchestrator"
	"go.uber.org/zap"
)

type fakeExecutor struct {
	mu           sync.Mutex
	executed     bool
	stoppedScan  string
	execCalledCh chan struct{}
}

func (f *fakeExecutor) Execute(_ context.Context, _ ScanExecutionRequest, progress func(int)) error {
	f.mu.Lock()
	f.executed = true
	f.mu.Unlock()
	progress(50)
	progress(100)
	select {
	case f.execCalledCh <- struct{}{}:
	default:
	}
	return nil
}

func (f *fakeExecutor) MarkStopped(_ context.Context, scanID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stoppedScan = scanID
	return nil
}

func TestServerStartScan_ExecutesPipeline(t *testing.T) {
	exec := &fakeExecutor{execCalledCh: make(chan struct{}, 1)}
	s := NewServer(exec, zap.NewNop())

	resp, err := s.StartScan(context.Background(), &pb.StartScanRequest{
		ScanId:         "11111111-1111-1111-1111-111111111111",
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

	select {
	case <-exec.execCalledCh:
	case <-time.After(1 * time.Second):
		t.Fatal("executor was not called")
	}

	status, _ := s.manager.GetStatus("11111111-1111-1111-1111-111111111111")
	if status != ScanStatusCompleted {
		t.Fatalf("status = %s, want %s", status, ScanStatusCompleted)
	}
}

func TestServerStopScan_CallsMarkStopped(t *testing.T) {
	exec := &fakeExecutor{execCalledCh: make(chan struct{}, 1)}
	s := NewServer(exec, zap.NewNop())

	_, _ = s.StartScan(context.Background(), &pb.StartScanRequest{
		ScanId:         "22222222-2222-2222-2222-222222222222",
		TargetEndpoint: "https://example.com",
		Mode:           "red_team",
		AttackTypes:    []string{"prompt_injection"},
	})

	resp, err := s.StopScan(context.Background(), &pb.StopScanRequest{ScanId: "22222222-2222-2222-2222-222222222222"})
	if err != nil {
		t.Fatalf("stop scan error: %v", err)
	}
	if !resp.Stopped {
		t.Fatalf("expected stopped=true, got %+v", resp)
	}

	exec.mu.Lock()
	stopped := exec.stoppedScan
	exec.mu.Unlock()
	if stopped != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("stopped scan = %s", stopped)
	}
}
