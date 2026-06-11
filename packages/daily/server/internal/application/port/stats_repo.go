package port

import (
	"context"
	"daily/internal/domain/entity"
)

type StatsRepository interface {
	GetStats(ctx context.Context, userID int64) (memos, tags, resources int64, err error)
	GetDailyHeatmap(ctx context.Context, userID int64) ([]entity.DailyStat, error)
}
