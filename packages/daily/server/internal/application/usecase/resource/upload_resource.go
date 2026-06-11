package resource

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"github.com/google/uuid"
)

type UploadResourceUseCase struct {
	repo      port.ResourceRepository
	blobStore port.BlobStore
}

func NewUploadResourceUseCase(repo port.ResourceRepository, blobStore port.BlobStore) *UploadResourceUseCase {
	return &UploadResourceUseCase{
		repo:      repo,
		blobStore: blobStore,
	}
}

func (uc *UploadResourceUseCase) Execute(ctx context.Context, userID int64, input dto.UploadResourceInput) (*dto.ResourceResponse, error) {
	// 1. 读取并计算哈希 (SHA-256)
	hasher := sha256.New()
	buf := new(bytes.Buffer)
	tee := io.TeeReader(input.Content, hasher)

	if _, err := io.Copy(buf, tee); err != nil {
		return nil, fmt.Errorf("read file stream: %w", err)
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// 2. 生成存储路径 (assets/YYYY/MM/hash.ext)
	now := time.Now()
	relPath := fmt.Sprintf("%d/%02d/%s%s",
		now.Year(), now.Month(), hash, filepath.Ext(input.FileName))

	// 3. 保存物理文件
	if err := uc.blobStore.Put(ctx, relPath, buf); err != nil {
		return nil, fmt.Errorf("store physical file: %w", err)
	}

	// 4. 保存数据库记录
	res := &entity.Resource{
		ID:           uuid.Must(uuid.NewV7()).String(),
		FileName:     input.FileName,
		Hash:         hash,
		Size:         input.Size,
		MimeType:     input.MimeType,
		InternalPath: relPath,
		CreatedAt:    now,
	}

	if err := uc.repo.Save(ctx, res, userID); err != nil {
		return nil, fmt.Errorf("save database record: %w", err)
	}

	return &dto.ResourceResponse{
		ID:        res.ID,
		FileName:  res.FileName,
		Size:      res.Size,
		MimeType:  res.MimeType,
		CreatedAt: res.CreatedAt,
	}, nil
}
