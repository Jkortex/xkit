package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"daily/internal/infrastructure/api/middleware"
	"daily/internal/interfaces/controller"
	"daily/internal/interfaces/presenter"
	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	ctrl      *controller.TagController
	presenter presenter.IMemoPresenter
	logger    *slog.Logger
}

func NewTagHandler(ctrl *controller.TagController, pres presenter.IMemoPresenter, l *slog.Logger) *TagHandler {
	return &TagHandler{
		ctrl:      ctrl,
		presenter: pres,
		logger:    l,
	}
}

func (h *TagHandler) Tags(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	res, err := h.ctrl.ListTags(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentTags(res))
}

func (h *TagHandler) RenameTag(c *gin.Context) {
	var input struct {
		From string `json:"from"`
		To   string `json:"to"`
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
	res, err := h.ctrl.RenameTag(c.Request.Context(), user.ID, input.From, input.To)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "tag renamed", "user_id", user.ID, "from", input.From, "to", input.To)
	c.JSON(http.StatusOK, res)
}

func (h *TagHandler) MergeTags(c *gin.Context) {
	var input struct {
		Sources []string `json:"sources"`
		Target  string   `json:"target"`
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
	res, err := h.ctrl.MergeTags(c.Request.Context(), user.ID, input.Sources, input.Target)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "tags merged", "user_id", user.ID, "sources", input.Sources, "target", input.Target)
	c.JSON(http.StatusOK, res)
}

func (h *TagHandler) UpsertTagAlias(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	var input struct {
		Alias     string `json:"alias"`
		Canonical string `json:"canonical"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	res, err := h.ctrl.UpsertTagAlias(c.Request.Context(), user.ID, input.Alias, input.Canonical)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "tag alias upserted", "user_id", user.ID, "alias", input.Alias, "canonical", input.Canonical)
	c.JSON(http.StatusOK, res)
}

func (h *TagHandler) ListTagAliases(c *gin.Context) {
	res, err := h.ctrl.ListTagAliases(c.Request.Context())
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *TagHandler) DeleteTagAlias(c *gin.Context) {
	alias := c.Param("alias")
	if strings.TrimSpace(alias) == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("invalid alias")))
		return
	}
	if err := h.ctrl.DeleteTagAlias(c.Request.Context(), alias); err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	h.logger.InfoContext(c.Request.Context(), "tag alias deleted", "alias", alias)
	c.Status(http.StatusNoContent)
}

func (h *TagHandler) TagAudits(c *gin.Context) {
	limit, err := parseLimitOffset(c.DefaultQuery("limit", "20"), 20, 1, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}
	action := strings.TrimSpace(c.Query("action"))
	res, err := h.ctrl.ListTagAudits(c.Request.Context(), limit, action)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, res)
}
