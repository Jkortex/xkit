package dto

import (
	"time"
)

// CreateMemoRequest 用户创建笔记的请求负载
type CreateMemoRequest struct {
	Content     string   `json:"content" binding:"required"`
	Tags        []string `json:"tags"`
	ResourceIDs []string `json:"resource_ids"`
	TimeToLive  string   `json:"ttl"` // 可选: 3d, 1h 等
}

type UpdateMemoRequest struct {
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
	ResourceIDs []string `json:"resource_ids"`
	TimeToLive  string   `json:"ttl"`
}

// MemoResponse 统一的笔记响应对象，防止 Domain Entity 泄漏
type MemoResponse struct {
	UUID      string              `json:"uuid"`
	Content   string              `json:"content"`
	RowStatus string              `json:"row_status"`
	Tags      []string            `json:"tags"`
	Resources []*ResourceResponse `json:"resources,omitempty"`
	ExpiresAt *time.Time          `json:"expires_at,omitempty"`
	Headline  string              `json:"headline,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// ResourceResponse 统一的资源响应对象
type ResourceResponse struct {
	ID        string    `json:"id"`
	FileName  string    `json:"filename"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

type TagStatResponse struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type RenameTagResponse struct {
	From          string `json:"from"`
	To            string `json:"to"`
	AffectedMemos int64  `json:"affected_memos"`
	Merged        bool   `json:"merged"`
}

type MergeTagsResponse struct {
	Sources        []string `json:"sources"`
	Target         string   `json:"target"`
	AffectedMemos  int64    `json:"affected_memos"`
	MergedSources  int      `json:"merged_sources"`
	SkippedSources []string `json:"skipped_sources"`
}

type TagAliasResponse struct {
	Alias     string `json:"alias"`
	Canonical string `json:"canonical"`
}

type TagAuditResponse struct {
	Action        string    `json:"action"`
	Summary       string    `json:"summary"`
	AffectedMemos int64     `json:"affected_memos"`
	CreatedAt     time.Time `json:"created_at"`
}

type MemoHistoryResponse struct {
	ID          string    `json:"id"`
	MemoUUID    string    `json:"memo_uuid"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	ResourceIDs []string  `json:"resource_ids"`
	CreatedAt   time.Time `json:"created_at"`
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Succeeded []string     `json:"succeeded"`
	Failed    []FailedItem `json:"failed"`
}

// FailedItem represents a failed item in a batch operation
type FailedItem struct {
	UUID   string `json:"uuid"`
	Reason string `json:"reason"` // e.g., "not_found", "forbidden"
}

func FromMemoEntity(m any) *MemoResponse {
	// We use any and then cast to avoid circular dependency if possible,
	// but here we are in application/dto, and domain/entity is fine to import.
	// Actually, let's just use the concrete type.
	return nil // placeholder
}
