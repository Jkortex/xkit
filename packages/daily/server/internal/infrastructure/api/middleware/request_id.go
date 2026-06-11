package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const RequestIdKey contextKey = "request_id"

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.Must(uuid.NewV7()).String()
		}

		c.Set(string(RequestIdKey), rid)
		ctx := context.WithValue(c.Request.Context(), RequestIdKey, rid)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
