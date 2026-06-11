package container

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"daily/internal/application/port"
	authuc "daily/internal/application/usecase/auth"
	memouc "daily/internal/application/usecase/memo"
	"daily/internal/application/usecase/resource"
	tagsetuc "daily/internal/application/usecase/tag_set"
	"daily/internal/infrastructure/config"
	"daily/internal/infrastructure/persistence/sqlite"
	"daily/internal/infrastructure/storage"
	"daily/internal/infrastructure/tokenizer"

	_ "modernc.org/sqlite"
)

// Container 容器包含了应用所需的所有基础设施组件
type Container struct {
	MemoRepo        port.MemoRepository
	ResRepo         port.ResourceRepository
	UserRepo        port.UserRepository
	TagSetGroupRepo port.TagSetGroupRepository
	TagSetRepo      port.TagSetRepository
	Tokenizer       port.Tokenizer
	BlobStore       port.BlobStore

	// UseCases
	ArchiveExpiredMemosUC *memouc.ArchiveExpiredMemosUseCase
}

// NewContainer 初始化底层基础设施
func NewContainer(ctx context.Context, cfg *config.Config, l *slog.Logger) (*Container, func(), error) {
	var (
		memoRepo        port.MemoRepository
		resRepo         port.ResourceRepository
		userRepo        port.UserRepository
		tagSetGroupRepo port.TagSetGroupRepository
		tagSetRepo      port.TagSetRepository
		cleanup         func()
	)

	db, err := sql.Open("sqlite", cfg.SQLiteDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite Concurrency: use SetMaxOpenConns(1) to serialize writes and prevent "database is locked" errors.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("ping sqlite: %w", err)
	}

	cleanup = func() { db.Close() }

	memoRepo = sqlite.NewSqliteMemoRepository(db)
	resRepo = sqlite.NewSqliteResourceRepository(db)
	userRepo = sqlite.NewSqliteUserRepository(db)
	tagSetGroupRepo = sqlite.NewSqliteTagSetGroupRepository(db)
	tagSetRepo = sqlite.NewSqliteTagSetRepository(db)

	// 2. 初始化核心组件
	gseTokenizer, err := tokenizer.NewGseTokenizer()
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("init tokenizer: %w", err)
	}
	blobStore, err := storage.NewLocalBlobStore(cfg.StorageDir)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("init blob store: %w", err)
	}

	// 3. 执行启动时的清理/引导逻辑
	identitySvc := authuc.NewIdentityService(userRepo)
	tagSetSvc := tagsetuc.NewService(tagSetGroupRepo, tagSetRepo)
	runBootstrapTasks(ctx, l, memoRepo, resRepo, blobStore, identitySvc, userRepo, tagSetSvc, cfg)

	c := &Container{
		MemoRepo:              memoRepo,
		ResRepo:               resRepo,
		UserRepo:              userRepo,
		TagSetGroupRepo:       tagSetGroupRepo,
		TagSetRepo:            tagSetRepo,
		Tokenizer:             gseTokenizer,
		BlobStore:             blobStore,
		ArchiveExpiredMemosUC: memouc.NewArchiveExpiredMemosUseCase(memoRepo),
	}

	return c, cleanup, nil
}

func runBootstrapTasks(
	ctx context.Context,
	l *slog.Logger,
	memoRepo port.MemoRepository,
	resRepo port.ResourceRepository,
	blobStore port.BlobStore,
	identitySvc *authuc.IdentityService,
	userRepo port.UserRepository,
	tagSetSvc *tagsetuc.Service,
	cfg *config.Config,
) {
	// 启动时清理过期笔记
	archiveUC := memouc.NewArchiveExpiredMemosUseCase(memoRepo)
	if err := archiveUC.Execute(ctx); err != nil {
		l.Warn("failed to archive expired memos on startup", "error", err)
	}

	// 清理无用资源
	cleanupUC := resource.NewCleanupResourcesUseCase(resRepo, blobStore)
	if cleaned, err := cleanupUC.Execute(ctx); err != nil {
		l.Warn("resource cleanup failed", "error", err)
	} else if cleaned > 0 {
		l.Info("startup resource cleanup completed", "files_removed", cleaned)
	}

	// 确保 Admin 用户存在
	if err := identitySvc.EnsureBootstrapAdmin(ctx, cfg.BootstrapAdminUsername, cfg.BootstrapAdminPassword); err != nil {
		l.Error("failed to ensure bootstrap admin", "error", err)
	}

	// 为 Admin 用户创建 Demo 笔记
	if cfg.BootstrapDemoMemos {
		memouc.BootstrapDemoMemos(ctx, l, memoRepo, userRepo, cfg.BootstrapAdminUsername, cfg.BootstrapDemoMemosPath)
	}

	// 为 Admin 用户创建默认标签预设
	if cfg.BootstrapAdminUsername != "" {
		admin, err := userRepo.GetByUsername(ctx, cfg.BootstrapAdminUsername)
		if err != nil {
			l.Warn("failed to fetch admin user for tag set bootstrap", "error", err)
		} else if err := tagSetSvc.BootstrapDefaults(ctx, admin.ID); err != nil {
			l.Warn("failed to bootstrap default tag sets", "error", err)
		} else {
			l.Info("default tag sets bootstrapped for admin user")
		}
	}
}
