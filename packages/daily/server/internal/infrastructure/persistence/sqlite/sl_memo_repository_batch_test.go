package sqlite

import (
	"context"
	"testing"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

func TestSqliteMemoRepository_BatchArchive(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	repo := NewSqliteMemoRepository(db)
	ctx := context.Background()
	userID := int64(1)

	// Create test memos
	memo1 := entity.NewMemo("m1")
	memo2 := entity.NewMemo("m2")
	repo.Create(ctx, memo1, userID, nil, nil, port.SearchIndex{})
	repo.Create(ctx, memo2, userID, nil, nil, port.SearchIndex{})

	uuids := []string{memo1.UUID, memo2.UUID}

	archived, err := repo.BatchArchive(ctx, userID, uuids)
	if err != nil {
		t.Fatal(err)
	}

	if len(archived) != 2 {
		t.Fatalf("expected 2 archived, got %d", len(archived))
	}

	// Verify status
	m1, _ := repo.GetByUUID(ctx, memo1.UUID, userID)
	if m1.RowStatus != entity.RowStatusArchived {
		t.Errorf("m1 status expected archived, got %s", m1.RowStatus)
	}
}

func TestSqliteMemoRepository_BatchDelete(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	repo := NewSqliteMemoRepository(db)
	ctx := context.Background()
	userID := int64(1)

	memo1 := entity.NewMemo("m1")
	repo.Create(ctx, memo1, userID, nil, nil, port.SearchIndex{})

	deleted, err := repo.BatchDelete(ctx, userID, []string{memo1.UUID})
	if err != nil {
		t.Fatal(err)
	}

	if len(deleted) != 1 {
		t.Fatalf("expected 1 deleted, got %d", len(deleted))
	}

	_, err = repo.GetByUUID(ctx, memo1.UUID, userID)
	if err == nil {
		t.Errorf("expected error not found")
	}
}

func TestSqliteMemoRepository_BatchTag(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	repo := NewSqliteMemoRepository(db)
	ctx := context.Background()
	userID := int64(1)

	memo1 := entity.NewMemo("m1")
	memo2 := entity.NewMemo("m2")
	repo.Create(ctx, memo1, userID, []string{"t1"}, nil, port.SearchIndex{})
	repo.Create(ctx, memo2, userID, []string{"t1", "t2"}, nil, port.SearchIndex{})

	uuids := []string{memo1.UUID, memo2.UUID}

	tagged, err := repo.BatchTag(ctx, userID, uuids, []string{"newTag"}, []string{"t1"})
	if err != nil {
		t.Fatal(err)
	}

	if len(tagged) != 2 {
		t.Fatalf("expected 2 tagged, got %d", len(tagged))
	}

	m1, _ := repo.GetByUUID(ctx, memo1.UUID, userID)
	if len(m1.Tags) != 1 || m1.Tags[0] != "newTag" {
		t.Errorf("expected tags [newTag], got %v", m1.Tags)
	}

	m2, _ := repo.GetByUUID(ctx, memo2.UUID, userID)
	if len(m2.Tags) != 2 { // newTag + t2
		t.Errorf("expected 2 tags, got %v", m2.Tags)
	}
}
