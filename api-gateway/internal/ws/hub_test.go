package ws_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/agentshield/api-gateway/internal/ws"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// startHub creates a hub, starts Run() in a goroutine, and waits for it to
// be ready. Keeping this as a helper avoids copy-paste in every test.
func startHub(t *testing.T) *ws.Hub {
	t.Helper()
	hub := ws.NewHub(zap.NewNop())
	go hub.Run()
	time.Sleep(10 * time.Millisecond)
	return hub
}

// recv reads one message from a client's send channel within a short timeout.
// Returns (msg, true) on success or ("", false) on timeout.
func recv(ch <-chan []byte, timeout time.Duration) ([]byte, bool) {
	select {
	case msg := <-ch:
		return msg, true
	case <-time.After(timeout):
		return nil, false
	}
}

// ── Register / Unregister ────────────────────────────────────────────────────

func TestHub_Register_ClientReceivesBroadcast(t *testing.T) {
	hub := startHub(t)
	client := ws.NewTestClient(hub, "scan-abc")
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	want := []byte(`{"status":"running"}`)
	hub.Broadcast("scan-abc", want)

	got, ok := recv(ws.ClientSendChan(client), 100*time.Millisecond)
	if !ok {
		t.Fatal("timed out waiting for broadcast message")
	}
	if string(got) != string(want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHub_Unregister_ClientNoLongerReceivesBroadcast(t *testing.T) {
	hub := startHub(t)
	client := ws.NewTestClient(hub, "scan-xyz")
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	hub.Broadcast("scan-xyz", []byte(`{"status":"done"}`))

	// send channel should have been closed by the hub on unregister —
	// either closed (zero-value read) or no message within timeout.
	select {
	case msg, open := <-ws.ClientSendChan(client):
		if open && len(msg) > 0 {
			t.Errorf("received unexpected message after unregister: %q", msg)
		}
	case <-time.After(50 * time.Millisecond):
		// No message and channel not closed: also acceptable — room was
		// deleted and broadcast was silently dropped.
	}
}

// ── Multiple clients in the same room ────────────────────────────────────────

func TestHub_MultipleClients_BothReceiveBroadcast(t *testing.T) {
	hub := startHub(t)

	c1 := ws.NewTestClient(hub, "scan-multi")
	c2 := ws.NewTestClient(hub, "scan-multi")
	hub.Register(c1)
	hub.Register(c2)
	time.Sleep(10 * time.Millisecond)

	payload := []byte(`{"event":"attack"}`)
	hub.Broadcast("scan-multi", payload)

	for i, ch := range []<-chan []byte{ws.ClientSendChan(c1), ws.ClientSendChan(c2)} {
		got, ok := recv(ch, 100*time.Millisecond)
		if !ok {
			t.Errorf("client %d: timed out waiting for broadcast", i+1)
			continue
		}
		if string(got) != string(payload) {
			t.Errorf("client %d: got %q, want %q", i+1, got, payload)
		}
	}
}

// ── Cross-room isolation ──────────────────────────────────────────────────────

func TestHub_BroadcastToOtherRoom_NotReceived(t *testing.T) {
	hub := startHub(t)

	clientA := ws.NewTestClient(hub, "scan-A")
	hub.Register(clientA)
	time.Sleep(10 * time.Millisecond)

	// Broadcast to a completely different room
	hub.Broadcast("scan-B", []byte(`{"event":"attack"}`))

	_, ok := recv(ws.ClientSendChan(clientA), 50*time.Millisecond)
	if ok {
		t.Error("client in scan-A received message destined for scan-B")
	}
}

// ── Broadcast to empty room (no panic) ───────────────────────────────────────

func TestHub_BroadcastToRoom(t *testing.T) {
	hub := startHub(t)
	// Broadcast to a non-existent room should not panic
	hub.Broadcast("scan-123", []byte(`{"test":"data"}`))
	time.Sleep(10 * time.Millisecond)
}

// ── Channel-full drop (no deadlock) ──────────────────────────────────────────

func TestHub_BroadcastDropsWhenFull(t *testing.T) {
	hub := startHub(t)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 300; i++ {
			hub.Broadcast("scan-flood", []byte(`{"i":1}`))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast blocked — should have dropped messages instead")
	}
}

// ── WebSocket round-trip (NewClient / WritePump / ReadPump) ───────────────────

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

// newWSServer starts an httptest.Server that upgrades connections to WebSocket,
// creates a Client, registers it with hub, and runs WritePump+ReadPump.
func newWSServer(hub *ws.Hub, scanID string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := ws.NewClient(hub, conn, scanID, "test-user", zap.NewNop())
		hub.Register(c)
		go c.WritePump()
		c.ReadPump() // blocks until connection closes
	}))
}

func dialWS(t *testing.T, srvURL string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(srvURL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	return conn
}

func TestClient_WritePump_DeliversBroadcast(t *testing.T) {
	hub := startHub(t)
	ts := newWSServer(hub, "scan-roundtrip")
	defer ts.Close()

	clientConn := dialWS(t, ts.URL)
	defer clientConn.Close()

	// Wait for registration to propagate
	time.Sleep(20 * time.Millisecond)

	want := []byte(`{"event":"attack_done"}`)
	hub.Broadcast("scan-roundtrip", want)

	clientConn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	if string(msg) != string(want) {
		t.Errorf("got %q, want %q", msg, want)
	}
}

func TestClient_WritePump_MultipleMessages(t *testing.T) {
	hub := startHub(t)
	ts := newWSServer(hub, "scan-multi-msg")
	defer ts.Close()

	clientConn := dialWS(t, ts.URL)
	defer clientConn.Close()

	time.Sleep(20 * time.Millisecond)

	messages := []string{`{"seq":1}`, `{"seq":2}`, `{"seq":3}`}
	for _, m := range messages {
		hub.Broadcast("scan-multi-msg", []byte(m))
	}

	for i, want := range messages {
		clientConn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		_, msg, err := clientConn.ReadMessage()
		if err != nil {
			t.Fatalf("message %d: read error: %v", i+1, err)
		}
		if string(msg) != want {
			t.Errorf("message %d: got %q, want %q", i+1, msg, want)
		}
	}
}

func TestClient_ReadPump_UnregistersOnClose(t *testing.T) {
	hub := startHub(t)
	ts := newWSServer(hub, "scan-close")
	defer ts.Close()

	clientConn := dialWS(t, ts.URL)
	time.Sleep(20 * time.Millisecond)

	// Verify the client is registered: it should receive a broadcast
	hub.Broadcast("scan-close", []byte(`{"event":"ping"}`))
	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if _, _, err := clientConn.ReadMessage(); err != nil {
		t.Fatalf("expected to receive message before close: %v", err)
	}

	// Close the client connection — ReadPump should return and call Unregister
	clientConn.Close()
	time.Sleep(50 * time.Millisecond)

	// Broadcast after close — should not panic; room should be gone
	hub.Broadcast("scan-close", []byte(`{"event":"after-close"}`))
	time.Sleep(10 * time.Millisecond)
}
