package resource

import (
	"archive/zip"
	"context"
	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type ImportDataUseCase struct {
	memoRepo  port.MemoRepository
	resRepo   port.ResourceRepository
	blobStore port.BlobStore
}

func NewImportDataUseCase(memoRepo port.MemoRepository, resRepo port.ResourceRepository, blobStore port.BlobStore) *ImportDataUseCase {
	return &ImportDataUseCase{
		memoRepo:  memoRepo,
		resRepo:   resRepo,
		blobStore: blobStore,
	}
}

// isValidInternalPath 检查路径是否安全，防止目录穿越
func isValidInternalPath(path string) bool {
	// 拒绝空路径和绝对路径
	if path == "" || strings.HasPrefix(path, "/") {
		return false
	}

	// 使用 filepath.Clean 清理路径，例如解析 ../
	cleaned := filepath.Clean(path)

	// 如果清理后的路径仍以 ../ 开头，说明试图越界
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return false
	}

	return true
}

func (uc *ImportDataUseCase) Execute(ctx context.Context, userID int64, r io.ReaderAt, size int64) (*dto.ImportReport, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("invalid zip file: %w", err)
	}
	report := &dto.ImportReport{
		Memos: dto.ImportSectionReport{
			Details: make([]dto.ImportSkipDetail, 0),
		},
		Resources: dto.ImportSectionReport{
			Details: make([]dto.ImportSkipDetail, 0),
		},
	}

	existingResources, err := uc.resRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list existing resources: %w", err)
	}
	resourceByID := make(map[string]struct{}, len(existingResources))
	resourceByHash := make(map[string]struct{}, len(existingResources))
	resourceByPath := make(map[string]struct{}, len(existingResources))
	for _, res := range existingResources {
		resourceByID[res.ID] = struct{}{}
		if res.Hash != "" {
			resourceByHash[res.Hash] = struct{}{}
		}
		if res.InternalPath != "" {
			resourceByPath[res.InternalPath] = struct{}{}
		}
	}

	existingMemos, err := uc.memoRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list existing memos: %w", err)
	}
	memoByUUID := make(map[string]string, len(existingMemos))
	for _, memo := range existingMemos {
		if memo.UUID != "" {
			memoByUUID[memo.UUID] = memo.UUID
		}
	}

	memoFile, err := zr.Open("memos.json")
	if err != nil {
		return nil, fmt.Errorf("missing memos.json in archive: %w", err)
	}
	defer memoFile.Close()

	var memos []entity.Memo
	if err := json.NewDecoder(memoFile).Decode(&memos); err != nil {
		return nil, fmt.Errorf("decode memos.json: %w", err)
	}
	memoUUIDMap := make(map[string]string, len(memos))

	for _, m := range memos {
		sourceUUID := m.UUID
		if m.UUID == "" {
			appendSkip(&report.Memos, "memo", sourceUUID, "invalid_memo_uuid")
			continue
		}
		if _, exists := memoByUUID[m.UUID]; exists {
			appendSkip(&report.Memos, "memo", m.UUID, "duplicate_by_memo_uuid")
			if sourceUUID != "" {
				memoUUIDMap[sourceUUID] = m.UUID
			}
			continue
		}

		si := port.SearchIndex{
			BodyTokens:  m.Content,
			IsEphemeral: false,
		}
		if err := uc.memoRepo.Create(ctx, &m, userID, m.Tags, nil, si); err != nil {
			return nil, fmt.Errorf("restore memo %s: %w", sourceUUID, err)
		}

		memoByUUID[m.UUID] = m.UUID
		if sourceUUID != "" {
			memoUUIDMap[sourceUUID] = m.UUID
		}
		report.Memos.Imported++
	}

	// 2. 恢复资源元数据与物理文件
	resFile, err := zr.Open("resources.json")
	if err != nil {
		return nil, fmt.Errorf("missing resources.json in archive: %w", err)
	}
	defer resFile.Close()

	var resources []entity.Resource
	if err := json.NewDecoder(resFile).Decode(&resources); err != nil {
		return nil, fmt.Errorf("decode resources.json: %w", err)
	}

	for _, res := range resources {
		if _, exists := resourceByID[res.ID]; exists {
			appendSkip(&report.Resources, "resource", res.ID, "duplicate_by_id")
			continue
		}
		if res.Hash != "" {
			if _, exists := resourceByHash[res.Hash]; exists {
				appendSkip(&report.Resources, "resource", res.Hash, "duplicate_by_hash")
				continue
			}
		}
		if res.InternalPath != "" {
			if _, exists := resourceByPath[res.InternalPath]; exists {
				appendSkip(&report.Resources, "resource", res.InternalPath, "duplicate_by_path")
				continue
			}
		}
		if res.ID == "" || res.InternalPath == "" {
			appendSkip(&report.Resources, "resource", res.ID, "invalid_metadata")
			continue
		}

		// Security: Validate internal_path to prevent Path Traversal
		if !isValidInternalPath(res.InternalPath) {
			appendSkip(&report.Resources, "resource", res.InternalPath, "path_traversal_attempt")
			continue
		}

		if res.MemoUUID != "" {
			mappedUUID, exists := memoUUIDMap[res.MemoUUID]
			if !exists {
				appendSkip(&report.Resources, "resource", res.ID, "invalid_memo_reference")
				continue
			}
			res.MemoUUID = mappedUUID
		}

		zipPath := fmt.Sprintf("assets/%s", res.InternalPath)
		zf, err := zr.Open(zipPath)
		if err != nil {
			return nil, fmt.Errorf("missing resource file %s in archive: %w", zipPath, err)
		}

		if err := uc.blobStore.Put(ctx, res.InternalPath, zf); err != nil {
			_ = zf.Close()
			return nil, fmt.Errorf("restore resource file %s: %w", res.InternalPath, err)
		}
		if err := zf.Close(); err != nil {
			return nil, fmt.Errorf("close resource file %s from archive: %w", zipPath, err)
		}

		if err := uc.resRepo.Save(ctx, &res, userID); err != nil {
			return nil, fmt.Errorf("save resource record %s: %w", res.ID, err)
		}
		resourceByID[res.ID] = struct{}{}
		if res.Hash != "" {
			resourceByHash[res.Hash] = struct{}{}
		}
		resourceByPath[res.InternalPath] = struct{}{}
		report.Resources.Imported++
	}

	report.Memos.Skipped = len(report.Memos.Details)
	report.Resources.Skipped = len(report.Resources.Details)
	return report, nil
}

func appendSkip(section *dto.ImportSectionReport, entity, key, reason string) {
	section.Details = append(section.Details, dto.ImportSkipDetail{
		Entity: entity,
		Key:    key,
		Reason: reason,
	})
}
