package cmd

import (
	"fmt"

	"daily/internal/infrastructure/config"
	tui "daily/internal/interfaces/tui"
)

type configItem = tui.ConfigItem

func buildConfigItems(cfg *config.Config, sources config.ConfigSources) []configItem {
	maskKey := func(k string) string {
		if len(k) <= 4 {
			return "****"
		}
		return "****" + k[len(k)-4:]
	}

	items := []configItem{
		{Label: "Theme", Key: "theme", Value: cfg.Theme, Type: "select", Options: ThemeNames(), Source: sources["theme"]},
		{Label: "─ TTS ─", Key: "_separator", Type: "separator"},
		{Label: "Provider", Key: "tts.provider", Value: cfg.TTS.Provider, Type: "select", Options: []string{"mimo", "edge", "os"}, Source: sources["tts.provider"]},
		{Label: "Voice", Key: "tts.voice", Value: cfg.TTS.Voice, Type: "select", Options: []string{"苏打", "白桦", "冰糖", "茉莉", "Mia", "Chloe", "Milo", "Dean"}, Source: sources["tts.voice"]},
		{Label: "API Key", Key: "tts.api_key", Value: maskKey(cfg.TTS.APIKey), Type: "string", Source: sources["tts.api_key"]},
		{Label: "Style", Key: "tts.style", Value: cfg.TTS.Style, Type: "select", Options: []string{"", "磁性", "温柔", "活泼", "严肃", "慵懒", "新闻播报"}, Source: sources["tts.style"]},
		{Label: "Auto Play", Key: "tts.auto_play", Value: fmt.Sprintf("%v", cfg.TTS.AutoPlay), Type: "bool", Source: sources["tts.auto_play"]},
	}

	// Mark env-sourced as read-only, skip separators
	for i := range items {
		if items[i].Type == "separator" {
			continue
		}
		if items[i].Source == config.SourceEnv {
			items[i].ReadOnly = true
		}
		// Empty value placeholder
		if items[i].Value == "" {
			items[i].Value = "(empty)"
		}
	}

	return items
}
