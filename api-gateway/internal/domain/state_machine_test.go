package domain_test

import (
	"testing"

	"github.com/agentshield/api-gateway/internal/domain"
)

// TestCanTransition verifies the complete scan status state machine.
//
// Each row in the table is an explicit (from, to, want) triple.
// The full set of cases is divided into three groups:
//
//  1. Valid transitions — the system must permit these moves.
//  2. Terminal-state exits — completed / failed / stopped have no
//     outbound edges; every attempted exit must be rejected.
//  3. Invalid moves — backwards, nonsensical, or skipped transitions
//     that the system must reject even though neither state is terminal.
func TestCanTransition(t *testing.T) {
	cases := []struct {
		name string
		from domain.ScanStatus
		to   domain.ScanStatus
		want bool
	}{
		// ── Valid transitions (want: true) ───────────────────────────────────
		// A freshly-created scan can be queued (orchestrator unavailable path).
		{
			name: "pending_to_queued",
			from: domain.StatusPending,
			to:   domain.StatusQueued,
			want: true,
		},
		// A freshly-created scan can go directly to running (orchestrator
		// accepts immediately).
		{
			name: "pending_to_running",
			from: domain.StatusPending,
			to:   domain.StatusRunning,
			want: true,
		},
		// A queued scan is picked up by the orchestrator and started.
		{
			name: "queued_to_running",
			from: domain.StatusQueued,
			to:   domain.StatusRunning,
			want: true,
		},
		// A queued scan can be cancelled before the orchestrator picks it up.
		{
			name: "queued_to_stopped",
			from: domain.StatusQueued,
			to:   domain.StatusStopped,
			want: true,
		},
		// An active scan finishes successfully.
		{
			name: "running_to_completed",
			from: domain.StatusRunning,
			to:   domain.StatusCompleted,
			want: true,
		},
		// An active scan is terminated by an agent or infra error.
		{
			name: "running_to_failed",
			from: domain.StatusRunning,
			to:   domain.StatusFailed,
			want: true,
		},
		// An active scan is explicitly cancelled by the user.
		{
			name: "running_to_stopped",
			from: domain.StatusRunning,
			to:   domain.StatusStopped,
			want: true,
		},

		// ── Terminal-state exits (want: false) ───────────────────────────────
		// Once completed, a scan cannot be restarted or mutated.
		{
			name: "completed_to_running",
			from: domain.StatusCompleted,
			to:   domain.StatusRunning,
			want: false,
		},
		{
			name: "completed_to_pending",
			from: domain.StatusCompleted,
			to:   domain.StatusPending,
			want: false,
		},
		{
			name: "completed_to_failed",
			from: domain.StatusCompleted,
			to:   domain.StatusFailed,
			want: false,
		},
		// Once failed, a scan cannot be retried or restarted.
		{
			name: "failed_to_running",
			from: domain.StatusFailed,
			to:   domain.StatusRunning,
			want: false,
		},
		{
			name: "failed_to_pending",
			from: domain.StatusFailed,
			to:   domain.StatusPending,
			want: false,
		},
		{
			name: "failed_to_stopped",
			from: domain.StatusFailed,
			to:   domain.StatusStopped,
			want: false,
		},
		// Once stopped, a scan cannot be resumed.
		{
			name: "stopped_to_running",
			from: domain.StatusStopped,
			to:   domain.StatusRunning,
			want: false,
		},
		{
			name: "stopped_to_pending",
			from: domain.StatusStopped,
			to:   domain.StatusPending,
			want: false,
		},
		{
			name: "stopped_to_queued",
			from: domain.StatusStopped,
			to:   domain.StatusQueued,
			want: false,
		},

		// ── Invalid moves — neither state is terminal (want: false) ──────────
		// A running scan cannot go backwards; it must complete, fail, or stop.
		{
			name: "running_to_pending",
			from: domain.StatusRunning,
			to:   domain.StatusPending,
			want: false,
		},
		{
			name: "running_to_queued",
			from: domain.StatusRunning,
			to:   domain.StatusQueued,
			want: false,
		},
		// A queued scan cannot skip to completed or failed without running.
		{
			name: "queued_to_completed",
			from: domain.StatusQueued,
			to:   domain.StatusCompleted,
			want: false,
		},
		{
			name: "queued_to_failed",
			from: domain.StatusQueued,
			to:   domain.StatusFailed,
			want: false,
		},
		// A pending scan cannot skip directly to terminal states.
		{
			name: "pending_to_completed",
			from: domain.StatusPending,
			to:   domain.StatusCompleted,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.CanTransition(tc.from, tc.to)
			if got != tc.want {
				t.Errorf(
					"CanTransition(%q, %q) = %v, want %v",
					tc.from, tc.to, got, tc.want,
				)
			}
		})
	}
}
