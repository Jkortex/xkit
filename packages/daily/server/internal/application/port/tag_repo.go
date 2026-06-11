package port

import (
	"context"
	"daily/internal/domain/entity"
)

type TagStat struct {
	Name  string
	Count int
}

type TagRenameResult struct {
	AffectedMemos int64
	Merged        bool
}

type TagMergeResult struct {
	AffectedMemos  int64
	MergedSources  int
	SkippedSources []string
}

type TagAlias struct {
	AliasName     string
	CanonicalName string
}

type TagAuditRecord struct {
	Action        string
	Summary       string
	AffectedMemos int64
	CreatedAt     string
}

type TagRepository interface {
	// 基本操作
	ReplaceMemoTags(ctx context.Context, memoUUID string, tags []string) error
	SaveTag(ctx context.Context, tagName string) error
	LinkMemoTag(ctx context.Context, memoUUID string, tagName string) error
	CleanupOrphanTags(ctx context.Context) error

	// 统计操作
	ListTagsWithCount(ctx context.Context, userID int64) ([]entity.TagStat, error)

	// 管理操作
	RenameTag(ctx context.Context, userID int64, from, to string) (*TagRenameResult, error)
	MergeTags(ctx context.Context, userID int64, sources []string, target string) (*TagMergeResult, error)
	SaveTagAlias(ctx context.Context, alias, canonical string) error
	DeleteTagAlias(ctx context.Context, alias string) error
	ListTagAliases(ctx context.Context) ([]TagAlias, error)
	ResolveCanonicalTag(ctx context.Context, tag string) (string, error)
	AppendTagAudit(
		ctx context.Context,
		action,
		summary string,
		affectedMemos int64,
	) error
	ListTagAudits(
		ctx context.Context,
		limit int,
		action string,
	) ([]TagAuditRecord, error)
}
