package kafka

import (
	"encoding/json"

	"github.com/agentshield/api-gateway/internal/ws"
	"go.uber.org/zap"
)

// ScanEvent is the minimal structure we need from Kafka messages to route to the right room.
type ScanEvent struct {
	ScanID string          `json:"scan_id"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// Dispatcher routes Kafka messages to the WebSocket hub.
type Dispatcher struct {
	hub    *ws.Hub
	logger *zap.Logger
}

func NewDispatcher(hub *ws.Hub, logger *zap.Logger) *Dispatcher {
	return &Dispatcher{hub: hub, logger: logger}
}

// Dispatch parses a raw Kafka message and broadcasts it to the matching scan room.
func (d *Dispatcher) Dispatch(topic string, payload []byte) {
	var event ScanEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		d.logger.Warn("kafka dispatcher: failed to parse message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return
	}

	if event.ScanID == "" {
		d.logger.Warn("kafka dispatcher: missing scan_id", zap.String("topic", topic))
		return
	}

	// Re-marshal as a WebSocket event envelope
	envelope := map[string]interface{}{
		"topic":   topic,
		"scan_id": event.ScanID,
		"data":    event.Data,
	}
	msg, err := json.Marshal(envelope)
	if err != nil {
		d.logger.Error("kafka dispatcher: marshal envelope", zap.Error(err))
		return
	}

	d.hub.Broadcast(event.ScanID, msg)
}
