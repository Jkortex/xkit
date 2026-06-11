package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

	"daily/internal/application/port"
	sldb "daily/internal/infrastructure/persistence/sqlite/db"
)

func (r *SqliteMemoRepository) SaveHistory(ctx context.Context, record *port.MemoHistoryRecord, userID int64) error {
	tagsJSON, err := json.Marshal(record.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags for history: %w", err)
	}
	resourceIDsJSON, err := json.Marshal(record.ResourceIDs)
	if err != nil {
		return fmt.Errorf("marshal resource ids for history: %w", err)
	}

	return r.queries.CreateMemoHistory(ctx, sldb.CreateMemoHistoryParams{
		ID:          record.ID,
		MemoUuid:    record.MemoUUID,
		OwnerUserID: userID,
		Content:     record.Content,
		Tags:        string(tagsJSON),
		ResourceIds: string(resourceIDsJSON),
	})
}

func (r *SqliteMemoRepository) ListHistory(ctx context.Context, memoUUID string, userID int64) ([]*port.MemoHistoryRecord, error) {
	rows, err := r.queries.ListMemoHistory(ctx, sldb.ListMemoHistoryParams{
		MemoUuid:    memoUUID,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, err
	}

	results := make([]*port.MemoHistoryRecord, 0, len(rows))
	for _, row := range rows {
		var tags []string
		_ = json.Unmarshal([]byte(row.Tags), &tags)
		var resourceIDs []string
		_ = json.Unmarshal([]byte(row.ResourceIds), &resourceIDs)

		results = append(results, &port.MemoHistoryRecord{
			ID:          row.ID,
			MemoUUID:    row.MemoUuid,
			Content:     row.Content,
			Tags:        tags,
			ResourceIDs: resourceIDs,
			CreatedAt:   row.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return results, nil
}

func (r *SqliteMemoRepository) GetHistoryByID(ctx context.Context, id string, userID int64) (*port.MemoHistoryRecord, error) {
	row, err := r.queries.GetMemoHistoryByID(ctx, sldb.GetMemoHistoryByIDParams{
		ID:          id,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, err
	}

	var tags []string
	_ = json.Unmarshal([]byte(row.Tags), &tags)
	var resourceIDs []string
	_ = json.Unmarshal([]byte(row.ResourceIds), &resourceIDs)

	return &port.MemoHistoryRecord{
		ID:          row.ID,
		MemoUUID:    row.MemoUuid,
		Content:     row.Content,
		Tags:        tags,
		ResourceIDs: resourceIDs,
		CreatedAt:   row.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (r *SqliteMemoRepository) DeleteOldHistory(ctx context.Context, memoUUID string, keepLimit int) (int64, error) {
	return r.queries.DeleteOldMemoHistory(ctx, sldb.DeleteOldMemoHistoryParams{
		MemoUuid:   memoUUID,
		MemoUuid_2: memoUUID,
		Limit:      int64(keepLimit),
	})
}
