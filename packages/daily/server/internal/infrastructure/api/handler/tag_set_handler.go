package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"daily/internal/application/dto"
	"daily/internal/infrastructure/api/middleware"
	"daily/internal/interfaces/controller"
	"daily/internal/interfaces/presenter"
	"github.com/gin-gonic/gin"
)

type TagSetHandler struct {
	ctrl      *controller.TagSetController
	presenter presenter.IMemoPresenter
	logger    *slog.Logger
}

func NewTagSetHandler(ctrl *controller.TagSetController, pres presenter.IMemoPresenter, l *slog.Logger) *TagSetHandler {
	return &TagSetHandler{
		ctrl:      ctrl,
		presenter: pres,
		logger:    l,
	}
}

// --- Group ---

func (h *TagSetHandler) ListGroups(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	res, err := h.ctrl.ListGroups(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentTagSetGroups(res))
}

func (h *TagSetHandler) CreateGroup(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	var req dto.CreateTagSetGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}
	res, err := h.ctrl.CreateGroup(c.Request.Context(), user.ID, req)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusCreated, h.presenter.PresentTagSetGroup(res))
}

func (h *TagSetHandler) UpdateGroup(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("missing group id")))
		return
	}
	var req dto.UpdateTagSetGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}
	res, err := h.ctrl.UpdateGroup(c.Request.Context(), user.ID, id, req)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentTagSetGroup(res))
}

func (h *TagSetHandler) DeleteGroup(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("missing group id")))
		return
	}
	if err := h.ctrl.DeleteGroup(c.Request.Context(), user.ID, id); err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// --- TagSet ---

func (h *TagSetHandler) ListTagSets(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	var groupID *string
	if g := strings.TrimSpace(c.Query("group_id")); g != "" {
		groupID = &g
	}
	res, err := h.ctrl.ListTagSets(c.Request.Context(), user.ID, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentTagSets(res))
}

func (h *TagSetHandler) CreateTagSet(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	var req dto.CreateTagSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}
	res, err := h.ctrl.CreateTagSet(c.Request.Context(), user.ID, req)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusCreated, h.presenter.PresentTagSet(res))
}

func (h *TagSetHandler) GetTagSet(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("missing tag set id")))
		return
	}
	res, err := h.ctrl.GetTagSet(c.Request.Context(), user.ID, id)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentTagSet(res))
}

func (h *TagSetHandler) UpdateTagSet(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("missing tag set id")))
		return
	}
	var req dto.UpdateTagSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}
	res, err := h.ctrl.UpdateTagSet(c.Request.Context(), user.ID, id, req)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusOK, h.presenter.PresentTagSet(res))
}

func (h *TagSetHandler) DeleteTagSet(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("missing tag set id")))
		return
	}
	if err := h.ctrl.DeleteTagSet(c.Request.Context(), user.ID, id); err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *TagSetHandler) TouchTagSet(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(fmt.Errorf("missing tag set id")))
		return
	}
	if err := h.ctrl.TouchTagSet(c.Request.Context(), user.ID, id); err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
