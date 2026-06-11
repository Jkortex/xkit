package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemoCli_CreateAndUpdateWithFile(t *testing.T) {
	// Create a temp directory for DB and test files to ensure isolation
	tmpDir, err := os.MkdirTemp("", "daily-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbFile := filepath.Join(tmpDir, "test.db")

	// Set DSN env so that getApp() uses it
	os.Setenv("DAILY_SQLITE_DSN", dbFile)
	defer os.Unsetenv("DAILY_SQLITE_DSN")

	// Set dbPath flag to default so it triggers env var lookup
	dbPath = "~/.daily/daily.db"

	// Mock the migrations directory.
	migrationsDir := ""
	candidates := []string{
		"../../../migrations/sqlite",
		"../../../../migrations/sqlite",
		"../../migrations/sqlite",
		"migrations/sqlite",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			migrationsDir = c
			break
		}
	}
	if migrationsDir == "" {
		t.Fatalf("failed to find migrations directory")
	}

	// Save original CWD
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Create "migrations/sqlite" under tmpDir and copy migration files there
	testMigrationsDir := filepath.Join(tmpDir, "migrations", "sqlite")
	if err := os.MkdirAll(testMigrationsDir, 0755); err != nil {
		t.Fatalf("failed to create test migrations dir: %v", err)
	}

	// Copy migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("failed to read migrations dir: %v", err)
	}
	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(migrationsDir, f.Name()))
		if err != nil {
			t.Fatalf("failed to read migration file %s: %v", f.Name(), err)
		}
		err = os.WriteFile(filepath.Join(testMigrationsDir, f.Name()), content, 0644)
		if err != nil {
			t.Fatalf("failed to write migration file %s: %v", f.Name(), err)
		}
	}

	// Change working directory to tmpDir so that daily-cli finds "migrations/sqlite"
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(oldCwd)

	// Reset global app context state
	appCtx = nil

	// 1. Test Memo Create with --file flag
	testContent := "DAP Integration Test\n#sync #DAP"
	contentFile := filepath.Join(tmpDir, "memo-content.md")
	if err := os.WriteFile(contentFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write content file: %v", err)
	}

	// Execute memo create --file command while capturing stdout
	rootCmd.SetArgs([]string{"memo", "create", "--file", contentFile, "--tag", "integration"})
	output, err := captureStdout(func() error {
		return rootCmd.ExecuteContext(context.Background())
	})
	if err != nil {
		t.Fatalf("failed to execute memo create: %v", err)
	}

	if !strings.Contains(output, `"uuid"`) {
		t.Fatalf("expected JSON output containing uuid, got: %s", output)
	}

	// Parse UUID from JSON output
	type MemoRes struct {
		UUID string `json:"uuid"`
	}
	var res MemoRes

	// Find JSON block in output using curly braces
	start := strings.Index(output, "{")
	end := strings.LastIndex(output, "}")
	if start == -1 || end == -1 || start >= end {
		t.Fatalf("could not find JSON block in output: %s", output)
	}
	jsonStr := output[start : end+1]
	if err := json.Unmarshal([]byte(jsonStr), &res); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v, raw output: %s", err, output)
	}
	if res.UUID == "" {
		t.Fatalf("returned UUID is empty, output: %s", output)
	}

	// 2. Test Memo Update with --file flag
	updatedContent := "DAP Integration Test Updated\n#sync #DAP #Updated"
	updatedFile := filepath.Join(tmpDir, "memo-content-updated.md")
	if err := os.WriteFile(updatedFile, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("failed to write updated content file: %v", err)
	}

	rootCmd.SetArgs([]string{"memo", "update", res.UUID, "--file", updatedFile, "--tag", "updated"})
	output, err = captureStdout(func() error {
		return rootCmd.ExecuteContext(context.Background())
	})
	if err != nil {
		t.Fatalf("failed to execute memo update: %v", err)
	}

	if !strings.Contains(output, `"uuid"`) {
		t.Fatalf("expected JSON output containing uuid, got: %s", output)
	}
}

func captureStdout(f func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}
