package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/domain/entity"
)

type SqliteTagSetGroupRepository struct {
	db *sql.DB
}

func NewSqliteTagSetGroupRepository(db *sql.DB) *SqliteTagSetGroupRepository {
	return &SqliteTagSetGroupRepository{db: db}
}

func (r *SqliteTagSetGroupRepository) ListByUser(ctx context.Context, userID int64) ([]*entity.TagSetGroup, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, weight, created_at, updated_at
		FROM tag_set_group
		WHERE user_id = ?
		ORDER BY weight DESC, created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list tag set groups: %w", err)
	}
	defer rows.Close()
	return scanSlGroups(rows)
}

func (r *SqliteTagSetGroupRepository) GetByID(ctx context.Context, userID int64, id string) (*entity.TagSetGroup, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, weight, created_at, updated_at
		FROM tag_set_group
		WHERE id = ? AND user_id = ?
	`, id, userID)
	g, err := scanSlGroupRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: tag set group not found", apperr.ErrNotFound)
		}
		return nil, fmt.Errorf("get tag set group: %w", err)
	}
	return g, nil
}

func (r *SqliteTagSetGroupRepository) Create(ctx context.Context, g *entity.TagSetGroup) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tag_set_group (id, user_id, name, weight, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, g.ID, g.UserID, g.Name, g.Weight, g.CreatedAt, g.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create tag set group: %w", err)
	}
	return nil
}

func (r *SqliteTagSetGroupRepository) Update(ctx context.Context, g *entity.TagSetGroup) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tag_set_group SET name = ?, weight = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, g.Name, g.Weight, g.UpdatedAt, g.ID, g.UserID)
	if err != nil {
		return fmt.Errorf("update tag set group: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("%w: tag set group not found", apperr.ErrNotFound)
	}
	return nil
}

func (r *SqliteTagSetGroupRepository) Delete(ctx context.Context, userID int64, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tag_set_group WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete tag set group: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("%w: tag set group not found", apperr.ErrNotFound)
	}
	return nil
}

func scanSlGroupRow(row interface{ Scan(dest ...any) error }) (*entity.TagSetGroup, error) {
	var g entity.TagSetGroup
	if err := row.Scan(&g.ID, &g.UserID, &g.Name, &g.Weight, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return nil, err
	}
	return &g, nil
}

func scanSlGroups(rows *sql.Rows) ([]*entity.TagSetGroup, error) {
	var results []*entity.TagSetGroup
	for rows.Next() {
		g, err := scanSlGroupRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, g)
	}
	return results, rows.Err()
}

// --- TagSet ---

type SqliteTagSetRepository struct {
	db *sql.DB
}

func NewSqliteTagSetRepository(db *sql.DB) *SqliteTagSetRepository {
	return &SqliteTagSetRepository{db: db}
}

func (r *SqliteTagSetRepository) ListByUser(ctx context.Context, userID int64, groupID *string) ([]*entity.TagSet, error) {
	var rows *sql.Rows
	var err error
	if groupID != nil {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, user_id, group_id, name, tags_any, tags_all, tags_exclude,
			       weight, last_used_at, created_at, updated_at
			FROM tag_set
			WHERE user_id = ? AND group_id = ?
			ORDER BY weight DESC, created_at ASC
		`, userID, *groupID)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, user_id, group_id, name, tags_any, tags_all, tags_exclude,
			       weight, last_used_at, created_at, updated_at
			FROM tag_set
			WHERE user_id = ?
			ORDER BY weight DESC, created_at ASC
		`, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("list tag sets: %w", err)
	}
	defer rows.Close()
	return scanSlTagSets(rows)
}

func (r *SqliteTagSetRepository) GetByID(ctx context.Context, userID int64, id string) (*entity.TagSet, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, group_id, name, tags_any, tags_all, tags_exclude,
		       weight, last_used_at, created_at, updated_at
		FROM tag_set
		WHERE id = ? AND user_id = ?
	`, id, userID)
	ts, err := scanSlTagSetRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: tag set not found", apperr.ErrNotFound)
		}
		return nil, fmt.Errorf("get tag set: %w", err)
	}
	return ts, nil
}

func (r *SqliteTagSetRepository) Create(ctx context.Context, ts *entity.TagSet) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tag_set (id, user_id, group_id, name, tags_any, tags_all, tags_exclude,
		                     weight, last_used_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ts.ID, ts.UserID, ts.GroupID, ts.Name, ts.TagsAny, ts.TagsAll, ts.TagsExclude,
		ts.Weight, ts.LastUsedAt, ts.CreatedAt, ts.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create tag set: %w", err)
	}
	return nil
}

func (r *SqliteTagSetRepository) Update(ctx context.Context, ts *entity.TagSet) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tag_set
		SET group_id = ?, name = ?, tags_any = ?, tags_all = ?,
		    tags_exclude = ?, weight = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, ts.GroupID, ts.Name, ts.TagsAny, ts.TagsAll, ts.TagsExclude,
		ts.Weight, ts.UpdatedAt, ts.ID, ts.UserID)
	if err != nil {
		return fmt.Errorf("update tag set: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("%w: tag set not found", apperr.ErrNotFound)
	}
	return nil
}

func (r *SqliteTagSetRepository) Delete(ctx context.Context, userID int64, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tag_set WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete tag set: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("%w: tag set not found", apperr.ErrNotFound)
	}
	return nil
}

func (r *SqliteTagSetRepository) TouchLastUsed(ctx context.Context, id string, userID int64) error {
	now := time.Now().UTC()
	weight := now.Unix()
	res, err := r.db.ExecContext(ctx, `
		UPDATE tag_set SET last_used_at = ?, weight = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, now, weight, now, id, userID)
	if err != nil {
		return fmt.Errorf("touch tag set: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("%w: tag set not found", apperr.ErrNotFound)
	}
	return nil
}

func scanSlTagSetRow(row interface{ Scan(dest ...any) error }) (*entity.TagSet, error) {
	var ts entity.TagSet
	if err := row.Scan(&ts.ID, &ts.UserID, &ts.GroupID, &ts.Name,
		&ts.TagsAny, &ts.TagsAll, &ts.TagsExclude,
		&ts.Weight, &ts.LastUsedAt, &ts.CreatedAt, &ts.UpdatedAt); err != nil {
		return nil, err
	}
	return &ts, nil
}

func scanSlTagSets(rows *sql.Rows) ([]*entity.TagSet, error) {
	var results []*entity.TagSet
	for rows.Next() {
		ts, err := scanSlTagSetRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, ts)
	}
	return results, rows.Err()
}
