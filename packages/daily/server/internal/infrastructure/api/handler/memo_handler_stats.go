package handler

import (
	"fmt"
	"net/http"

	"daily/internal/infrastructure/api/middleware"
	"github.com/gin-gonic/gin"
)

func (h *MemoHandler) Stats(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	res, err := h.ctrl.GetStats(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentStats(res))
}
