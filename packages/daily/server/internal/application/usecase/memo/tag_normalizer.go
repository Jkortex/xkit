package memo

import (
	"context"
	"daily/internal/application/port"
)

type TagNormalizer struct {
	tagRepo port.TagRepository
}

func NewTagNormalizer(tagRepo port.TagRepository) *TagNormalizer {
	return &TagNormalizer{tagRepo: tagRepo}
}

func (n *TagNormalizer) NormalizeTag(ctx context.Context, tag string) (string, error) {
	canonical, err := n.tagRepo.ResolveCanonicalTag(ctx, tag)
	if err != nil {
		return tag, err
	}
	return canonical, nil
}

func (n *TagNormalizer) NormalizeTags(ctx context.Context, tags []string) []string {
	normalized := make([]string, len(tags))
	for i, tag := range tags {
		canonical, err := n.NormalizeTag(ctx, tag)
		if err == nil {
			normalized[i] = canonical
		} else {
			normalized[i] = tag
		}
	}
	return normalized
}
