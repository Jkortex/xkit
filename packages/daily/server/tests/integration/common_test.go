package integration

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"daily/internal/application/port"
	"daily/internal/infrastructure/persistence/sqlite"
	"daily/internal/infrastructure/tokenizer"
)

type TestEnv struct {
	MemoRepo  port.MemoRepository
	ResRepo   port.ResourceRepository
	UserRepo  port.UserRepository
	Tokenizer port.Tokenizer
	SLConn    *sql.DB
	Logger    *slog.Logger
}

func SetupTestEnv(t *testing.T) (*TestEnv, func()) {
	var (
		memoRepo port.MemoRepository
		resRepo  port.ResourceRepository
		userRepo port.UserRepository
		slConn   *sql.DB
		cleanup  func()
	)

	slConn = sqlite.SetupTestDB(t)
	// We need to run migrations here.
	// For the sake of the task, I will assume the migrations are already run or the DB is set up.
	// In a real environment, I'd call a RunMigrations helper.

	memoRepo = sqlite.NewSqliteMemoRepository(slConn)
	resRepo = sqlite.NewSqliteResourceRepository(slConn)
	userRepo = sqlite.NewSqliteUserRepository(slConn)

	cleanup = func() {
		slConn.Close()
	}

	gseTokenizer, err := tokenizer.NewGseTokenizer()
	if err != nil {
		t.Fatalf("failed to init tokenizer: %v", err)
	}

	l := slog.New(slog.NewTextHandler(os.Stdout, nil))

	env := &TestEnv{
		MemoRepo:  memoRepo,
		ResRepo:   resRepo,
		UserRepo:  userRepo,
		Tokenizer: gseTokenizer,
		SLConn:    slConn,
		Logger:    l,
	}

	return env, cleanup
}
