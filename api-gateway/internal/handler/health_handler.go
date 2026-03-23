package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HealthHandler struct {
	startTime time.Time
	version   string
}

func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{startTime: time.Now(), version: version}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": h.version,
		"uptime":  time.Since(h.startTime).String(),
	})
}

func (h *HealthHandler) Metrics() gin.HandlerFunc {
	promHandler := promhttp.Handler()
	return func(c *gin.Context) {
		promHandler.ServeHTTP(c.Writer, c.Request)
	}
}
