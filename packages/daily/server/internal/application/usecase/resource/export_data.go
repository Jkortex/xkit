package resource

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"daily/internal/application/port"
)

type ExportDataUseCase struct {
	memoRepo  port.MemoRepository
	resRepo   port.ResourceRepository
	blobStore port.BlobStore
}

func NewExportDataUseCase(memoRepo port.MemoRepository, resRepo port.ResourceRepository, blobStore port.BlobStore) *ExportDataUseCase {
	return &ExportDataUseCase{
		memoRepo:  memoRepo,
		resRepo:   resRepo,
		blobStore: blobStore,
	}
}

func (uc *ExportDataUseCase) Execute(ctx context.Context, userID int64, w io.Writer) error {
	zw := zip.NewWriter(w)

	// 1. 导出笔记数据
	memos, err := uc.memoRepo.ListAll(ctx, userID)
	if err != nil {
		return fmt.Errorf("list memos: %w", err)
	}
	memoData, err := json.MarshalIndent(memos, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal memos: %w", err)
	}
	mf, err := zw.Create("memos.json")
	if err != nil {
		return fmt.Errorf("create memos.json in zip: %w", err)
	}
	if _, err := mf.Write(memoData); err != nil {
		return fmt.Errorf("write memos.json in zip: %w", err)
	}

	// 2. 导出资源元数据
	resources, err := uc.resRepo.ListAll(ctx, userID)
	if err != nil {
		return fmt.Errorf("list resources: %w", err)
	}
	resData, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal resources: %w", err)
	}
	rf, err := zw.Create("resources.json")
	if err != nil {
		return fmt.Errorf("create resources.json in zip: %w", err)
	}
	if _, err := rf.Write(resData); err != nil {
		return fmt.Errorf("write resources.json in zip: %w", err)
	}

	// 3. 导出原始附件文件
	for _, res := range resources {
		zipPath := fmt.Sprintf("assets/%s", res.InternalPath)
		zf, err := zw.Create(zipPath)
		if err != nil {
			return fmt.Errorf("create resource entry %s in zip: %w", zipPath, err)
		}

		rc, err := uc.blobStore.Get(ctx, res.InternalPath)
		if err != nil {
			return fmt.Errorf("read resource file %s: %w", res.InternalPath, err)
		}
		if _, err := io.Copy(zf, rc); err != nil {
			_ = rc.Close()
			return fmt.Errorf("write resource file %s into zip: %w", res.InternalPath, err)
		}
		if err := rc.Close(); err != nil {
			return fmt.Errorf("close resource file %s: %w", res.InternalPath, err)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("finalize export zip: %w", err)
	}
	return nil
}
