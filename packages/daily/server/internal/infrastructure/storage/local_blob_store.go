package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalBlobStore struct {
	baseDir string
}

func NewLocalBlobStore(baseDir string) (*LocalBlobStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create base storage dir: %w", err)
	}
	return &LocalBlobStore{baseDir: baseDir}, nil
}

func (s *LocalBlobStore) Put(ctx context.Context, relPath string, reader io.Reader) error {
	fullPath := filepath.Join(s.baseDir, relPath)

	// 确保子目录存在 (例如 2024/05/)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("ensure subdir: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("copy stream: %w", err)
	}

	return nil
}

func (s *LocalBlobStore) Get(ctx context.Context, relPath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, relPath)
	return os.Open(fullPath)
}

func (s *LocalBlobStore) Delete(ctx context.Context, relPath string) error {
	fullPath := filepath.Join(s.baseDir, relPath)
	return os.Remove(fullPath)
}

func (s *LocalBlobStore) ListAll(ctx context.Context) ([]string, error) {
	var files []string
	err := filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(s.baseDir, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}
