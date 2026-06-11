package memo

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"github.com/google/uuid"
)

// MemoService integrates all memo-related logic into a deep module.
type MemoService struct {
	repo       port.MemoRepository
	resRepo    port.ResourceRepository
	processor  *MemoProcessor
	normalizer *TagNormalizer
	tokenizer  port.Tokenizer
}

func NewMemoService(
	repo port.MemoRepository,
	resRepo port.ResourceRepository,
	tagRepo port.TagRepository,
	tokenizer port.Tokenizer,
) *MemoService {
	return &MemoService{
		repo:       repo,
		resRepo:    resRepo,
		processor:  NewMemoProcessor(tagRepo, tokenizer),
		normalizer: NewTagNormalizer(tagRepo),
		tokenizer:  tokenizer,
	}
}

// Create handles the full flow of creating a memo.
func (s *MemoService) Create(ctx context.Context, userID int64, input dto.CreateMemoRequest) (*dto.MemoResponse, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: memo content cannot be empty", apperr.ErrInvalidInput)
	}
	if len(content) > 100000 {
		return nil, fmt.Errorf("%w: content exceeds maximum length", apperr.ErrInvalidInput)
	}

	resourceNames, err := s.validateResourceOwnership(ctx, userID, input.ResourceIDs)
	if err != nil {
		return nil, err
	}

	tags, cleanedContent, expiresAt, si, err := s.processor.Process(ctx, content, input.Tags, input.TimeToLive, resourceNames)
	if err != nil {
		return nil, err
	}

	memo := entity.NewMemo(cleanedContent)
	memo.Tags = tags
	memo.ExpiresAt = expiresAt

	if err := s.repo.Create(ctx, memo, userID, tags, input.ResourceIDs, si); err != nil {
		return nil, fmt.Errorf("failed to create memo: %w", err)
	}

	return s.getMemoResponse(ctx, memo.UUID, userID)
}

// Update handles updating a memo with history preservation.
func (s *MemoService) Update(
	ctx context.Context,
	userID int64,
	memoUUID string,
	input dto.UpdateMemoRequest,
) (*dto.MemoResponse, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content cannot be empty", apperr.ErrInvalidInput)
	}

	resourceNames, err := s.validateResourceOwnership(ctx, userID, input.ResourceIDs)
	if err != nil {
		return nil, err
	}

	// 1. Process first to get cleaned content and tags
	tags, cleanedContent, expiresAt, si, err := s.processor.Process(ctx, content, input.Tags, input.TimeToLive, resourceNames)
	if err != nil {
		return nil, err
	}

	// 2. Fetch current state to check for changes
	oldMemo, err := s.repo.GetByUUID(ctx, memoUUID, userID)
	if err != nil {
		return nil, err
	}
	oldResources, _ := s.resRepo.ListByMemoUUID(ctx, memoUUID, userID)
	oldResourceIDs := make([]string, 0, len(oldResources))
	for _, r := range oldResources {
		oldResourceIDs = append(oldResourceIDs, r.ID)
	}

	// 3. Compare with cleaned content and tags
	if s.isChanged(oldMemo, oldResourceIDs, cleanedContent, tags, input.ResourceIDs) {
		history := &port.MemoHistoryRecord{
			ID:          uuid.Must(uuid.NewV7()).String(),
			MemoUUID:    memoUUID,
			Content:     oldMemo.Content,
			Tags:        oldMemo.Tags,
			ResourceIDs: oldResourceIDs,
		}
		if err := s.repo.SaveHistory(ctx, history, userID); err != nil {
			slog.Warn("failed to save history", "error", err)
		} else {
			_, _ = s.repo.DeleteOldHistory(ctx, memoUUID, 20)
		}
	}

	// 4. Update
	if err := s.repo.Update(ctx, memoUUID, userID, cleanedContent, si, tags, input.ResourceIDs, expiresAt); err != nil {
		return nil, err
	}

	return s.getMemoResponse(ctx, memoUUID, userID)
}

// Delete removes a memo and cleans up orphan tags.
func (s *MemoService) Delete(ctx context.Context, userID int64, memoUUID string) error {
	if err := s.repo.Delete(ctx, memoUUID, userID); err != nil {
		return err
	}
	return s.repo.CleanupOrphanTags(ctx)
}

// Get returns a single memo by UUID.
func (s *MemoService) Get(ctx context.Context, userID int64, memoUUID string) (*dto.MemoResponse, error) {
	return s.getMemoResponse(ctx, memoUUID, userID)
}

// List returns filtered memos.
func (s *MemoService) List(ctx context.Context, userID int64, filter port.MemoFilter) ([]*dto.MemoResponse, error) {
	// 1. DSL Parsing
	if filter.Search != nil && *filter.Search != "" {
		parsed := ParseSearchDSL(*filter.Search)
		if len(parsed.Tags) > 0 {
			filter.TagsAll = append(filter.TagsAll, parsed.Tags...)
		}
		if len(parsed.TagsExclude) > 0 {
			filter.TagsExclude = append(filter.TagsExclude, parsed.TagsExclude...)
		}
		if parsed.HasResource != nil {
			filter.HasResource = parsed.HasResource
		}
		if parsed.RowStatus != nil {
			filter.RowStatus = parsed.RowStatus
		}
		if parsed.FromDate != nil {
			filter.FromDate = parsed.FromDate
		}
		if parsed.ToDate != nil {
			filter.ToDate = parsed.ToDate
		}

		if parsed.Search != "" {
			processed := s.tokenizer.Tokenize(parsed.Search)
			terms := strings.Fields(processed)
			if len(terms) > 0 {
				var queryParts []string
				for _, t := range terms {
					if strings.HasPrefix(t, "-") {
						queryParts = append(queryParts, "NOT "+strings.TrimPrefix(t, "-"))
					} else {
						queryParts = append(queryParts, t)
					}
				}
				searchQuery := strings.Join(queryParts, " AND ")
				filter.Search = &searchQuery
			}
		} else {
			filter.Search = nil
		}
	}

	// 2. Normalization
	if filter.Tag != nil && *filter.Tag != "" {
		canonical, _ := s.normalizer.NormalizeTag(ctx, *filter.Tag)
		filter.Tag = &canonical
	}
	filter.TagsAll = s.normalizer.NormalizeTags(ctx, filter.TagsAll)
	filter.TagsExclude = s.normalizer.NormalizeTags(ctx, filter.TagsExclude)

	// 3. Query
	memos, err := s.repo.List(ctx, filter, userID)
	if err != nil {
		return nil, err
	}

	results := make([]*dto.MemoResponse, 0, len(memos))
	for _, m := range memos {
		results = append(results, s.mapToDTO(m))
	}
	return results, nil
}

// TransitionTask handles task status transitions.
func (s *MemoService) TransitionTask(ctx context.Context, userID int64, memoUUID, targetStatus, agentID string) (*dto.MemoResponse, error) {
	m, err := s.repo.GetByUUID(ctx, memoUUID, userID)
	if err != nil {
		return nil, err
	}

	canTransition := false
	var newTags []string

	switch targetStatus {
	case "doing":
		if s.hasTag(m.Tags, "todo") {
			canTransition = true
			newTags = s.replaceTag(m.Tags, "todo", "doing")
			newTags = append(newTags, fmt.Sprintf("by/%s", agentID))
		}
	case "done", "failed":
		if s.hasTag(m.Tags, "doing") || s.hasTag(m.Tags, "todo") {
			canTransition = true
			newTags = s.replaceTag(m.Tags, "doing", targetStatus)
			newTags = s.replaceTag(newTags, "todo", targetStatus)
		}
	}

	if !canTransition {
		return nil, fmt.Errorf("%w: cannot transition to %s", apperr.ErrInvalidTransition, targetStatus)
	}

	si := port.SearchIndex{BodyTokens: m.Content}
	if err := s.repo.Update(ctx, memoUUID, userID, m.Content, si, newTags, nil, m.ExpiresAt); err != nil {
		return nil, err
	}

	return s.getMemoResponse(ctx, memoUUID, userID)
}

// GetRandom returns a random memo for the user.
func (s *MemoService) GetRandom(ctx context.Context, userID int64) (*dto.MemoResponse, error) {
	m, err := s.repo.GetRandom(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.mapToDTO(m), nil
}

// GetStats returns usage statistics.
func (s *MemoService) GetStats(ctx context.Context, userID int64) (*dto.StatsResponse, error) {
	m, t, r, err := s.repo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.StatsResponse{
		MemosTotal:     m,
		TagsTotal:      t,
		ResourcesTotal: r,
	}, nil
}

// BatchArchive archives multiple memos in a single operation.
func (s *MemoService) BatchArchive(ctx context.Context, userID int64, uuids []string) (*dto.BatchResult, error) {
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: uuids cannot be empty", apperr.ErrInvalidInput)
	}
	if len(uuids) > 100 {
		return nil, fmt.Errorf("%w: max 100 uuids per batch", apperr.ErrInvalidInput)
	}

	// Perform batch archive
	archived, err := s.repo.BatchArchive(ctx, userID, uuids)
	if err != nil {
		return nil, fmt.Errorf("batch archive: %w", err)
	}

	succeededMap := make(map[string]bool)
	for _, u := range archived {
		succeededMap[u] = true
	}

	failedItems := make([]dto.FailedItem, 0)
	for _, u := range uuids {
		if !succeededMap[u] {
			failedItems = append(failedItems, dto.FailedItem{UUID: u, Reason: "not_found_or_forbidden"})
		}
	}

	return &dto.BatchResult{
		Succeeded: archived,
		Failed:    failedItems,
	}, nil
}

// BatchDelete deletes multiple memos in a single operation.
func (s *MemoService) BatchDelete(ctx context.Context, userID int64, uuids []string) (*dto.BatchResult, error) {
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: uuids cannot be empty", apperr.ErrInvalidInput)
	}
	if len(uuids) > 100 {
		return nil, fmt.Errorf("%w: max 100 uuids per batch", apperr.ErrInvalidInput)
	}

	// Perform batch delete
	deleted, err := s.repo.BatchDelete(ctx, userID, uuids)
	if err != nil {
		return nil, fmt.Errorf("batch delete: %w", err)
	}

	succeededMap := make(map[string]bool)
	for _, u := range deleted {
		succeededMap[u] = true
	}

	failedItems := make([]dto.FailedItem, 0)
	for _, u := range uuids {
		if !succeededMap[u] {
			failedItems = append(failedItems, dto.FailedItem{UUID: u, Reason: "not_found_or_forbidden"})
		}
	}

	return &dto.BatchResult{
		Succeeded: deleted,
		Failed:    failedItems,
	}, nil
}

// BatchTag adds or removes tags from multiple memos in a single operation.
func (s *MemoService) BatchTag(ctx context.Context, userID int64, uuids []string, addTags []string, removeTags []string) (*dto.BatchResult, error) {
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: uuids cannot be empty", apperr.ErrInvalidInput)
	}
	if len(uuids) > 100 {
		return nil, fmt.Errorf("%w: max 100 uuids per batch", apperr.ErrInvalidInput)
	}

	// Perform batch tag operation
	validUUIDs, err := s.repo.BatchTag(ctx, userID, uuids, addTags, removeTags)
	if err != nil {
		return nil, fmt.Errorf("batch tag: %w", err)
	}

	succeededMap := make(map[string]bool)
	for _, u := range validUUIDs {
		succeededMap[u] = true
	}

	failedItems := make([]dto.FailedItem, 0)
	for _, u := range uuids {
		if !succeededMap[u] {
			failedItems = append(failedItems, dto.FailedItem{UUID: u, Reason: "not_found_or_forbidden"})
		}
	}

	return &dto.BatchResult{
		Succeeded: validUUIDs,
		Failed:    failedItems,
	}, nil
}

// ListHistory returns version history for a memo.
func (s *MemoService) ListHistory(ctx context.Context, userID int64, memoUUID string) ([]*dto.MemoHistoryResponse, error) {
	m, err := s.repo.GetByUUID(ctx, memoUUID, userID)
	if err != nil {
		return nil, err
	}
	records, err := s.repo.ListHistory(ctx, m.UUID, userID)
	if err != nil {
		return nil, err
	}
	results := make([]*dto.MemoHistoryResponse, 0, len(records))
	for _, rec := range records {
		createdAt, _ := time.Parse(time.RFC3339, rec.CreatedAt)
		results = append(results, &dto.MemoHistoryResponse{
			ID:          rec.ID,
			MemoUUID:    rec.MemoUUID,
			Content:     rec.Content,
			Tags:        rec.Tags,
			ResourceIDs: rec.ResourceIDs,
			CreatedAt:   createdAt,
		})
	}
	return results, nil
}

// Rollback restores a memo to a previous version.
func (s *MemoService) Rollback(ctx context.Context, userID int64, memoUUID, historyID string) (*dto.MemoResponse, error) {
	history, err := s.repo.GetHistoryByID(ctx, historyID, userID)
	if err != nil {
		return nil, err
	}

	current, err := s.repo.GetByUUID(ctx, memoUUID, userID)
	if err != nil {
		return nil, err
	}
	currentResources, _ := s.resRepo.ListByMemoUUID(ctx, memoUUID, userID)
	currentResourceIDs := make([]string, 0, len(currentResources))
	for _, r := range currentResources {
		currentResourceIDs = append(currentResourceIDs, r.ID)
	}

	// Backup current state
	backup := &port.MemoHistoryRecord{
		ID:          uuid.Must(uuid.NewV7()).String(),
		MemoUUID:    memoUUID,
		Content:     current.Content,
		Tags:        current.Tags,
		ResourceIDs: currentResourceIDs,
	}
	_ = s.repo.SaveHistory(ctx, backup, userID)

	resourceNames, _ := s.validateResourceOwnership(ctx, userID, history.ResourceIDs)
	tags, cleanedContent, _, si, err := s.processor.Process(ctx, history.Content, nil, "", resourceNames)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, memoUUID, userID, cleanedContent, si, tags, history.ResourceIDs, current.ExpiresAt); err != nil {
		return nil, err
	}

	return s.getMemoResponse(ctx, memoUUID, userID)
}

// Helpers

func (s *MemoService) validateResourceOwnership(ctx context.Context, userID int64, resourceIDs []string) ([]string, error) {
	names := make([]string, 0, len(resourceIDs))
	for _, rid := range resourceIDs {
		res, err := s.resRepo.GetByID(ctx, rid, userID)
		if err != nil {
			return nil, fmt.Errorf("%w: resource %s not found", apperr.ErrNotFound, rid)
		}
		names = append(names, res.FileName)
	}
	return names, nil
}

func (s *MemoService) getMemoResponse(ctx context.Context, uuid string, userID int64) (*dto.MemoResponse, error) {
	m, err := s.repo.GetByUUID(ctx, uuid, userID)
	if err != nil {
		return nil, err
	}
	return s.mapToDTO(m), nil
}

func (s *MemoService) mapToDTO(m *entity.Memo) *dto.MemoResponse {
	resources := make([]*dto.ResourceResponse, 0, len(m.Resources))
	for _, res := range m.Resources {
		resources = append(resources, &dto.ResourceResponse{
			ID:        res.ID,
			FileName:  res.FileName,
			Size:      res.Size,
			MimeType:  res.MimeType,
			CreatedAt: res.CreatedAt,
		})
	}
	return &dto.MemoResponse{
		UUID:      m.UUID,
		Content:   m.Content,
		RowStatus: string(m.RowStatus),
		Tags:      m.Tags,
		Resources: resources,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (s *MemoService) isChanged(
	oldMemo *entity.Memo,
	oldResIDs []string,
	newContent string,
	newTags []string,
	newResIDs []string,
) bool {
	if oldMemo.Content != newContent {
		return true
	}

	// Compare tags
	if len(oldMemo.Tags) != len(newTags) {
		return true
	}
	tagMap := make(map[string]struct{})
	for _, t := range oldMemo.Tags {
		tagMap[t] = struct{}{}
	}
	for _, t := range newTags {
		if _, exists := tagMap[t]; !exists {
			return true
		}
	}

	// Compare resources
	if len(oldResIDs) != len(newResIDs) {
		return true
	}
	oldMap := make(map[string]struct{})
	for _, id := range oldResIDs {
		oldMap[id] = struct{}{}
	}
	for _, id := range newResIDs {
		if _, exists := oldMap[id]; !exists {
			return true
		}
	}
	return false
}

func (s *MemoService) hasTag(tags []string, target string) bool {
	for _, t := range tags {
		if strings.TrimPrefix(t, "#") == target {
			return true
		}
	}
	return false
}

func (s *MemoService) replaceTag(tags []string, oldTag, newTag string) []string {
	var result []string
	for _, t := range tags {
		if strings.TrimPrefix(t, "#") == oldTag {
			result = append(result, newTag)
		} else {
			result = append(result, t)
		}
	}
	return result
}
