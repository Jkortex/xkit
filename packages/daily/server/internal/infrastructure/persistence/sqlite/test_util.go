package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

type Fataler interface {
	Fatalf(format string, args ...interface{})
}

// SetupTestDB initializes a sqlite test database and runs migrations.
func SetupTestDB(t Fataler) *sql.DB {
	// Use a temporary file for each test run to ensure isolation
	tmpDir, err := os.MkdirTemp("", "daily-sqlite-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir for sqlite test: %v", err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")

	// DSN with WAL mode and foreign keys enabled
	dsn := fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	db.SetMaxOpenConns(1)

	// Run migrations
	// We are in internal/infrastructure/persistence/sqlite/test_util.go
	// Project root is 4 levels up. Migrations are in migrations/sqlite
	// But tests might run from packages/daily/server/tests/integration
	// Let's try to find it.

	migrationsDir := ""
	candidates := []string{
		"../../../../migrations/sqlite",
		"../../../migrations/sqlite",
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

	if err := runMigrations(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func runMigrations(db *sql.DB, dir string) error {
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for i, file := range files {
		version := i + 1
		content, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return err
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("file %s: %w", file, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
