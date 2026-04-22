package ws

// NewTestClient creates a Client for hub unit tests.
// conn is nil — only hub routing logic is exercised (no WebSocket I/O).
func NewTestClient(hub *Hub, scanID string) *Client {
	return &Client{
		hub:    hub,
		send:   make(chan []byte, 8),
		scanID: scanID,
	}
}

// ClientSendChan exposes the send channel of a Client so test assertions
// can inspect whether a broadcast message was delivered.
func ClientSendChan(c *Client) <-chan []byte { return c.send }
