package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
	sldb "daily/internal/infrastructure/persistence/sqlite/db"
)

func (r *SqliteMemoRepository) ReplaceMemoTags(ctx context.Context, memoUUID string, tags []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace tags tx: %w", err)
	}
	defer tx.Rollback()

	queries := sldb.New(tx)

	// 1. Delete existing
	if err := queries.DeleteMemoTagsByMemoUUID(ctx, memoUUID); err != nil {
		return fmt.Errorf("delete existing tags: %w", err)
	}

	// 2. Insert new
	for _, tag := range tags {
		if strings.TrimSpace(tag) == "" {
			continue
		}
		// Ensure tag exists in tag table
		if err := queries.SaveTag(ctx, tag); err != nil {
			return fmt.Errorf("ensure tag %s: %w", tag, err)
		}
		// Link to memo
		if err := queries.LinkMemoTag(ctx, sldb.LinkMemoTagParams{
			MemoUuid: memoUUID,
			TagName:  tag,
		}); err != nil {
			return fmt.Errorf("link tag %s to memo: %w", tag, err)
		}
	}

	return tx.Commit()
}

func (r *SqliteMemoRepository) ListTagsWithCount(ctx context.Context, userID int64) ([]entity.TagStat, error) {
	rows, err := r.queries.ListTagsWithCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("db list tags with count: %w", err)
	}

	results := make([]entity.TagStat, 0, len(rows))
	for _, row := range rows {
		results = append(results, entity.TagStat{
			Name:  row.Name,
			Count: int(row.Count),
		})
	}
	return results, nil
}

func (r *SqliteMemoRepository) RenameTag(ctx context.Context, userID int64, from, to string) (*port.TagRenameResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin rename tag tx: %w", err)
	}
	defer tx.Rollback()

	queries := sldb.New(tx)

	// Check if target tag exists
	if err := queries.SaveTag(ctx, to); err != nil {
		return nil, fmt.Errorf("ensure target tag %s: %w", to, err)
	}

	// 1. Delete links to target tag for memos that also have the from tag
	// This prevents UNIQUE constraint failed: memo_tag.memo_uuid, memo_tag.tag_name
	_, err = queries.DeleteDuplicateMemoTagLinksByOwner(ctx, sldb.DeleteDuplicateMemoTagLinksByOwnerParams{
		TagName:     to,
		TagName_2:   from,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("pre-cleanup for rename: %w", err)
	}

	// 2. Move remaining links
	affected, err := queries.MoveMemoTagLinksByOwner(ctx, sldb.MoveMemoTagLinksByOwnerParams{
		TagName:     to,
		TagName_2:   from,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("move memo tags from %s to %s: %w", from, to, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit rename tag tx: %w", err)
	}

	return &port.TagRenameResult{
		AffectedMemos: affected,
		Merged:        true,
	}, nil
}

func (r *SqliteMemoRepository) MergeTags(ctx context.Context, userID int64, sources []string, target string) (*port.TagMergeResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin merge tags tx: %w", err)
	}
	defer tx.Rollback()

	queries := sldb.New(tx)

	// Ensure target exists
	if err := queries.SaveTag(ctx, target); err != nil {
		return nil, fmt.Errorf("ensure target tag %s: %w", target, err)
	}

	totalAffected := int64(0)
	mergedCount := 0
	var skipped []string

	for _, source := range sources {
		if strings.EqualFold(source, target) {
			skipped = append(skipped, source)
			continue
		}

		// Delete duplicates first to avoid primary key conflict on (memo_uuid, tag_name)
		_, err := queries.DeleteDuplicateMemoTagLinksByOwner(ctx, sldb.DeleteDuplicateMemoTagLinksByOwnerParams{
			TagName:     target,
			TagName_2:   source,
			OwnerUserID: userID,
		})
		if err != nil {
			skipped = append(skipped, source)
			continue
		}

		// Move remaining links
		aff, err := queries.MoveMemoTagLinksByOwner(ctx, sldb.MoveMemoTagLinksByOwnerParams{
			TagName:     target,
			TagName_2:   source,
			OwnerUserID: userID,
		})
		if err != nil {
			skipped = append(skipped, source)
			continue
		}

		totalAffected += aff
		mergedCount++
	}

	if err := queries.CleanupOrphanTags(ctx); err != nil {
		return nil, fmt.Errorf("cleanup after merge: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit merge tags tx: %w", err)
	}

	return &port.TagMergeResult{
		AffectedMemos:  totalAffected,
		MergedSources:  mergedCount,
		SkippedSources: skipped,
	}, nil
}

func (r *SqliteMemoRepository) ResolveCanonicalTag(ctx context.Context, tag string) (string, error) {
	canonical, err := r.queries.GetCanonicalTagAlias(ctx, tag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tag, nil
		}
		return "", err
	}
	return canonical, nil
}

func (r *SqliteMemoRepository) SaveTagAlias(ctx context.Context, alias, canonical string) error {
	return r.queries.UpsertTagAlias(ctx, sldb.UpsertTagAliasParams{
		AliasName:     alias,
		CanonicalName: canonical,
	})
}

func (r *SqliteMemoRepository) ListTagAliases(ctx context.Context) ([]port.TagAlias, error) {
	rows, err := r.queries.ListTagAliases(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]port.TagAlias, 0, len(rows))
	for _, row := range rows {
		results = append(results, port.TagAlias{
			AliasName:     row.AliasName,
			CanonicalName: row.CanonicalName,
		})
	}
	return results, nil
}

func (r *SqliteMemoRepository) DeleteTagAlias(ctx context.Context, alias string) error {
	_, err := r.queries.DeleteTagAliasByName(ctx, alias)
	return err
}

func (r *SqliteMemoRepository) AppendTagAudit(ctx context.Context, action, summary string, affectedMemos int64) error {
	return r.queries.AppendTagGovernanceAudit(ctx, sldb.AppendTagGovernanceAuditParams{
		Action:        action,
		Summary:       summary,
		AffectedMemos: affectedMemos,
	})
}

func (r *SqliteMemoRepository) ListTagAudits(ctx context.Context, limit int, action string) ([]port.TagAuditRecord, error) {
	rows, err := r.queries.ListTagAudits(ctx, sldb.ListTagAuditsParams{
		Column1: action,
		Action:  action,
		Limit:   int64(limit),
	})
	if err != nil {
		return nil, err
	}
	results := make([]port.TagAuditRecord, 0, len(rows))
	for _, row := range rows {
		results = append(results, port.TagAuditRecord{
			Action:        row.Action,
			Summary:       row.Summary,
			AffectedMemos: row.AffectedMemos,
			CreatedAt:     row.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}
	return results, nil
}

func (r *SqliteMemoRepository) SaveTag(ctx context.Context, tagName string) error {
	return r.queries.SaveTag(ctx, tagName)
}

func (r *SqliteMemoRepository) LinkMemoTag(ctx context.Context, memoUUID string, tagName string) error {
	return r.queries.LinkMemoTag(ctx, sldb.LinkMemoTagParams{
		MemoUuid: memoUUID,
		TagName:  tagName,
	})
}

func (r *SqliteMemoRepository) CleanupOrphanTags(ctx context.Context) error {
	return r.queries.CleanupOrphanTags(ctx)
}
