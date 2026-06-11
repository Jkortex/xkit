package memo

import (
	"context"
	"log/slog"

	"daily/internal/application/port"
)

// ArchiveExpiredMemosUseCase 负责定期归档过期的临时笔记
type ArchiveExpiredMemosUseCase struct {
	repo port.MemoWriteRepository
}

func NewArchiveExpiredMemosUseCase(repo port.MemoWriteRepository) *ArchiveExpiredMemosUseCase {
	return &ArchiveExpiredMemosUseCase{repo: repo}
}

func (uc *ArchiveExpiredMemosUseCase) Execute(ctx context.Context) error {
	affected, err := uc.repo.ArchiveExpired(ctx)
	if err != nil {
		return err
	}
	if affected > 0 {
		slog.Info("auto-archived expired memos", "count", affected)
	}
	return nil
}
