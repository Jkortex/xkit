package cmd

import (
	"fmt"
	"strings"

	"daily/internal/infrastructure/config"
	tui "daily/internal/interfaces/tui"

	"github.com/charmbracelet/lipgloss"
)

// topFixedLines is the number of lines in the top section that are not content:
// titleBar(1) + hr(1) + drawTabs(1) + empty(1)
const topFixedLines = 4

func (m tuiModel) cmdPaletteLinesCount() int {
	filtered := m.getFilteredCommands()
	if len(filtered) == 0 {
		return 5 // input(1), hr(1), "No matching"(1), hr(1), hint(1)
	}
	lines := 4 // input(1), hr(1), hr(1), hint(1)
	maxShow := 4
	startIdx := 0
	if m.cmdSelIdx >= maxShow {
		startIdx = m.cmdSelIdx - maxShow + 1
	}
	if startIdx > 0 {
		lines++
	}
	showCount := len(filtered) - startIdx
	if showCount > maxShow {
		showCount = maxShow
	}
	if showCount < 0 {
		showCount = 0
	}
	lines += showCount
	remaining := len(filtered) - (startIdx + maxShow)
	if remaining > 0 {
		lines++
	}
	return lines
}

func (m tuiModel) bottomOverhead() int {
	h := 0
	// 1. prompt / command palette
	if m.mode == modeCmd || m.mode == modeDSLInput || m.mode == modeTagInput || m.mode == modeCreate {
		h += 1 // dynamicHr before prompt
		if m.mode == modeCmd {
			h += m.cmdPaletteLinesCount()
		} else if m.mode == modeDSLInput {
			h += 2 // "DSL Filter:" input line + bottomLine(dynamicHr)
			if m.cmdHint() != "" {
				h++ // hint line
			}
			h++ // "Esc cancel · Enter filter" hint footer
		} else if m.mode == modeTagInput {
			h += 3 // "Tags:" input line + bottomLine(dynamicHr) + hint footer
		} else if m.mode == modeCreate {
			h += 3 // "New Memo:" input line + bottomLine(dynamicHr) + hint footer
		}
	}

	// 2. footer keys
	h += 1 // dynamicHr before footer
	h += m.footerLinesCount()

	// 3. status message
	if m.statusMsg != "" {
		h++
	}

	return h
}

func (m tuiModel) contentLines() int {
	if m.height <= 0 {
		return 10
	}
	avail := m.height - topFixedLines - m.bottomOverhead()
	if avail < 1 {
		return 1
	}
	return avail
}

// memoHeight returns how many terminal lines memo at index i consumes.
func (m tuiModel) memoHeight(i int) int {
	n := 1 // summary line
	if m.memos[i].Expanded {
		n += 2 // hr lines
		cw := m.width - 8
		if cw < 20 {
			cw = 20
		}
		n += len(tui.RenderMarkdown(m.memos[i].Content, cw, markdownStyleFromTheme(m.theme)))
	}
	return n
}

// cursorLine returns the 0-based line number of the cursor memo's first line.
func (m tuiModel) cursorLine() int {
	line := 0
	for i := 0; i < m.cursorIdx; i++ {
		line += m.memoHeight(i)
	}
	return line
}

// computeOffset returns the line offset that keeps the cursor on screen.
func (m tuiModel) computeOffset() int {
	if m.activeTab == tabTags {
		cl := m.contentLines()
		if cl < 1 {
			cl = 1
		}
		avail := cl
		if len(m.tagsWithCount) > cl {
			avail = cl - 2
			if avail < 1 {
				avail = 1
			}
		}
		off := m.offset
		if m.cursorIdx < off {
			off = m.cursorIdx
		}
		if m.cursorIdx >= off+avail {
			off = m.cursorIdx - avail + 1
		}
		if off < 0 {
			off = 0
		}
		return off
	}

	if len(m.memos) == 0 {
		return 0
	}
	cl := m.contentLines()
	if cl < 1 {
		cl = 1
	}

	totalHeight := m.totalContentHeight()
	if totalHeight <= cl {
		return 0
	}

	cLine := m.cursorLine()
	cHeight := m.memoHeight(m.cursorIdx)

	getMemoAvail := func(o int) int {
		if o == 0 {
			return cl - 1
		}
		if o >= totalHeight-cl+1 {
			return cl - 1
		}
		return cl - 2
	}

	isCursorVisible := func(o int) bool {
		avail := getMemoAvail(o)
		if cHeight >= avail {
			return o == cLine
		}
		return o <= cLine && cLine+cHeight <= o+avail
	}

	if isCursorVisible(m.offset) {
		return m.offset
	}

	if cLine < m.offset {
		if cLine < 0 {
			return 0
		}
		return cLine
	}

	startO := m.offset
	if cLine+cHeight-cl > startO {
		startO = cLine + cHeight - cl
	}
	if startO < 0 {
		startO = 0
	}
	for o := startO; o < totalHeight; o++ {
		if isCursorVisible(o) {
			return o
		}
	}
	return cLine
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}

	dynamicHr := m.dynamicHR()

	var topB strings.Builder
	topLine := func(s string) { topB.WriteString(s); topB.WriteByte('\n') }

	topLine(m.renderTitleBar())
	topLine(dynamicHr)

	// ── config mode ──
	if m.mode == modeConfig || m.mode == modeConfigEdit {
		topLine("  " + m.theme.TopBar.Render("TTS Configuration"))
		topLine(dynamicHr)
		topLine("")

		sourceLabel := func(s config.ConfigSource) string {
			switch s {
			case config.SourceEnv:
				return m.theme.Dim.Render("[env]")
			case config.SourceFile:
				return m.theme.Dim.Render("[file]")
			default:
				return m.theme.Dim.Render("[default]")
			}
		}

		for i, item := range m.configItems {
			if item.Type == "separator" {
				topLine("  " + m.theme.Dim.Render(item.Label))
				continue
			}

			cursor := "  "
			label := item.Label
			val := item.Value
			src := sourceLabel(item.Source)

			if i == m.configIdx {
				cursor = m.theme.Cursor.String() + " "
				label = m.theme.Sel.Render(label)
				if m.mode == modeConfigEdit && i == m.configIdx {
					val = m.theme.Sel.Render(m.configEditBuf) + m.theme.Dim.Render("█")
				} else if item.Type == "select" && i == m.configIdx {
					val = m.theme.Sel.Render(val) + m.theme.Dim.Render("  ← →")
				} else {
					val = m.theme.Sel.Render(val)
				}
			} else {
				label = lipgloss.NewStyle().Foreground(m.theme.Foreground).Render(label)
				val = lipgloss.NewStyle().Foreground(m.theme.Foreground).Render(val)
			}

			ro := ""
			if item.ReadOnly && i == m.configIdx {
				ro = " " + m.theme.Dim.Render("(read-only)")
			}

			topLine(fmt.Sprintf("  %s%-14s %s  %s%s", cursor, label, val, src, ro))
		}

		// Calculate padding so that footer instructions sit at the absolute bottom
		// top: title(1) + hr(1) + empty(1) + config items + empty(1) = 4 + len(items)
		// bottom: hr(1) + hint(1) = 2 lines
		hTop := 4 + len(m.configItems)
		hBottom := 2
		pad := 0
		if m.height > 0 {
			pad = m.height - hTop - hBottom
		}
		if pad < 0 {
			pad = 0
		}
		if pad > 0 {
			topB.WriteString(strings.Repeat("\n", pad))
		}

		topLine(dynamicHr)
		if m.mode == modeConfigEdit {
			topLine(m.theme.Dim.Render("  Type to edit · Enter to save · Esc to cancel"))
		} else {
			topLine(m.theme.Dim.Render("  Esc back · ↑↓ navigate · Enter edit/cycle · s toggle"))
		}
		return topB.String()
	}

	// ── main content ──
	topLine(m.drawTabs())
	topLine("")
	topLine(m.renderMainContent())

	// Create bottom section builder
	var bottomB strings.Builder
	bottomLine := func(s string) { bottomB.WriteString(s); bottomB.WriteByte('\n') }

	// ── prompt / command palette fixed at bottom ──
	if promptSection := m.renderPromptSection(dynamicHr); promptSection != "" {
		bottomLine(dynamicHr)
		bottomB.WriteString(promptSection)
	}

	// ── footer ──
	bottomLine(dynamicHr)
	bottomB.WriteString(m.renderFooterSection())

	// Join them together, adding vertical padding to push bottomSection to the absolute bottom of the terminal screen
	topStr := topB.String()
	bottomStr := bottomB.String()

	hTop := len(strings.Split(strings.TrimSuffix(topStr, "\n"), "\n"))
	hBottom := len(strings.Split(strings.TrimSuffix(bottomStr, "\n"), "\n"))

	pad := 0
	if m.height > 0 {
		pad = m.height - hTop - hBottom
	}
	if pad < 0 {
		pad = 0
	}

	var finalB strings.Builder
	finalB.WriteString(topStr)
	if pad > 0 {
		finalB.WriteString(strings.Repeat("\n", pad))
	}
	finalB.WriteString(bottomStr)

	return finalB.String()
}

// dynamicHR returns a horizontal rule spanning the full terminal width.
func (m tuiModel) dynamicHR() string {
	if m.width > 0 {
		return m.theme.HR.Render(strings.Repeat("─", m.width))
	}
	return m.theme.HRRule.String()
}

// renderTitleBar builds the top line: title + memo count + filter badges.
func (m tuiModel) renderTitleBar() string {
	title := m.theme.TopBar.Render("◉  daily tui")
	count := m.theme.Count.Render(fmt.Sprintf("%d memos", len(m.memos)))

	var badges []string
	if m.tagFilter != "" {
		badges = append(badges, m.theme.Filter.Render("@"+m.tagFilter))
	}
	if len(m.tagOr) > 0 {
		badges = append(badges, m.theme.Filter.Render("or:"+strings.Join(m.tagOr, ",")))
	}
	if len(m.tagAnd) > 0 {
		badges = append(badges, m.theme.Filter.Render("and:"+strings.Join(m.tagAnd, ",")))
	}
	if len(m.tagExcl) > 0 {
		badges = append(badges, m.theme.Filter.Render("not:"+strings.Join(m.tagExcl, ",")))
	}
	if m.searchTxt != "" {
		badges = append(badges, m.theme.Filter.Render("/"+m.searchTxt))
	}
	if m.fromDt != "" {
		badges = append(badges, m.theme.Count.Render("from:"+m.fromDt))
	}
	if m.toDt != "" {
		badges = append(badges, m.theme.Count.Render("to:"+m.toDt))
	}
	if m.maxItems > 0 && m.maxItems < 100 {
		badges = append(badges, m.theme.Count.Render(fmt.Sprintf("last:%d", m.maxItems)))
	}

	selCount := 0
	for _, sel := range m.selectedMap {
		if sel {
			selCount++
		}
	}
	if selCount > 0 {
		badges = append(badges, m.theme.Create.Render(fmt.Sprintf("selected:%d", selCount)))
	}

	badgeLine := strings.Join(badges, " ")
	return lipgloss.JoinHorizontal(lipgloss.Top, title, "  ", count, "  ", badgeLine)
}

// renderMainContent renders tabs + the active tab's content (or detail mode).
func (m tuiModel) renderMainContent() string {
	showSplit := m.width >= 80 && (m.activeTab == tabInbox || m.activeTab == tabArchived)
	if showSplit {
		leftW := m.width / 3
		if leftW < 25 {
			leftW = 25
		}
		rightW := m.width - leftW - 3
		contentH := m.contentLines()

		leftContent := m.renderMemosListOnly(leftW)
		rightContent := m.renderRightPane(rightW, m.mode == modeDetail, contentH)

		splitView := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(leftW).Height(contentH).Render(leftContent),
			m.theme.HR.Render(" │ "),
			lipgloss.NewStyle().Width(rightW).Height(contentH).Render(rightContent),
		)
		return splitView
	}

	if m.mode == modeDetail {
		return m.renderDetailMode()
	}

	switch m.activeTab {
	case tabInbox, tabArchived:
		return m.renderMemosList()
	case tabTags:
		return m.renderTagsTab()
	case tabStats:
		return m.renderStatsTab()
	default:
		return ""
	}
}

// renderPromptSection renders the bottom prompt overlay (command palette, DSL filter, etc.).
// Returns empty string when no prompt is active.
func (m tuiModel) renderPromptSection(dynamicHr string) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	switch m.mode {
	case modeCmd:
		line("  " + m.theme.Filter.Render("Command Palette:") + " " + m.cmdBuf + m.theme.Dim.Render("█"))
		line(dynamicHr)

		filtered := m.getFilteredCommands()
		if len(filtered) == 0 {
			line("    " + m.theme.Dim.Render("No matching actions found"))
		} else {
			maxShow := 4
			startIdx := 0
			if m.cmdSelIdx >= maxShow {
				startIdx = m.cmdSelIdx - maxShow + 1
			}

			if startIdx > 0 {
				line("    " + m.theme.Dim.Render(fmt.Sprintf("... and %d more actions above ...", startIdx)))
			}

			for i := startIdx; i < len(filtered) && i < startIdx+maxShow; i++ {
				cmd := filtered[i]
				isSel := i == m.cmdSelIdx
				prefix := "  "
				labelStr := cmd.label
				descStr := " - " + cmd.desc
				keyStr := fmt.Sprintf(" [%s] ", cmd.key)

				if isSel {
					prefix = m.theme.Cursor.String() + " "
					labelStr = m.theme.Sel.Render(labelStr)
					descStr = lipgloss.NewStyle().Foreground(m.theme.Foreground).Render(descStr)
					keyStr = m.theme.Filter.Render(keyStr)
				} else {
					labelStr = lipgloss.NewStyle().Foreground(m.theme.Foreground).Render(labelStr)
					descStr = m.theme.Dim.Render(descStr)
					keyStr = m.theme.Dim.Render(keyStr)
				}

				line(prefix + labelStr + descStr + " " + keyStr)
			}

			remaining := len(filtered) - (startIdx + maxShow)
			if remaining > 0 {
				line("    " + m.theme.Dim.Render(fmt.Sprintf("... and %d more actions below ...", remaining)))
			}
		}
		line(dynamicHr)
		line(m.theme.Dim.Render("  Esc exit · ↑↓/Ctrl+p/n nav · Enter execute"))
	case modeDSLInput:
		line("  " + m.theme.Filter.Render("DSL Filter:") + " " + m.cmdBuf + m.theme.Dim.Render("█"))
		hint := m.cmdHint()
		if hint != "" {
			line("  " + hint)
		}
		line(dynamicHr)
		line(m.theme.Dim.Render("  Esc cancel · Enter filter memos"))
	case modeTagInput:
		line("  " + m.theme.HelpKey.Render("Tags:") + " " + m.tagInputBuf + m.theme.Dim.Render("█"))
		line(dynamicHr)
		line(m.theme.Dim.Render("  Esc cancel · Enter save (space separated)"))
	case modeCreate:
		line("  " + m.theme.Create.Render("New Memo:") + " " + m.createBuf + m.theme.Dim.Render("█"))
		line(dynamicHr)
		line(m.theme.Dim.Render("  Esc cancel · Enter submit"))
	}

	return b.String()
}

// renderFooterSection renders footer keys and status message.
func (m tuiModel) renderFooterSection() string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	keys := m.footerKeys()

	var footerLines []string
	currentLine := ""
	currentLineWidth := 0
	sep := m.theme.HelpSep.String()
	sepWidth := lipgloss.Width(sep)

	for _, kv := range keys {
		item := m.theme.HelpKey.Render(kv.k) + m.theme.Dim.Render(" "+kv.d)
		itemWidth := lipgloss.Width(item)

		if currentLine == "" {
			currentLine = item
			currentLineWidth = itemWidth
		} else {
			if m.width > 0 && currentLineWidth+sepWidth+itemWidth > m.width-2 {
				footerLines = append(footerLines, currentLine)
				currentLine = item
				currentLineWidth = itemWidth
			} else {
				currentLine += sep + item
				currentLineWidth += sepWidth + itemWidth
			}
		}
	}
	if currentLine != "" {
		footerLines = append(footerLines, currentLine)
	}

	for _, fl := range footerLines {
		line(fl)
	}

	if m.statusMsg != "" {
		isErr := strings.HasPrefix(m.statusMsg, "Delete") || strings.HasPrefix(m.statusMsg, "Error") || strings.HasPrefix(m.statusMsg, "failed")
		if isErr {
			line("  " + m.theme.Error.Render(m.statusMsg))
		} else {
			line("  " + m.theme.Status.Render(m.statusMsg))
		}
	}

	return b.String()
}

func (m tuiModel) drawTabs() string {
	tabs := []struct {
		t    tabType
		name string
	}{
		{tabInbox, "Inbox (Active)"},
		{tabArchived, "Archived"},
		{tabTags, "Tags list"},
		{tabStats, "Statistics"},
	}

	var renderedTabs []string
	for _, tab := range tabs {
		isActive := m.activeTab == tab.t
		style := m.theme.TabInactive
		if isActive {
			style = m.theme.TabActive
		}
		renderedTabs = append(renderedTabs, style.Render(tab.name))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m tuiModel) renderMemosList() string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	if len(m.memos) == 0 {
		line(m.theme.Dim.Render("  (no memos)"))
		return b.String()
	}

	cl := m.contentLines()
	if cl < 1 {
		cl = 1
	}
	total := m.totalContentHeight()
	remaining := cl
	moreBelow := false

	if m.offset > 0 {
		line(m.theme.Dim.Render(fmt.Sprintf("  ↑ %d lines ..", m.offset)))
		remaining--
	}

	skip := m.offset
	for i := range m.memos {
		h := m.memoHeight(i)
		if skip >= h {
			skip -= h
			continue
		}
		hActual := h
		if skip > 0 {
			hActual = h - skip
		}
		if len(m.memos)-i > 1 && remaining < hActual+1 && i != m.cursorIdx {
			moreBelow = true
			break
		}
		if remaining <= 0 {
			moreBelow = true
			break
		}

		memo := m.memos[i]
		isSel := i == m.cursorIdx

		prefix := "   "
		if isSel {
			prefix = m.theme.Cursor.String() + "  "
		}

		checkMark := ""
		if m.selectedMap != nil && m.selectedMap[memo.UUID] {
			checkMark = m.theme.CheckOn.String()
		}

		summary := tui.GetShortSummary(memo.Content)
		if isSel {
			summary = m.theme.Sel.Render(summary)
		} else {
			summary = tui.RenderInline(summary, &tui.InlineState{}, markdownStyleFromTheme(m.theme))
		}
		extra := ""
		if memo.Expanded {
			extra = "  " + m.theme.ExpIcon.String()
		}

		availWidth := m.width - lipgloss.Width(prefix) - lipgloss.Width(checkMark) - lipgloss.Width(extra) - 2
		if availWidth < 10 {
			availWidth = 10
		}
		if lipgloss.Width(summary) > availWidth {
			summary = tui.TruncateToWidth(summary, availWidth-3) + "..."
		}

		lineContent := prefix + checkMark + summary + extra
		if isSel {
			line(m.theme.SelBg.Render(lineContent))
		} else {
			line(lineContent)
		}
		remaining--

		if memo.Expanded {
			cw := m.width - 8
			if cw < 20 {
				cw = 20
			}
			line(m.theme.DetailHR.String())
			remaining--
			rendered := tui.RenderMarkdown(memo.Content, cw, markdownStyleFromTheme(m.theme))
			maxLines := remaining
			if maxLines < 0 {
				maxLines = 0
			}
			for idx, wl := range rendered {
				if idx >= maxLines {
					moreBelow = true
					break
				}
				line("   " + wl)
				remaining--
			}
			if !moreBelow {
				line(m.theme.DetailHR.String())
				remaining--
			}
		}
	}

	if moreBelow {
		rendered := cl - remaining
		if rendered > cl {
			rendered = cl
		}
		line(m.theme.Dim.Render(fmt.Sprintf("  ↓ %d lines ..", total-m.offset-rendered)))
		remaining--
	}

	// Pad to cl lines so the list height is consistent — avoids layout jitter.
	for remaining > 0 {
		line("")
		remaining--
	}

	return b.String()
}

func (m tuiModel) renderTagsTab() string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	if len(m.tagsWithCount) == 0 {
		line(m.theme.Dim.Render("  (no tags available)"))
		return b.String()
	}

	cl := m.contentLines()
	if cl < 1 {
		cl = 1
	}

	start := m.offset
	if start < 0 {
		start = 0
	}

	remaining := cl
	limit := start + cl
	if len(m.tagsWithCount) > cl {
		limit = start + cl - 2
		if limit < start+1 {
			limit = start + 1
		}
	}

	for i := start; i < len(m.tagsWithCount) && i < limit; i++ {
		tag := m.tagsWithCount[i]
		isSel := i == m.cursorIdx

		prefix := "  "
		tagName := m.theme.Filter.Render("#" + tag.Name)
		tagCount := m.theme.Count.Render(fmt.Sprintf("(%d memos)", tag.Count))

		if isSel {
			prefix = m.theme.Cursor.String() + " "
			tagName = m.theme.Sel.Render("#" + tag.Name)
		} else {
			prefix = "  "
		}

		line(prefix + tagName + " " + tagCount)
		remaining--
	}

	if len(m.tagsWithCount) > cl {
		line("")
		line(m.theme.Dim.Render(fmt.Sprintf("  Total tags: %d (Use j/k to browse, Enter to filter)", len(m.tagsWithCount))))
		remaining -= 2
	}

	for remaining > 0 {
		line("")
		remaining--
	}

	return b.String()
}

func (m tuiModel) renderStatsTab() string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	if m.statsData == nil {
		line(m.theme.Dim.Render("  Loading stats..."))
		return b.String()
	}

	statsBox := fmt.Sprintf(
		"\n"+
			"   Total Memos:     %s\n"+
			"   Total Tags:      %s\n"+
			"   Total Resources: %s\n"+
			"\n"+
			"   Press [r] to reset statistics.",
		m.theme.Sel.Render(fmt.Sprintf("%d", m.statsData.MemosTotal)),
		m.theme.Filter.Render(fmt.Sprintf("%d", m.statsData.TagsTotal)),
		m.theme.Create.Render(fmt.Sprintf("%d", m.statsData.ResourcesTotal)),
	)

	line(m.theme.StatsBorder.Render(statsBox))
	return b.String()
}

func (m tuiModel) renderDetailMode() string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	if len(m.memos) == 0 || m.cursorIdx < 0 || m.cursorIdx >= len(m.memos) {
		line(m.theme.Dim.Render("  No memo selected"))
		return b.String()
	}

	innerW := m.width - 6
	if innerW < 20 {
		innerW = 20
	}
	availH := m.contentLines() - 4
	if availH < 1 {
		availH = 1
	}
	content := m.renderRightPane(innerW, true, availH)

	line(m.theme.DetailBox.Copy().Width(m.width - 2).Render(content))
	return b.String()
}

func (m tuiModel) totalContentHeight() int {
	h := 0
	for i := range m.memos {
		h += m.memoHeight(i)
	}
	return h
}

func (m tuiModel) footerLinesCount() int {
	keys := m.footerKeys()

	lines := 0
	currentLineWidth := 0
	sepWidth := lipgloss.Width(m.theme.HelpSep.String())

	for i, kv := range keys {
		itemWidth := lipgloss.Width(m.theme.HelpKey.Render(kv.k) + " " + kv.d)
		if i == 0 {
			lines = 1
			currentLineWidth = itemWidth
		} else {
			if m.width > 0 && currentLineWidth+sepWidth+itemWidth > m.width-2 {
				lines++
				currentLineWidth = itemWidth
			} else {
				currentLineWidth += sepWidth + itemWidth
			}
		}
	}
	if lines == 0 {
		return 1
	}
	return lines
}

func (m tuiModel) footerKeys() []struct{ k, d string } {
	if m.mode == modeDetail {
		return []struct{ k, d string }{
			{"esc", "back"}, {"e", "edit"}, {"d", "del"}, {"a", "archive"}, {"p", "play"},
		}
	} else if m.mode == modeDeleteConfirm {
		return []struct{ k, d string }{
			{"y", "confirm delete"}, {"n/esc", "cancel"},
		}
	} else if m.activeTab == tabTags {
		return []struct{ k, d string }{
			{"q", "quit"}, {"h/l", "prev/next tab"}, {"enter", "filter tag"}, {"r", "reset"},
		}
	} else if m.activeTab == tabStats {
		return []struct{ k, d string }{
			{"q", "quit"}, {"h/l", "prev/next tab"}, {"r", "reset"},
		}
	}
	return []struct{ k, d string }{
		{"q", "quit"}, {"h/l", "nav"}, {"enter", "open"}, {"space", "expand"}, {":", "cmd"},
		{"e", "edit"}, {"d", "del"}, {"S", "cfg"}, {"r", "reset"}, {"?", "help"},
	}
}

func (m tuiModel) renderMemosListOnly(width int) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	if len(m.memos) == 0 {
		line(m.theme.Dim.Render("  (no memos)"))
		return b.String()
	}

	cl := m.contentLines()
	if cl < 1 {
		cl = 1
	}

	remaining := cl
	moreBelow := false
	lastRendered := -1

	if m.offset > 0 {
		line(m.theme.Dim.Render(fmt.Sprintf("  ↑ %d items ..", m.offset)))
		remaining--
	}

	skip := m.offset
	// In split view, offset is item-based (each memo = 1 line).
	// If offset exceeds memo count (e.g. toggled from full view where
	// offset was line-based with expanded height), clamp to cursor position.
	if skip >= len(m.memos) {
		skip = m.cursorIdx
		if skip < 0 {
			skip = 0
		}
	}
	for i := range m.memos {
		if skip > 0 {
			skip--
			continue
		}
		if len(m.memos)-i > 1 && remaining == 1 && i != m.cursorIdx {
			moreBelow = true
			break
		}
		if remaining <= 0 {
			moreBelow = true
			break
		}

		memo := m.memos[i]
		isSel := i == m.cursorIdx

		prefix := "   "
		if isSel {
			prefix = m.theme.Cursor.String() + "  "
		}

		checkMark := ""
		if m.selectedMap != nil && m.selectedMap[memo.UUID] {
			checkMark = m.theme.CheckOn.String()
		}

		summary := tui.GetShortSummary(memo.Content)

		availWidth := width - lipgloss.Width(prefix) - lipgloss.Width(checkMark) - 2
		if availWidth < 5 {
			availWidth = 5
		}

		if lipgloss.Width(summary) > availWidth {
			summary = tui.TruncateToWidth(summary, availWidth-3) + "..."
		}

		if isSel {
			summary = m.theme.Sel.Render(summary)
		} else {
			summary = tui.RenderInline(summary, &tui.InlineState{}, markdownStyleFromTheme(m.theme))
		}

		lineContent := prefix + checkMark + summary

		if isSel {
			line(m.theme.SelBg.Render(lineContent))
		} else {
			line(lineContent)
		}
		lastRendered = i
		remaining--
	}

	if moreBelow {
		belowCount := len(m.memos) - lastRendered - 1
		line(m.theme.Dim.Render(fmt.Sprintf("  ↓ %d items ..", belowCount)))
		remaining--
	}

	for remaining > 0 {
		line("")
		remaining--
	}

	return b.String()
}

func (m *tuiModel) renderRightPane(width int, highlightBlock bool, maxContentH int) string {
	if len(m.memos) == 0 || m.cursorIdx < 0 || m.cursorIdx >= len(m.memos) {
		return m.theme.Dim.Render("No memo selected")
	}
	memo := m.memos[m.cursorIdx]
	blocks := tui.ParseMarkdownBlocks(memo.Content)

	var lines []string

	shortUUID := memo.UUID
	if len(shortUUID) > 8 {
		shortUUID = shortUUID[:8]
	}
	idStr := m.theme.Sel.Render("ID: " + shortUUID)
	var statusBadge string
	if memo.RowStatus == "archived" {
		statusBadge = m.theme.StatusArchived.String()
	} else {
		statusBadge = m.theme.StatusActive.String()
	}

	headerSpacing := width - lipgloss.Width(idStr) - lipgloss.Width(statusBadge)
	if headerSpacing < 2 {
		headerSpacing = 2
	}

	lines = append(lines, idStr+strings.Repeat(" ", headerSpacing)+statusBadge)
	lines = append(lines, m.theme.Dim.Render("Created: "+memo.CreatedAt))

	if len(memo.Tags) > 0 {
		tagStrs := make([]string, len(memo.Tags))
		for ti, t := range memo.Tags {
			tagStrs[ti] = m.theme.Filter.Render("#" + t)
		}
		lines = append(lines, m.theme.Dim.Render("Tags: ")+strings.Join(tagStrs, " "))
	} else {
		lines = append(lines, m.theme.Dim.Render("Tags: (none)"))
	}
	lines = append(lines, m.theme.HR.Render(strings.Repeat("─", width)))

	type lineInfo struct {
		text     string
		blockIdx int
	}
	var renderedLines []lineInfo

	for bIdx, block := range blocks {
		blockLines := tui.RenderMarkdown(block.Raw, width, markdownStyleFromTheme(m.theme))
		style := lipgloss.NewStyle()
		if highlightBlock && bIdx == m.detailBlockIdx {
			style = m.theme.BlockHighlight
		}

		for _, bl := range blockLines {
			renderedLines = append(renderedLines, lineInfo{
				text:     style.Render(bl),
				blockIdx: bIdx,
			})
		}
		if bIdx < len(blocks)-1 {
			renderedLines = append(renderedLines, lineInfo{
				text:     "",
				blockIdx: -1,
			})
		}
	}

	// Calculate available height for details block rendering
	headerH := 4
	// If the rendered lines exceed maxContentH - headerH, we will need the scroll info footer (2 lines)
	availH := maxContentH - headerH
	if len(renderedLines) > availH {
		availH = maxContentH - headerH - 2
	}
	if availH < 1 {
		availH = 1
	}

	if highlightBlock && m.detailBlockIdx >= 0 && m.detailBlockIdx < len(blocks) {
		firstLineOfBlock := -1
		lastLineOfBlock := -1
		for idx, rl := range renderedLines {
			if rl.blockIdx == m.detailBlockIdx {
				if firstLineOfBlock == -1 {
					firstLineOfBlock = idx
				}
				lastLineOfBlock = idx
			}
		}

		if firstLineOfBlock != -1 {
			if firstLineOfBlock < m.previewScroll {
				m.previewScroll = firstLineOfBlock
			} else if lastLineOfBlock >= m.previewScroll+availH {
				m.previewScroll = lastLineOfBlock - availH + 1
			}
		}
	}

	if m.previewScroll < 0 {
		m.previewScroll = 0
	}
	if m.previewScroll > len(renderedLines)-availH {
		m.previewScroll = len(renderedLines) - availH
	}
	if m.previewScroll < 0 {
		m.previewScroll = 0
	}

	for idx := m.previewScroll; idx < m.previewScroll+availH && idx < len(renderedLines); idx++ {
		lines = append(lines, renderedLines[idx].text)
	}

	if len(renderedLines) > availH {
		scrollStr := fmt.Sprintf(" Block %d/%d · Line %d/%d ", m.detailBlockIdx+1, len(blocks), m.previewScroll+1, len(renderedLines))
		scrollLine := m.theme.Dim.Render("Scroll:") + " " + m.theme.Count.Render(scrollStr)
		lines = append(lines, "", scrollLine)
	}

	return strings.Join(lines, "\n")
}

func markdownStyleFromTheme(t *Theme) *tui.MarkdownStyle {
	return &tui.MarkdownStyle{
		H1Style:         t.H1Style,
		H2Style:         t.H2Style,
		H3Style:         t.H3Style,
		H4Style:         t.H4Style,
		BoldStyle:       t.BoldStyle,
		ItalicStyle:     t.ItalicStyle,
		InlineCodeStyle: t.InlineCodeStyle,
		StrikeStyle:     t.StrikeStyle,
		LinkStyle:       t.LinkStyle,
		URLStyle:        t.URLStyle,
		TagStyle:        t.TagStyle,
		CodeBlockStyle:  t.CodeBlockStyle,
		QuoteStyle:      t.QuoteStyle,
		QuoteBorder:     t.QuoteBorder,
		BulletStyle:     t.BulletStyle,
		NumStyle:        t.NumStyle,
		TaskDoneStyle:   t.TaskDoneStyle,
		TaskTodoStyle:   t.TaskTodoStyle,
		HR:              t.HR,
	}
}
