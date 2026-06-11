package memo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"daily/internal/domain/service"
)

// MemoProcessor 封装了创建和更新笔记时的共有业务逻辑
type MemoProcessor struct {
	tagRepo      port.TagRepository
	tokenizer    port.Tokenizer
	tagExtractor *service.TagExtractor
}

func NewMemoProcessor(tagRepo port.TagRepository, tokenizer port.Tokenizer) *MemoProcessor {
	return &MemoProcessor{
		tagRepo:      tagRepo,
		tokenizer:    tokenizer,
		tagExtractor: service.NewTagExtractor(),
	}
}

// Process 处理内容提取标签、计算过期时间以及生成搜索索引
func (p *MemoProcessor) Process(
	ctx context.Context,
	content string,
	explicitTags []string,
	ttl string,
	resourceNames []string,
) ([]string, string, *time.Time, port.SearchIndex, error) {
	extractedTags, cleanedContent, err := p.tagExtractor.Extract(content)
	if err != nil {
		return nil, "", nil, port.SearchIndex{}, err
	}

	// Merge extracted tags with explicit tags
	allInputTags := append(extractedTags, explicitTags...)

	canonicalTags, err := p.resolveCanonicalTags(ctx, allInputTags)
	if err != nil {
		return nil, "", nil, port.SearchIndex{}, err
	}

	expiresAt, err := p.calculateExpiration(canonicalTags, ttl)
	if err != nil {
		return nil, "", nil, port.SearchIndex{}, err
	}

	idx := port.SearchIndex{IsEphemeral: expiresAt != nil}
	if !idx.IsEphemeral {
		idx.TagsTokens = p.tokenizer.Tokenize(strings.Join(canonicalTags, " "))
		idx.FilesTokens = p.tokenizer.Tokenize(strings.Join(resourceNames, " "))
		idx.BodyTokens = p.tokenizer.Tokenize(cleanedContent)
	}

	return canonicalTags, cleanedContent, expiresAt, idx, nil
}

func (p *MemoProcessor) resolveCanonicalTags(ctx context.Context, tags []string) ([]string, error) {
	results := make([]string, 0, len(tags))
	uniq := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		canonical, err := p.tagRepo.ResolveCanonicalTag(ctx, tag)
		if err != nil {
			return nil, err
		}
		if _, exists := uniq[canonical]; exists {
			continue
		}
		uniq[canonical] = struct{}{}
		results = append(results, canonical)
	}
	return results, nil
}

func (p *MemoProcessor) calculateExpiration(tags []string, ttl string) (*time.Time, error) {
	hasTempTag := false
	for _, t := range tags {
		if t == entity.EphemeralTag {
			hasTempTag = true
			break
		}
	}

	if ttl == "" && hasTempTag {
		ttl = entity.DefaultEphemeralTTL
	}

	if ttl == "" {
		return nil, nil
	}

	duration, err := parseDuration(ttl)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ttl format: %s", apperr.ErrInvalidInput, ttl)
	}

	expiresAt := time.Now().Add(duration)
	return &expiresAt, nil
}

func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
