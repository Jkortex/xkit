package resource

import (
	"context"
	"fmt"
	"io"

	"daily/internal/application/dto"
	"daily/internal/application/port"
)

type GetResourceOutput struct {
	Resource *dto.ResourceResponse
	Content  io.ReadCloser
}

type GetResourceUseCase struct {
	repo      port.ResourceRepository
	blobStore port.BlobStore
}

func NewGetResourceUseCase(repo port.ResourceRepository, blobStore port.BlobStore) *GetResourceUseCase {
	return &GetResourceUseCase{
		repo:      repo,
		blobStore: blobStore,
	}
}

func (uc *GetResourceUseCase) Execute(ctx context.Context, userID int64, id string) (*GetResourceOutput, error) {
	// 1. 获取元数据
	res, err := uc.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// 2. 获取文件流
	content, err := uc.blobStore.Get(ctx, res.InternalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return &GetResourceOutput{
		Resource: &dto.ResourceResponse{
			ID:        res.ID,
			FileName:  res.FileName,
			Size:      res.Size,
			MimeType:  res.MimeType,
			CreatedAt: res.CreatedAt,
		},
		Content: content,
	}, nil
}
