package memo

import (
	"context"
	"testing"
	"time"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

// MockTokenizer 简单模拟分词
type MockTokenizer struct{}

func (t *MockTokenizer) Tokenize(text string) string { return text }
func (t *MockTokenizer) Reload() error               { return nil }

// MockMemoRepository 实现 port.MemoRepository 接口
type MockMemoRepository struct {
	port.MemoRepository
	CreatedMemo       *entity.Memo
	LinkedResourceIDs []string
}

func (m *MockMemoRepository) Create(
	ctx context.Context,
	memo *entity.Memo,
	userID int64,
	tags,
	resourceIDs []string,
	si port.SearchIndex,
) error {
	m.CreatedMemo = memo
	memo.Tags = tags
	m.LinkedResourceIDs = append([]string{}, resourceIDs...)
	return nil
}

func (m *MockMemoRepository) GetByUUID(ctx context.Context, uuid string, userID int64) (*entity.Memo, error) {
	if m.CreatedMemo != nil && m.CreatedMemo.UUID == uuid {
		return m.CreatedMemo, nil
	}
	return &entity.Memo{UUID: uuid}, nil
}

func (m *MockMemoRepository) CleanupOrphanTags(ctx context.Context) error { return nil }

func (m *MockMemoRepository) SaveTag(ctx context.Context, tag string) error { return nil }
func (m *MockMemoRepository) LinkMemoTag(ctx context.Context, memoUUID string, tag string) error {
	return nil
}
func (m *MockMemoRepository) ReplaceMemoTags(ctx context.Context, memoUUID string, tags []string) error {
	return nil
}
func (m *MockMemoRepository) ResolveCanonicalTag(ctx context.Context, tag string) (string, error) {
	return tag, nil
}
func (m *MockMemoRepository) RenameTag(ctx context.Context, userID int64, from, to string) (*port.TagRenameResult, error) {
	return nil, nil
}
func (m *MockMemoRepository) MergeTags(ctx context.Context, userID int64, sources []string, target string) (*port.TagMergeResult, error) {
	return nil, nil
}
func (m *MockMemoRepository) SaveTagAlias(ctx context.Context, alias, canonical string) error {
	return nil
}
func (m *MockMemoRepository) DeleteTagAlias(ctx context.Context, alias string) error {
	return nil
}
func (m *MockMemoRepository) ListTagAliases(ctx context.Context) ([]port.TagAlias, error) {
	return nil, nil
}
func (m *MockMemoRepository) AppendTagAudit(ctx context.Context, action, summary string, affectedMemos int64) error {
	return nil
}
func (m *MockMemoRepository) ListTagAudits(
	ctx context.Context,
	limit int,
	action string,
) ([]port.TagAuditRecord, error) {
	return nil, nil
}

// MockResourceRepository 实现 port.ResourceRepository 接口
type MockResourceRepository struct {
	port.ResourceRepository
	LinkedResourceIDs []string
	MemoRepo          *MockMemoRepository
}

func (m *MockResourceRepository) LinkToMemo(ctx context.Context, resID string, memoUUID string, userID int64) error {
	m.LinkedResourceIDs = append(m.LinkedResourceIDs, resID)
	return nil
}

func (m *MockResourceRepository) ListByMemoUUID(ctx context.Context, memoUUID string, userID int64) ([]*entity.Resource, error) {
	linked := m.LinkedResourceIDs
	if m.MemoRepo != nil {
		linked = m.MemoRepo.LinkedResourceIDs
	}
	results := make([]*entity.Resource, 0, len(linked))
	for _, id := range linked {
		results = append(results, &entity.Resource{
			ID:        id,
			FileName:  id + ".png",
			Size:      1,
			MimeType:  "image/png",
			CreatedAt: time.Now(),
		})
	}
	return results, nil
}

func (m *MockResourceRepository) GetByID(ctx context.Context, id string, userID int64) (*entity.Resource, error) {
	return &entity.Resource{ID: id, FileName: id + ".png"}, nil
}

func (m *MockResourceRepository) UnlinkByMemoUUID(ctx context.Context, memoUUID string, userID int64) error {
	return nil
}

func TestCreateMemoUseCase_Execute(t *testing.T) {
	// 1. 初始化
	memoRepo := &MockMemoRepository{}
	resRepo := &MockResourceRepository{MemoRepo: memoRepo}
	svc := NewMemoService(memoRepo, resRepo, memoRepo, &MockTokenizer{})

	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		input := dto.CreateMemoRequest{
			Content:     "Hello!\n#Daily",
			ResourceIDs: []string{"res-1"},
		}

		res, err := svc.Create(ctx, 1, input)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// 2. 断言响应 DTO 内容
		expectedContent := "Hello!"
		if res.Content != expectedContent {
			t.Errorf("Expected content %s, got %s", expectedContent, res.Content)
		}

		// 3. 断言标签提取 (通过响应检查)
		if len(res.Tags) != 1 || res.Tags[0] != "Daily" {
			t.Errorf("Tags extraction failed: %v", res.Tags)
		}

		// 4. 断言资源关联 (通过 Mock 检查)
		if len(memoRepo.LinkedResourceIDs) != 1 || memoRepo.LinkedResourceIDs[0] != "res-1" {
			t.Errorf("Resource linking failed: %v", memoRepo.LinkedResourceIDs)
		}
	})

	t.Run("Create empty content should fail", func(t *testing.T) {
		input := dto.CreateMemoRequest{Content: ""}
		_, err := svc.Create(ctx, 1, input)
		if err == nil {
			t.Error("Empty content should return error")
		}
	})
}
