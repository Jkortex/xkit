package memo

import (
	"daily/internal/domain/entity"
	"regexp"
	"strings"
)

var (
	tagIncludePattern = regexp.MustCompile(`\btag:([^\s]+)`)
	tagExcludePattern = regexp.MustCompile(`-tag:([^\s]+)`)
	hasAttachPattern  = regexp.MustCompile(`\bhas:attachment\b`)
	isArchivedPattern = regexp.MustCompile(`\bis:archived\b`)
	isNormalPattern   = regexp.MustCompile(`\bis:normal\b`)
	afterPattern      = regexp.MustCompile(`\b(after|from):(\d{4}-\d{2}-\d{2})`)
	beforePattern     = regexp.MustCompile(`\b(before|to):(\d{4}-\d{2}-\d{2})`)
)

// ParsedQuery 解析后的查询对象
type ParsedQuery struct {
	Tags        []string
	TagsExclude []string
	HasResource *bool
	RowStatus   *entity.RowStatus
	FromDate    *string
	ToDate      *string
	Search      string
}

// ParseSearchDSL 解析增强型查询语法
func ParseSearchDSL(input string) ParsedQuery {
	res := ParsedQuery{
		Tags:        make([]string, 0),
		TagsExclude: make([]string, 0),
	}

	// 1. 提取标签排除 -tag:xxx
	matches := tagExcludePattern.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		res.TagsExclude = append(res.TagsExclude, match[1])
		input = strings.Replace(input, match[0], "", 1)
	}

	// 2. 提取标签包含 tag:xxx
	matches = tagIncludePattern.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		res.Tags = append(res.Tags, match[1])
		input = strings.Replace(input, match[0], "", 1)
	}

	// 3. 提取 has:attachment
	if hasAttachPattern.MatchString(input) {
		val := true
		res.HasResource = &val
		input = hasAttachPattern.ReplaceAllString(input, "")
	}

	// 4. 提取 is:archived / is:normal
	if isArchivedPattern.MatchString(input) {
		status := entity.RowStatusArchived
		res.RowStatus = &status
		input = isArchivedPattern.ReplaceAllString(input, "")
	} else if isNormalPattern.MatchString(input) {
		status := entity.RowStatusNormal
		res.RowStatus = &status
		input = isNormalPattern.ReplaceAllString(input, "")
	}

	// 5. 提取 after:YYYY-MM-DD / from:YYYY-MM-DD
	if match := afterPattern.FindStringSubmatch(input); len(match) > 2 {
		res.FromDate = &match[2]
		input = strings.Replace(input, match[0], "", 1)
	}

	// 6. 提取 before:YYYY-MM-DD / to:YYYY-MM-DD
	if match := beforePattern.FindStringSubmatch(input); len(match) > 2 {
		res.ToDate = &match[2]
		input = strings.Replace(input, match[0], "", 1)
	}

	// 7. 剩余部分作为全文检索词 (保留引号和减号，由 FTS 处理)
	res.Search = strings.TrimSpace(input)

	return res
}
