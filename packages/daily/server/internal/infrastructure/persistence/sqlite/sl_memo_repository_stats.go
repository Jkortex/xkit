package sqlite

import (
	"context"
	"daily/internal/domain/entity"
)

func (r *SqliteMemoRepository) GetStats(ctx context.Context, userID int64) (memos, tags, resources int64, err error) {
	memos, _ = r.queries.CountMemos(ctx, userID)
	tags, _ = r.queries.CountTags(ctx, userID)
	resources, _ = r.queries.CountResources(ctx, userID)
	return memos, tags, resources, nil
}

func (r *SqliteMemoRepository) GetDailyHeatmap(ctx context.Context, userID int64) ([]entity.DailyStat, error) {
	rows, err := r.queries.GetDailyHeatmap(ctx, userID)
	if err != nil {
		return nil, err
	}
	results := make([]entity.DailyStat, 0, len(rows))
	for _, row := range rows {
		dateStr := ""
		if s, ok := row.Date.(string); ok {
			dateStr = s
		} else if b, ok := row.Date.([]byte); ok {
			dateStr = string(b)
		}

		results = append(results, entity.DailyStat{
			Date:  dateStr,
			Count: int(row.Count),
		})
	}
	return results, nil
}
