package middleware

import (
	"net/http"

	"daily/internal/application/dto"
	"daily/internal/interfaces/controller"
	"github.com/gin-gonic/gin"
)

const currentUserContextKey = "current_user"

func SessionAuthMiddleware(ctrl *controller.AuthController, adminUsername string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := ctrl.GetDefaultAdmin(c.Request.Context(), adminUsername)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to resolve default admin user",
				"code":  "INTERNAL_ERROR",
			})
			return
		}
		SetCurrentUser(c, user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*dto.UserResponse, bool) {
	v, ok := c.Get(currentUserContextKey)
	if !ok {
		return nil, false
	}
	user, ok := v.(*dto.UserResponse)
	return user, ok
}

func SetCurrentUser(c *gin.Context, user *dto.UserResponse) {
	c.Set(currentUserContextKey, user)
}
