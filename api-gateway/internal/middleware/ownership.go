package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/agentshield/api-gateway/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OwnershipRepo is the minimal interface used to verify scan ownership.
type OwnershipRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (ownerID uuid.UUID, err error)
}

// Ownership verifies that the authenticated user owns the scan referenced by :id.
func Ownership(repo OwnershipRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		scanIDStr := c.Param("id")
		scanID, err := uuid.Parse(scanIDStr)
		if err != nil {
			abortForbidden(c, "invalid scan id")
			return
		}

		userIDStr := c.GetString(UserIDKey)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			abortForbidden(c, "invalid user id in token")
			return
		}

		ownerID, err := repo.GetByID(c.Request.Context(), scanID)
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				rid, _ := c.Get(RequestIDKey)
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"error":       "scan not found",
					"code":        "NOT_FOUND",
					"status_code": http.StatusNotFound,
					"timestamp":   time.Now().UTC(),
					"request_id":  rid,
				})
				return
			}
			abortForbidden(c, "could not verify ownership")
			return
		}

		if ownerID != userID {
			abortForbidden(c, "access denied")
			return
		}

		c.Next()
	}
}

func abortForbidden(c *gin.Context, msg string) {
	rid, _ := c.Get(RequestIDKey)
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error":       msg,
		"code":        "FORBIDDEN",
		"status_code": http.StatusForbidden,
		"timestamp":   time.Now().UTC(),
		"request_id":  rid,
	})
}
