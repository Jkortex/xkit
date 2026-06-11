package controller

import (
	"context"
	"daily/internal/application/dto"
	"daily/internal/application/usecase/memo"
)

type MemoHistoryController struct {
	memoSvc *memo.MemoService
}

func NewMemoHistoryController(
	memoSvc *memo.MemoService,
) *MemoHistoryController {
	return &MemoHistoryController{
		memoSvc: memoSvc,
	}
}

func (ctrl *MemoHistoryController) ListHistory(
	ctx context.Context,
	userID int64,
	memoUUID string,
) ([]*dto.MemoHistoryResponse, error) {
	return ctrl.memoSvc.ListHistory(ctx, userID, memoUUID)
}

func (ctrl *MemoHistoryController) Rollback(
	ctx context.Context,
	userID int64,
	memoUUID string,
	historyID string,
) (*dto.MemoResponse, error) {
	return ctrl.memoSvc.Rollback(ctx, userID, memoUUID, historyID)
}
