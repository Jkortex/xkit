package memo

import (
	"context"
	"testing"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

type MockBatchRepo struct {
	port.MemoRepository
	partialSuccess bool
}

func (m *MockBatchRepo) GetByUUID(ctx context.Context, uuid string, userID int64) (*entity.Memo, error) {
	return &entity.Memo{UUID: uuid}, nil
}

func (m *MockBatchRepo) BatchArchive(ctx context.Context, userID int64, uuids []string) ([]string, error) {
	if m.partialSuccess && len(uuids) > 1 {
		return uuids[:1], nil
	}
	return uuids, nil
}

func (m *MockBatchRepo) BatchDelete(ctx context.Context, userID int64, uuids []string) ([]string, error) {
	if m.partialSuccess && len(uuids) > 1 {
		return uuids[:1], nil
	}
	return uuids, nil
}

func (m *MockBatchRepo) BatchTag(ctx context.Context, userID int64, uuids []string, addTags []string, removeTags []string) ([]string, error) {
	if m.partialSuccess && len(uuids) > 1 {
		return uuids[:1], nil
	}
	return uuids, nil
}

func TestBatchArchive(t *testing.T) {
	repo := &MockBatchRepo{}
	svc := NewMemoService(repo, nil, nil, nil)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		res, err := svc.BatchArchive(ctx, 1, []string{"u1", "u2"})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Succeeded) != 2 {
			t.Errorf("Expected 2 succeeded")
		}
		if len(res.Failed) != 0 {
			t.Errorf("Expected 0 failed")
		}
	})

	t.Run("Empty", func(t *testing.T) {
		_, err := svc.BatchArchive(ctx, 1, []string{})
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("ExceedLimit", func(t *testing.T) {
		var uuids []string
		for i := 0; i < 101; i++ {
			uuids = append(uuids, "uuid")
		}
		_, err := svc.BatchArchive(ctx, 1, uuids)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("PartialSuccess", func(t *testing.T) {
		repo.partialSuccess = true
		defer func() { repo.partialSuccess = false }()
		res, err := svc.BatchArchive(ctx, 1, []string{"u1", "u2"})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Succeeded) != 1 {
			t.Errorf("Expected 1 succeeded")
		}
		if len(res.Failed) != 1 || res.Failed[0].UUID != "u2" {
			t.Errorf("Expected 1 failed with UUID u2, got %v", res.Failed)
		}
	})
}

func TestBatchDelete(t *testing.T) {
	repo := &MockBatchRepo{}
	svc := NewMemoService(repo, nil, nil, nil)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		res, err := svc.BatchDelete(ctx, 1, []string{"u1", "u2"})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Succeeded) != 2 {
			t.Errorf("Expected 2 succeeded")
		}
	})

	t.Run("PartialSuccess", func(t *testing.T) {
		repo.partialSuccess = true
		defer func() { repo.partialSuccess = false }()
		res, err := svc.BatchDelete(ctx, 1, []string{"u1", "u2"})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Succeeded) != 1 {
			t.Errorf("Expected 1 succeeded")
		}
		if len(res.Failed) != 1 || res.Failed[0].UUID != "u2" {
			t.Errorf("Expected 1 failed with UUID u2, got %v", res.Failed)
		}
	})
}

func TestBatchTag(t *testing.T) {
	repo := &MockBatchRepo{}
	svc := NewMemoService(repo, nil, nil, nil)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		res, err := svc.BatchTag(ctx, 1, []string{"u1", "u2"}, []string{"tag1"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Succeeded) != 2 {
			t.Errorf("Expected 2 succeeded")
		}
	})

	t.Run("PartialSuccess", func(t *testing.T) {
		repo.partialSuccess = true
		defer func() { repo.partialSuccess = false }()
		res, err := svc.BatchTag(ctx, 1, []string{"u1", "u2"}, []string{"tag1"}, []string{"tag2"})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Succeeded) != 1 {
			t.Errorf("Expected 1 succeeded")
		}
		if len(res.Failed) != 1 || res.Failed[0].UUID != "u2" {
			t.Errorf("Expected 1 failed with UUID u2, got %v", res.Failed)
		}
	})
}
