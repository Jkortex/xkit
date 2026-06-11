package tui

import (
	"fmt"

	"daily/internal/infrastructure/config"
)

func VoiceOptions(provider string) []string {
	switch provider {
	case "edge", "edge-tts":
		return []string{"zh-CN-XiaoxiaoNeural", "zh-CN-YunxiNeural", "zh-CN-XiaoyiNeural", "zh-CN-YunjianNeural", "en-US-EmmaMultilingualNeural", "en-US-GuyNeural"}
	case "os":
		return []string{"(system default)"}
	default:
		return []string{"苏打", "白桦", "冰糖", "茉莉", "Mia", "Chloe", "Milo", "Dean"}
	}
}

func ApplyConfigItem(cfg *config.Config, item ConfigItem, value string) {
	switch item.Key {
	case "theme":
		cfg.Theme = value
	case "tts.provider":
		cfg.TTS.Provider = value
		if opts := VoiceOptions(value); len(opts) > 0 {
			cfg.TTS.Voice = opts[0]
		}
	case "tts.voice":
		cfg.TTS.Voice = value
	case "tts.api_key":
		cfg.TTS.APIKey = value
	case "tts.style":
		cfg.TTS.Style = value
	case "tts.auto_play":
		cfg.TTS.AutoPlay = value == "true" || value == "True" || value == "1"
	}
}

func ToggleConfigItem(cfg *config.Config, item ConfigItem) {
	if item.Key == "tts.auto_play" {
		cfg.TTS.AutoPlay = !cfg.TTS.AutoPlay
	}
}

func configValueForKey(cfg *config.Config, key string) string {
	switch key {
	case "theme":
		return cfg.Theme
	case "tts.provider":
		return cfg.TTS.Provider
	case "tts.voice":
		return cfg.TTS.Voice
	case "tts.api_key":
		return cfg.TTS.APIKey
	case "tts.style":
		return cfg.TTS.Style
	case "tts.auto_play":
		return fmt.Sprintf("%v", cfg.TTS.AutoPlay)
	}
	return ""
}

func CycleSelectItem(cfg *config.Config, item ConfigItem) string {
	if len(item.Options) == 0 {
		return item.Value
	}
	current := configValueForKey(cfg, item.Key)
	idx := -1
	for i, opt := range item.Options {
		if opt == current || (opt == "" && current == "") {
			idx = i
			break
		}
	}
	if idx < 0 {
		idx = 0
	}
	next := (idx + 1) % len(item.Options)
	newVal := item.Options[next]
	ApplyConfigItem(cfg, item, newVal)
	return newVal
}
