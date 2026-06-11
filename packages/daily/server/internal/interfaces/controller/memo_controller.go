package controller

import (
	"context"
	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/application/usecase/memo"
)

type MemoController struct {
	memoSvc *memo.MemoService
}

func NewMemoController(
	memoSvc *memo.MemoService,
) *MemoController {
	return &MemoController{
		memoSvc: memoSvc,
	}
}

func (ctrl *MemoController) TransitionTask(ctx context.Context, userID int64, uuid string, status string, agentID string) (*dto.MemoResponse, error) {
	return ctrl.memoSvc.TransitionTask(ctx, userID, uuid, status, agentID)
}

func (ctrl *MemoController) Get(ctx context.Context, userID int64, uuid string) (*dto.MemoResponse, error) {
	return ctrl.memoSvc.Get(ctx, userID, uuid)
}

func (ctrl *MemoController) Update(
	ctx context.Context,
	userID int64,
	uuid string,
	input dto.UpdateMemoRequest,
) (*dto.MemoResponse, error) {
	return ctrl.memoSvc.Update(ctx, userID, uuid, input)
}

func (ctrl *MemoController) Delete(ctx context.Context, userID int64, uuid string) error {
	return ctrl.memoSvc.Delete(ctx, userID, uuid)
}

func (ctrl *MemoController) GetRandom(ctx context.Context, userID int64) (*dto.MemoResponse, error) {
	return ctrl.memoSvc.GetRandom(ctx, userID)
}

func (ctrl *MemoController) Create(ctx context.Context, userID int64, content string, tags []string, resourceIDs []string, ttl string) (*dto.MemoResponse, error) {
	input := dto.CreateMemoRequest{
		Content:     content,
		Tags:        tags,
		ResourceIDs: resourceIDs,
		TimeToLive:  ttl,
	}
	return ctrl.memoSvc.Create(ctx, userID, input)
}

func (ctrl *MemoController) List(ctx context.Context, userID int64, filter port.MemoFilter) ([]*dto.MemoResponse, error) {
	return ctrl.memoSvc.List(ctx, userID, filter)
}

func (ctrl *MemoController) GetStats(ctx context.Context, userID int64) (*dto.StatsResponse, error) {
	return ctrl.memoSvc.GetStats(ctx, userID)
}

func (ctrl *MemoController) BatchArchive(ctx context.Context, userID int64, uuids []string) (*dto.BatchResult, error) {
	return ctrl.memoSvc.BatchArchive(ctx, userID, uuids)
}

func (ctrl *MemoController) BatchDelete(ctx context.Context, userID int64, uuids []string) (*dto.BatchResult, error) {
	return ctrl.memoSvc.BatchDelete(ctx, userID, uuids)
}

func (ctrl *MemoController) BatchTag(ctx context.Context, userID int64, uuids []string, addTags []string, removeTags []string) (*dto.BatchResult, error) {
	return ctrl.memoSvc.BatchTag(ctx, userID, uuids, addTags, removeTags)
}
