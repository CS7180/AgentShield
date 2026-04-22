package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentshield/api-gateway/internal/handler"
	"github.com/agentshield/api-gateway/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// neverKeyFunc is a jwt.Keyfunc that always returns an error.
// Used for tests that should never reach token validation.
func neverKeyFunc(_ *jwt.Token) (interface{}, error) {
	return nil, fmt.Errorf("keyFunc should not be called in this test")
}

// rejectKeyFunc rejects every token regardless of content.
func rejectKeyFunc(_ *jwt.Token) (interface{}, error) {
	return nil, fmt.Errorf("no valid key")
}

func newWSHandlerRouter(h *handler.WSHandler, withScanIDParam bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if withScanIDParam {
		r.GET("/ws/scans/:id/status", h.HandleScanStatus)
	} else {
		// Route without :id — scanID will be empty string
		r.GET("/ws/status", h.HandleScanStatus)
	}
	return r
}

func newWSHub() *ws.Hub {
	hub := ws.NewHub(zap.NewNop())
	go hub.Run()
	return hub
}

func TestWSHandler_MissingScanID_Returns400(t *testing.T) {
	h := handler.NewWSHandler(newWSHub(), neverKeyFunc, zap.NewNop())
	r := newWSHandlerRouter(h, false) // no :id param — scanID will be ""

	req := httptest.NewRequest(http.MethodGet, "/ws/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestWSHandler_MissingToken_Returns401(t *testing.T) {
	h := handler.NewWSHandler(newWSHub(), neverKeyFunc, zap.NewNop())
	r := newWSHandlerRouter(h, true)

	req := httptest.NewRequest(http.MethodGet, "/ws/scans/scan-abc/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestWSHandler_InvalidToken_Returns401(t *testing.T) {
	h := handler.NewWSHandler(newWSHub(), rejectKeyFunc, zap.NewNop())
	r := newWSHandlerRouter(h, true)

	req := httptest.NewRequest(http.MethodGet, "/ws/scans/scan-abc/status?token=not.a.jwt", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d (body: %s)", w.Code, w.Body.String())
	}
}
