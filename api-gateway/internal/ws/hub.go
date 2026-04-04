package ws

import (
	"sync"

	"go.uber.org/zap"
)

// Message is a JSON payload broadcast to a specific scan room.
type Message struct {
	ScanID  string
	Payload []byte
}

// Hub manages room-based WebSocket connections.
// rooms maps scan_id → set of clients.
// Hub.Run() is the sole goroutine that writes to the rooms map.
type Hub struct {
	rooms      map[string]map[*Client]struct{}
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *zap.Logger
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]struct{}),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		logger:     logger,
	}
}

// Run processes hub events. Must be called in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.rooms[client.scanID]; !ok {
				h.rooms[client.scanID] = make(map[*Client]struct{})
			}
			h.rooms[client.scanID][client] = struct{}{}
			h.logger.Debug("ws client registered",
				zap.String("scan_id", client.scanID),
				zap.Int("room_size", len(h.rooms[client.scanID])),
			)

		case client := <-h.unregister:
			if room, ok := h.rooms[client.scanID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.send)
				}
				if len(room) == 0 {
					delete(h.rooms, client.scanID)
				}
			}
			h.logger.Debug("ws client unregistered", zap.String("scan_id", client.scanID))

		case msg := <-h.broadcast:
			room, ok := h.rooms[msg.ScanID]
			if !ok {
				continue
			}
			for client := range room {
				select {
				case client.send <- msg.Payload:
				default:
					// Client's send buffer full — drop it
					delete(room, client)
					close(client.send)
					h.logger.Warn("ws client dropped: send buffer full",
						zap.String("scan_id", msg.ScanID),
					)
				}
			}
		}
	}
}

// Broadcast sends a payload to all clients subscribed to the given scan.
// Safe to call from any goroutine. Drops the message if the broadcast channel is full.
func (h *Hub) Broadcast(scanID string, payload []byte) {
	msg := Message{ScanID: scanID, Payload: payload}
	select {
	case h.broadcast <- msg:
	default:
		h.logger.Warn("ws broadcast channel full, dropping message",
			zap.String("scan_id", scanID),
		)
	}
}

func (h *Hub) Register(c *Client)   { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }
