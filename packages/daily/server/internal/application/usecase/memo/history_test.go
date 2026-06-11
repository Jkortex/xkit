package memo

import (
	"context"
	"testing"
	"time"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"github.com/google/uuid"
)

type MockHistoryMemoRepo struct {
	port.MemoRepository
	Histories map[string][]*port.MemoHistoryRecord
	Current   *entity.Memo
}

func (m *MockHistoryMemoRepo) GetByUUID(ctx context.Context, uuid string, userID int64) (*entity.Memo, error) {
	return m.Current, nil
}

func (m *MockHistoryMemoRepo) ListHistory(ctx context.Context, memoUUID string, userID int64) ([]*port.MemoHistoryRecord, error) {
	return m.Histories[memoUUID], nil
}

func (m *MockHistoryMemoRepo) SaveHistory(ctx context.Context, rec *port.MemoHistoryRecord, userID int64) error {
	m.Histories[rec.MemoUUID] = append(m.Histories[rec.MemoUUID], rec)
	return nil
}

func (m *MockHistoryMemoRepo) GetHistoryByID(ctx context.Context, hid string, userID int64) (*port.MemoHistoryRecord, error) {
	for _, list := range m.Histories {
		for _, h := range list {
			if h.ID == hid {
				return h, nil
			}
		}
	}
	return nil, nil
}

func (m *MockHistoryMemoRepo) Update(
	ctx context.Context,
	uuid string,
	userID int64,
	content string,
	si port.SearchIndex,
	tags []string,
	resourceIDs []string,
	expiresAt *time.Time,
) error {
	m.Current.Content = content
	m.Current.Tags = tags
	m.Current.ExpiresAt = expiresAt
	return nil
}

func (m *MockHistoryMemoRepo) Create(
	ctx context.Context,
	memo *entity.Memo,
	userID int64,
	tags []string,
	resourceIDs []string,
	si port.SearchIndex,
) error {
	m.Current = memo
	return nil
}

func (m *MockHistoryMemoRepo) DeleteOldHistory(ctx context.Context, memoUUID string, keepLimit int) (int64, error) {
	return 0, nil
}

func (m *MockHistoryMemoRepo) ListAll(ctx context.Context, userID int64) ([]*entity.Memo, error) {
	return nil, nil
}

func (m *MockHistoryMemoRepo) GetRandom(ctx context.Context, userID int64) (*entity.Memo, error) {
	return nil, nil
}

func (m *MockHistoryMemoRepo) ArchiveExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockHistoryMemoRepo) SaveTag(ctx context.Context, tag string) error {
	return nil
}

func (m *MockHistoryMemoRepo) LinkMemoTag(ctx context.Context, memoUUID string, tag string) error {
	return nil
}

func (m *MockHistoryMemoRepo) ReplaceMemoTags(ctx context.Context, memoUUID string, tags []string) error {
	return nil
}

func (m *MockHistoryMemoRepo) CleanupOrphanTags(ctx context.Context) error {
	return nil
}

func (m *MockHistoryMemoRepo) ListTagsWithCount(ctx context.Context, userID int64) ([]entity.TagStat, error) {
	return nil, nil
}

func (m *MockHistoryMemoRepo) ResolveCanonicalTag(ctx context.Context, tag string) (string, error) {
	return tag, nil
}

func (m *MockHistoryMemoRepo) RenameTag(ctx context.Context, userID int64, from, to string) (*port.TagRenameResult, error) {
	return nil, nil
}
func (m *MockHistoryMemoRepo) MergeTags(ctx context.Context, userID int64, sources []string, target string) (*port.TagMergeResult, error) {
	return nil, nil
}
func (m *MockHistoryMemoRepo) SaveTagAlias(ctx context.Context, alias, canonical string) error {
	return nil
}
func (m *MockHistoryMemoRepo) DeleteTagAlias(ctx context.Context, alias string) error {
	return nil
}
func (m *MockHistoryMemoRepo) ListTagAliases(ctx context.Context) ([]port.TagAlias, error) {
	return nil, nil
}
func (m *MockHistoryMemoRepo) AppendTagAudit(ctx context.Context, action, summary string, affectedMemos int64) error {
	return nil
}
func (m *MockHistoryMemoRepo) ListTagAudits(ctx context.Context, limit int, action string) ([]port.TagAuditRecord, error) {
	return nil, nil
}

func TestListMemoHistoryUseCase(t *testing.T) {
	uuid1 := uuid.NewString()
	repo := &MockHistoryMemoRepo{
		Current: &entity.Memo{UUID: uuid1},
		Histories: map[string][]*port.MemoHistoryRecord{
			uuid1: {
				{ID: "h1", MemoUUID: uuid1, Content: "v1"},
			},
		},
	}
	svc := NewMemoService(repo, nil, repo, &MockTokenizer{})

	res, err := svc.ListHistory(context.Background(), 1, uuid1)
	if err != nil {
		t.Fatalf("ListHistory failed: %v", err)
	}
	if len(res) != 1 || res[0].Content != "v1" {
		t.Errorf("Unexpected history result: %+v", res)
	}
}

func TestRollbackMemoUseCase(t *testing.T) {
	uuid1 := uuid.NewString()
	repo := &MockHistoryMemoRepo{
		Current: &entity.Memo{UUID: uuid1, Content: "current"},
		Histories: map[string][]*port.MemoHistoryRecord{
			uuid1: {
				{ID: "h1", MemoUUID: uuid1, Content: "old version", Tags: []string{"T1"}},
			},
		},
	}
	resRepo := &MockResourceRepository{} // Reuse from create_memo_test.go
	svc := NewMemoService(repo, resRepo, repo, &MockTokenizer{})

	res, err := svc.Rollback(context.Background(), 1, uuid1, "h1")
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	if res.Content != "old version" {
		t.Errorf("Rollback content mismatch: %s", res.Content)
	}

	// Check if current was backed up before rollback
	histories := repo.Histories[uuid1]
	foundBackup := false
	for _, h := range histories {
		if h.Content == "current" {
			foundBackup = true
			break
		}
	}
	if !foundBackup {
		t.Error("Current state was not backed up before rollback")
	}
}
