package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"daily/internal/infrastructure/api/middleware"
	"daily/internal/interfaces/controller"
	"daily/internal/interfaces/presenter"
	"github.com/gin-gonic/gin"
)

var historyUUIDRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func parseHistoryMemoUUID(c *gin.Context) string {
	uuid := c.Param("uuid")
	if uuid == "" || !historyUUIDRegex.MatchString(uuid) {
		return ""
	}
	return uuid
}

type MemoHistoryHandler struct {
	ctrl      *controller.MemoHistoryController
	presenter presenter.IMemoPresenter
	logger    *slog.Logger
}

func NewMemoHistoryHandler(ctrl *controller.MemoHistoryController, pres presenter.IMemoPresenter, l *slog.Logger) *MemoHistoryHandler {
	return &MemoHistoryHandler{
		ctrl:      ctrl,
		presenter: pres,
		logger:    l,
	}
}

func (h *MemoHistoryHandler) ListHistory(c *gin.Context) {
	uuid := parseHistoryMemoUUID(c)
	if uuid == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid memo uuid")))
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}

	res, err := h.ctrl.ListHistory(c.Request.Context(), user.ID, uuid)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *MemoHistoryHandler) Rollback(c *gin.Context) {
	uuid := parseHistoryMemoUUID(c)
	if uuid == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid memo uuid")))
		return
	}
	historyID := c.Param("hid")
	if strings.TrimSpace(historyID) == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid history id")))
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}

	res, err := h.ctrl.Rollback(c.Request.Context(), user.ID, uuid, historyID)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "memo rolled back", "user_id", user.ID, "memo_uuid", uuid, "history_id", historyID)
	c.JSON(http.StatusOK, h.presenter.PresentMemo(res))
}
