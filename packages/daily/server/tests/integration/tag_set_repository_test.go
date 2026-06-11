package integration

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"daily/internal/domain/entity"
	persistence "daily/internal/infrastructure/persistence/sqlite"
)

var testUserSeq int

func createTestUser(t *testing.T, pool *sql.DB) *entity.User {
	t.Helper()
	testUserSeq++
	user := &entity.User{
		Username:     fmt.Sprintf("tags-test-user-%d", testUserSeq),
		PasswordHash: "fakehash",
		Role:         entity.UserRoleMember,
		Status:       entity.UserStatusActive,
	}
	userRepo := persistence.NewSqliteUserRepository(pool)
	if err := userRepo.Create(t.Context(), user); err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return user
}

func TestTagSetGroupRepository_Create(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetGroupRepository(pool)

	g := &entity.TagSetGroup{
		ID:        "group-1",
		UserID:    user.ID,
		Name:      "Favorites",
		Weight:    10,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(t.Context(), g); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestTagSetGroupRepository_GetByID(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetGroupRepository(pool)

	g := &entity.TagSetGroup{
		ID:        "group-get",
		UserID:    user.ID,
		Name:      "Work",
		Weight:    5,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(t.Context(), g); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	got, err := repo.GetByID(t.Context(), user.ID, "group-get")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != "group-get" {
		t.Errorf("got ID %q, want %q", got.ID, "group-get")
	}
	if got.Name != "Work" {
		t.Errorf("got Name %q, want %q", got.Name, "Work")
	}
	if got.UserID != user.ID {
		t.Errorf("got UserID %d, want %d", got.UserID, user.ID)
	}
	if got.Weight != 5 {
		t.Errorf("got Weight %d, want %d", got.Weight, 5)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestTagSetGroupRepository_GetByID_NotFound(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	repo := persistence.NewSqliteTagSetGroupRepository(pool)
	_, err := repo.GetByID(t.Context(), 99999, "nonexistent")
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestTagSetGroupRepository_ListByUser(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetGroupRepository(pool)

	now := time.Now().UTC()
	groups := []*entity.TagSetGroup{
		{ID: "g1", UserID: user.ID, Name: "First", Weight: 10, CreatedAt: now, UpdatedAt: now},
		{ID: "g2", UserID: user.ID, Name: "Second", Weight: 20, CreatedAt: now, UpdatedAt: now},
		{ID: "g3", UserID: user.ID, Name: "Third", Weight: 5, CreatedAt: now, UpdatedAt: now},
	}
	for _, g := range groups {
		if err := repo.Create(t.Context(), g); err != nil {
			t.Fatalf("create setup: %v", err)
		}
	}

	results, err := repo.ListByUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d groups, want 3", len(results))
	}
	if results[0].Weight != 20 {
		t.Errorf("expected highest weight first, got %d", results[0].Weight)
	}
}

func TestTagSetGroupRepository_ListByUser_OtherUserIsolation(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	u1 := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetGroupRepository(pool)

	now := time.Now().UTC()
	if err := repo.Create(t.Context(), &entity.TagSetGroup{
		ID: "u1-only", UserID: u1.ID, Name: "U1", Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	u2 := createTestUser(t, pool)
	results, err := repo.ListByUser(t.Context(), u2.ID)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 groups for other user, got %d", len(results))
	}
}

func TestTagSetGroupRepository_Update(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetGroupRepository(pool)

	now := time.Now().UTC()
	g := &entity.TagSetGroup{
		ID: "group-upd", UserID: user.ID, Name: "Original", Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(t.Context(), g); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	g.Name = "Updated"
	g.Weight = 99
	if err := repo.Update(t.Context(), g); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.GetByID(t.Context(), user.ID, "group-upd")
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "Updated" {
		t.Errorf("got Name %q, want %q", got.Name, "Updated")
	}
	if got.Weight != 99 {
		t.Errorf("got Weight %d, want %d", got.Weight, 99)
	}
}

func TestTagSetGroupRepository_Update_NotFound(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	repo := persistence.NewSqliteTagSetGroupRepository(pool)
	now := time.Now().UTC()
	err := repo.Update(t.Context(), &entity.TagSetGroup{
		ID: "nonexistent", UserID: 99999, Name: "X", Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	})
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestTagSetGroupRepository_Delete(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetGroupRepository(pool)

	now := time.Now().UTC()
	g := &entity.TagSetGroup{
		ID: "group-del", UserID: user.ID, Name: "DeleteMe", Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(t.Context(), g); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	if err := repo.Delete(t.Context(), user.ID, "group-del"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetByID(t.Context(), user.ID, "group-del")
	if err == nil {
		t.Error("expected not found after delete")
	}
}

func TestTagSetGroupRepository_Delete_NotFound(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	repo := persistence.NewSqliteTagSetGroupRepository(pool)
	err := repo.Delete(t.Context(), 99999, "nonexistent")
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestTagSetRepository_Create(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	ts := &entity.TagSet{
		ID: "ts-1", UserID: user.ID, Name: "My Set",
		TagsAny: `["tag1","tag2"]`, TagsAll: `["tag3"]`, TagsExclude: `[]`,
		Weight: 10, CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(t.Context(), ts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestTagSetRepository_Create_WithGroup(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	groupRepo := persistence.NewSqliteTagSetGroupRepository(pool)
	tagSetRepo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	group := &entity.TagSetGroup{
		ID: "ts-group", UserID: user.ID, Name: "TestGroup", Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := groupRepo.Create(t.Context(), group); err != nil {
		t.Fatalf("create group: %v", err)
	}

	groupID := group.ID
	ts := &entity.TagSet{
		ID: "ts-with-group", UserID: user.ID, GroupID: &groupID, Name: "Grouped",
		TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 5,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := tagSetRepo.Create(t.Context(), ts); err != nil {
		t.Fatalf("Create with group failed: %v", err)
	}

	got, err := tagSetRepo.GetByID(t.Context(), user.ID, "ts-with-group")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.GroupID == nil {
		t.Fatal("expected GroupID to be set")
	}
	if *got.GroupID != groupID {
		t.Errorf("got GroupID %q, want %q", *got.GroupID, groupID)
	}
}

func TestTagSetRepository_GetByID(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	ts := &entity.TagSet{
		ID: "ts-get", UserID: user.ID, Name: "GetTest",
		TagsAny: `["a"]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 3,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(t.Context(), ts); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	got, err := repo.GetByID(t.Context(), user.ID, "ts-get")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != "ts-get" {
		t.Errorf("got ID %q", got.ID)
	}
	if got.Name != "GetTest" {
		t.Errorf("got Name %q", got.Name)
	}
	if got.UserID != user.ID {
		t.Errorf("got UserID %d", got.UserID)
	}
	if got.Weight != 3 {
		t.Errorf("got Weight %d", got.Weight)
	}
}

func TestTagSetRepository_GetByID_NotFound(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	repo := persistence.NewSqliteTagSetRepository(pool)
	_, err := repo.GetByID(t.Context(), 99999, "nonexistent")
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestTagSetRepository_ListByUser(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	sets := []*entity.TagSet{
		{ID: "s1", UserID: user.ID, Name: "S1", TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 1, CreatedAt: now, UpdatedAt: now},
		{ID: "s2", UserID: user.ID, Name: "S2", TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 2, CreatedAt: now, UpdatedAt: now},
	}
	for _, s := range sets {
		if err := repo.Create(t.Context(), s); err != nil {
			t.Fatalf("create setup: %v", err)
		}
	}

	results, err := repo.ListByUser(t.Context(), user.ID, nil)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d, want 2", len(results))
	}
	if results[0].Weight != 2 {
		t.Errorf("expected higher weight first, got %d", results[0].Weight)
	}
}

func TestTagSetRepository_ListByGroup(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	groupRepo := persistence.NewSqliteTagSetGroupRepository(pool)
	tagSetRepo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	group := &entity.TagSetGroup{
		ID: "list-group", UserID: user.ID, Name: "ListGroup", Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := groupRepo.Create(t.Context(), group); err != nil {
		t.Fatalf("create group: %v", err)
	}

	groupID := group.ID
	sets := []*entity.TagSet{
		{ID: "sg1", UserID: user.ID, GroupID: &groupID, Name: "InGroup1", TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 1, CreatedAt: now, UpdatedAt: now},
		{ID: "sg2", UserID: user.ID, GroupID: &groupID, Name: "InGroup2", TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 2, CreatedAt: now, UpdatedAt: now},
	}
	for _, s := range sets {
		if err := tagSetRepo.Create(t.Context(), s); err != nil {
			t.Fatalf("create setup: %v", err)
		}
	}

	results, err := tagSetRepo.ListByUser(t.Context(), user.ID, &groupID)
	if err != nil {
		t.Fatalf("ListByUser with groupID failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d, want 2", len(results))
	}
}

func TestTagSetRepository_Update(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	ts := &entity.TagSet{
		ID: "ts-upd", UserID: user.ID, Name: "Before",
		TagsAny: `["old"]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(t.Context(), ts); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	ts.Name = "After"
	ts.TagsAny = `["new"]`
	ts.Weight = 99
	if err := repo.Update(t.Context(), ts); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.GetByID(t.Context(), user.ID, "ts-upd")
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "After" {
		t.Errorf("got Name %q", got.Name)
	}
	if got.TagsAny != `["new"]` {
		t.Errorf("got TagsAny %q", got.TagsAny)
	}
	if got.Weight != 99 {
		t.Errorf("got Weight %d", got.Weight)
	}
}

func TestTagSetRepository_Update_NotFound(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	repo := persistence.NewSqliteTagSetRepository(pool)
	now := time.Now().UTC()
	err := repo.Update(t.Context(), &entity.TagSet{
		ID: "nonexistent", UserID: 99999, Name: "X",
		TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	})
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestTagSetRepository_Delete(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	ts := &entity.TagSet{
		ID: "ts-del", UserID: user.ID, Name: "DeleteMe",
		TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(t.Context(), ts); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	if err := repo.Delete(t.Context(), user.ID, "ts-del"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetByID(t.Context(), user.ID, "ts-del")
	if err == nil {
		t.Error("expected not found after delete")
	}
}

func TestTagSetRepository_Delete_NotFound(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	repo := persistence.NewSqliteTagSetRepository(pool)
	err := repo.Delete(t.Context(), 99999, "nonexistent")
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestTagSetRepository_TouchLastUsed(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	user := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	ts := &entity.TagSet{
		ID: "ts-touch", UserID: user.ID, Name: "TouchMe",
		TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(t.Context(), ts); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	if err := repo.TouchLastUsed(t.Context(), "ts-touch", user.ID); err != nil {
		t.Fatalf("TouchLastUsed failed: %v", err)
	}

	got, err := repo.GetByID(t.Context(), user.ID, "ts-touch")
	if err != nil {
		t.Fatalf("GetByID after touch: %v", err)
	}
	if got.LastUsedAt == nil {
		t.Fatal("expected LastUsedAt to be set after touch")
	}
	if got.LastUsedAt.IsZero() {
		t.Error("LastUsedAt is zero")
	}
}

func TestTagSetRepository_TouchLastUsed_NotFound(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	repo := persistence.NewSqliteTagSetRepository(pool)
	err := repo.TouchLastUsed(t.Context(), "nonexistent", 99999)
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestTagSetRepository_UserIsolation(t *testing.T) {
	pool := persistence.SetupTestDB(t)
	defer pool.Close()

	u1 := createTestUser(t, pool)
	u2 := createTestUser(t, pool)
	repo := persistence.NewSqliteTagSetRepository(pool)

	now := time.Now().UTC()
	if err := repo.Create(t.Context(), &entity.TagSet{
		ID: "u1-only", UserID: u1.ID, Name: "U1",
		TagsAny: `[]`, TagsAll: `[]`, TagsExclude: `[]`, Weight: 1,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	results, err := repo.ListByUser(t.Context(), u2.ID, nil)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 tag sets for other user, got %d", len(results))
	}
}
