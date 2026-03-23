package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				rid, _ := c.Get(RequestIDKey)
				logger.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("request_id", rid.(string)),
					zap.String("path", c.FullPath()),
				)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "internal server error",
					"code":       "INTERNAL_ERROR",
					"status_code": http.StatusInternalServerError,
					"timestamp":  time.Now().UTC(),
					"request_id": rid,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
