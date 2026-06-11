package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	sldb "daily/internal/infrastructure/persistence/sqlite/db"
)

func (r *SqliteMemoRepository) getMemoTags(ctx context.Context, memoUUID string) ([]string, error) {
	rows, err := r.queries.GetMemoTags(ctx, memoUUID)
	if err != nil {
		return nil, fmt.Errorf("list memo tags for memo %s: %w", memoUUID, err)
	}
	return rows, nil
}

func (r *SqliteMemoRepository) getMemoTagsBatch(ctx context.Context, memoUUIDs []string) (map[string][]string, error) {
	result := make(map[string][]string, len(memoUUIDs))
	if len(memoUUIDs) == 0 {
		return result, nil
	}

	for _, uid := range memoUUIDs {
		result[uid] = []string{}
	}

	rows, err := r.queries.GetMemoTagsBatch(ctx, memoUUIDs)
	if err != nil {
		return nil, fmt.Errorf("list memo tags batch: %w", err)
	}
	for _, row := range rows {
		result[row.MemoUuid] = append(result[row.MemoUuid], row.TagName)
	}
	return result, nil
}

func (r *SqliteMemoRepository) getMemoResourcesBatch(
	ctx context.Context,
	memoUUIDs []string,
) (map[string][]*entity.Resource, error) {
	result := make(map[string][]*entity.Resource, len(memoUUIDs))
	if len(memoUUIDs) == 0 {
		return result, nil
	}
	for _, uid := range memoUUIDs {
		result[uid] = make([]*entity.Resource, 0)
	}

	// Fix method name from GetMemoResourcesBatch to ListMemoResourcesBatch
	nullUUIDs := make([]sql.NullString, len(memoUUIDs))
	for i, uid := range memoUUIDs {
		nullUUIDs[i] = sql.NullString{String: uid, Valid: true}
	}
	rows, err := r.queries.ListMemoResourcesBatch(ctx, nullUUIDs)
	if err != nil {
		return nil, fmt.Errorf("list memo resources batch: %w", err)
	}

	for _, row := range rows {
		res := &entity.Resource{
			ID:           row.ID,
			MemoUUID:     fromSlTextNull(row.MemoUuid),
			FileName:     row.Filename,
			Hash:         row.Hash,
			Size:         row.Size,
			MimeType:     row.MimeType,
			InternalPath: row.InternalPath,
			CreatedAt:    row.CreatedAt,
		}
		result[row.MemoUuid.String] = append(result[row.MemoUuid.String], res)
	}
	return result, nil
}

func applyCreateMemoRow(memo *entity.Memo, row sldb.CreateMemoRow) {
	memo.UUID = row.MemoUuid
	memo.ExpiresAt = fromSlTime(row.ExpiresAt)
	memo.CreatedAt = row.CreatedAt
	memo.UpdatedAt = row.UpdatedAt
}

func formatSearchText(si port.SearchIndex) string {
	return strings.Join([]string{si.TagsTokens, si.FilesTokens, si.BodyTokens}, " ")
}

func updateMemoContentAndExpiresTx(
	ctx context.Context,
	tx *sql.Tx,
	memoUUID string,
	userID int64,
	content string,
	expiresAt *time.Time,
	si port.SearchIndex,
) error {
	queries := sldb.New(tx)
	aff, err := queries.UpdateMemoContentAndExpires(ctx, sldb.UpdateMemoContentAndExpiresParams{
		Content:     content,
		ExpiresAt:   toSlTime(expiresAt),
		SearchText:  toSlTextNull(formatSearchText(si)),
		MemoUuid:    memoUUID,
		OwnerUserID: userID,
	})
	if err != nil {
		return fmt.Errorf("update memo content and expires: %w", err)
	}
	if aff == 0 {
		return fmt.Errorf("%w: memo %s", apperr.ErrNotFound, memoUUID)
	}
	return nil
}

func replaceMemoTagsTx(ctx context.Context, tx *sql.Tx, memoUUID string, tags []string) error {
	queries := sldb.New(tx)
	if err := queries.DeleteMemoTagsByMemoUUID(ctx, memoUUID); err != nil {
		return fmt.Errorf("clear memo tags: %w", err)
	}
	for _, tagName := range normalizeTags(tags) {
		if err := queries.SaveTag(ctx, tagName); err != nil {
			return fmt.Errorf("save tag %s: %w", tagName, err)
		}
		if err := queries.LinkMemoTag(ctx, sldb.LinkMemoTagParams{
			MemoUuid: memoUUID,
			TagName:  tagName,
		}); err != nil {
			return fmt.Errorf("link memo %s tag %s: %w", memoUUID, tagName, err)
		}
	}
	return nil
}

func replaceMemoResourcesTx(
	ctx context.Context,
	tx *sql.Tx,
	memoUUID string,
	userID int64,
	resourceIDs []string,
) error {
	normalizedIDs := normalizeResourceIDs(resourceIDs)
	for _, resourceID := range normalizedIDs {
		if err := ensureResourceExistsTx(ctx, tx, resourceID, userID); err != nil {
			return err
		}
	}
	queries := sldb.New(tx)
	if _, err := queries.UnlinkMemoResourcesByOwner(ctx, sldb.UnlinkMemoResourcesByOwnerParams{
		MemoUuid:    toSlTextNull(memoUUID),
		OwnerUserID: userID,
	}); err != nil {
		return fmt.Errorf("unlink memo resources: %w", err)
	}
	for _, resourceID := range normalizedIDs {
		if err := queries.LinkResourceToMemo(ctx, sldb.LinkResourceToMemoParams{
			MemoUuid:    toSlTextNull(memoUUID),
			ID:          resourceID,
			OwnerUserID: userID,
		}); err != nil {
			return fmt.Errorf("link resource %s to memo %s: %w", resourceID, memoUUID, err)
		}
	}
	return nil
}

func ensureResourceExistsTx(
	ctx context.Context,
	tx *sql.Tx,
	resourceID string,
	userID int64,
) error {
	exists, err := sldb.New(tx).ResourceExistsForOwner(ctx, sldb.ResourceExistsForOwnerParams{
		ID:          resourceID,
		OwnerUserID: userID,
	})
	if err != nil {
		return fmt.Errorf("query resource %s: %w", resourceID, err)
	}
	if exists == 0 {
		return fmt.Errorf("%w: resource %s", apperr.ErrInvalidInput, resourceID)
	}
	return nil
}

func cleanupOrphanTagsTx(ctx context.Context, tx *sql.Tx) error {
	return sldb.New(tx).CleanupOrphanTags(ctx)
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(tags))
	results := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, exists := unique[t]; !exists {
			unique[t] = struct{}{}
			results = append(results, t)
		}
	}
	sort.Strings(results)
	return results
}

func normalizeResourceIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(ids))
	results := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := unique[id]; !exists {
			unique[id] = struct{}{}
			results = append(results, id)
		}
	}
	return results
}
