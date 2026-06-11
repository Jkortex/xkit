package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"daily/internal/application/apperr"
	"daily/internal/domain/entity"
	sldb "daily/internal/infrastructure/persistence/sqlite/db"
)

type SqliteResourceRepository struct {
	queries *sldb.Queries
	db      *sql.DB
}

func NewSqliteResourceRepository(db *sql.DB) *SqliteResourceRepository {
	return &SqliteResourceRepository{
		queries: sldb.New(db),
		db:      db,
	}
}

func (r *SqliteResourceRepository) Save(ctx context.Context, res *entity.Resource, userID int64) error {
	return r.queries.CreateResource(ctx, sldb.CreateResourceParams{
		ID:           res.ID,
		MemoUuid:     toSlTextNull(res.MemoUUID),
		OwnerUserID:  userID,
		Filename:     res.FileName,
		Hash:         res.Hash,
		Size:         res.Size,
		MimeType:     res.MimeType,
		InternalPath: res.InternalPath,
	})
}

func (r *SqliteResourceRepository) GetByID(ctx context.Context, id string, userID int64) (*entity.Resource, error) {
	row, err := r.queries.GetResourceByID(ctx, sldb.GetResourceByIDParams{
		ID:          id,
		OwnerUserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: resource %s", apperr.ErrNotFound, id)
		}
		return nil, err
	}
	return &entity.Resource{
		ID:           row.ID,
		MemoUUID:     fromSlTextNull(row.MemoUuid),
		FileName:     row.Filename,
		Hash:         row.Hash,
		Size:         row.Size,
		MimeType:     row.MimeType,
		InternalPath: row.InternalPath,
		CreatedAt:    row.CreatedAt,
	}, nil
}

func (r *SqliteResourceRepository) ListByMemoUUID(ctx context.Context, memoUUID string, userID int64) ([]*entity.Resource, error) {
	rows, err := r.queries.ListResourcesByMemoUUID(ctx, sldb.ListResourcesByMemoUUIDParams{
		MemoUuid:    toSlTextNull(memoUUID),
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, err
	}
	results := make([]*entity.Resource, 0, len(rows))
	for _, row := range rows {
		results = append(results, &entity.Resource{
			ID:           row.ID,
			MemoUUID:     fromSlTextNull(row.MemoUuid),
			FileName:     row.Filename,
			Hash:         row.Hash,
			Size:         row.Size,
			MimeType:     row.MimeType,
			InternalPath: row.InternalPath,
			CreatedAt:    row.CreatedAt,
		})
	}
	return results, nil
}

func (r *SqliteResourceRepository) LinkToMemo(ctx context.Context, resourceID string, memoUUID string, userID int64) error {
	return r.queries.LinkResourceToMemo(ctx, sldb.LinkResourceToMemoParams{
		ID:          resourceID,
		MemoUuid:    toSlTextNull(memoUUID),
		OwnerUserID: userID,
	})
}

func (r *SqliteResourceRepository) UnlinkByMemoUUID(ctx context.Context, memoUUID string, userID int64) error {
	_, err := r.queries.UnlinkMemoResourcesByOwner(ctx, sldb.UnlinkMemoResourcesByOwnerParams{
		MemoUuid:    toSlTextNull(memoUUID),
		OwnerUserID: userID,
	})
	return err
}

func (r *SqliteResourceRepository) ListAll(ctx context.Context, userID int64) ([]*entity.Resource, error) {
	rows, err := r.queries.ListAllResources(ctx, userID)
	if err != nil {
		return nil, err
	}
	results := make([]*entity.Resource, 0, len(rows))
	for _, row := range rows {
		results = append(results, &entity.Resource{
			ID:           row.ID,
			MemoUUID:     fromSlTextNull(row.MemoUuid),
			FileName:     row.Filename,
			Hash:         row.Hash,
			Size:         row.Size,
			MimeType:     row.MimeType,
			InternalPath: row.InternalPath,
			CreatedAt:    row.CreatedAt,
		})
	}
	return results, nil
}

func (r *SqliteResourceRepository) ListTrackedPaths(ctx context.Context) ([]string, error) {
	return r.queries.ListTrackedPaths(ctx)
}

func (r *SqliteResourceRepository) CleanupOrphanResources(ctx context.Context) (int64, error) {
	return r.queries.CleanupOrphanResources(ctx)
}
