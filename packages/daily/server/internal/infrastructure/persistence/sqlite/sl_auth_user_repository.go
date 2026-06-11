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

type SqliteUserRepository struct {
	queries *sldb.Queries
	db      *sql.DB
}

func NewSqliteUserRepository(db *sql.DB) *SqliteUserRepository {
	return &SqliteUserRepository{
		queries: sldb.New(db),
		db:      db,
	}
}

func (r *SqliteUserRepository) Create(ctx context.Context, user *entity.User) error {
	row, err := r.queries.CreateAuthUser(ctx, sldb.CreateAuthUserParams{
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
		Status:       string(user.Status),
	})
	if err != nil {
		return fmt.Errorf("db create user: %w", err)
	}
	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *SqliteUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	row, err := r.queries.GetAuthUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: user not found", apperr.ErrNotFound)
		}
		return nil, err
	}
	return r.mapRowToEntity(row), nil
}

func (r *SqliteUserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	row, err := r.queries.GetAuthUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: user not found", apperr.ErrNotFound)
		}
		return nil, err
	}
	return r.mapRowToEntity(row), nil
}

func (r *SqliteUserRepository) DeleteByID(ctx context.Context, id int64) error {
	aff, err := r.queries.DeleteAuthUserByID(ctx, id)
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("%w: user not found", apperr.ErrNotFound)
	}
	return nil
}

func (r *SqliteUserRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountAuthUsers(ctx)
}

func (r *SqliteUserRepository) mapRowToEntity(row sldb.AuthUser) *entity.User {
	return &entity.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Role:         entity.UserRole(row.Role),
		Status:       entity.UserStatus(row.Status),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
