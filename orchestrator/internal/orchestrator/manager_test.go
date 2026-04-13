package orchestrator

import (
	"testing"
)

func TestManagerStartAndComplete(t *testing.T) {
	m := NewManager()

	_, accepted, _ := m.StartScan("scan-1")
	if !accepted {
		t.Fatal("expected scan to be accepted")
	}

	if err := m.SetProgress("scan-1", 65); err != nil {
		t.Fatalf("set progress: %v", err)
	}
	if err := m.MarkCompleted("scan-1"); err != nil {
		t.Fatalf("mark completed: %v", err)
	}

	status, progress := m.GetStatus("scan-1")
	if status != ScanStatusCompleted || progress != 100 {
		t.Fatalf("status/progress = %s/%d, want completed/100", status, progress)
	}
}

func TestManagerStop(t *testing.T) {
	m := NewManager()
	_, accepted, _ := m.StartScan("scan-2")
	if !accepted {
		t.Fatal("expected scan to be accepted")
	}

	stopped, _ := m.StopScan("scan-2")
	if !stopped {
		t.Fatal("expected scan to stop")
	}

	status, _ := m.GetStatus("scan-2")
	if status != ScanStatusStopped {
		t.Fatalf("status = %s, want %s", status, ScanStatusStopped)
	}
}

func TestManagerDuplicateRunningStart(t *testing.T) {
	m := NewManager()
	_, accepted, _ := m.StartScan("scan-3")
	if !accepted {
		t.Fatal("expected first start accepted")
	}

	_, accepted, _ = m.StartScan("scan-3")
	if accepted {
		t.Fatal("expected second start rejected while running")
	}
}

func TestManagerMarkFailed(t *testing.T) {
	m := NewManager()
	_, accepted, _ := m.StartScan("scan-4")
	if !accepted {
		t.Fatal("expected first start accepted")
	}
	if err := m.MarkFailed("scan-4"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	status, _ := m.GetStatus("scan-4")
	if status != ScanStatusFailed {
		t.Fatalf("status = %s, want failed", status)
	}
}
