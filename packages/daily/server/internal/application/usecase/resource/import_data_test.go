package resource

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

type ImportMemoRepoMock struct {
	port.MemoRepository
	Existing []*entity.Memo
	Created  []*entity.Memo
}

func (m *ImportMemoRepoMock) ListAll(ctx context.Context, userID int64) ([]*entity.Memo, error) {
	return m.Existing, nil
}

func (m *ImportMemoRepoMock) Create(
	ctx context.Context,
	memo *entity.Memo,
	userID int64,
	tags []string,
	resourceIDs []string,
	si port.SearchIndex,
) error {
	memo.Tags = tags
	m.Created = append(m.Created, memo)
	return nil
}

type ImportResourceRepoMock struct {
	port.ResourceRepository
	Existing []*entity.Resource
	Saved    []*entity.Resource
}

func (m *ImportResourceRepoMock) ListAll(ctx context.Context, userID int64) ([]*entity.Resource, error) {
	return m.Existing, nil
}

func (m *ImportResourceRepoMock) Save(ctx context.Context, res *entity.Resource, userID int64) error {
	m.Saved = append(m.Saved, res)
	return nil
}

type ImportBlobStoreMock struct {
	port.BlobStore
}

func (m *ImportBlobStoreMock) Put(ctx context.Context, relPath string, reader io.Reader) error {
	_, err := io.ReadAll(reader)
	return err
}

func TestImportDataUseCase_Execute_Report(t *testing.T) {
	memoRepo := &ImportMemoRepoMock{
		Existing: []*entity.Memo{
			{UUID: "memo-existing"},
		},
	}
	resRepo := &ImportResourceRepoMock{
		Existing: []*entity.Resource{
			{ID: "res-existing-id", Hash: "hash-existing", InternalPath: "2026/03/existing.png"},
		},
	}
	store := &ImportBlobStoreMock{}
	uc := NewImportDataUseCase(memoRepo, resRepo, store)

	archive, err := buildImportArchive()
	if err != nil {
		t.Fatalf("build archive failed: %v", err)
	}

	report, err := uc.Execute(context.Background(), 1, bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if report.Resources.Imported != 1 || report.Resources.Skipped != 4 {
		t.Fatalf("unexpected resources report: %+v", report.Resources)
	}
	if report.Memos.Imported != 1 || report.Memos.Skipped != 2 {
		t.Fatalf("unexpected memos report: %+v", report.Memos)
	}
	if len(memoRepo.Created) != 1 || memoRepo.Created[0].Content != "imported memo content" {
		t.Fatalf("unexpected created memos: %+v", memoRepo.Created)
	}
	if len(resRepo.Saved) != 1 || resRepo.Saved[0].ID != "res-imported" {
		t.Fatalf("unexpected saved resources: %+v", resRepo.Saved)
	}
}

func buildImportArchive() ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	resources := []entity.Resource{
		{ID: "res-existing-id", Hash: "hash-new", InternalPath: "2026/03/new-id.png"},
		{ID: "res-dup-hash", Hash: "hash-existing", InternalPath: "2026/03/new-hash.png"},
		{ID: "res-dup-path", Hash: "hash-dup-path", InternalPath: "2026/03/existing.png"},
		{ID: "", Hash: "hash-invalid", InternalPath: ""},
		{ID: "res-imported", Hash: "hash-imported", InternalPath: "2026/03/imported.png"},
	}
	resFile, err := zw.Create("resources.json")
	if err != nil {
		return nil, err
	}
	if err := json.NewEncoder(resFile).Encode(resources); err != nil {
		return nil, err
	}

	assetFile, err := zw.Create("assets/2026/03/imported.png")
	if err != nil {
		return nil, err
	}
	if _, err := assetFile.Write([]byte("imported")); err != nil {
		return nil, err
	}

	memos := []entity.Memo{
		{UUID: "memo-existing", Content: "same-id memo"},
		{Content: "missing uuid memo"},
		{UUID: "memo-imported", Content: "imported memo content"},
	}
	memoFile, err := zw.Create("memos.json")
	if err != nil {
		return nil, err
	}
	if err := json.NewEncoder(memoFile).Encode(memos); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
