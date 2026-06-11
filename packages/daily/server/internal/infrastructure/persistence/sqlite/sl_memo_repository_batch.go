package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

	"daily/internal/application/apperr"
	sldb "daily/internal/infrastructure/persistence/sqlite/db"
)

// BatchArchive implements port.MemoRepository.
func (r *SqliteMemoRepository) BatchArchive(ctx context.Context, userID int64, uuids []string) ([]string, error) {
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: uuids cannot be empty", apperr.ErrInvalidInput)
	}

	return r.queries.BatchArchive(ctx, sldb.BatchArchiveParams{
		Uuids:       uuids,
		OwnerUserID: userID,
	})
}

// BatchDelete implements port.MemoRepository.
func (r *SqliteMemoRepository) BatchDelete(ctx context.Context, userID int64, uuids []string) ([]string, error) {
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: uuids cannot be empty", apperr.ErrInvalidInput)
	}

	return r.queries.BatchDelete(ctx, sldb.BatchDeleteParams{
		Uuids:       uuids,
		OwnerUserID: userID,
	})
}

// BatchTag implements port.MemoRepository.
func (r *SqliteMemoRepository) BatchTag(ctx context.Context, userID int64, uuids []string, addTags []string, removeTags []string) ([]string, error) {
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: uuids cannot be empty", apperr.ErrInvalidInput)
	}

	// Use transaction to ensure consistency
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	qtx := sldb.New(tx)

	validUuids, err := qtx.BatchPrecheck(ctx, sldb.BatchPrecheckParams{
		Uuids:       uuids,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("precheck: %w", err)
	}

	if len(validUuids) == 0 {
		tx.Rollback()
		return nil, nil
	}

	if len(addTags) > 0 {
		tagsJSON, _ := json.Marshal(addTags)
		if err := qtx.BatchSaveTags(ctx, string(tagsJSON)); err != nil {
			return nil, fmt.Errorf("batch save tags: %w", err)
		}

		uuidsJSON, _ := json.Marshal(validUuids)
		if err := qtx.BatchTagAdd(ctx, sldb.BatchTagAddParams{
			Uuids: string(uuidsJSON),
			Tags:  string(tagsJSON),
		}); err != nil {
			return nil, fmt.Errorf("batch tag add: %w", err)
		}
	}

	if len(removeTags) > 0 {
		if err := qtx.BatchTagRemove(ctx, sldb.BatchTagRemoveParams{
			Uuids: validUuids,
			Tags:  removeTags,
		}); err != nil {
			return nil, fmt.Errorf("batch tag remove: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return validUuids, nil
}
