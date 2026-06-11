package controller

import (
	"context"
	"daily/internal/application/dto"
	"daily/internal/application/usecase/memo"
)

type TagController struct {
	tagSvc *memo.TagService
}

func NewTagController(
	tagSvc *memo.TagService,
) *TagController {
	return &TagController{
		tagSvc: tagSvc,
	}
}

func (ctrl *TagController) ListTags(ctx context.Context, userID int64) ([]dto.TagStatResponse, error) {
	return ctrl.tagSvc.ListTags(ctx, userID)
}

func (ctrl *TagController) RenameTag(
	ctx context.Context,
	userID int64,
	from, to string,
) (*dto.RenameTagResponse, error) {
	return ctrl.tagSvc.RenameTag(ctx, userID, from, to)
}

func (ctrl *TagController) MergeTags(
	ctx context.Context,
	userID int64,
	sources []string,
	target string,
) (*dto.MergeTagsResponse, error) {
	return ctrl.tagSvc.MergeTags(ctx, userID, sources, target)
}

func (ctrl *TagController) UpsertTagAlias(
	ctx context.Context,
	userID int64,
	alias,
	canonical string,
) (*dto.TagAliasResponse, error) {
	return ctrl.tagSvc.UpsertTagAlias(ctx, userID, alias, canonical)
}

func (ctrl *TagController) ListTagAliases(
	ctx context.Context,
) ([]dto.TagAliasResponse, error) {
	return ctrl.tagSvc.ListTagAliases(ctx)
}

func (ctrl *TagController) DeleteTagAlias(ctx context.Context, alias string) error {
	return ctrl.tagSvc.DeleteTagAlias(ctx, alias)
}

func (ctrl *TagController) ListTagAudits(
	ctx context.Context,
	limit int,
	action string,
) ([]dto.TagAuditResponse, error) {
	return ctrl.tagSvc.ListTagAudits(ctx, limit, action)
}
