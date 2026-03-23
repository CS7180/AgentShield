package middleware

import (
	"strconv"
	"time"

	"github.com/agentshield/api-gateway/internal/metrics"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method
		statusStr := strconv.Itoa(status)

		metrics.HTTPRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()

		logger.Info("http request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("request_id", c.GetString(RequestIDKey)),
			zap.String("ip", c.ClientIP()),
		)
	}
}
