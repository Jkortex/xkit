package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	sldb "daily/internal/infrastructure/persistence/sqlite/db"
	"github.com/google/uuid"
)

func (r *SqliteMemoRepository) Create(
	ctx context.Context,
	memo *entity.Memo,
	userID int64,
	tags,
	resourceIDs []string,
	si port.SearchIndex,
) error {
	if strings.TrimSpace(memo.UUID) == "" {
		memo.UUID = uuid.Must(uuid.NewV7()).String()
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create memo graph tx: %w", err)
	}
	defer tx.Rollback()

	queries := sldb.New(tx)
	row, err := queries.CreateMemo(ctx, sldb.CreateMemoParams{
		MemoUuid:    memo.UUID,
		OwnerUserID: userID,
		Content:     memo.Content,
		RowStatus:   string(memo.RowStatus),
		ExpiresAt:   toSlTime(memo.ExpiresAt),
		SearchText:  toSlTextNull(formatSearchText(si)),
	})
	if err != nil {
		return fmt.Errorf("insert memo in tx: %w", err)
	}

	if err := replaceMemoTagsTx(ctx, tx, memo.UUID, tags); err != nil {
		return err
	}
	if err := replaceMemoResourcesTx(ctx, tx, memo.UUID, userID, resourceIDs); err != nil {
		return err
	}
	if err := cleanupOrphanTagsTx(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create memo graph tx: %w", err)
	}

	applyCreateMemoRow(memo, row)
	return nil
}

func (r *SqliteMemoRepository) GetByUUID(ctx context.Context, uuid string, userID int64) (*entity.Memo, error) {
	row, err := r.queries.GetMemoByUUID(ctx, sldb.GetMemoByUUIDParams{
		MemoUuid:    uuid,
		OwnerUserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: memo %s", apperr.ErrNotFound, uuid)
		}
		return nil, fmt.Errorf("db get memo: %w", err)
	}

	tags, err := r.getMemoTags(ctx, row.MemoUuid)
	if err != nil {
		return nil, err
	}

	return &entity.Memo{
		UUID:      row.MemoUuid,
		Content:   row.Content,
		RowStatus: entity.RowStatus(row.RowStatus),
		Tags:      tags,
		ExpiresAt: fromSlTime(row.ExpiresAt),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *SqliteMemoRepository) List(ctx context.Context, filter port.MemoFilter, userID int64) ([]*entity.Memo, error) {
	query, args := buildListQuery(filter, userID)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db list memos: %w", err)
	}
	defer rows.Close()

	memos := make([]*entity.Memo, 0)
	uuids := make([]string, 0)
	for rows.Next() {
		m := &entity.Memo{}
		var expiresAt sql.NullTime
		if err := rows.Scan(
			&m.UUID,
			&m.Content,
			&m.RowStatus,
			&expiresAt,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan memo row: %w", err)
		}
		m.ExpiresAt = fromSlTime(expiresAt)
		uuids = append(uuids, m.UUID)
		memos = append(memos, m)
	}

	if err := r.attachTagsAndResources(ctx, memos, uuids, filter.IncludeResources); err != nil {
		return nil, err
	}

	return memos, nil
}

func (r *SqliteMemoRepository) attachTagsAndResources(ctx context.Context, memos []*entity.Memo, uuids []string, includeResources bool) error {
	tagsByMemoUUID, err := r.getMemoTagsBatch(ctx, uuids)
	if err != nil {
		return err
	}

	for _, memo := range memos {
		memo.Tags = tagsByMemoUUID[memo.UUID]
	}
	if includeResources {
		resourcesByMemoUUID, err := r.getMemoResourcesBatch(ctx, uuids)
		if err != nil {
			return err
		}
		for _, memo := range memos {
			memo.Resources = resourcesByMemoUUID[memo.UUID]
		}
	}
	return nil
}

func (r *SqliteMemoRepository) ListAll(ctx context.Context, userID int64) ([]*entity.Memo, error) {
	return r.List(ctx, port.MemoFilter{}, userID)
}

func (r *SqliteMemoRepository) GetRandom(ctx context.Context, userID int64) (*entity.Memo, error) {
	row, err := r.queries.GetRandomMemo(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: no memos found", apperr.ErrNotFound)
		}
		return nil, fmt.Errorf("db get random memo: %w", err)
	}

	tags, err := r.getMemoTags(ctx, row.MemoUuid)
	if err != nil {
		return nil, err
	}

	return &entity.Memo{
		UUID:      row.MemoUuid,
		Content:   row.Content,
		RowStatus: entity.RowStatus(row.RowStatus),
		Tags:      tags,
		ExpiresAt: fromSlTime(row.ExpiresAt),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *SqliteMemoRepository) Update(
	ctx context.Context,
	uuid string,
	userID int64,
	content string,
	si port.SearchIndex,
	tags []string,
	resourceIDs []string,
	expiresAt *time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update memo full tx: %w", err)
	}
	defer tx.Rollback()

	if err := updateMemoContentAndExpiresTx(ctx, tx, uuid, userID, content, expiresAt, si); err != nil {
		return err
	}
	if err := replaceMemoTagsTx(ctx, tx, uuid, tags); err != nil {
		return err
	}
	if err := replaceMemoResourcesTx(ctx, tx, uuid, userID, resourceIDs); err != nil {
		return err
	}
	if err := cleanupOrphanTagsTx(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SqliteMemoRepository) Delete(ctx context.Context, memoUUID string, userID int64) error {
	aff, err := r.queries.DeleteMemoByUUID(ctx, sldb.DeleteMemoByUUIDParams{
		MemoUuid:    memoUUID,
		OwnerUserID: userID,
	})
	if err != nil {
		return fmt.Errorf("db delete memo: %w", err)
	}
	if aff == 0 {
		return fmt.Errorf("%w: memo %s", apperr.ErrNotFound, memoUUID)
	}
	return nil
}

func (r *SqliteMemoRepository) ArchiveExpired(ctx context.Context) (int64, error) {
	return r.queries.ArchiveExpiredMemosBefore(ctx, sldb.ArchiveExpiredMemosBeforeParams{
		UpdatedAt: time.Now(),
		ExpiresAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
}
