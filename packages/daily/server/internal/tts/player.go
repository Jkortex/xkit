package tts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PlayAudio plays a WAV file using the platform-native player.
// Supports cancellation via context.
func PlayAudio(ctx context.Context, filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		if isWSL() {
			// WSL: use aplay if available, otherwise fall back to PowerShell
			cmd = exec.CommandContext(ctx, "aplay", filePath)
		} else {
			// Native Windows: use PowerShell System.Windows.Media.MediaPlayer to support both WAV and MP3 robustly
			psScript := fmt.Sprintf(
				"Add-Type -AssemblyName PresentationCore; "+
					"$p = New-Object System.Windows.Media.MediaPlayer; "+
					"$p.Open([Uri]'%s'); "+
					"$p.Play(); "+
					"$start = Get-Date; "+
					"while ($true) { "+
					"  $pos = $p.Position; "+
					"  $duration = $p.NaturalDuration; "+
					"  if ($duration.HasTimeSpan) { "+
					"    if ($pos -ge $duration.TimeSpan) { break }; "+
					"  } else { "+
					"    if (((Get-Date) - $start).TotalSeconds -gt 5) { break }; "+
					"  }; "+
					"  if (((Get-Date) - $start).TotalSeconds -gt 300) { break }; "+
					"  Start-Sleep -Milliseconds 100; "+
					"}",
				strings.ReplaceAll(filePath, "'", "''"),
			)
			cmd = exec.CommandContext(ctx, "powershell.exe", "-Command", psScript)
		}
	case "darwin":
		cmd = exec.CommandContext(ctx, "afplay", filePath)
	default: // linux
		if _, err := exec.LookPath("ffplay"); err == nil {
			// ffplay handles both WAV and MP3 — needed for edge-tts (MP3)
			cmd = exec.CommandContext(ctx, "ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", filePath)
		} else {
			cmd = exec.CommandContext(ctx, "aplay", filePath)
		}
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("play audio: %v, output: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// SpeakOSTTS speaks text using the platform-native TTS engine (fallback).
func SpeakOSTTS(text string) {
	if text == "" {
		return
	}
	isChinese := false
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			isChinese = true
			break
		}
	}

	go func() {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			if isWSL() {
				lang := "en"
				if isChinese {
					lang = "zh"
				}
				cmd = exec.Command("spd-say", "-l", lang, text)
			} else {
				ttsText := strings.ReplaceAll(text, "'", "''")
				langPattern := "en-*"
				if isChinese {
					langPattern = "zh-*"
				}
				psCmd := fmt.Sprintf(
					"Add-Type -AssemblyName System.Speech; "+
						"$s = New-Object System.Speech.Synthesis.SpeechSynthesizer; "+
						"$v = $s.GetInstalledVoices() | Where-Object { $_.VoiceInfo.Culture.Name -like '%s' } | Select-Object -First 1; "+
						"if ($v) { $s.SelectVoice($v.VoiceInfo.Name) }; "+
						"$s.Speak('%s')",
					langPattern,
					ttsText,
				)
				cmd = exec.Command("powershell.exe", "-Command", psCmd)
			}
		case "darwin":
			cmd = exec.Command("say", text)
		default:
			lang := "en"
			if isChinese {
				lang = "zh"
			}
			cmd = exec.Command("spd-say", "-l", lang, text)
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "TTS fallback error: %v, output: %s\n", err, string(out))
		}
	}()
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	return err == nil && strings.Contains(strings.ToLower(string(data)), "microsoft")
}
