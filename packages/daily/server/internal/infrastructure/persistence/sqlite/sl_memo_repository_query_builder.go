package sqlite

import (
	"strings"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"github.com/Masterminds/squirrel"
)

func buildListQuery(filter port.MemoFilter, userID int64) (string, []any) {
	// SQLite uses ? as placeholder
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	query := sb.Select("m.memo_uuid", "m.content", "m.row_status", "m.expires_at", "m.created_at", "m.updated_at").
		From("memo m")

	query = query.Where(squirrel.Eq{"m.owner_user_id": userID})

	rowStatus := string(entity.RowStatusNormal)
	if filter.RowStatus != nil {
		rowStatus = string(*filter.RowStatus)
	}
	query = query.Where(squirrel.Eq{"m.row_status": rowStatus})

	// Search (SQLite FTS5)
	if filter.Search != nil && *filter.Search != "" {
		// SQLite FTS5 MATCH syntax
		// UseCase passes "term1 AND term2 AND NOT term3"
		// SQLite FTS5 understands AND, OR, NOT (case sensitive)
		searchTerm := *filter.Search
		query = query.Where("m.memo_uuid IN (SELECT memo_uuid FROM memo_fts WHERE content MATCH ?)", searchTerm)
	}

	// Single Tag
	if filter.Tag != nil && *filter.Tag != "" {
		query = query.Where("EXISTS (SELECT 1 FROM memo_tag mt WHERE mt.memo_uuid = m.memo_uuid AND mt.tag_name = ?)", *filter.Tag)
	}

	// Tags Any
	if len(filter.TagsAny) > 0 {
		placeholders := make([]string, len(filter.TagsAny))
		args := make([]any, len(filter.TagsAny))
		for i, tag := range filter.TagsAny {
			placeholders[i] = "?"
			args[i] = tag
		}
		query = query.Where("EXISTS (SELECT 1 FROM memo_tag mt_any WHERE mt_any.memo_uuid = m.memo_uuid AND mt_any.tag_name IN ("+strings.Join(placeholders, ",")+"))", args...)
	}

	// Tags All
	for _, tag := range filter.TagsAll {
		query = query.Where("EXISTS (SELECT 1 FROM memo_tag mt_all WHERE mt_all.memo_uuid = m.memo_uuid AND mt_all.tag_name = ?)", tag)
	}

	// Tags Exclude
	for _, tag := range filter.TagsExclude {
		query = query.Where("NOT EXISTS (SELECT 1 FROM memo_tag mt_exc WHERE mt_exc.memo_uuid = m.memo_uuid AND mt_exc.tag_name = ?)", tag)
	}

	// Has Resource
	if filter.HasResource != nil {
		if *filter.HasResource {
			query = query.Where("EXISTS (SELECT 1 FROM resource r WHERE r.memo_uuid = m.memo_uuid)")
		} else {
			query = query.Where("NOT EXISTS (SELECT 1 FROM resource r WHERE r.memo_uuid = m.memo_uuid)")
		}
	}

	// Date Filters
	if filter.FromDate != nil {
		query = query.Where("m.created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		// SQLite date arithmetic: date(created_at) < date(?, '+1 day')
		query = query.Where("date(m.created_at) <= date(?)", *filter.ToDate)
	}

	// Order By
	sortBy := strings.TrimSpace(filter.Sort)
	if sortBy == "" {
		sortBy = "created_at_desc"
	}

	switch sortBy {
	case "created_at_asc":
		query = query.OrderBy("m.created_at ASC")
	case "updated_at_desc":
		query = query.OrderBy("m.updated_at DESC")
	default:
		// SQLite FTS5 rank is not directly comparable to ts_rank.
		// For now, keep it simple.
		query = query.OrderBy("m.created_at DESC")
	}

	// Limit & Offset
	limit := uint64(20)
	if filter.Limit > 0 {
		limit = uint64(filter.Limit)
	}
	query = query.Limit(limit).Offset(uint64(filter.Offset))

	sql, args, err := query.ToSql()
	if err != nil {
		return "", nil
	}

	return sql, args
}
