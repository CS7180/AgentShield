package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentshield/api-gateway/internal/handler"
	"github.com/gin-gonic/gin"
)

func healthRouter(version string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handler.NewHealthHandler(version)
	r := gin.New()
	r.GET("/health", h.Health)
	r.GET("/metrics", h.Metrics())
	return r
}

func TestHealth_Returns200_WithStatusOK(t *testing.T) {
	r := healthRouter("v1.2.3")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
	if body["version"] != "v1.2.3" {
		t.Errorf("version = %q, want v1.2.3", body["version"])
	}
	if _, ok := body["uptime"]; !ok {
		t.Error("uptime field missing from response")
	}
}

func TestMetrics_Returns200(t *testing.T) {
	r := healthRouter("v0.0.1")
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
