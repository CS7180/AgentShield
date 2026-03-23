package handler

import (
	"net/http"

	"github.com/agentshield/api-gateway/internal/auth"
	"github.com/agentshield/api-gateway/internal/metrics"
	"github.com/agentshield/api-gateway/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Origin check is handled by the CORS middleware upstream
		return true
	},
}

type WSHandler struct {
	hub       *ws.Hub
	jwtSecret []byte
	logger    *zap.Logger
}

func NewWSHandler(hub *ws.Hub, jwtSecret []byte, logger *zap.Logger) *WSHandler {
	return &WSHandler{hub: hub, jwtSecret: jwtSecret, logger: logger}
}

// HandleScanStatus upgrades to WebSocket and subscribes the client to a scan's events.
// Auth is done via ?token=<supabase_jwt> because browsers cannot set Authorization headers on WS.
func (h *WSHandler) HandleScanStatus(c *gin.Context) {
	scanID := c.Param("id")
	if scanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing scan_id"})
		return
	}

	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := auth.ParseSupabaseToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", zap.Error(err))
		return
	}

	metrics.WebSocketConnectionsActive.Inc()

	client := ws.NewClient(h.hub, conn, scanID, claims.Subject, h.logger)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
