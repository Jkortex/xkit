package middleware

import (
	"github.com/gin-gonic/gin"
)

func RateLimitGlobal() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func RateLimitByIP(isAuthEndpoint bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func RateLimitAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
