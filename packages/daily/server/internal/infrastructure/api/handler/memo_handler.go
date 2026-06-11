package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/dto"
	"daily/internal/infrastructure/api/middleware"
	"daily/internal/infrastructure/notify"
	"daily/internal/interfaces/controller"
	"daily/internal/interfaces/presenter"
	"github.com/gin-gonic/gin"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type MemoHandler struct {
	ctrl      *controller.MemoController
	presenter presenter.IMemoPresenter
	logger    *slog.Logger
}

func NewMemoHandler(ctrl *controller.MemoController, pres presenter.IMemoPresenter, l *slog.Logger) *MemoHandler {
	return &MemoHandler{
		ctrl:      ctrl,
		presenter: pres,
		logger:    l,
	}
}

func (h *MemoHandler) Create(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	var input struct {
		Content     string   `json:"content"`
		Tags        []string `json:"tags"`
		ResourceIDs []string `json:"resource_ids"`
		TTL         string   `json:"ttl"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	if strings.TrimSpace(input.Content) == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("%w: content cannot be empty", apperr.ErrInvalidInput)))
		return
	}

	res, err := h.ctrl.Create(c.Request.Context(), user.ID, input.Content, input.Tags, input.ResourceIDs, input.TTL)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}

	h.logger.InfoContext(c.Request.Context(), "memo created", "user_id", user.ID, "memo_uuid", res.UUID)
	notify.Touch()
	c.JSON(http.StatusCreated, h.presenter.PresentMemo(res))
}

func parseMemoUUID(c *gin.Context) string {
	uuid := c.Param("uuid")
	if uuid == "" || !uuidRegex.MatchString(uuid) {
		return ""
	}
	return uuid
}

func (h *MemoHandler) Get(c *gin.Context) {
	uuid := parseMemoUUID(c)
	if uuid == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid memo uuid")))
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	res, err := h.ctrl.Get(c.Request.Context(), user.ID, uuid)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentMemo(res))
}

func (h *MemoHandler) List(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	filter, err := buildMemoFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}
	start := time.Now()
	res, err := h.ctrl.List(c.Request.Context(), user.ID, filter)
	if err != nil {
		status := statusFromError(err)
		c.JSON(status, h.presenter.PresentError(err))
		h.logger.InfoContext(c.Request.Context(), "memo list query failed", "user_id", user.ID, "status", status, "error", err)
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentMemos(res))
	h.logger.InfoContext(c.Request.Context(), "memo list query",
		"user_id", user.ID,
		"count", len(res),
		"duration_ms", time.Since(start).Milliseconds(),
		"filter_tag", filter.Tag,
		"filter_search", filter.Search,
	)
}

func (h *MemoHandler) Random(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	res, err := h.ctrl.GetRandom(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentMemo(res))
}

func (h *MemoHandler) TransitionTask(c *gin.Context) {
	uuid := parseMemoUUID(c)
	if uuid == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid memo uuid")))
		return
	}

	var input struct {
		Status  string `json:"status"`
		AgentID string `json:"agent_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	res, err := h.ctrl.TransitionTask(c.Request.Context(), user.ID, uuid, input.Status, input.AgentID)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	notify.Touch()
	c.JSON(http.StatusOK, h.presenter.PresentMemo(res))
}

func (h *MemoHandler) Update(c *gin.Context) {
	uuid := parseMemoUUID(c)
	if uuid == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid memo uuid")))
		return
	}

	var input dto.UpdateMemoRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	if strings.TrimSpace(input.Content) == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("%w: content cannot be empty", apperr.ErrInvalidInput)))
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	res, err := h.ctrl.Update(c.Request.Context(), user.ID, uuid, input)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "memo updated", "user_id", user.ID, "memo_uuid", uuid)
	notify.Touch()
	c.JSON(http.StatusOK, h.presenter.PresentMemo(res))
}

func (h *MemoHandler) Delete(c *gin.Context) {
	uuid := parseMemoUUID(c)
	if uuid == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid memo uuid")))
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	if err := h.ctrl.Delete(c.Request.Context(), user.ID, uuid); err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "memo deleted", "user_id", user.ID, "memo_uuid", uuid)
	notify.Touch()
	c.JSON(http.StatusNoContent, nil)
}

// BatchArchive archives multiple memos in a single operation.
func (h *MemoHandler) BatchArchive(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}

	var input struct {
		UUIDs []string `json:"uuids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	res, err := h.ctrl.BatchArchive(c.Request.Context(), user.ID, input.UUIDs)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "batch archive", "user_id", user.ID, "count", len(input.UUIDs))
	notify.Touch()
	c.JSON(http.StatusOK, h.presenter.PresentBatchResult(res))
}

// BatchDelete deletes multiple memos in a single operation.
func (h *MemoHandler) BatchDelete(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}

	var input struct {
		UUIDs []string `json:"uuids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	res, err := h.ctrl.BatchDelete(c.Request.Context(), user.ID, input.UUIDs)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "batch delete", "user_id", user.ID, "count", len(input.UUIDs))
	notify.Touch()
	c.JSON(http.StatusOK, h.presenter.PresentBatchResult(res))
}

// BatchTag adds or removes tags from multiple memos in a single operation.
func (h *MemoHandler) BatchTag(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}

	var input struct {
		UUIDs  []string `json:"uuids" binding:"required"`
		Add    []string `json:"add,omitempty"`
		Remove []string `json:"remove,omitempty"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	res, err := h.ctrl.BatchTag(c.Request.Context(), user.ID, input.UUIDs, input.Add, input.Remove)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "batch tag", "user_id", user.ID, "count", len(input.UUIDs))
	notify.Touch()
	c.JSON(http.StatusOK, h.presenter.PresentBatchResult(res))
}
