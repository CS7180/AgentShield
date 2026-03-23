package ws_test

import (
	"testing"
	"time"

	"github.com/agentshield/api-gateway/internal/ws"
	"go.uber.org/zap"
)

func TestHub_BroadcastToRoom(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := ws.NewHub(logger)
	go hub.Run()

	// Give the hub goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Broadcast to a non-existent room should not panic
	hub.Broadcast("scan-123", []byte(`{"test":"data"}`))

	// Small delay to let the hub process
	time.Sleep(10 * time.Millisecond)
}

func TestHub_BroadcastDropsWhenFull(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := ws.NewHub(logger)
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	// Flood the broadcast channel — should not block
	done := make(chan struct{})
	go func() {
		for i := 0; i < 300; i++ {
			hub.Broadcast("scan-flood", []byte(`{"i":1}`))
		}
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast blocked — should have dropped messages instead")
	}
}
