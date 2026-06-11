package memo

import (
	"context"
	"testing"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

type MockListMemoRepo struct {
	port.MemoRepository
	LastFilter port.MemoFilter
}

func (m *MockListMemoRepo) List(ctx context.Context, filter port.MemoFilter, userID int64) ([]*entity.Memo, error) {
	m.LastFilter = filter
	return []*entity.Memo{{UUID: "test-uuid-1", Content: "Test"}}, nil
}

func (m *MockListMemoRepo) ResolveCanonicalTag(ctx context.Context, tag string) (string, error) {
	if tag == "SRE" {
		return "Ops", nil
	}
	return tag, nil
}
func (m *MockListMemoRepo) RenameTag(ctx context.Context, userID int64, from, to string) (*port.TagRenameResult, error) {
	return nil, nil
}
func (m *MockListMemoRepo) MergeTags(ctx context.Context, userID int64, sources []string, target string) (*port.TagMergeResult, error) {
	return nil, nil
}
func (m *MockListMemoRepo) SaveTagAlias(ctx context.Context, alias, canonical string) error {
	return nil
}
func (m *MockListMemoRepo) DeleteTagAlias(ctx context.Context, alias string) error {
	return nil
}
func (m *MockListMemoRepo) ListTagAliases(ctx context.Context) ([]port.TagAlias, error) {
	return nil, nil
}
func (m *MockListMemoRepo) AppendTagAudit(ctx context.Context, action, summary string, affectedMemos int64) error {
	return nil
}
func (m *MockListMemoRepo) ListTagAudits(
	ctx context.Context,
	limit int,
	action string,
) ([]port.TagAuditRecord, error) {
	return nil, nil
}

func TestListMemosUseCase_Execute(t *testing.T) {
	repo := &MockListMemoRepo{}
	tokenizer := &MockTokenizer{} // 使用 create_memo_test.go 中定义的 MockTokenizer
	svc := NewMemoService(repo, nil, repo, tokenizer)

	t.Run("Search term should be processed into AND pattern", func(t *testing.T) {
		searchTerm := "Go Clean"
		filter := port.MemoFilter{Search: &searchTerm}

		_, _ = svc.List(context.Background(), 1, filter)

		// 验证分词后的 AND 拼接逻辑
		expected := "Go AND Clean"
		if *repo.LastFilter.Search != expected {
			t.Errorf("Expected search pattern %s, got %s", expected, *repo.LastFilter.Search)
		}
	})

	t.Run("Empty search term should remain empty", func(t *testing.T) {
		filter := port.MemoFilter{}
		_, _ = svc.List(context.Background(), 1, filter)
		if repo.LastFilter.Search != nil {
			t.Error("Search filter should be nil")
		}
	})

	t.Run("Tag filter should be canonicalized by alias", func(t *testing.T) {
		tag := "SRE"
		filter := port.MemoFilter{Tag: &tag}
		_, _ = svc.List(context.Background(), 1, filter)
		if repo.LastFilter.Tag == nil || *repo.LastFilter.Tag != "Ops" {
			t.Fatalf("expected canonical tag Ops, got %+v", repo.LastFilter.Tag)
		}
	})
}
