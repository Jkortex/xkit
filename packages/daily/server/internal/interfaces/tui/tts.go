package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"daily/internal/infrastructure/config"
	"daily/internal/tts"
	edgetts "github.com/hecx333/edge-tts-go"
)

func SpeakText(text string) {
	SpeakWithCache(text, "", nil, "T", nil)
}

func SpeakWithCache(text string, memoUUID string, expiresAt *time.Time, section string, cfg *config.TTSConfig) {
	reg := regexp.MustCompile(`[^\p{Han}a-zA-Z0-9\s.,'!?，。？！（）“”《》、：；-]`)
	sanitized := reg.ReplaceAllString(text, "")
	if sanitized == "" {
		return
	}

	if section == "" {
		section = "T"
	}

	provider := "mimo"
	apiKey := ""
	voice := "苏打"
	style := ""

	if cfg != nil {
		provider = cfg.Provider
		apiKey = cfg.APIKey
		voice = cfg.Voice
		style = cfg.Style
	} else {
		provider = os.Getenv("TTS_PROVIDER")
		if provider == "" {
			provider = "mimo"
		}
		apiKey = os.Getenv("TTS_API_KEY")
		voice = os.Getenv("TTS_VOICE")
		if voice == "" {
			voice = "苏打"
		}
		style = os.Getenv("TTS_STYLE")
	}

	if provider == "os" || (provider == "mimo" && apiKey == "") {
		tts.SpeakOSTTS(sanitized)
		return
	}

	cacheDir, err := tts.CacheDir()
	if err != nil {
		tts.SpeakOSTTS(sanitized)
		return
	}

	if memoUUID != "" {
		if cached := tts.FindCache(cacheDir, memoUUID, section); cached != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			tts.PlayAudio(ctx, cached)
			return
		}
	}

	var wavBytes []byte
	var synthesisErr error

	if provider == "edge" || provider == "edge-tts" {
		edgeVoice := voice
		if !strings.Contains(edgeVoice, "Neural") {
			if IsChineseText(sanitized) {
				edgeVoice = "zh-CN-XiaoxiaoNeural"
			} else {
				edgeVoice = "en-US-EmmaMultilingualNeural"
			}
		}
		tObj := edgetts.NewTTS(edgetts.WithVoice(edgeVoice))
		wavBytes, synthesisErr = tObj.Speak(sanitized)
	} else {
		wavBytes, synthesisErr = tts.CallMiMoTTS(sanitized, apiKey, voice, style)
	}

	if synthesisErr != nil {
		fmt.Fprintf(os.Stderr, "TTS synthesis error: %v, falling back to OS TTS\n", synthesisErr)
		tts.SpeakOSTTS(sanitized)
		return
	}

	if memoUUID != "" {
		cacheKey := tts.CacheKey(memoUUID, section, expiresAt)
		cachePath := filepath.Join(cacheDir, cacheKey)
		_ = os.WriteFile(cachePath, wavBytes, 0644)
	}

	tmpFile, err := os.CreateTemp("", "daily-tts-*.wav")
	if err != nil {
		tts.SpeakOSTTS(sanitized)
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(wavBytes); err != nil {
		tmpFile.Close()
		tts.SpeakOSTTS(sanitized)
		return
	}
	tmpFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	tts.PlayAudio(ctx, tmpPath)
}

func ShowCacheStatus(memoUUID, section string, expiresAt *time.Time) string {
	cacheDir, err := tts.CacheDir()
	if err != nil {
		return "TTS cache: unavailable"
	}

	status := tts.GetCacheStatus(cacheDir, memoUUID, section)
	if !status.IsValid {
		return "TTS cache: not cached"
	}

	sizeKB := float64(status.Size) / 1024
	if status.ExpiresAt > 0 {
		remaining := time.Until(time.Unix(status.ExpiresAt, 0))
		if remaining < time.Hour {
			return fmt.Sprintf("TTS cache: cached (%.0fKB, expires in %dm)", sizeKB, int(remaining.Minutes()))
		}
		return fmt.Sprintf("TTS cache: cached (%.0fKB, expires in %dh)", sizeKB, int(remaining.Hours()))
	}
	info, err := os.Stat(status.FilePath)
	if err == nil {
		age := time.Since(info.ModTime())
		return fmt.Sprintf("TTS cache: cached (%.0fKB, age %dh, 7d default TTL)", sizeKB, int(age.Hours()))
	}
	return fmt.Sprintf("TTS cache: cached (%.0fKB)", sizeKB)
}


