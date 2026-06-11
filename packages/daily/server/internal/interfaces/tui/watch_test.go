package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWatchFileReturnsCmd(t *testing.T) {
	cmd := WatchFile("/nonexistent/path")
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd")
	}
}

func TestEnsureEventPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test", "event")
	err := EnsureEventPath(path)
	if err != nil {
		t.Fatalf("EnsureEventPath failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected event file to exist at %s", path)
	}
}

func TestEnsureEventPathExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing-event")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	err := EnsureEventPath(path)
	if err != nil {
		t.Fatalf("EnsureEventPath on existing file failed: %v", err)
	}
}
