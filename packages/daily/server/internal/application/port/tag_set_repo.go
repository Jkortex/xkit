package port

import (
	"context"
	"daily/internal/domain/entity"
)

type TagSetGroupRepository interface {
	ListByUser(ctx context.Context, userID int64) ([]*entity.TagSetGroup, error)
	GetByID(ctx context.Context, userID int64, id string) (*entity.TagSetGroup, error)
	Create(ctx context.Context, g *entity.TagSetGroup) error
	Update(ctx context.Context, g *entity.TagSetGroup) error
	Delete(ctx context.Context, userID int64, id string) error
}

type TagSetRepository interface {
	ListByUser(ctx context.Context, userID int64, groupID *string) ([]*entity.TagSet, error)
	GetByID(ctx context.Context, userID int64, id string) (*entity.TagSet, error)
	Create(ctx context.Context, ts *entity.TagSet) error
	Update(ctx context.Context, ts *entity.TagSet) error
	Delete(ctx context.Context, userID int64, id string) error
	TouchLastUsed(ctx context.Context, id string, userID int64) error
}
