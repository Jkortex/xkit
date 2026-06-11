package tui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type MarkdownBlock struct {
	Type    string
	Content string
	Raw     string
}

func SimplifyTime(t string) string {
	layouts := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, t); err == nil {
			local := parsed.Local()
			return local.Format("01-02 15:04")
		}
	}
	t = strings.ReplaceAll(t, "T", " ")
	t = strings.TrimSuffix(t, "Z")
	if idx := strings.Index(t, "."); idx != -1 {
		t = t[:idx]
	}
	if idx := strings.Index(t, "+"); idx != -1 {
		t = t[:idx]
	}
	if len(t) >= 19 {
		return t[5:19]
	}
	return t
}

func SplitTags(s string) []string {
	var tags []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func AnySlice[T any](s []T) []any {
	r := make([]any, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r
}

func IsChineseText(text string) bool {
	var chineseCount int
	var totalCount int
	for _, r := range text {
		if r >= 0x20 && r <= 0x7e {
			totalCount++
			continue
		}
		if r >= 0x4e00 && r <= 0x9fff {
			chineseCount++
			totalCount++
		}
	}
	if totalCount == 0 {
		return false
	}
	return float64(chineseCount)/float64(totalCount) > 0.15
}

func IsAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r > 127
}

func IsTagChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '/' || r == ':' || r > 127
}

func CleanMarkdownForTTS(text string) string {
	text = regexp.MustCompile("```[a-zA-Z0-9]*\n?").ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "```", "")

	text = regexp.MustCompile(`(?m)^#+\s*`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`(?m)^\s*[-*+]\s+`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`(?m)^\s*\d+\.\s+`).ReplaceAllString(text, "")

	text = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(text, "$1")

	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")

	text = regexp.MustCompile(`\[[ xX]\]\s*`).ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "☑", "")
	text = strings.ReplaceAll(text, "☐", "")

	return strings.TrimSpace(text)
}

func ExtractTranslation(content string) string {
	var transMarker, vocabMarker, syntaxMarker string
	if strings.Contains(content, "[T]") {
		transMarker = "[T]"
		vocabMarker = "[V]"
		syntaxMarker = "[S]"
	} else if strings.Contains(content, "## Translation") {
		transMarker = "## Translation"
		vocabMarker = "## Vocabulary"
		syntaxMarker = "## Grammar"
	} else if strings.Contains(content, "## 翻译") {
		transMarker = "## 翻译"
		vocabMarker = "## 词汇"
		syntaxMarker = "## 语法"
	} else {
		return ""
	}

	idx := strings.Index(content, transMarker)
	if idx == -1 {
		return ""
	}
	sub := content[idx+len(transMarker):]

	vocabIdx := strings.Index(sub, vocabMarker)
	syntaxIdx := strings.Index(sub, syntaxMarker)

	endIdx := len(sub)
	if vocabIdx != -1 && vocabIdx < endIdx {
		endIdx = vocabIdx
	}
	if syntaxIdx != -1 && syntaxIdx < endIdx {
		endIdx = syntaxIdx
	}

	return strings.TrimSpace(sub[:endIdx])
}

func GetShortSummary(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed == "[O]" || trimmed == "[T]" || trimmed == "[V]" || trimmed == "[S]" ||
			trimmed == "[Original]" || trimmed == "[Translation]" || trimmed == "[Vocabulary]" || trimmed == "[Syntax]" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			headerText := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			lowerHeader := strings.ToLower(headerText)
			if lowerHeader == "translation" || lowerHeader == "vocabulary" || lowerHeader == "grammar" || lowerHeader == "details" ||
				lowerHeader == "翻译" || lowerHeader == "词汇" || lowerHeader == "语法" || lowerHeader == "详解" ||
				lowerHeader == "original" || lowerHeader == "original text" || lowerHeader == "原文" {
				continue
			}
		}
		summary := strings.TrimLeft(trimmed, "#* \t")
		if summary != "" {
			return summary
		}
	}
	return "Empty Memo"
}

func ParseMarkdownBlocks(content string) []MarkdownBlock {
	var blocks []MarkdownBlock
	lines := strings.Split(content, "\n")

	var currentBlock []string
	currentType := "paragraph"
	inCode := false

	flush := func() {
		if len(currentBlock) == 0 {
			return
		}
		raw := strings.Join(currentBlock, "\n")
		cleaned := CleanMarkdownForTTS(raw)
		if strings.TrimSpace(cleaned) != "" {
			blocks = append(blocks, MarkdownBlock{
				Type:    currentType,
				Content: cleaned,
				Raw:     raw,
			})
		}
		currentBlock = nil
		currentType = "paragraph"
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				currentBlock = append(currentBlock, line)
				flush()
				inCode = false
			} else {
				flush()
				inCode = true
				currentType = "code"
				currentBlock = append(currentBlock, line)
			}
			continue
		}

		if inCode {
			currentBlock = append(currentBlock, line)
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			flush()
			currentType = "header"
			currentBlock = append(currentBlock, line)
			flush()
			continue
		}

		if trimmed == "" {
			flush()
			continue
		}

		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") ||
			(strings.Contains(trimmed, ". ") && len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9') {
			flush()
			currentType = "list"
			currentBlock = append(currentBlock, line)
			flush()
			continue
		}

		currentBlock = append(currentBlock, line)
	}
	flush()
	return blocks
}

func WrapLine(line string, max int) []string {
	if max <= 0 {
		return []string{line}
	}
	if strings.TrimSpace(line) == "" {
		return []string{""}
	}

	leadingSpaces := ""
	for _, r := range line {
		if r == ' ' || r == '\t' {
			leadingSpaces += string(r)
		} else {
			break
		}
	}

	content := line[len(leadingSpaces):]
	words := strings.Split(content, " ")

	var result []string
	var curr strings.Builder
	curr.WriteString(leadingSpaces)

	wordWidth := func(w string) int { return lipgloss.Width(w) }

	for _, word := range words {
		if word == "" {
			if curr.Len() > len(leadingSpaces) {
				curr.WriteByte(' ')
			}
			continue
		}

		ww := wordWidth(word)
		avail := max - wordWidth(leadingSpaces)

		if ww > avail {
			if curr.Len() > len(leadingSpaces) {
				result = append(result, curr.String())
				curr.Reset()
				curr.WriteString(leadingSpaces)
			}
			r := []rune(word)
			for len(r) > 0 {
				limit := 0
				for limit < len(r) {
					nextW := wordWidth(string(r[:limit+1]))
					if nextW > avail {
						break
					}
					limit++
				}
				if limit == 0 {
					limit = 1
				}
				result = append(result, leadingSpaces+string(r[:limit]))
				r = r[limit:]
			}
			curr.WriteString(leadingSpaces)
			continue
		}

		spaceNeed := 0
		if curr.Len() > len(leadingSpaces) {
			spaceNeed = 1
		}
		if wordWidth(curr.String())+spaceNeed+ww > max {
			result = append(result, curr.String())
			curr.Reset()
			curr.WriteString(leadingSpaces)
			curr.WriteString(word)
		} else {
			if spaceNeed > 0 {
				curr.WriteByte(' ')
			}
			curr.WriteString(word)
		}
	}

	if curr.Len() > len(leadingSpaces) {
		result = append(result, curr.String())
	}
	return result
}

func TruncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	w := 0
	runes := []rune(s)
	for i, r := range runes {
		rw := lipgloss.Width(string(r))
		if w+rw > maxWidth {
			return string(runes[:i])
		}
		w += rw
	}
	return s
}

func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("clip")
	} else if IsWSL() {
		cmd = exec.Command("clip.exe")
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("pbcopy")
	} else {
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--input", "--clipboard")
		} else {
			return fmt.Errorf("no clipboard tool found (install xclip or xsel)")
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := fmt.Fprint(stdin, text); err != nil {
		return err
	}
	stdin.Close()
	return cmd.Wait()
}
