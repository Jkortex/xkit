package port

import (
	"context"
	"daily/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	DeleteByID(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
}
