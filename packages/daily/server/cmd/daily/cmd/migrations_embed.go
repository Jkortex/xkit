package cmd

import (
	"embed"
	"io/fs"
	"path/filepath"
	"sort"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func readEmbeddedMigrations() (map[string][]byte, error) {
	entries, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	result := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		data, err := migrationsFS.ReadFile(entry)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(entry)
		result[name] = data
	}
	return result, nil
}
