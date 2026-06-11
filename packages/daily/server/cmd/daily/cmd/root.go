package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"daily/internal/application/usecase/memo"
	"daily/internal/infrastructure/config"
	"daily/internal/infrastructure/container"
	"daily/internal/infrastructure/logger"
	"daily/internal/interfaces/controller"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// AppContext 持有 CLI 共享状态（容器 + 已解析的管理员用户 ID）
type AppContext struct {
	Container     *container.Container
	Config        *config.Config
	ConfigSources config.ConfigSources
	AdminID       int64
	MemoCtrl      *controller.MemoController
	TagCtrl       *controller.TagController
	DB            *sql.DB
}

var (
	dbPath    string
	adminUser string
	adminPass string
)

var rootCmd = &cobra.Command{
	Use:   "daily",
	Short: "Daily — local-first memo management",
	Long: `Daily — a local-first tool for personal knowledge management.

  daily tui     Interactive terminal UI
  daily web     HTTP server with embedded web UI
  daily memo    Non-interactive memo operations (for scripts/tools)
  daily watch   Real-time memo monitor (for pipelines)
  daily tag     Tag management
  daily tree    Tree document protocol

Operates directly on a SQLite database.
All operations run as the configured admin user.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "~/.daily/daily.db", "SQLite database path")
	rootCmd.PersistentFlags().StringVar(&adminUser, "admin-user", "admin", "Admin username (auto-created if not exists)")
	rootCmd.PersistentFlags().StringVar(&adminPass, "admin-pass", "admin", "Admin password")
}

// 全局懒加载的应用上下文
var appCtx *AppContext

func getApp() (*AppContext, error) {
	if appCtx != nil {
		return appCtx, nil
	}

	// 1. 加载配置（优先级: 配置文件 > 环境变量 > 默认值）
	cfg, sources, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// 2. CLI flags 覆盖配置
	if cfg.SQLiteDSN == "" || cfg.SQLiteDSN == "./data/daily.db" {
		cfg.SQLiteDSN = expandHome(dbPath)
	}

	// 3. 确保路径是绝对路径
	absDB, err := filepath.Abs(cfg.SQLiteDSN)
	if err != nil {
		return nil, fmt.Errorf("resolve db path: %w", err)
	}
	cfg.SQLiteDSN = absDB

	// 4. 运行数据库迁移（幂等，CREATE IF NOT EXISTS）
	if err := runMigrations(absDB); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	// 5. 初始化容器（内部会创建 DB 连接、repos、执行 bootstrap 任务）
	l := logger.Init("warn")
	c, cleanup, err := container.NewContainer(context.Background(), cfg, l)
	if err != nil {
		return nil, fmt.Errorf("container: %w", err)
	}

	// 6. 解析管理员用户 ID（container.NewContainer 的 EnsureBootstrapAdmin 已确保存在）
	user, err := c.UserRepo.GetByUsername(context.Background(), cfg.BootstrapAdminUsername)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("resolve admin user %q: %w", cfg.BootstrapAdminUsername, err)
	}
	adminID := user.ID

	// 7. 组装控制器（与 handler/factory.go 一致，复用现有业务逻辑）
	memoSvc := memo.NewMemoService(c.MemoRepo, c.ResRepo, c.MemoRepo, c.Tokenizer)
	tagSvc := memo.NewTagService(c.MemoRepo)

	memoCtrl := controller.NewMemoController(memoSvc)
	tagCtrl := controller.NewTagController(tagSvc)

	// Open DB connection for direct CLI queries
	directDB, err := sql.Open("sqlite", absDB)
	if err != nil {
		return nil, fmt.Errorf("open direct db: %w", err)
	}

	appCtx = &AppContext{
		Container:     c,
		Config:        cfg,
		ConfigSources: sources,
		AdminID:       adminID,
		MemoCtrl:      memoCtrl,
		TagCtrl:       tagCtrl,
		DB:            directDB,
	}

	return appCtx, nil
}

// runMigrations 对 SQLite 数据库执行 schema 迁移（幂等）
// 迁移文件通过 //go:embed 内嵌在二进制中，不依赖 CWD
func runMigrations(dbPath string) error {
	// 打开数据库（确保目录存在）
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db dir %s: %w", dir, err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite for migration: %w", err)
	}
	defer db.Close()

	migrations, err := readEmbeddedMigrations()
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	// 顺序执行迁移文件（已按文件名排序）
	for _, name := range sortedKeys(migrations) {
		if _, err := db.Exec(string(migrations[name])); err != nil {
			return fmt.Errorf("exec migration %s: %w", name, err)
		}
	}

	return nil
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
