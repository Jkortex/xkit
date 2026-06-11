package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tui "daily/internal/interfaces/tui"

	"github.com/charmbracelet/bubbletea"
)

type paletteCmd struct {
	label string
	desc  string
	key   string
	run   func(m tuiModel) (tuiModel, tea.Cmd)
}

func (m tuiModel) cmdHint() string {
	cb := m.cmdBuf
	if cb == "" {
		return m.theme.Dim.Render("/text  t:tag  t+:a,b  t-:tag  from:date  to:date  sort:field  l:N  clear")
	}
	prefixes := []string{"/text", "t:", "tag:", "t+:", "tag+:", "t-:", "tag-:", "from:", "to:", "sort:", "l:", "latest:", "clear", "reset"}
	for _, p := range prefixes {
		if strings.HasPrefix(cb, p) || strings.HasPrefix(p, cb) {
			switch {
			case strings.HasPrefix(cb, "/"):
				return m.theme.Dim.Render("search content with LIKE ·  e.g. /golang")
			case strings.HasPrefix(cb, "t+:"), strings.HasPrefix(cb, "tag+:"):
				tags := m.allTagHint(strings.TrimPrefix(strings.TrimPrefix(cb, "t+:"), "tag+:"))
				return m.theme.Dim.Render("must have ALL tags  ·  e.g. t+:work,important" + tags)
			case strings.HasPrefix(cb, "t-:"), strings.HasPrefix(cb, "tag-:"):
				tags := m.allTagHint(strings.TrimPrefix(strings.TrimPrefix(cb, "t-:"), "tag-:"))
				return m.theme.Dim.Render("exclude tags  ·  e.g. t-:personal" + tags)
			case strings.HasPrefix(cb, "t:"), strings.HasPrefix(cb, "tag:"):
				tags := m.allTagHint(strings.TrimPrefix(strings.TrimPrefix(cb, "t:"), "tag:"))
				return m.theme.Dim.Render("match ANY tag  ·  e.g. t:work,lingo" + tags)
			case strings.HasPrefix(cb, "from:"):
				return m.theme.Dim.Render("start date  ·  e.g. from:2026-06-01")
			case strings.HasPrefix(cb, "to:"):
				return m.theme.Dim.Render("end date  ·  e.g. to:2026-06-03")
			case strings.HasPrefix(cb, "sort:"):
				return m.theme.Dim.Render("created_at / created_at_asc / updated_at / updated_at_desc")
			case strings.HasPrefix(cb, "l:"), strings.HasPrefix(cb, "latest:"):
				return m.theme.Dim.Render("show latest N memos  ·  e.g. l:5")
			case cb == "clear" || cb == "reset":
				return m.theme.Dim.Render("clear all filters: tag, search, date, sort, limit")
			}
		}
	}
	return m.theme.Dim.Render("unknown command  ·  /text  t:tag  t+:a,b  t-:tag  from:date  to:date  sort:field  l:N  clear")
}

func (m tuiModel) allTagHint(prefix string) string {
	if len(m.allTags) == 0 {
		return ""
	}
	searchPrefix := prefix
	parts := strings.Split(prefix, ",")
	if len(parts) > 0 {
		searchPrefix = parts[len(parts)-1]
	}
	var matched []string
	for _, t := range m.allTags {
		if searchPrefix == "" || strings.HasPrefix(t, searchPrefix) {
			matched = append(matched, t)
		}
		if len(matched) >= 5 {
			break
		}
	}
	if len(matched) == 0 {
		return ""
	}
	return "  " + m.theme.Dim.Render("tags:") + " " + m.theme.Filter.Render(strings.Join(matched, "  "))
}

func (m tuiModel) launchEditorCreate() (tuiModel, tea.Cmd) {
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "daily-new-*.md")
	if err != nil {
		m.statusMsg = fmt.Sprintf("Failed to create temp file: %v", err)
		return m, nil
	}
	defer tmpFile.Close()
	_, _ = tmpFile.WriteString("# Write your memo here\n\n")
	tempPath := tmpFile.Name()

	editor := tui.GetEditor()
	c := exec.Command(editor, tempPath)
	return m, tea.Exec(tui.ExecCommandWrapper{c}, func(err error) tea.Msg {
		if err != nil {
			return tui.EditFinishedMsg{Err: err, TempPath: tempPath}
		}
		data, err := os.ReadFile(tempPath)
		if err != nil {
			return tui.EditFinishedMsg{Err: err, TempPath: tempPath}
		}
		return tui.EditFinishedMsg{
			Content:  string(data),
			MemoUUID: "",
			TempPath: tempPath,
		}
	})
}

func (m tuiModel) launchEditorEdit() (tuiModel, tea.Cmd) {
	if m.cursorIdx < 0 || m.cursorIdx >= len(m.memos) {
		m.statusMsg = "No memo selected"
		return m, nil
	}
	memo := m.memos[m.cursorIdx]
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "daily-memo-*.md")
	if err != nil {
		m.statusMsg = fmt.Sprintf("Failed to create temp file: %v", err)
		return m, nil
	}
	defer tmpFile.Close()
	_, _ = tmpFile.WriteString(memo.Content)
	tempPath := tmpFile.Name()

	editor := tui.GetEditor()
	c := exec.Command(editor, tempPath)
	return m, tea.Exec(tui.ExecCommandWrapper{c}, func(err error) tea.Msg {
		if err != nil {
			return tui.EditFinishedMsg{Err: err, TempPath: tempPath}
		}
		data, err := os.ReadFile(tempPath)
		if err != nil {
			return tui.EditFinishedMsg{Err: err, TempPath: tempPath}
		}
		return tui.EditFinishedMsg{
			Content:  string(data),
			MemoUUID: memo.UUID,
			TempPath: tempPath,
		}
	})
}

func (m tuiModel) getPaletteCommands() []paletteCmd {
	return []paletteCmd{
		{
			label: "Filter Memos by DSL Query",
			desc:  "Enter search keyword, tag filters (t:work), date range or sort criteria",
			key:   ":",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				m.mode = modeDSLInput
				m.cmdBuf = ""
				m.statusMsg = "dsl filter: enter search/sort query (Esc to cancel)"
				return m, nil
			},
		},
		{
			label: "Create Memo (via System Editor)",
			desc:  "Open system editor to write a multi-line memo",
			key:   "C",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				return m.launchEditorCreate()
			},
		},
		{
			label: "Create Memo (Inline Text)",
			desc:  "Quickly compose a single-line memo inside the terminal",
			key:   "c",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				m.mode = modeCreate
				m.createBuf = ""
				m.statusMsg = "Enter content (Esc to cancel, Enter to submit)"
				return m, nil
			},
		},
		{
			label: "Edit Selected Memo",
			desc:  "Open the selected memo in the system editor to modify it",
			key:   "e",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				return m.launchEditorEdit()
			},
		},
		{
			label: "Edit Memo Tags",
			desc:  "Modify tags for selected memos or the current memo under cursor",
			key:   "T",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
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
				return m, nil
			},
		},
		{
			label: "Delete Memo(s)",
			desc:  "Permanently delete selected or highlighted memo(s)",
			key:   "d",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
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
				return m, nil
			},
		},
		{
			label: "Archive / Restore Memo(s)",
			desc:  "Archive active memo(s) or restore archived memo(s) to Inbox",
			key:   "a",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				return m.executeArchiveRestore()
			},
		},
		{
			label: "Play Memo (Text-to-Speech)",
			desc:  "Read aloud the full memo content",
			key:   "p",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				if m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
					text := tui.CleanMarkdownForTTS(m.memos[m.cursorIdx].Content)
					if text != "" {
						go tui.SpeakWithCache(text, m.memos[m.cursorIdx].UUID, m.memos[m.cursorIdx].ExpiresAt, "T", &m.app.Config.TTS)
						m.statusMsg = fmt.Sprintf("▶ (%s)", m.memos[m.cursorIdx].UUID[:8])
					} else {
						m.statusMsg = "Empty memo"
					}
				}
				return m, nil
			},
		},
		{
			label: "Show TTS Cache Status",
			desc:  "Display cache info for the current memo's translation audio",
			key:   "s",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				if m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
					m.statusMsg = tui.ShowCacheStatus(m.memos[m.cursorIdx].UUID, "T", m.memos[m.cursorIdx].ExpiresAt)
				}
				return m, nil
			},
		},
		{
			label: "TTS Configuration",
			desc:  "View and edit TTS settings (provider, voice, style, etc.)",
			key:   "S",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				m.configItems = buildConfigItems(m.app.Config, m.app.ConfigSources)
				m.configIdx = 0
				m.mode = modeConfig
				m.statusMsg = ""
				return m, nil
			},
		},
		{
			label: "Toggle Memo Selection Checkbox",
			desc:  "Mark memo for batch operations (archive, delete, tag)",
			key:   "x",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				if m.cursorIdx >= 0 && m.cursorIdx < len(m.memos) {
					uid := m.memos[m.cursorIdx].UUID
					if m.selectedMap == nil {
						m.selectedMap = make(map[string]bool)
					}
					m.selectedMap[uid] = !m.selectedMap[uid]
					if !m.selectedMap[uid] {
						delete(m.selectedMap, uid)
					}
				}
				return m, nil
			},
		},
		{
			label: "Clear All Active Filters",
			desc:  "Reset tag, search query, date filters, sorting, and selections",
			key:   "esc",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
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
				m.statusMsg = "Filters cleared"
				m.refreshAll()
				return m, nil
			},
		},
		{
			label: "Switch View Tab: Inbox",
			desc:  "Show active memos dashboard",
			key:   "1",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				model, cmd := m.switchTab(tabInbox)
				return model, cmd
			},
		},
		{
			label: "Switch View Tab: Archived",
			desc:  "Show archived memos storage",
			key:   "2",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				model, cmd := m.switchTab(tabArchived)
				return model, cmd
			},
		},
		{
			label: "Switch View Tab: Tags list",
			desc:  "Browse tag cloud with statistics",
			key:   "3",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				model, cmd := m.switchTab(tabTags)
				return model, cmd
			},
		},
		{
			label: "Switch View Tab: Statistics",
			desc:  "Show memo and tag metrics dashboard",
			key:   "4",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				model, cmd := m.switchTab(tabStats)
				return model, cmd
			},
		},
		{
			label: "Reset View / Reload",
			desc:  "Sync database and reload memo lists",
			key:   "r",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				m.statusMsg = "Reset"
				model, cmd := m.switchTab(m.activeTab)
				return model, cmd
			},
		},
		{
			label: "Quit Daily TUI",
			desc:  "Exit interactive terminal application",
			key:   "q",
			run: func(m tuiModel) (tuiModel, tea.Cmd) {
				m.quitting = true
				return m, tea.Quit
			},
		},
	}
}

func (m tuiModel) getFilteredCommands() []paletteCmd {
	all := m.getPaletteCommands()
	if m.cmdBuf == "" {
		return all
	}
	var filtered []paletteCmd
	for _, c := range all {
		if strings.Contains(strings.ToLower(c.label), strings.ToLower(m.cmdBuf)) ||
			strings.Contains(strings.ToLower(c.desc), strings.ToLower(m.cmdBuf)) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
