package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderInline(text string, state *InlineState, style *MarkdownStyle) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	n := len(runes)
	var sb strings.Builder

	boldStyle := style.BoldStyle
	italicStyle := style.ItalicStyle
	codeStyle := style.InlineCodeStyle
	strikeStyle := style.StrikeStyle

	var currentChunk []rune
	lastInBold := state.InBold
	lastInItalic := state.InItalic
	lastInCode := state.InCode
	lastInStrike := state.InStrike

	flush := func() {
		if len(currentChunk) == 0 {
			return
		}
		chunkStr := string(currentChunk)
		currentChunk = nil

		if lastInCode {
			sb.WriteString(codeStyle.Render(chunkStr))
			return
		}

		var st lipgloss.Style
		hasStyle := false
		if lastInBold {
			st = boldStyle
			hasStyle = true
		}
		if lastInItalic {
			if hasStyle {
				st = st.Italic(true)
			} else {
				st = italicStyle
				hasStyle = true
			}
		}
		if lastInStrike {
			if hasStyle {
				st = st.Strikethrough(true).Foreground(lipgloss.Color("#6b7280"))
			} else {
				st = strikeStyle
				hasStyle = true
			}
		}

		if hasStyle {
			sb.WriteString(st.Render(chunkStr))
		} else {
			sb.WriteString(chunkStr)
		}
	}

	i := 0
	for i < n {
		var change = false
		var newInBold = state.InBold
		var newInItalic = state.InItalic
		var newInCode = state.InCode
		var newInStrike = state.InStrike
		var skipCount = 0

		if runes[i] == '`' {
			newInCode = !state.InCode
			change = true
			skipCount = 1
		} else if state.InCode {
		} else if i+1 < n && runes[i] == '~' && runes[i+1] == '~' {
			newInStrike = !state.InStrike
			change = true
			skipCount = 2
		} else if i+1 < n && runes[i] == '*' && runes[i+1] == '*' {
			newInBold = !state.InBold
			change = true
			skipCount = 2
		} else if i+1 < n && runes[i] == '_' && runes[i+1] == '_' {
			newInBold = !state.InBold
			change = true
			skipCount = 2
		} else if runes[i] == '*' {
			newInItalic = !state.InItalic
			change = true
			skipCount = 1
		} else if runes[i] == '_' {
			newInItalic = !state.InItalic
			change = true
			skipCount = 1
		} else if runes[i] == '[' {
			closeBracketIdx := -1
			for j := i + 1; j < n; j++ {
				if runes[j] == ']' {
					closeBracketIdx = j
					break
				}
			}
			if closeBracketIdx != -1 && closeBracketIdx+1 < n && runes[closeBracketIdx+1] == '(' {
				closeParenIdx := -1
				for j := closeBracketIdx + 2; j < n; j++ {
					if runes[j] == ')' {
						closeParenIdx = j
						break
					}
				}
				if closeParenIdx != -1 {
					flush()
					linkText := string(runes[i+1 : closeBracketIdx])
					linkURL := string(runes[closeBracketIdx+2 : closeParenIdx])

					sb.WriteString(style.LinkStyle.Render(linkText))
					sb.WriteString(" ")
					sb.WriteString(style.URLStyle.Render("(" + linkURL + ")"))

					i = closeParenIdx + 1
					continue
				}
			}
		} else if runes[i] == '#' {
			isTagStart := i == 0 || runes[i-1] == ' ' || runes[i-1] == '\t' || runes[i-1] == '\n'
			if isTagStart && i+1 < n && IsAlphaNum(runes[i+1]) {
				flush()
				var tagRunes []rune
				tagRunes = append(tagRunes, '#')
				i++
				for i < n && IsTagChar(runes[i]) {
					tagRunes = append(tagRunes, runes[i])
					i++
				}
				sb.WriteString(style.TagStyle.Render(string(tagRunes)))
				continue
			}
		}

		if change {
			flush()
			state.InBold = newInBold
			state.InItalic = newInItalic
			state.InCode = newInCode
			state.InStrike = newInStrike

			lastInBold = state.InBold
			lastInItalic = state.InItalic
			lastInCode = state.InCode
			lastInStrike = state.InStrike

			i += skipCount
		} else {
			currentChunk = append(currentChunk, runes[i])
			i++
		}
	}
	flush()

	return sb.String()
}

func RenderMarkdown(content string, width int, style *MarkdownStyle) []string {
	lines := strings.Split(content, "\n")
	var result []string

	inCodeBlock := false
	var codeBlockLines []string

	h1Style := style.H1Style
	h2Style := style.H2Style
	h3Style := style.H3Style
	h4Style := style.H4Style

	codeBg := style.CodeBlockStyle

	quoteStyle := style.QuoteStyle
	quoteBorder := style.QuoteBorder

	bulletColor := style.BulletStyle
	numColor := style.NumStyle

	taskDone := style.TaskDoneStyle
	taskTodo := style.TaskTodoStyle

	for _, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)

		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				inCodeBlock = false
				for _, cl := range codeBlockLines {
					wrapped := WrapLine(cl, width-6)
					for _, wl := range wrapped {
						wLen := lipgloss.Width(wl)
						padSize := width - 6 - wLen
						if padSize < 0 {
							padSize = 0
						}
						padded := wl + strings.Repeat(" ", padSize)
						result = append(result, "    "+codeBg.Render(padded))
					}
				}
				codeBlockLines = nil
			} else {
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, rawLine)
			continue
		}

		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			ruleWidth := width - 4
			if ruleWidth < 10 {
				ruleWidth = 10
			}
			result = append(result, "  "+style.HR.Render(strings.Repeat("─", ruleWidth)))
			continue
		}

		if strings.HasPrefix(rawLine, "# ") {
			text := strings.TrimPrefix(rawLine, "# ")
			wrapped := WrapLine(text, width-6)
			for idx, wl := range wrapped {
				if idx == 0 {
					result = append(result, "  ◈ "+h1Style.Render(wl))
				} else {
					result = append(result, "    "+h1Style.Render(wl))
				}
			}
			continue
		}
		if strings.HasPrefix(rawLine, "## ") {
			text := strings.TrimPrefix(rawLine, "## ")
			wrapped := WrapLine(text, width-6)
			for idx, wl := range wrapped {
				if idx == 0 {
					result = append(result, "  ◇ "+h2Style.Render(wl))
				} else {
					result = append(result, "    "+h2Style.Render(wl))
				}
			}
			continue
		}
		if strings.HasPrefix(rawLine, "### ") {
			text := strings.TrimPrefix(rawLine, "### ")
			wrapped := WrapLine(text, width-6)
			for idx, wl := range wrapped {
				if idx == 0 {
					result = append(result, "  ▪ "+h3Style.Render(wl))
				} else {
					result = append(result, "    "+h3Style.Render(wl))
				}
			}
			continue
		}
		if strings.HasPrefix(rawLine, "#### ") {
			text := strings.TrimPrefix(rawLine, "#### ")
			wrapped := WrapLine(text, width-6)
			for idx, wl := range wrapped {
				if idx == 0 {
					result = append(result, "  ▫ "+h4Style.Render(wl))
				} else {
					result = append(result, "    "+h4Style.Render(wl))
				}
			}
			continue
		}

		if strings.HasPrefix(trimmed, ">") {
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			wrapped := WrapLine(text, width-6)
			state := &InlineState{}
			for _, wl := range wrapped {
				result = append(result, "  "+quoteBorder.String()+quoteStyle.Render(RenderInline(wl, state, style)))
			}
			continue
		}

		isTask := false
		isDone := false
		taskText := ""
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "* [ ] ") {
			isTask = true
			isDone = false
			taskText = trimmed[6:]
		} else if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "* [x] ") ||
			strings.HasPrefix(trimmed, "- [X] ") || strings.HasPrefix(trimmed, "* [X] ") {
			isTask = true
			isDone = true
			taskText = trimmed[6:]
		}

		if isTask {
			wrapped := WrapLine(taskText, width-6)
			state := &InlineState{}
			for idx, wl := range wrapped {
				renderedWl := RenderInline(wl, state, style)
				if idx == 0 {
					box := taskTodo.String()
					if isDone {
						box = taskDone.String()
					}
					result = append(result, "  "+box+renderedWl)
				} else {
					result = append(result, "    "+renderedWl)
				}
			}
			continue
		}

		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			text := trimmed[2:]
			wrapped := WrapLine(text, width-6)
			state := &InlineState{}
			for idx, wl := range wrapped {
				renderedWl := RenderInline(wl, state, style)
				if idx == 0 {
					result = append(result, "  "+bulletColor.String()+renderedWl)
				} else {
					result = append(result, "    "+renderedWl)
				}
			}
			continue
		}

		isNumList := false
		numPrefix := ""
		numText := ""
		if dotIdx := strings.Index(trimmed, ". "); dotIdx > 0 && dotIdx < 5 {
			numPart := trimmed[:dotIdx]
			isAllDigits := true
			for _, r := range numPart {
				if r < '0' || r > '9' {
					isAllDigits = false
					break
				}
			}
			if isAllDigits {
				isNumList = true
				numPrefix = numPart + ". "
				numText = trimmed[dotIdx+2:]
			}
		}

		if isNumList {
			wrapped := WrapLine(numText, width-6)
			state := &InlineState{}
			for idx, wl := range wrapped {
				renderedWl := RenderInline(wl, state, style)
				if idx == 0 {
					prefixStr := numColor.Render(numPrefix)
					result = append(result, "  "+prefixStr+renderedWl)
				} else {
					padding := strings.Repeat(" ", len(numPrefix)+2)
					result = append(result, padding+renderedWl)
				}
			}
			continue
		}

		if trimmed == "" {
			result = append(result, "")
		} else {
			wrapped := WrapLine(rawLine, width-2)
			state := &InlineState{}
			for _, wl := range wrapped {
				result = append(result, "  "+RenderInline(wl, state, style))
			}
		}
	}

	if inCodeBlock && len(codeBlockLines) > 0 {
		for _, cl := range codeBlockLines {
			wrapped := WrapLine(cl, width-6)
			for _, wl := range wrapped {
				wLen := lipgloss.Width(wl)
				padSize := width - 6 - wLen
				if padSize < 0 {
					padSize = 0
				}
				padded := wl + strings.Repeat(" ", padSize)
				result = append(result, "    "+codeBg.Render(padded))
			}
		}
	}

	return result
}
