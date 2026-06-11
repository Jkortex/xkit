package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Init 初始化日志系统
func Init(level string) *slog.Logger {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// 1. Stdout Handler: 本地 JSON 打印
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})

	logger := slog.New(h)
	slog.SetDefault(logger)

	return logger
}
