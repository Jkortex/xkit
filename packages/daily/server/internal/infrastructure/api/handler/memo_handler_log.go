package handler

import (
	"strconv"
	"time"

	"daily/internal/application/port"
	"github.com/gin-gonic/gin"
	"log/slog"
)

func logMemoQuery(
	c *gin.Context,
	filter port.MemoFilter,
	resultCount int,
	duration time.Duration,
	status int,
	queryErr error,
) {
	slog.Info("memo_list_query",
		"path", c.FullPath(),
		"status", status,
		"result_count", resultCount,
		"duration_ms", duration.Milliseconds(),
		"search", valueOrEmpty(filter.Search),
		"tag", valueOrEmpty(filter.Tag),
		"tags_any_count", len(filter.TagsAny),
		"tags_all_count", len(filter.TagsAll),
		"has_resource", boolPtrToValue(filter.HasResource),
		"from", valueOrEmpty(filter.FromDate),
		"to", valueOrEmpty(filter.ToDate),
		"sort", filter.Sort,
		"limit", filter.Limit,
		"offset", filter.Offset,
		"error", errorToValue(queryErr),
	)
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func boolPtrToValue(v *bool) string {
	if v == nil {
		return ""
	}
	return strconv.FormatBool(*v)
}

func errorToValue(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
