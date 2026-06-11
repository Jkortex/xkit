package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"daily/internal/infrastructure/api"
	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/config"
	"daily/internal/infrastructure/container"
	"daily/internal/infrastructure/logger"
	"daily/internal/infrastructure/notify"
	"daily/internal/tts"

	"github.com/spf13/cobra"
)

var servePort int

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the Daily web server (embedded UI)",
	Long: `Start the Daily HTTP API server with the embedded web UI.

Pairs with 'daily tui' as the two interaction modes for Daily.
Uses the same SQLite database as other CLI commands.
Environment variables (DAILY_PORT, DAILY_LOG_LEVEL, etc.) are also respected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. 加载配置（优先级: 配置文件 > 环境变量 > 默认值）
		cfg, _, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		// 2. CLI flags 覆盖配置
		if servePort != 8080 {
			cfg.Port = servePort
		}
		if dbPath != "~/.daily/daily.db" {
			cfg.SQLiteDSN = expandHome(dbPath)
		}
		if adminUser != "admin" {
			cfg.BootstrapAdminUsername = adminUser
		}
		if adminPass != "admin" {
			cfg.BootstrapAdminPassword = adminPass
		}

		// 3. 确保路径是绝对路径
		absDB, err := filepath.Abs(cfg.SQLiteDSN)
		if err != nil {
			return fmt.Errorf("resolve db path: %w", err)
		}
		cfg.SQLiteDSN = absDB

		if err := runMigrations(absDB); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
		notify.Init(absDB)

		l := logger.Init(cfg.LogLevel)
		l.Info("bootstrapping daily server")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, cleanup, err := container.NewContainer(ctx, cfg, l)
		if err != nil {
			return fmt.Errorf("container: %w", err)
		}
		defer cleanup()

		handlers := handler.NewHandlers(c, l)
		srv := api.NewServer(cfg, handlers)

		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			l.Info("background archive worker started")
			for {
				select {
				case <-ticker.C:
					if err := c.ArchiveExpiredMemosUC.Execute(ctx); err != nil {
						l.Warn("background archive expired memos failed", "error", err)
					}
					// Clean TTS cache
					if cacheDir, err := tts.CacheDir(); err == nil {
						if removed, err := tts.CleanupCache(cacheDir); err == nil && removed > 0 {
							l.Info("TTS cache cleanup", "removed", removed)
						}
					}
				case <-ctx.Done():
					l.Info("background archive worker shutting down")
					return
				}
			}
		}()

		errCh := make(chan error, 1)
		go func() {
			l.Info("server starting", "addr", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-errCh:
			return fmt.Errorf("listen: %w", err)
		case sig := <-quit:
			l.Info("shutdown signal received", "signal", sig.String())
		}

		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		return srv.Shutdown(shutdownCtx)
	},
}

func init() {
	webCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "HTTP server port")
	rootCmd.AddCommand(webCmd)
}
