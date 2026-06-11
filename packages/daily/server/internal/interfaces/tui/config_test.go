package tui

import (
	"testing"

	"daily/internal/infrastructure/config"
)

func TestVoiceOptions(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantMin  int
	}{
		{name: "edge", provider: "edge", wantMin: 3},
		{name: "edge-tts", provider: "edge-tts", wantMin: 3},
		{name: "os", provider: "os", wantMin: 1},
		{name: "mimo", provider: "mimo", wantMin: 3},
		{name: "unknown", provider: "unknown", wantMin: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VoiceOptions(tt.provider)
			if len(got) < tt.wantMin {
				t.Errorf("VoiceOptions(%q) returned %d options, want at least %d", tt.provider, len(got), tt.wantMin)
			}
		})
	}
}

func TestApplyConfigItem(t *testing.T) {
	cfg := &config.Config{}
	cfg.Theme = "ocean"
	cfg.TTS.Provider = "edge"
	cfg.TTS.Voice = "zh-CN-XiaoxiaoNeural"

	ApplyConfigItem(cfg, ConfigItem{Key: "theme", Value: "ocean"}, "sunset")
	if cfg.Theme != "sunset" {
		t.Errorf("theme = %q, want %q", cfg.Theme, "sunset")
	}

	ApplyConfigItem(cfg, ConfigItem{Key: "tts.provider", Value: "edge"}, "mimo")
	if cfg.TTS.Provider != "mimo" {
		t.Errorf("provider = %q, want %q", cfg.TTS.Provider, "mimo")
	}
	if cfg.TTS.Voice == "" {
		t.Error("voice should be reset to first option for new provider")
	}

	ApplyConfigItem(cfg, ConfigItem{Key: "tts.auto_play", Value: "false"}, "true")
	if !cfg.TTS.AutoPlay {
		t.Error("auto_play should be true")
	}
}

func TestToggleConfigItem(t *testing.T) {
	cfg := &config.Config{}
	cfg.TTS.AutoPlay = false

	ToggleConfigItem(cfg, ConfigItem{Key: "tts.auto_play"})
	if !cfg.TTS.AutoPlay {
		t.Error("auto_play should be toggled to true")
	}

	ToggleConfigItem(cfg, ConfigItem{Key: "tts.auto_play"})
	if cfg.TTS.AutoPlay {
		t.Error("auto_play should be toggled back to false")
	}
}

func TestCycleSelectItem(t *testing.T) {
	cfg := &config.Config{}
	cfg.TTS.Provider = "edge"

	item := ConfigItem{
		Key:     "tts.provider",
		Value:   "edge",
		Options: []string{"edge", "mimo", "os"},
	}

	result := CycleSelectItem(cfg, item)
	if result != "mimo" {
		t.Errorf("first cycle: expected mimo, got %s", result)
	}
	if cfg.TTS.Provider != "mimo" {
		t.Errorf("cfg.TTS.Provider = %q, want %q", cfg.TTS.Provider, "mimo")
	}

	result = CycleSelectItem(cfg, item)
	if result != "os" {
		t.Errorf("second cycle: expected os, got %s", result)
	}

	result = CycleSelectItem(cfg, item)
	if result != "edge" {
		t.Errorf("third cycle back to edge, got %s", result)
	}
}
