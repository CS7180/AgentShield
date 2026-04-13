package orchestrator

import (
	"context"
	"fmt"
	"sync"
)

const (
	ScanStatusRunning   = "running"
	ScanStatusCompleted = "completed"
	ScanStatusFailed    = "failed"
	ScanStatusStopped   = "stopped"
	ScanStatusNotFound  = "not_found"
)

type scanState struct {
	status   string
	progress int
	cancel   context.CancelFunc
	ctx      context.Context
}

type Manager struct {
	mu    sync.RWMutex
	scans map[string]*scanState
}

func NewManager() *Manager {
	return &Manager{
		scans: make(map[string]*scanState),
	}
}

func (m *Manager) StartScan(scanID string) (context.Context, bool, string) {
	m.mu.Lock()
	if existing, ok := m.scans[scanID]; ok && existing.status == ScanStatusRunning {
		m.mu.Unlock()
		return nil, false, "scan already running"
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.scans[scanID] = &scanState{
		status:   ScanStatusRunning,
		progress: 0,
		cancel:   cancel,
		ctx:      ctx,
	}
	m.mu.Unlock()

	return ctx, true, "scan accepted"
}

func (m *Manager) StopScan(scanID string) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.scans[scanID]
	if !ok {
		return false, "scan not found"
	}
	if state.status == ScanStatusCompleted {
		return false, "scan already completed"
	}
	if state.status == ScanStatusFailed {
		return false, "scan already failed"
	}
	if state.status == ScanStatusStopped {
		return false, "scan already stopped"
	}

	state.status = ScanStatusStopped
	if state.cancel != nil {
		state.cancel()
	}
	return true, "scan stopped"
}

func (m *Manager) SetProgress(scanID string, progress int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.scans[scanID]
	if !ok {
		return fmt.Errorf("scan not found")
	}
	if state.status != ScanStatusRunning {
		return nil
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	state.progress = progress
	return nil
}

func (m *Manager) MarkCompleted(scanID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	state, ok := m.scans[scanID]
	if !ok {
		return fmt.Errorf("scan not found")
	}
	if state.status == ScanStatusStopped {
		return nil
	}
	state.status = ScanStatusCompleted
	state.progress = 100
	return nil
}

func (m *Manager) MarkFailed(scanID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	state, ok := m.scans[scanID]
	if !ok {
		return fmt.Errorf("scan not found")
	}
	if state.status == ScanStatusStopped {
		return nil
	}
	state.status = ScanStatusFailed
	return nil
}

func (m *Manager) GetStatus(scanID string) (string, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.scans[scanID]
	if !ok {
		return ScanStatusNotFound, 0
	}
	return state.status, state.progress
}
