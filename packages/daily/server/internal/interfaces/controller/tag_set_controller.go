package controller

import (
	"context"
	"daily/internal/application/dto"
	tag_set "daily/internal/application/usecase/tag_set"
)

type TagSetController struct {
	svc *tag_set.Service
}

func NewTagSetController(svc *tag_set.Service) *TagSetController {
	return &TagSetController{svc: svc}
}

// --- Group ---

func (ctrl *TagSetController) ListGroups(ctx context.Context, userID int64) ([]*dto.TagSetGroupResponse, error) {
	return ctrl.svc.ListGroups(ctx, userID)
}

func (ctrl *TagSetController) CreateGroup(ctx context.Context, userID int64, req dto.CreateTagSetGroupRequest) (*dto.TagSetGroupResponse, error) {
	return ctrl.svc.CreateGroup(ctx, userID, req)
}

func (ctrl *TagSetController) UpdateGroup(ctx context.Context, userID int64, id string, req dto.UpdateTagSetGroupRequest) (*dto.TagSetGroupResponse, error) {
	return ctrl.svc.UpdateGroup(ctx, userID, id, req)
}

func (ctrl *TagSetController) DeleteGroup(ctx context.Context, userID int64, id string) error {
	return ctrl.svc.DeleteGroup(ctx, userID, id)
}

// --- TagSet ---

func (ctrl *TagSetController) ListTagSets(ctx context.Context, userID int64, groupID *string) ([]*dto.TagSetResponse, error) {
	return ctrl.svc.ListTagSets(ctx, userID, groupID)
}

func (ctrl *TagSetController) CreateTagSet(ctx context.Context, userID int64, req dto.CreateTagSetRequest) (*dto.TagSetResponse, error) {
	return ctrl.svc.CreateTagSet(ctx, userID, req)
}

func (ctrl *TagSetController) GetTagSet(ctx context.Context, userID int64, id string) (*dto.TagSetResponse, error) {
	return ctrl.svc.GetTagSet(ctx, userID, id)
}

func (ctrl *TagSetController) UpdateTagSet(ctx context.Context, userID int64, id string, req dto.UpdateTagSetRequest) (*dto.TagSetResponse, error) {
	return ctrl.svc.UpdateTagSet(ctx, userID, id, req)
}

func (ctrl *TagSetController) DeleteTagSet(ctx context.Context, userID int64, id string) error {
	return ctrl.svc.DeleteTagSet(ctx, userID, id)
}

func (ctrl *TagSetController) TouchTagSet(ctx context.Context, userID int64, id string) error {
	return ctrl.svc.TouchTagSet(ctx, userID, id)
}
