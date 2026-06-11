package port

import (
	"context"
	"daily/internal/domain/entity"
	"time"
)

// MemoFilter 用于列表查询的过滤条件
type MemoFilter struct {
	RowStatus        *entity.RowStatus
	Tag              *string
	Search           *string // 全文检索关键词
	FromDate         *string
	ToDate           *string
	HasResource      *bool
	IncludeResources bool
	TagsAny          []string
	TagsAll          []string
	TagsExclude      []string
	Sort             string
	Limit            int
	Offset           int
}

// SearchIndex 封装写入搜索索引所需的 token 文本
type SearchIndex struct {
	TagsTokens  string // 标签分词，weight A
	FilesTokens string // 文件名分词，weight B
	BodyTokens  string // 正文分词，weight D
	IsEphemeral bool   // true 时 search_vector 置 NULL
}

// MemoRepository 定义了聚合的笔记持久化契约 (组合模式)
type MemoRepository interface {
	MemoReadRepository
	MemoWriteRepository
	TagRepository
	MemoHistoryRepository
	StatsRepository
}

type MemoReadRepository interface {
	GetByUUID(ctx context.Context, uuid string, userID int64) (*entity.Memo, error)
	List(ctx context.Context, filter MemoFilter, userID int64) ([]*entity.Memo, error)
	ListAll(ctx context.Context, userID int64) ([]*entity.Memo, error)
	GetRandom(ctx context.Context, userID int64) (*entity.Memo, error)
}

type MemoWriteRepository interface {
	Create(
		ctx context.Context,
		memo *entity.Memo,
		userID int64,
		tags []string,
		resourceIDs []string,
		si SearchIndex,
	) error
	Update(
		ctx context.Context,
		uuid string,
		userID int64,
		content string,
		si SearchIndex,
		tags []string,
		resourceIDs []string,
		expiresAt *time.Time,
	) error
	Delete(ctx context.Context, uuid string, userID int64) error
	ArchiveExpired(ctx context.Context) (int64, error) // 自动归档过期笔记

	BatchArchive(ctx context.Context, userID int64, uuids []string) ([]string, error)
	BatchDelete(ctx context.Context, userID int64, uuids []string) ([]string, error)
	BatchTag(ctx context.Context, userID int64, uuids []string, addTags []string, removeTags []string) ([]string, error)
}

// ResourceRepository 定义了附件持久化的契约
type ResourceRepository interface {
	Save(ctx context.Context, res *entity.Resource, userID int64) error
	GetByID(ctx context.Context, id string, userID int64) (*entity.Resource, error)
	ListByMemoUUID(ctx context.Context, memoUUID string, userID int64) ([]*entity.Resource, error)
	LinkToMemo(ctx context.Context, resourceID string, memoUUID string, userID int64) error
	UnlinkByMemoUUID(ctx context.Context, memoUUID string, userID int64) error
	ListAll(ctx context.Context, userID int64) ([]*entity.Resource, error)
	ListTrackedPaths(ctx context.Context) ([]string, error)
	CleanupOrphanResources(ctx context.Context) (int64, error)
}
