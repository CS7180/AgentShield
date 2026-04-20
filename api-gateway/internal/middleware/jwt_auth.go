package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/agentshield/api-gateway/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey = "user_id"
const UserEmailKey = "user_email"

func JWTAuth(keyFunc jwt.Keyfunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortUnauthorized(c, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortUnauthorized(c, "invalid Authorization header format")
			return
		}

		claims, err := auth.ParseSupabaseTokenWithKeyFunc(parts[1], keyFunc)
		if err != nil {
			abortUnauthorized(c, "invalid token: "+err.Error())
			return
		}

		c.Set(UserIDKey, claims.Subject)
		c.Set(UserEmailKey, claims.Email)
		c.Next()
	}
}

func abortUnauthorized(c *gin.Context, msg string) {
	rid, _ := c.Get(RequestIDKey)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":       msg,
		"code":        "UNAUTHORIZED",
		"status_code": http.StatusUnauthorized,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}
