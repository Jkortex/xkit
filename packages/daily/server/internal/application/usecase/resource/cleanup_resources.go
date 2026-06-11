package resource

import (
	"context"
	"log/slog"

	"daily/internal/application/port"
)

type CleanupResourcesUseCase struct {
	repo      port.ResourceRepository
	blobStore port.BlobStore
}

func NewCleanupResourcesUseCase(repo port.ResourceRepository, blobStore port.BlobStore) *CleanupResourcesUseCase {
	return &CleanupResourcesUseCase{
		repo:      repo,
		blobStore: blobStore,
	}
}

func (uc *CleanupResourcesUseCase) Execute(ctx context.Context) (int, error) {
	// 0. 先清理数据库中无引用的孤儿资源记录 (保护历史引用的逻辑已包含在 repository 中)
	if _, err := uc.repo.CleanupOrphanResources(ctx); err != nil {
		slog.Error("failed to cleanup orphan resource records", "error", err)
	}

	// 1. 获取数据库中登记的所有文件路径
	paths, err := uc.repo.ListTrackedPaths(ctx)
	if err != nil {
		return 0, err
	}
	dbPaths := make(map[string]struct{})
	for _, path := range paths {
		dbPaths[path] = struct{}{}
	}

	// 2. 获取磁盘上所有的物理文件路径
	physicalPaths, err := uc.blobStore.ListAll(ctx)
	if err != nil {
		return 0, err
	}

	// 3. 对比并删除孤儿文件
	count := 0
	for _, path := range physicalPaths {
		if _, exists := dbPaths[path]; !exists {
			slog.Warn("deleting orphan resource file", "path", path)
			if err := uc.blobStore.Delete(ctx, path); err != nil {
				slog.Error("failed to delete orphan file", "path", path, "error", err)
				continue
			}
			count++
		}
	}

	return count, nil
}
