package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/agentshield/api-gateway/internal/metrics"
	"github.com/gin-gonic/gin"
)

type rateLimiter interface {
	AllowRequest(ctx context.Context, key string, maxTokens int, refillPerSec float64) (bool, error)
}

// GlobalRateLimit applies 100 req/min per user (token bucket: 100 tokens, refill ~1.67/sec).
func GlobalRateLimit(limiter rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString(UserIDKey)
		if userID == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:global:%s", userID)
		allowed, err := limiter.AllowRequest(c.Request.Context(), key, 100, 100.0/60.0)
		if err != nil || !allowed {
			if err == nil {
				metrics.RateLimitExceededTotal.WithLabelValues("global").Inc()
			}
			abortRateLimited(c, "rate limit exceeded")
			return
		}
		c.Next()
	}
}

// ScanCreateRateLimit applies 10 scans/hr per user (token bucket: 10 tokens, refill 1/360 per sec).
func ScanCreateRateLimit(limiter rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString(UserIDKey)
		if userID == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:scan_create:%s", userID)
		allowed, err := limiter.AllowRequest(c.Request.Context(), key, 10, 10.0/3600.0)
		if err != nil || !allowed {
			if err == nil {
				metrics.RateLimitExceededTotal.WithLabelValues("scan_create").Inc()
			}
			abortRateLimited(c, "scan creation rate limit exceeded: max 10 scans per hour")
			return
		}
		c.Next()
	}
}

func abortRateLimited(c *gin.Context, msg string) {
	rid, _ := c.Get(RequestIDKey)
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":       msg,
		"code":        "RATE_LIMIT_EXCEEDED",
		"status_code": http.StatusTooManyRequests,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}
