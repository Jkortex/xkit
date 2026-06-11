package handler

import (
	"log/slog"
	"net/http"

	"daily/internal/infrastructure/api/middleware"
	"daily/internal/interfaces/controller"
	"daily/internal/interfaces/presenter"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	ctrl      *controller.AuthController
	presenter presenter.IMemoPresenter
	logger    *slog.Logger
}

func NewAuthHandler(ctrl *controller.AuthController, pres presenter.IMemoPresenter, l *slog.Logger) *AuthHandler {
	return &AuthHandler{
		ctrl:      ctrl,
		presenter: pres,
		logger:    l,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(nil))
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(nil))
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
