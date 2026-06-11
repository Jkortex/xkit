package resource

import (
	"bytes"
	"context"
	"io"
	"testing"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

type MockBlobStore struct {
	port.BlobStore
	PutCalled bool
}

func (m *MockBlobStore) Put(ctx context.Context, path string, r io.Reader) error {
	m.PutCalled = true
	return nil
}

func (m *MockBlobStore) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte("test content"))), nil
}

type MockResourceRepo struct {
	port.ResourceRepository
	SavedResource *entity.Resource
}

func (m *MockResourceRepo) Save(ctx context.Context, res *entity.Resource, userID int64) error {
	m.SavedResource = res
	return nil
}

func (m *MockResourceRepo) GetByID(ctx context.Context, id string, userID int64) (*entity.Resource, error) {
	return &entity.Resource{ID: id, InternalPath: "test/path"}, nil
}

func TestUploadResourceUseCase_Execute(t *testing.T) {
	repo := &MockResourceRepo{}
	store := &MockBlobStore{}
	uc := NewUploadResourceUseCase(repo, store)

	input := dto.UploadResourceInput{
		FileName: "test.png",
		Content:  bytes.NewReader([]byte("fake image content")),
		Size:     10,
		MimeType: "image/png",
	}

	res, err := uc.Execute(context.Background(), 1, input)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if !store.PutCalled {
		t.Error("BlobStore.Put was not called")
	}

	if res.ID == "" {
		t.Error("Resource ID should not be empty")
	}
}

func TestGetResourceUseCase_Execute(t *testing.T) {
	repo := &MockResourceRepo{}
	store := &MockBlobStore{}
	uc := NewGetResourceUseCase(repo, store)

	out, err := uc.Execute(context.Background(), 1, "res-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if out.Resource.ID != "res-123" {
		t.Errorf("Expected ID res-123, got %s", out.Resource.ID)
	}

	content, _ := io.ReadAll(out.Content)
	if string(content) != "test content" {
		t.Errorf("Content mismatch, got %s", string(content))
	}
}
