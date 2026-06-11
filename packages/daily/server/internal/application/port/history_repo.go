package port

import (
	"context"
)

type MemoHistoryRecord struct {
	ID          string
	MemoUUID    string
	Content     string
	Tags        []string
	ResourceIDs []string
	CreatedAt   string
}

type MemoHistoryRepository interface {
	SaveHistory(ctx context.Context, record *MemoHistoryRecord, userID int64) error
	ListHistory(ctx context.Context, memoUUID string, userID int64) ([]*MemoHistoryRecord, error)
	GetHistoryByID(ctx context.Context, historyID string, userID int64) (*MemoHistoryRecord, error)
	DeleteOldHistory(ctx context.Context, memoUUID string, keepLimit int) (int64, error)
}
