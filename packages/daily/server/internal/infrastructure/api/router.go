package api

import (
	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/api/middleware"
	"daily/internal/infrastructure/api/static"
	"daily/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// NewRouter 定义 API 路由规则
func NewRouter(cfg *config.Config, handlers *handler.Handlers) *gin.Engine {
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIdMiddleware())

	api := r.Group("/api/v1")
	{
		authSession := api.Group("/auth", middleware.SessionAuthMiddleware(handlers.AuthCtrl, cfg.BootstrapAdminUsername))
		{
			authSession.POST("/login", handlers.Auth.Login)
			authSession.POST("/logout", handlers.Auth.Logout)
			authSession.GET("/me", handlers.Auth.Me)
		}

		secured := api.Group("", middleware.SessionAuthMiddleware(handlers.AuthCtrl, cfg.BootstrapAdminUsername))
		{
			secured.GET("/stats", handlers.Memo.Stats)
			secured.GET("/tags", handlers.Tag.Tags)
			secured.POST("/tags/rename", handlers.Tag.RenameTag)
			secured.POST("/tags/merge", handlers.Tag.MergeTags)
			secured.GET("/tags/aliases", handlers.Tag.ListTagAliases)
			secured.POST("/tags/aliases", handlers.Tag.UpsertTagAlias)
			secured.DELETE("/tags/aliases/:alias", handlers.Tag.DeleteTagAlias)
			secured.GET("/tags/audits", handlers.Tag.TagAudits)

			secured.POST("/memos", handlers.Memo.Create)
			secured.GET("/memos", handlers.Memo.List)
			secured.GET("/memos/:uuid", handlers.Memo.Get)
			secured.POST("/memos/:uuid/transition", handlers.Memo.TransitionTask)
			secured.GET("/memos/random", handlers.Memo.Random)
			secured.PATCH("/memos/:uuid", handlers.Memo.Update)
			secured.DELETE("/memos/:uuid", handlers.Memo.Delete)
			secured.POST("/memos/batch/archive", handlers.Memo.BatchArchive)
			secured.POST("/memos/batch/delete", handlers.Memo.BatchDelete)
			secured.POST("/memos/batch/tag", handlers.Memo.BatchTag)
			secured.GET("/memos/:uuid/history", handlers.MemoHistory.ListHistory)
			secured.POST("/memos/:uuid/rollback/:hid", handlers.MemoHistory.Rollback)

			secured.POST("/resources", handlers.Res.Upload)
			secured.GET("/resources/:id", handlers.Res.Get)

			secured.GET("/tag-set-groups", handlers.TagSet.ListGroups)
			secured.POST("/tag-set-groups", handlers.TagSet.CreateGroup)
			secured.PATCH("/tag-set-groups/:id", handlers.TagSet.UpdateGroup)
			secured.DELETE("/tag-set-groups/:id", handlers.TagSet.DeleteGroup)

			secured.GET("/tag-sets", handlers.TagSet.ListTagSets)
			secured.POST("/tag-sets", handlers.TagSet.CreateTagSet)
			secured.GET("/tag-sets/:id", handlers.TagSet.GetTagSet)
			secured.PATCH("/tag-sets/:id", handlers.TagSet.UpdateTagSet)
			secured.DELETE("/tag-sets/:id", handlers.TagSet.DeleteTagSet)
			secured.POST("/tag-sets/:id/touch", handlers.TagSet.TouchTagSet)

			secured.GET("/system/export", handlers.Res.Export)
			secured.POST("/system/import", handlers.Res.Import)
		}
	}

	// SPA fallback: 非 API 路由返回前端页面
	r.NoRoute(static.GetGinHandler())

	return r
}
