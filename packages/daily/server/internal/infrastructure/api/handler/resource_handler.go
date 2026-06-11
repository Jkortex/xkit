package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"daily/internal/infrastructure/api/middleware"
	"daily/internal/interfaces/controller"
	"daily/internal/interfaces/presenter"
	"github.com/gin-gonic/gin"
)

type ResourceHandler struct {
	ctrl      *controller.ResourceController
	presenter presenter.IResourcePresenter
	logger    *slog.Logger
}

func NewResourceHandler(ctrl *controller.ResourceController, pres presenter.IResourcePresenter, l *slog.Logger) *ResourceHandler {
	return &ResourceHandler{
		ctrl:      ctrl,
		presenter: pres,
		logger:    l,
	}
}

func (h *ResourceHandler) Upload(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	defer f.Close()

	res, err := h.ctrl.Upload(c.Request.Context(), user.ID, file.Filename, file.Header.Get("Content-Type"), file.Size, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}

	h.logger.InfoContext(c, "resource uploaded", "user_id", user.ID, "resource_id", res.ID, "filename", res.FileName)
	c.JSON(http.StatusCreated, h.presenter.PresentResource(res))
}

func (h *ResourceHandler) Get(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	id := c.Param("id")
	out, err := h.ctrl.Get(c.Request.Context(), user.ID, id)
	if err != nil {
		c.JSON(statusFromError(err), h.presenter.PresentError(err))
		return
	}
	defer out.Content.Close()

	c.Header("Content-Type", out.Resource.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", out.Resource.Size))
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", out.Resource.FileName))

	c.DataFromReader(http.StatusOK, out.Resource.Size, out.Resource.MimeType, out.Content, nil)
}

func (h *ResourceHandler) Export(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	filename := fmt.Sprintf("daily_export_%s.zip", time.Now().UTC().Format("20060102T150405Z"))
	tmpFile, err := os.CreateTemp("", "daily-export-*.zip")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := h.ctrl.Export(c.Request.Context(), user.ID, tmpFile); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}

	stat, err := tmpFile.Stat()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}

	h.logger.InfoContext(c, "data exported", "user_id", user.ID)
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.DataFromReader(http.StatusOK, stat.Size(), "application/zip", tmpFile, nil)
}

func (h *ResourceHandler) Import(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, h.presenter.PresentError(fmt.Errorf("unauthorized")))
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, h.presenter.PresentError(err))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}
	defer f.Close()

	report, err := h.ctrl.Import(c.Request.Context(), user.ID, f, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, h.presenter.PresentError(err))
		return
	}

	h.logger.InfoContext(c, "data imported", "user_id", user.ID, "memos_imported", report.Memos.Imported)
	c.JSON(http.StatusOK, gin.H{
		"message":            "导入成功",
		"memos_imported":     report.Memos.Imported,
		"resources_imported": report.Resources.Imported,
		"memos_skipped":      report.Memos.Skipped,
		"resources_skipped":  report.Resources.Skipped,
		"report":             report,
	})
}
