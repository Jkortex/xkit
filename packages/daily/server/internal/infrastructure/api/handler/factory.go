package handler

import (
	"log/slog"

	authuc "daily/internal/application/usecase/auth"
	"daily/internal/application/usecase/memo"
	"daily/internal/application/usecase/resource"
	tag_set "daily/internal/application/usecase/tag_set"
	api_presenter "daily/internal/infrastructure/api/presenter"
	"daily/internal/infrastructure/container"
	"daily/internal/interfaces/controller"
)

// Handlers 包装了所有接口处理器
type Handlers struct {
	Memo        *MemoHandler
	Tag         *TagHandler
	MemoHistory *MemoHistoryHandler
	Res         *ResourceHandler
	Auth        *AuthHandler
	TagSet      *TagSetHandler
	AuthCtrl    *controller.AuthController
}

// NewHandlers 负责将业务用例注入并组装成处理器
func NewHandlers(c *container.Container, l *slog.Logger) *Handlers {
	// 1. 初始化 Services & UseCases
	memoSvc := memo.NewMemoService(c.MemoRepo, c.ResRepo, c.MemoRepo, c.Tokenizer)
	tagSvc := memo.NewTagService(c.MemoRepo)

	uploadUC := resource.NewUploadResourceUseCase(c.ResRepo, c.BlobStore)
	getResourceUC := resource.NewGetResourceUseCase(c.ResRepo, c.BlobStore)
	exportUC := resource.NewExportDataUseCase(c.MemoRepo, c.ResRepo, c.BlobStore)
	importUC := resource.NewImportDataUseCase(c.MemoRepo, c.ResRepo, c.BlobStore)

	identitySvc := authuc.NewIdentityService(c.UserRepo)

	// 2. 初始化 Controllers
	memoCtrl := controller.NewMemoController(memoSvc)
	tagCtrl := controller.NewTagController(tagSvc)
	memoHistoryCtrl := controller.NewMemoHistoryController(memoSvc)
	resCtrl := controller.NewResourceController(uploadUC, getResourceUC, exportUC, importUC)
	authCtrl := controller.NewAuthController(identitySvc)
	tagSetSvc := tag_set.NewService(c.TagSetGroupRepo, c.TagSetRepo)
	tagSetCtrl := controller.NewTagSetController(tagSetSvc)

	// 3. 构建 Handlers
	p := api_presenter.NewJsonPresenter()
	return &Handlers{
		Memo:        NewMemoHandler(memoCtrl, p, l),
		Tag:         NewTagHandler(tagCtrl, p, l),
		MemoHistory: NewMemoHistoryHandler(memoHistoryCtrl, p, l),
		Res:         NewResourceHandler(resCtrl, p, l),
		Auth:        NewAuthHandler(authCtrl, p, l),
		TagSet:      NewTagSetHandler(tagSetCtrl, p, l),
		AuthCtrl:    authCtrl,
	}
}
