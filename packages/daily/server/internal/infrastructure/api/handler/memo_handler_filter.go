package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"daily/internal/application/port"
	"github.com/gin-gonic/gin"
)

var nowFunc = time.Now
var sinceFunc = time.Since

func buildMemoFilter(c *gin.Context) (port.MemoFilter, error) {
	search := strings.TrimSpace(c.Query("search"))
	tag := strings.TrimSpace(c.Query("tag"))

	limit, err := parseLimitOffset(c.DefaultQuery("limit", "20"), 20, 1, 100)
	if err != nil {
		return port.MemoFilter{}, err
	}
	offset, err := parseLimitOffset(c.DefaultQuery("offset", "0"), 0, 0, 1000000)
	if err != nil {
		return port.MemoFilter{}, err
	}
	fromDate, err := parseOptionalDate(c.Query("from"))
	if err != nil {
		return port.MemoFilter{}, err
	}
	toDate, err := parseOptionalDate(c.Query("to"))
	if err != nil {
		return port.MemoFilter{}, err
	}
	if fromDate != nil && toDate != nil && *fromDate > *toDate {
		return port.MemoFilter{}, fmt.Errorf("from date must be <= to date")
	}
	hasResource, err := parseOptionalBool(c.Query("has_resource"))
	if err != nil {
		return port.MemoFilter{}, err
	}
	includeResources, err := parseOptionalBool(c.Query("include_resources"))
	if err != nil {
		return port.MemoFilter{}, err
	}
	sortBy := c.Query("sort")
	if !isValidSort(sortBy) {
		return port.MemoFilter{}, fmt.Errorf("invalid sort: %s", sortBy)
	}

	filter := port.MemoFilter{
		Limit:            limit,
		Offset:           offset,
		FromDate:         fromDate,
		ToDate:           toDate,
		HasResource:      hasResource,
		IncludeResources: includeResources != nil && *includeResources,
		TagsAny:          parseCSV(c.Query("tags_any")),
		TagsAll:          parseCSV(c.Query("tags_all")),
		TagsExclude:      parseCSV(c.Query("tags_exclude")),
		Sort:             sortBy,
	}
	if tag != "" {
		filter.Tag = &tag
	}
	if search != "" {
		filter.Search = &search
	}
	return filter, nil
}

func parseLimitOffset(raw string, defaultValue, minValue, maxValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric parameter: %s", raw)
	}
	if value < minValue || value > maxValue {
		return 0, fmt.Errorf("parameter out of range: %d", value)
	}
	return value, nil
}

func parseOptionalDate(raw string) (*string, error) {
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return nil, fmt.Errorf("invalid date format: %s", raw)
	}
	return &raw, nil
}

func parseOptionalBool(raw string) (*bool, error) {
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid boolean parameter: %s", raw)
	}
	return &v, nil
}

func parseCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	uniq := make(map[string]struct{}, len(parts))
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, exists := uniq[tag]; exists {
			continue
		}
		uniq[tag] = struct{}{}
		results = append(results, tag)
	}
	return results
}

func isValidSort(sortBy string) bool {
	switch sortBy {
	case "", "created_at_desc", "created_at_asc", "updated_at_desc":
		return true
	default:
		return false
	}
}
