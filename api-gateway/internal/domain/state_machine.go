package domain

// validTransitions encodes every permitted (from → to) status pair.
// States absent from this map (completed, failed, stopped) are terminal
// and have no outbound edges; a map lookup on a missing key returns nil,
// which evaluates to false in the inner bool lookup.
var validTransitions = map[ScanStatus]map[ScanStatus]bool{
	StatusPending: {StatusQueued: true, StatusRunning: true},
	StatusQueued:  {StatusRunning: true, StatusStopped: true},
	StatusRunning: {StatusCompleted: true, StatusFailed: true, StatusStopped: true},
}

// CanTransition reports whether a scan may move from status `from` to
// status `to`. It is the single source of truth for all lifecycle rules.
func CanTransition(from, to ScanStatus) bool {
	return validTransitions[from][to]
}
