package sqlite

import (
	sldb "daily/internal/infrastructure/persistence/sqlite/db"
	"database/sql"
)

type SqliteMemoRepository struct {
	queries *sldb.Queries
	db      *sql.DB
}

func NewSqliteMemoRepository(db *sql.DB) *SqliteMemoRepository {
	return &SqliteMemoRepository{
		queries: sldb.New(db),
		db:      db,
	}
}
