package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentshield/api-gateway/internal/handler"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newJudgeRouter(h *handler.JudgeHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := "00000000-0000-0000-0000-000000000001"

	r.POST("/judge/calibrate", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.Calibrate(c)
	})

	r.GET("/judge/calibration-report", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Set(middleware.RequestIDKey, uuid.New().String())
		h.CalibrationReport(c)
	})

	return r
}

func TestJudgeCalibrate_Returns200_AndStoresLatestReport(t *testing.T) {
	h := handler.NewJudgeHandler()
	r := newJudgeRouter(h)

	body := []byte(`{
		"samples": [
			{
				"attack_type": "prompt_injection",
				"expected_success": true,
				"predicted_success": true,
				"expected_severity": "critical",
				"predicted_severity": "critical",
				"human_confidence": 0.90,
				"judge_confidence": 0.85
			},
			{
				"attack_type": "prompt_injection",
				"expected_success": true,
				"predicted_success": false,
				"expected_severity": "high",
				"predicted_severity": "medium",
				"human_confidence": 0.75,
				"judge_confidence": 0.50
			},
			{
				"attack_type": "jailbreak",
				"expected_success": false,
				"predicted_success": false,
				"expected_severity": "low",
				"predicted_severity": "low",
				"human_confidence": 0.20,
				"judge_confidence": 0.30
			}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/judge/calibrate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)

	reportReq := httptest.NewRequest(http.MethodGet, "/judge/calibration-report", nil)
	reportW := httptest.NewRecorder()
	r.ServeHTTP(reportW, reportReq)

	assertStatus(t, reportW, http.StatusOK)

	var payload map[string]any
	if err := json.Unmarshal(reportW.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal calibration report: %v", err)
	}
	if count, _ := payload["sample_count"].(float64); int(count) != 3 {
		t.Fatalf("sample_count = %v, want 3", payload["sample_count"])
	}
	overall, ok := payload["overall"].(map[string]any)
	if !ok {
		t.Fatalf("overall missing: %s", reportW.Body.String())
	}
	if _, ok := overall["f1"]; !ok {
		t.Fatalf("overall f1 missing: %s", reportW.Body.String())
	}
}

func TestJudgeCalibrationReport_Returns404_WhenNoReport(t *testing.T) {
	h := handler.NewJudgeHandler()
	r := newJudgeRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/judge/calibration-report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusNotFound)
	assertErrorCode(t, w, "NOT_FOUND")
}

func TestJudgeCalibrate_Returns400_WhenSamplesEmpty(t *testing.T) {
	h := handler.NewJudgeHandler()
	r := newJudgeRouter(h)

	body := []byte(`{"samples":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/judge/calibrate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_REQUEST")
}

func TestJudgeCalibrate_Returns400_WhenAttackTypeInvalid(t *testing.T) {
	h := handler.NewJudgeHandler()
	r := newJudgeRouter(h)

	body := []byte(`{
		"samples":[
			{
				"attack_type":"unknown",
				"expected_success":true,
				"predicted_success":true
			}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/judge/calibrate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusBadRequest)
	assertErrorCode(t, w, "INVALID_REQUEST")
}
