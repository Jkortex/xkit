package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"xkit/packages/strata"
)

// Re-export strata types for backward compatibility.
type ConfigSource = strata.Source
type ConfigSources = strata.Sources

const (
	SourceDefault = strata.SourceDefault
	SourceFile    = strata.SourceFile
	SourceEnv     = strata.SourceEnv
)

// TTSConfig holds text-to-speech configuration.
type TTSConfig struct {
	Provider string `json:"provider" env:"TTS_PROVIDER" envDefault:"edge"` // "edge" | "mimo" | "os"
	Voice    string `json:"voice" env:"TTS_VOICE" envDefault:"苏打"`
	APIKey   string `json:"api_key" env:"TTS_API_KEY"`
	Style    string `json:"style" env:"TTS_STYLE"`
	AutoPlay bool   `json:"auto_play" env:"TTS_AUTO_PLAY" envDefault:"true"`
}

type Config struct {
	Port                   int    `json:"port" env:"DAILY_PORT" envDefault:"8080"`
	SQLiteDSN              string `json:"sqlite_dsn" env:"DAILY_SQLITE_DSN" envDefault:"./data/daily.db"`
	StorageDir             string `json:"storage_dir" env:"DAILY_STORAGE_DIR" envDefault:"./data/storage"`
	LogLevel               string `json:"log_level" env:"DAILY_LOG_LEVEL" envDefault:"info"`
	BootstrapAdminUsername string `json:"bootstrap_admin_username" env:"DAILY_BOOTSTRAP_ADMIN_USERNAME" envDefault:"admin"`
	BootstrapAdminPassword string `json:"bootstrap_admin_password" env:"DAILY_BOOTSTRAP_ADMIN_PASSWORD" envDefault:"admin"`
	BootstrapDemoMemos     bool   `json:"bootstrap_demo_memos" env:"DAILY_BOOTSTRAP_DEMO_MEMOS" envDefault:"false"`
	BootstrapDemoMemosPath string `json:"bootstrap_demo_memos_path" env:"DAILY_BOOTSTRAP_DEMO_MEMOS_PATH" envDefault:"./docs/seeds/architecture_memos.json"`
	Theme                  string `json:"theme" env:"TUI_THEME" envDefault:"ocean"`

	TTS TTSConfig `json:"tts"`
}

// Load loads configuration via strata: defaults → .env → ~/.xkit/config.json["daily"] → ~/.daily/config.json → env vars.
func Load() (*Config, ConfigSources, error) {
	cfg := &Config{}

	home, _ := os.UserHomeDir()
	opts := strata.Options{
		Namespace: "daily",
	}
	if home != "" {
		opts.ConfigFile = filepath.Join(home, ".daily", "config.json")
	}

	sources, err := strata.Load(cfg, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	// Normalize paths
	cfg.SQLiteDSN = expandHome(cfg.SQLiteDSN)
	cfg.BootstrapDemoMemosPath = expandHome(cfg.BootstrapDemoMemosPath)

	// Auto-compute storage dir if not explicitly set
	if cfg.StorageDir == "./data/storage" && cfg.SQLiteDSN != "./data/daily.db" {
		cfg.StorageDir = filepath.Join(filepath.Dir(cfg.SQLiteDSN), "storage")
	}

	// Validate
	if cfg.SQLiteDSN == "" {
		return nil, nil, fmt.Errorf("DAILY_SQLITE_DSN is required")
	}

	return cfg, sources, nil
}

// Save writes the config to ~/.daily/config.json via strata.
func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	return strata.Save(cfg, strata.Options{
		Namespace:  "daily",
		ConfigFile: filepath.Join(home, ".daily", "config.json"),
	})
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
