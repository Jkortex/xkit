package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"daily/internal/infrastructure/config"
	tui "daily/internal/interfaces/tui"

	"github.com/charmbracelet/bubbletea"
)

func statusToRowStatus(s string) *entity.RowStatus {
	switch s {
	case "normal":
		rs := entity.RowStatus("normal")
		return &rs
	case "archived":
		rs := entity.RowStatus("archived")
		return &rs
	default:
		return nil
	}
}

func tuiSortToPortSort(fld, dir string) string {
	if fld == "" {
		fld = "created_at"
	}
	if dir == "" {
		dir = "DESC"
	}
	if dir == "ASC" {
		return fld + "_asc"
	}
	return fld + "_desc"
}

func memoResponseToUIMemo(r *dto.MemoResponse) tui.UIMemo {
	return tui.UIMemo{
		UUID:      r.UUID,
		Content:   r.Content,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Tags:      r.Tags,
		RowStatus: r.RowStatus,
		ExpiresAt: r.ExpiresAt,
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(tui.WatchFile(m.eventPath))
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case modeCreate:
			return m.updateCreate(msg)
		case modeDeleteConfirm:
			return m.updateDeleteConfirm(msg)
		case modeCmd:
			return m.updateCmd(msg)
		case modeTagInput:
			return m.updateTagInput(msg)
		case modeDetail:
			return m.updateDetail(msg)
		case modeDSLInput:
			return m.updateDSLInput(msg)
		case modeConfig:
			return m.updateConfig(msg)
		case modeConfigEdit:
			return m.updateConfigEdit(msg)
		default:
			return m.updateList(msg)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.offset = m.computeOffset()
	case tui.EventMsg:
		m.pullNew()
		return m, tui.WatchFile(m.eventPath)
	case tui.ErrMsg:
		m.statusMsg = fmt.Sprintf("Error: %v", msg.Err)
		return m, tui.WatchFile(m.eventPath)
	case tui.EditFinishedMsg:
		if msg.TempPath != "" {
			_ = os.Remove(msg.TempPath)
		}
		if msg.Err != nil {
			m.statusMsg = fmt.Sprintf("Edit failed: %v", msg.Err)
			return m, nil
		}
		content := strings.TrimSpace(msg.Content)
		if strings.HasPrefix(content, "# Write your memo here") {
			content = strings.TrimPrefix(content, "# Write your memo here")
			content = strings.TrimSpace(content)
		}
		if content == "" {
			m.statusMsg = "Cancelled: content cannot be empty"
			return m, nil
		}

		if msg.MemoUUID == "" {
			// Creating a new memo
			tags := []string{}
			if m.tagFilter != "" {
				tags = append(tags, m.tagFilter)
			}

			if _, err := m.app.MemoCtrl.Create(context.Background(), m.app.AdminID, content, tags, nil, ""); err != nil {
				m.statusMsg = fmt.Sprintf("Create failed: %v", err)
			} else {
				m.statusMsg = "Memo created"
				touchEventFile()
				m.refreshAll()
			}
		} else {
			// Editing an existing memo
			var oldMemo tui.UIMemo
			found := false
			for _, memo := range m.memos {
				if memo.UUID == msg.MemoUUID {
					oldMemo = memo
					found = true
					break
				}
			}
			if !found {
				m.statusMsg = "Error: Memo not found in list"
				return m, nil
			}

			input := dto.UpdateMemoRequest{
				Content: content,
				Tags:    oldMemo.Tags,
			}
			if _, err := m.app.MemoCtrl.Update(context.Background(), m.app.AdminID, msg.MemoUUID, input); err != nil {
				m.statusMsg = fmt.Sprintf("Update failed: %v", err)
			} else {
				m.statusMsg = "Memo updated"
				touchEventFile()
				m.refreshAll()
			}
		}
		return m, nil
	}
	return m, nil
}

func (m tuiModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		return m.switchTab(tabInbox)
	case "2":
		return m.switchTab(tabArchived)
	case "3":
		return m.switchTab(tabTags)
	case "4":
		return m.switchTab(tabStats)
	case "h", "left":
		prev := m.activeTab - 1
		if prev < 0 {
			prev = tabStats
		}
		return m.switchTab(prev)
	case "l", "right":
		next := m.activeTab + 1
		if next > tabStats {
			next = tabInbox
		}
		return m.switchTab(next)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		if m.tagFilter != "" || m.searchTxt != "" {
			m.tagFilter = ""
			m.searchTxt = ""
			m.statusMsg = "Filters cleared"
			m.refreshAll()
		}
		return m, nil
	case "up", "k":
		if m.cursorIdx > 0 {
			m.cursorIdx--
			m.statusMsg = ""
			m.offset = m.computeOffset()
		}
	case "down", "j":
		var maxIdx int
		if m.activeTab == tabTags {
			maxIdx = len(m.tagsWithCount) - 1
		} else {
			maxIdx = len(m.memos) - 1
		}
		if m.cursorIdx < maxIdx {
			m.cursorIdx++
			m.statusMsg = ""
			m.offset = m.computeOffset()
		}
	case "g":
		m.cursorIdx = 0
		m.offset = 0
	case "G":
		if m.activeTab == tabTags {
			m.cursorIdx = len(m.tagsWithCount) - 1
		} else {
			m.cursorIdx = len(m.memos) - 1
		}
		m.offset = m.computeOffset()
	case "r":
		m.clearAllFilters()
		m.statusMsg = "Reset"
		return m.switchTab(m.activeTab)
	case " ":
		if (m.activeTab == tabInbox || m.activeTab == tabArchived) && m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
			m.memos[m.cursorIdx].Expanded = !m.memos[m.cursorIdx].Expanded
			m.offset = m.computeOffset()
		}
	case "enter", "tab":
		if m.activeTab == tabTags {
			if m.cursorIdx >= 0 && m.cursorIdx < len(m.tagsWithCount) {
				m.tagFilter = m.tagsWithCount[m.cursorIdx].Name
				m.statusMsg = fmt.Sprintf("Filtered by @%s", m.tagFilter)
				return m.switchTab(tabInbox)
			}
		} else if m.activeTab == tabInbox || m.activeTab == tabArchived {
			if m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
				m.mode = modeDetail
				m.previewScroll = 0
				m.detailBlockIdx = 0
				m.detailBlocks = tui.ParseMarkdownBlocks(m.memos[m.cursorIdx].Content)
			}
		}
	case "y":
		if (m.activeTab == tabInbox || m.activeTab == tabArchived) && m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
			memo := m.memos[m.cursorIdx]
			textToCopy := tui.ExtractTranslation(memo.Content)
			if textToCopy == "" {
				textToCopy = tui.CleanMarkdownForTTS(memo.Content)
			}
			if err := tui.CopyToClipboard(textToCopy); err != nil {
				m.statusMsg = fmt.Sprintf("Copy failed: %v", err)
			} else {
				m.statusMsg = "Copied to clipboard"
			}
		}
	case "p":
		if (m.activeTab == tabInbox || m.activeTab == tabArchived) && m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
			text := tui.CleanMarkdownForTTS(m.memos[m.cursorIdx].Content)
			if text != "" {
				go tui.SpeakWithCache(text, m.memos[m.cursorIdx].UUID, m.memos[m.cursorIdx].ExpiresAt, "T", &m.app.Config.TTS)
				m.statusMsg = fmt.Sprintf("▶ (%s)", m.memos[m.cursorIdx].UUID[:8])
			} else {
				m.statusMsg = "Empty memo"
			}
		}
	case "s":
		if (m.activeTab == tabInbox || m.activeTab == tabArchived) && m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
			m.statusMsg = tui.ShowCacheStatus(m.memos[m.cursorIdx].UUID, "T", m.memos[m.cursorIdx].ExpiresAt)
		}
	case "x":
		if (m.activeTab == tabInbox || m.activeTab == tabArchived) && m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
			uid := m.memos[m.cursorIdx].UUID
			if m.selectedMap == nil {
				m.selectedMap = make(map[string]bool)
			}
			m.selectedMap[uid] = !m.selectedMap[uid]
			if !m.selectedMap[uid] {
				delete(m.selectedMap, uid)
			}
		}
	case "e":
		if m.activeTab == tabInbox || m.activeTab == tabArchived {
			return m.launchEditorEdit()
		}
	case "C":
		if m.activeTab == tabInbox || m.activeTab == tabArchived {
			return m.launchEditorCreate()
		}
	case "T":
		if m.activeTab == tabInbox || m.activeTab == tabArchived {
			var selectedUUIDs []string
			for uid, sel := range m.selectedMap {
				if sel {
					selectedUUIDs = append(selectedUUIDs, uid)
				}
			}
			var initialTags []string
			if len(selectedUUIDs) > 0 {
				for _, memo := range m.memos {
					if memo.UUID == selectedUUIDs[0] {
						initialTags = memo.Tags
						break
					}
				}
			} else if m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
				initialTags = m.memos[m.cursorIdx].Tags
			}
			m.tagInputBuf = strings.Join(initialTags, " ")
			m.mode = modeTagInput
			m.statusMsg = "Edit tags (Esc to cancel, Enter to save)"
		}
	case "c":
		if m.activeTab == tabInbox || m.activeTab == tabArchived {
			m.mode = modeCreate
			m.createBuf = ""
			m.statusMsg = "Enter content (Esc to cancel, Enter to submit)"
		}
	case "d":
		if m.activeTab == tabInbox || m.activeTab == tabArchived {
			var selectedUUIDs []string
			for uid, sel := range m.selectedMap {
				if sel {
					selectedUUIDs = append(selectedUUIDs, uid)
				}
			}
			if len(selectedUUIDs) > 0 {
				m.mode = modeDeleteConfirm
				m.statusMsg = fmt.Sprintf("Delete %d selected memos? (y/n)", len(selectedUUIDs))
			} else if m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
				m.mode = modeDeleteConfirm
				m.delUUID = m.memos[m.cursorIdx].UUID
				m.statusMsg = fmt.Sprintf("Delete %s? (y/n)", m.memos[m.cursorIdx].UUID[:8])
			}
		}
	case "a":
		if m.activeTab == tabInbox || m.activeTab == tabArchived {
			return m.executeArchiveRestore()
		}
	case "t":
		if m.activeTab == tabInbox || m.activeTab == tabArchived {
			if m.allTags == nil {
				tagResp, err := m.app.TagCtrl.ListTags(context.Background(), m.app.AdminID)
				if err == nil {
					m.allTags = make([]string, len(tagResp))
					for i, t := range tagResp {
						m.allTags[i] = t.Name
					}
				}
			}
			if len(m.allTags) == 0 {
				m.statusMsg = "No tags available"
			} else {
				m.tagIdx++
				if m.tagIdx >= len(m.allTags) {
					m.tagIdx = 0
					m.tagFilter = ""
					m.statusMsg = "Filter cleared"
				} else {
					m.tagFilter = m.allTags[m.tagIdx]
					m.statusMsg = fmt.Sprintf("Filter by @%s  (t again to cycle)", m.tagFilter)
				}
				m.refreshAll()
			}
		}
	case ":":
		m.mode = modeCmd
		m.cmdBuf = ""
		m.cmdSelIdx = 0
		m.statusMsg = ""
	case "S":
		m.configItems = buildConfigItems(m.app.Config, m.app.ConfigSources)
		m.configIdx = 0
		m.mode = modeConfig
		m.statusMsg = ""
	case "?":
		m.statusMsg = "h/l:tab  enter:open  space:expand  e:edit  d:del  p:play  S:config  :cmd  ?:help"
	}
	return m, nil
}

func (m tuiModel) switchTab(tab tabType) (tuiModel, tea.Cmd) {
	m.activeTab = tab
	m.cursorIdx = 0
	m.offset = 0
	m.previewScroll = 0
	m.statusMsg = ""

	switch tab {
	case tabInbox:
		m.statusFilter = "normal"
		m.refreshAll()
	case tabArchived:
		m.statusFilter = "archived"
		m.refreshAll()
	case tabTags:
		tagResp, err := m.app.TagCtrl.ListTags(context.Background(), m.app.AdminID)
		if err == nil {
			m.tagsWithCount = make([]tui.TagStat, len(tagResp))
			for i, t := range tagResp {
				m.tagsWithCount[i] = tui.TagStat{Name: t.Name, Count: t.Count}
			}
		}
	case tabStats:
		stats, err := m.app.MemoCtrl.GetStats(context.Background(), m.app.AdminID)
		if err == nil {
			m.statsData = stats
		}
	}
	return m, nil
}

func (m tuiModel) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
		m.statusMsg = ""
	case "enter":
		content := strings.TrimSpace(m.createBuf)
		if content != "" {
			tags := []string{}
			if m.tagFilter != "" {
				tags = append(tags, m.tagFilter)
			}

			if _, err := m.app.MemoCtrl.Create(context.Background(), m.app.AdminID, content, tags, nil, ""); err != nil {
				m.statusMsg = fmt.Sprintf("Create failed: %v", err)
			} else {
				m.statusMsg = "Memo created"
				touchEventFile()
				m.refreshAll()
			}
		}
		m.mode = modeList
	case "backspace":
		if len(m.createBuf) > 0 {
			runes := []rune(m.createBuf)
			m.createBuf = string(runes[:len(runes)-1])
		}
	default:
		if len(msg.Runes) > 0 {
			m.createBuf += string(msg.Runes)
		} else {
			k := msg.String()
			if len(k) == 1 && k[0] >= 32 {
				m.createBuf += k
			}
		}
	}
	return m, nil
}

func (m tuiModel) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		var selectedUUIDs []string
		for uid, sel := range m.selectedMap {
			if sel {
				selectedUUIDs = append(selectedUUIDs, uid)
			}
		}

		if len(selectedUUIDs) > 0 {
			if _, err := m.app.MemoCtrl.BatchDelete(context.Background(), m.app.AdminID, selectedUUIDs); err != nil {
				m.statusMsg = fmt.Sprintf("Batch delete failed: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Deleted %d memos", len(selectedUUIDs))
			}
			m.selectedMap = make(map[string]bool)
		} else if m.delUUID != "" {
			if err := m.app.MemoCtrl.Delete(context.Background(), m.app.AdminID, m.delUUID); err != nil {
				m.statusMsg = fmt.Sprintf("Delete failed: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Deleted %s", m.delUUID[:8])
			}
			m.delUUID = ""
		}
		touchEventFile()
		m.refreshAll()
		m.mode = modeList
	case "n", "N", "esc":
		m.mode = modeList
		m.statusMsg = ""
	}
	return m, nil
}

func (m tuiModel) updateCmd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.getFilteredCommands()

	switch msg.String() {
	case "esc":
		m.mode = modeList
		m.cmdBuf = ""
		m.statusMsg = ""
	case "up", "ctrl+p", "ctrl+k":
		if len(filtered) > 0 {
			m.cmdSelIdx--
			if m.cmdSelIdx < 0 {
				m.cmdSelIdx = len(filtered) - 1
			}
		}
	case "down", "ctrl+n", "ctrl+j":
		if len(filtered) > 0 {
			m.cmdSelIdx++
			if m.cmdSelIdx >= len(filtered) {
				m.cmdSelIdx = 0
			}
		}
	case "enter":
		if len(filtered) > 0 && m.cmdSelIdx >= 0 && m.cmdSelIdx < len(filtered) {
			selected := filtered[m.cmdSelIdx]
			m.mode = modeList
			m.cmdBuf = ""
			return selected.run(m)
		}
	case "backspace":
		if len(m.cmdBuf) > 0 {
			runes := []rune(m.cmdBuf)
			m.cmdBuf = string(runes[:len(runes)-1])
			m.cmdSelIdx = 0
		}
	default:
		if len(msg.Runes) > 0 {
			m.cmdBuf += string(msg.Runes)
			m.cmdSelIdx = 0
		}
	}
	return m, nil
}

func (m tuiModel) updateDSLInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
		m.cmdBuf = ""
		m.statusMsg = ""
	case "enter":
		cmd := strings.TrimSpace(m.cmdBuf)
		m.mode = modeList
		m.cmdBuf = ""
		if cmd == "" {
			return m, nil
		}
		m.applyDSLQuery(cmd)
	case "backspace":
		if len(m.cmdBuf) > 0 {
			r := []rune(m.cmdBuf)
			m.cmdBuf = string(r[:len(r)-1])
		}
	default:
		if len(msg.Runes) > 0 {
			m.cmdBuf += string(msg.Runes)
		} else {
			k := msg.String()
			if len(k) == 1 && k[0] >= 32 {
				m.cmdBuf += k
			}
		}
	}
	return m, nil
}

func (m tuiModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.memos) == 0 || m.cursorIdx < 0 || m.cursorIdx >= len(m.memos) {
		m.mode = modeList
		return m, nil
	}

	if len(m.detailBlocks) == 0 {
		m.detailBlocks = tui.ParseMarkdownBlocks(m.memos[m.cursorIdx].Content)
		m.detailBlockIdx = 0
	}

	switch msg.String() {
	case "esc", "tab":
		m.mode = modeList
		m.statusMsg = ""
		m.previewScroll = 0
		m.detailBlocks = nil
	case "q":
		m.mode = modeList
		m.statusMsg = ""
		m.previewScroll = 0
		m.detailBlocks = nil
	case "up", "k":
		if m.detailBlockIdx > 0 {
			m.detailBlockIdx--
		} else {
			m.detailBlockIdx = 0
		}
	case "down", "j":
		if m.detailBlockIdx < len(m.detailBlocks)-1 {
			m.detailBlockIdx++
		} else {
			m.detailBlockIdx = len(m.detailBlocks) - 1
		}
	case "ctrl+d":
		m.previewScroll += 5
	case "ctrl+u":
		m.previewScroll -= 5
		if m.previewScroll < 0 {
			m.previewScroll = 0
		}
	case "space", "p":
		if m.detailBlockIdx >= 0 && m.detailBlockIdx < len(m.detailBlocks) {
			block := m.detailBlocks[m.detailBlockIdx]
			go tui.SpeakWithCache(block.Content, m.memos[m.cursorIdx].UUID, m.memos[m.cursorIdx].ExpiresAt, fmt.Sprintf("B%d", m.detailBlockIdx), &m.app.Config.TTS)
			m.statusMsg = fmt.Sprintf("▶ Block %d/%d", m.detailBlockIdx+1, len(m.detailBlocks))
		}
	case "y":
		if m.detailBlockIdx >= 0 && m.detailBlockIdx < len(m.detailBlocks) {
			block := m.detailBlocks[m.detailBlockIdx]
			if err := tui.CopyToClipboard(block.Content); err != nil {
				m.statusMsg = fmt.Sprintf("Copy failed: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Copied block %d to clipboard", m.detailBlockIdx+1)
			}
		}
	case "s":
		m.statusMsg = tui.ShowCacheStatus(m.memos[m.cursorIdx].UUID, fmt.Sprintf("B%d", m.detailBlockIdx), m.memos[m.cursorIdx].ExpiresAt)
	case "e":
		m.detailBlocks = nil
		return m.launchEditorEdit()
	case "T":
		m.tagInputBuf = strings.Join(m.memos[m.cursorIdx].Tags, " ")
		m.mode = modeTagInput
		m.statusMsg = "Edit tags (Esc to cancel, Enter to save)"
	case "d":
		m.mode = modeDeleteConfirm
		m.delUUID = m.memos[m.cursorIdx].UUID
		m.statusMsg = fmt.Sprintf("Delete %s? (y/n)", m.memos[m.cursorIdx].UUID[:8])
	case "a":
		m.detailBlocks = nil
		return m.executeArchiveRestore()
	case "S":
		m.configItems = buildConfigItems(m.app.Config, m.app.ConfigSources)
		m.configIdx = 0
		m.mode = modeConfig
		m.statusMsg = ""
	}
	return m, nil
}

func (m tuiModel) updateConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.mode = modeList
		m.statusMsg = ""
	case "up", "k":
		if m.configIdx > 0 {
			m.configIdx--
		}
	case "down", "j":
		if m.configIdx < len(m.configItems)-1 {
			m.configIdx++
		}
	case "enter", "s":
		item := m.configItems[m.configIdx]
		if item.Type == "separator" {
			return m, nil
		}
		if item.ReadOnly {
			m.statusMsg = "Env vars are read-only — edit ~/.env or ~/.daily/config.json instead"
			return m, nil
		}
		if item.Type == "bool" {
			tui.ToggleConfigItem(m.app.Config, item)
			m.app.ConfigSources[item.Key] = config.SourceFile
			m.configItems = buildConfigItems(m.app.Config, m.app.ConfigSources)
			if err := config.Save(m.app.Config); err != nil {
				m.statusMsg = fmt.Sprintf("Save error: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("%s toggled", item.Label)
			}
			return m, nil
		}
		if item.Type == "select" {
			oldVal := item.Value
			newVal := tui.CycleSelectItem(m.app.Config, item)
			m.app.ConfigSources[item.Key] = config.SourceFile
			m.configItems = buildConfigItems(m.app.Config, m.app.ConfigSources)
			if err := config.Save(m.app.Config); err != nil {
				m.statusMsg = fmt.Sprintf("Save error: %v", err)
			} else {
				display := newVal
				if display == "" {
					display = "(empty)"
				}
				m.statusMsg = fmt.Sprintf("%s: %s → %s", item.Label, oldVal, display)
			}
			// Hot-reload theme if theme changed
			if item.Key == "theme" {
				m.theme = LoadTheme(m.app.Config.Theme)
				activeTheme = m.theme
			}
			return m, nil
		}
		// String field — enter edit mode
		m.configEditBuf = item.Value
		if m.configEditBuf == "(empty)" {
			m.configEditBuf = ""
		}
		m.mode = modeConfigEdit
		m.statusMsg = fmt.Sprintf("Edit %s (Enter to save, Esc to cancel)", item.Label)
	}
	return m, nil
}

func (m tuiModel) updateConfigEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeConfig
		m.statusMsg = "Edit cancelled"
	case "enter":
		item := m.configItems[m.configIdx]
		tui.ApplyConfigItem(m.app.Config, item, m.configEditBuf)
		m.app.ConfigSources[item.Key] = config.SourceFile
		// Save to config file
		if err := config.Save(m.app.Config); err != nil {
			m.statusMsg = fmt.Sprintf("Save error: %v", err)
		} else {
			m.statusMsg = fmt.Sprintf("%s saved: %s → %s", item.Label, item.Value, m.configEditBuf)
		}
		m.configItems = buildConfigItems(m.app.Config, m.app.ConfigSources)
		m.mode = modeConfig
	case "backspace":
		if len(m.configEditBuf) > 0 {
			m.configEditBuf = m.configEditBuf[:len(m.configEditBuf)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.configEditBuf += msg.String()
		}
	}
	return m, nil
}

func (m tuiModel) updateTagInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
		m.statusMsg = ""
	case "enter":
		tags := tui.SplitTags(strings.ReplaceAll(m.tagInputBuf, " ", ","))
		var selectedUUIDs []string
		for uid, sel := range m.selectedMap {
			if sel {
				selectedUUIDs = append(selectedUUIDs, uid)
			}
		}
		if len(selectedUUIDs) == 0 && m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
			selectedUUIDs = []string{m.memos[m.cursorIdx].UUID}
		}

		if len(selectedUUIDs) > 0 {
			for _, uid := range selectedUUIDs {
				var currentMemo tui.UIMemo
				for _, memo := range m.memos {
					if memo.UUID == uid {
						currentMemo = memo
						break
					}
				}
			if currentMemo.UUID == "" {
				resp, err := m.app.MemoCtrl.Get(context.Background(), m.app.AdminID, uid)
				if err == nil && resp != nil {
					currentMemo.Content = resp.Content
				}
			}
				input := dto.UpdateMemoRequest{
					Content: currentMemo.Content,
					Tags:    tags,
				}
				if _, err := m.app.MemoCtrl.Update(context.Background(), m.app.AdminID, uid, input); err != nil {
					m.statusMsg = fmt.Sprintf("Update tags failed: %v", err)
				}
			}
			m.statusMsg = fmt.Sprintf("Updated tags for %d memo(s)", len(selectedUUIDs))
			m.selectedMap = make(map[string]bool)
			touchEventFile()
			m.refreshAll()
		}
		m.mode = modeList
	case "backspace":
		if len(m.tagInputBuf) > 0 {
			runes := []rune(m.tagInputBuf)
			m.tagInputBuf = string(runes[:len(runes)-1])
		}
	default:
		if len(msg.Runes) > 0 {
			m.tagInputBuf += string(msg.Runes)
		} else {
			k := msg.String()
			if len(k) == 1 && k[0] >= 32 {
				m.tagInputBuf += k
			}
		}
	}
	return m, nil
}

func (m *tuiModel) pullNew() {
	resp, err := m.app.MemoCtrl.List(context.Background(), m.app.AdminID, m.filter())
	if err != nil {
		return
	}
	fresh := make([]tui.UIMemo, 0, len(resp))
	for _, r := range resp {
		fresh = append(fresh, memoResponseToUIMemo(r))
	}
	if len(fresh) > 0 {
		m.lastTime = fresh[len(fresh)-1].CreatedAt
	}

	// find truly new items by comparing counts forwards
	if len(fresh) <= len(m.memos) {
		return
	}
	// append only items not in current list
	existing := make(map[string]bool, len(m.memos))
	for _, em := range m.memos {
		existing[em.UUID] = true
	}
	for _, nm := range fresh {
		if existing[nm.UUID] {
			continue
		}
		for i := range m.memos {
			m.memos[i].Expanded = false
		}
		nm.Expanded = true
		m.memos = append(m.memos, nm)
		if tr := tui.ExtractTranslation(nm.Content); tr != "" {
			go tui.SpeakWithCache(tr, nm.UUID, nm.ExpiresAt, "T", &m.app.Config.TTS)
		}
	}
	if len(m.memos) > 100 {
		m.memos = m.memos[len(m.memos)-100:]
	}
	m.cursorIdx = len(m.memos) - 1
	m.offset = m.computeOffset()
	m.statusMsg = "New memo received!"
}

func (m *tuiModel) refreshAll() {
	resp, err := m.app.MemoCtrl.List(context.Background(), m.app.AdminID, m.filter())
	if err != nil {
		m.statusMsg = fmt.Sprintf("List error: %v", err)
		return
	}
	memos := make([]tui.UIMemo, 0, len(resp))
	var lastTime string
	for _, r := range resp {
		u := memoResponseToUIMemo(r)
		memos = append(memos, u)
		lastTime = u.CreatedAt
	}
	m.memos = memos
	m.lastTime = lastTime
	if m.cursorIdx >= len(m.memos) {
		m.cursorIdx = len(m.memos) - 1
	}
	if m.cursorIdx < 0 && len(m.memos) > 0 {
		m.cursorIdx = 0
	}
	m.offset = m.computeOffset()
}

func (m *tuiModel) filter() port.MemoFilter {
	mi := m.maxItems
	if mi <= 0 {
		mi = 50
	}
	var search *string
	if m.searchTxt != "" {
		search = &m.searchTxt
	}
	var tagFilter *string
	if m.tagFilter != "" {
		tagFilter = &m.tagFilter
	}
	return port.MemoFilter{
		Tag:         tagFilter,
		Search:      search,
		Limit:       mi,
		TagsAny:     m.tagOr,
		TagsAll:     m.tagAnd,
		TagsExclude: m.tagExcl,
		FromDate:    strPtr(m.fromDt),
		ToDate:      strPtr(m.toDt),
		Sort:        tuiSortToPortSort(m.sortFld, m.sortDir),
		RowStatus:   statusToRowStatus(m.statusFilter),
	}
}

func (m *tuiModel) clearAllFilters() {
	m.tagFilter = ""
	m.searchTxt = ""
	m.maxItems = 0
	m.tagIdx = 0
	m.tagOr = nil
	m.tagAnd = nil
	m.tagExcl = nil
	m.fromDt = ""
	m.toDt = ""
	m.sortFld = ""
	m.sortDir = ""
	m.statusFilter = "normal"
	m.selectedMap = make(map[string]bool)
}

func (m *tuiModel) applyDSLQuery(cmd string) {
	switch {
	case cmd == "clear" || cmd == "reset":
		m.clearAllFilters()
		m.statusMsg = "Filters cleared"
		m.refreshAll()
	case strings.HasPrefix(cmd, "/"):
		m.searchTxt = strings.TrimPrefix(cmd, "/")
		m.statusMsg = fmt.Sprintf("Search: %s", m.searchTxt)
		m.refreshAll()
	case strings.HasPrefix(cmd, "t+:"), strings.HasPrefix(cmd, "tag+:"):
		v := strings.TrimPrefix(strings.TrimPrefix(cmd, "t+:"), "tag+:")
		m.tagAnd = tui.SplitTags(v)
		m.tagFilter = ""
		m.statusMsg = fmt.Sprintf("Must have tags: %s", strings.Join(m.tagAnd, ", "))
		m.refreshAll()
	case strings.HasPrefix(cmd, "t-:"), strings.HasPrefix(cmd, "tag-:"):
		v := strings.TrimPrefix(strings.TrimPrefix(cmd, "t-:"), "tag-:")
		m.tagExcl = tui.SplitTags(v)
		m.statusMsg = fmt.Sprintf("Exclude tags: %s", strings.Join(m.tagExcl, ", "))
		m.refreshAll()
	case strings.HasPrefix(cmd, "t:"), strings.HasPrefix(cmd, "tag:"):
		v := strings.TrimPrefix(strings.TrimPrefix(cmd, "t:"), "tag:")
		m.tagOr = tui.SplitTags(v)
		m.tagFilter = ""
		m.statusMsg = fmt.Sprintf("Filter by tags: %s", strings.Join(m.tagOr, ", "))
		m.refreshAll()
	case strings.HasPrefix(cmd, "l:"), strings.HasPrefix(cmd, "latest:"):
		n := 0
		v := strings.TrimPrefix(strings.TrimPrefix(cmd, "l:"), "latest:")
		fmt.Sscanf(v, "%d", &n)
		if n > 0 {
			m.maxItems = n
			m.statusMsg = fmt.Sprintf("Show latest %d", n)
			m.refreshAll()
		}
	case strings.HasPrefix(cmd, "from:"):
		m.fromDt = strings.TrimPrefix(cmd, "from:")
		m.statusMsg = fmt.Sprintf("From: %s", m.fromDt)
		m.refreshAll()
	case strings.HasPrefix(cmd, "to:"):
		m.toDt = strings.TrimPrefix(cmd, "to:")
		m.statusMsg = fmt.Sprintf("To: %s", m.toDt)
		m.refreshAll()
	case strings.HasPrefix(cmd, "sort:"):
		v := strings.TrimPrefix(cmd, "sort:")
		switch v {
		case "created_at", "created_at_asc":
			m.sortFld = "created_at"
			m.sortDir = "ASC"
		case "created_at_desc":
			m.sortFld = "created_at"
			m.sortDir = "DESC"
		case "updated_at", "updated_at_asc":
			m.sortFld = "updated_at"
			m.sortDir = "ASC"
		case "updated_at_desc":
			m.sortFld = "updated_at"
			m.sortDir = "DESC"
		default:
			m.sortFld = "created_at"
			m.sortDir = "DESC"
		}
		m.statusMsg = fmt.Sprintf("Sort: %s %s", m.sortFld, m.sortDir)
		m.refreshAll()
	default:
		m.statusMsg = "Unknown cmd: " + cmd
	}
}

func (m tuiModel) executeArchiveRestore() (tuiModel, tea.Cmd) {
	var selectedUUIDs []string
	for uid, sel := range m.selectedMap {
		if sel {
			selectedUUIDs = append(selectedUUIDs, uid)
		}
	}

	if len(selectedUUIDs) > 0 {
		firstUUID := selectedUUIDs[0]
		var firstIsArchived bool
		for _, memo := range m.memos {
			if memo.UUID == firstUUID {
				firstIsArchived = (memo.RowStatus == "archived")
				break
			}
		}

		if firstIsArchived {
			for _, uid := range selectedUUIDs {
				_, _ = m.app.DB.ExecContext(context.Background(), "UPDATE memo SET row_status = 'normal', updated_at = CURRENT_TIMESTAMP WHERE memo_uuid = ?", uid)
			}
			m.statusMsg = fmt.Sprintf("Restored %d memos", len(selectedUUIDs))
		} else {
			if _, err := m.app.MemoCtrl.BatchArchive(context.Background(), m.app.AdminID, selectedUUIDs); err != nil {
				m.statusMsg = fmt.Sprintf("Batch archive failed: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Archived %d memos", len(selectedUUIDs))
			}
		}
		m.selectedMap = make(map[string]bool)
		touchEventFile()
		m.refreshAll()
	} else if m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
		uid := m.memos[m.cursorIdx].UUID
		status := m.memos[m.cursorIdx].RowStatus
		if status == "archived" {
			if _, err := m.app.DB.ExecContext(context.Background(), "UPDATE memo SET row_status = 'normal', updated_at = CURRENT_TIMESTAMP WHERE memo_uuid = ?", uid); err != nil {
				m.statusMsg = fmt.Sprintf("Restore failed: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Restored %s", uid[:8])
				touchEventFile()
				m.refreshAll()
			}
		} else {
			if _, err := m.app.MemoCtrl.BatchArchive(context.Background(), m.app.AdminID, []string{uid}); err != nil {
				m.statusMsg = fmt.Sprintf("Archive failed: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Archived %s", uid[:8])
				touchEventFile()
				m.refreshAll()
			}
		}
	}
	return m, nil
}
