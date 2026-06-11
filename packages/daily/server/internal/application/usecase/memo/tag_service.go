package memo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/dto"
	"daily/internal/application/port"
)

// TagService integrates all tag-related logic into a deep module.
type TagService struct {
	repo port.TagRepository
}

func NewTagService(repo port.TagRepository) *TagService {
	return &TagService{repo: repo}
}

// ListTags returns tag statistics for a user.
func (s *TagService) ListTags(ctx context.Context, userID int64) ([]dto.TagStatResponse, error) {
	stats, err := s.repo.ListTagsWithCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	results := make([]dto.TagStatResponse, 0, len(stats))
	for _, st := range stats {
		results = append(results, dto.TagStatResponse{
			Name:  st.Name,
			Count: st.Count,
		})
	}
	return results, nil
}

// RenameTag renames a tag and audits the change.
func (s *TagService) RenameTag(ctx context.Context, userID int64, from, to string) (*dto.RenameTagResponse, error) {
	if from == "" || to == "" {
		return nil, fmt.Errorf("%w: from/to cannot be empty", apperr.ErrInvalidInput)
	}
	if from == to {
		return nil, fmt.Errorf("%w: from and to cannot be same", apperr.ErrInvalidInput)
	}

	result, err := s.repo.RenameTag(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	_ = s.repo.AppendTagAudit(ctx, "rename", fmt.Sprintf("%s -> %s", from, to), result.AffectedMemos)

	return &dto.RenameTagResponse{
		From:          from,
		To:            to,
		AffectedMemos: result.AffectedMemos,
		Merged:        result.Merged,
	}, nil
}

// MergeTags merges multiple source tags into a target tag.
func (s *TagService) MergeTags(ctx context.Context, userID int64, sources []string, target string) (*dto.MergeTagsResponse, error) {
	if target == "" {
		return nil, fmt.Errorf("%w: target cannot be empty", apperr.ErrInvalidInput)
	}

	cleanSources := s.normalizeTagNames(sources)
	if len(cleanSources) == 0 {
		return nil, fmt.Errorf("%w: sources cannot be empty", apperr.ErrInvalidInput)
	}

	result, err := s.repo.MergeTags(ctx, userID, cleanSources, target)
	if err != nil {
		return nil, err
	}

	_ = s.repo.AppendTagAudit(ctx, "merge", fmt.Sprintf("%s -> %s", strings.Join(cleanSources, ","), target), result.AffectedMemos)

	return &dto.MergeTagsResponse{
		Sources:        cleanSources,
		Target:         target,
		AffectedMemos:  result.AffectedMemos,
		MergedSources:  result.MergedSources,
		SkippedSources: result.SkippedSources,
	}, nil
}

// UpsertTagAlias creates or updates a tag alias.
func (s *TagService) UpsertTagAlias(ctx context.Context, userID int64, alias, canonical string) (*dto.TagAliasResponse, error) {
	if alias == "" || canonical == "" {
		return nil, fmt.Errorf("%w: alias/canonical cannot be empty", apperr.ErrInvalidInput)
	}

	resolvedCanonical, err := s.repo.ResolveCanonicalTag(ctx, canonical)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(alias, resolvedCanonical) {
		return nil, fmt.Errorf("%w: alias and canonical cannot be same", apperr.ErrInvalidInput)
	}

	affectedMemos := int64(0)
	if renameResult, err := s.repo.RenameTag(ctx, userID, alias, resolvedCanonical); err == nil {
		affectedMemos = renameResult.AffectedMemos
	} else if !errors.Is(err, apperr.ErrNotFound) {
		return nil, err
	}

	if err := s.repo.SaveTagAlias(ctx, alias, resolvedCanonical); err != nil {
		return nil, err
	}

	_ = s.repo.AppendTagAudit(ctx, "alias_upsert", fmt.Sprintf("%s => %s", alias, resolvedCanonical), affectedMemos)

	return &dto.TagAliasResponse{
		Alias:     alias,
		Canonical: resolvedCanonical,
	}, nil
}

// ListTagAliases returns all defined tag aliases.
func (s *TagService) ListTagAliases(ctx context.Context) ([]dto.TagAliasResponse, error) {
	rows, err := s.repo.ListTagAliases(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]dto.TagAliasResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.TagAliasResponse{
			Alias:     row.AliasName,
			Canonical: row.CanonicalName,
		})
	}
	return results, nil
}

// DeleteTagAlias removes a tag alias.
func (s *TagService) DeleteTagAlias(ctx context.Context, alias string) error {
	if alias == "" {
		return fmt.Errorf("%w: alias cannot be empty", apperr.ErrInvalidInput)
	}
	if err := s.repo.DeleteTagAlias(ctx, alias); err != nil {
		return err
	}
	return s.repo.AppendTagAudit(ctx, "alias_delete", alias, 0)
}

// ListTagAudits returns tag management audit logs.
func (s *TagService) ListTagAudits(ctx context.Context, limit int, action string) ([]dto.TagAuditResponse, error) {
	rows, err := s.repo.ListTagAudits(ctx, limit, action)
	if err != nil {
		return nil, err
	}

	results := make([]dto.TagAuditResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.TagAuditResponse{
			Action:        row.Action,
			Summary:       row.Summary,
			AffectedMemos: row.AffectedMemos,
			CreatedAt:     s.parseTagAuditTime(row.CreatedAt),
		})
	}
	return results, nil
}

func (s *TagService) normalizeTagNames(tags []string) []string {
	uniq := make(map[string]struct{}, len(tags))
	results := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.TrimSpace(raw)
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

func (s *TagService) parseTagAuditTime(raw string) time.Time {
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}
