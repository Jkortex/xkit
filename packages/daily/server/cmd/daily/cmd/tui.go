package cmd

import (
	"context"
	"fmt"
	"os"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	tui "daily/internal/interfaces/tui"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// tuiCmd 表示交互式 TUI 子命令
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI for managing memos",
	Long:  `Launch an interactive terminal UI for browsing, creating, filtering, and managing memos in real-time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return cmd.Help()
		}
		app, err := getApp()
		if err != nil {
			return err
		}

		// Pre-resolve default editor in background to eliminate latency on first launch
		go tui.GetEditor()

		tui.EnsureEventPath(getEventFilePath())
		eventPath := getEventFilePath()

		ctx := context.Background()
		listResp, _ := app.MemoCtrl.List(ctx, app.AdminID, port.MemoFilter{Limit: 50, RowStatus: statusToRowStatus("normal")})
		memos := make([]tui.UIMemo, 0, len(listResp))
		var lastTime string
		for _, r := range listResp {
			u := memoResponseToUIMemo(r)
			memos = append(memos, u)
			lastTime = u.CreatedAt
		}
		cursorIdx := len(memos) - 1
		if cursorIdx < 0 {
			cursorIdx = 0
		}

		tagResp, _ := app.TagCtrl.ListTags(ctx, app.AdminID)
		allTags := make([]string, len(tagResp))
		for i, t := range tagResp {
			allTags[i] = t.Name
		}
		initOffset := 0
		if cursorIdx > 10 {
			initOffset = cursorIdx - 10 + 1
		}
		theme := LoadTheme(app.Config.Theme)
		activeTheme = theme

		m := tuiModel{
			app:          app,
			mode:         modeList,
			memos:        memos,
			cursorIdx:    cursorIdx,
			offset:       initOffset,
			statusMsg:    fmt.Sprintf("%d memos loaded", len(memos)),
			eventPath:    eventPath,
			lastTime:     lastTime,
			allTags:      allTags,
			statusFilter: "normal",
			selectedMap:  make(map[string]bool),
			activeTab:    tabInbox,
			theme:        theme,
		}

		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

// ── TUI model ──

type tuiMode int

const (
	modeList tuiMode = iota
	modeCreate
	modeDeleteConfirm
	modeCmd
	modeTagInput
	modeDetail
	modeDSLInput
	modeConfig
	modeConfigEdit
)

type tabType int

const (
	tabInbox tabType = iota
	tabArchived
	tabTags
	tabStats
)

type tuiModel struct {
	mode      tuiMode
	memos     []tui.UIMemo
	cursorIdx int
	statusMsg string
	tagFilter string // simple single-tag filter (from t key)
	searchTxt string
	maxItems  int
	allTags   []string
	tagIdx    int
	width     int
	offset    int
	height    int
	quitting  bool

	app       *AppContext
	eventPath string
	lastTime  string
	theme     *Theme

	createBuf string
	createTag string

	delUUID string
	cmdBuf  string

	// advanced filters (from command palette)
	tagOr    []string // tag:a,b — any of
	tagAnd   []string // tag+:a,b — all of
	tagExcl  []string // tag-:a   — exclude
	fromDt   string   // from:YYYY-MM-DD
	toDt     string   // to:YYYY-MM-DD
	sortFld  string   // sort:created_at|updated_at
	sortDir  string   // ASC|DESC

	statusFilter  string          // "normal", "archived", "all"
	selectedMap   map[string]bool // map of memo UUID -> selected status
	previewScroll int             // scroll line offset for preview pane
	tagInputBuf   string          // buffer for editing tags

	// redesigned Tab-based TUI fields
	activeTab     tabType
	tagsWithCount []tui.TagStat
	statsData     *dto.StatsResponse
	cmdSelIdx     int // selected index in command palette list

	// config mode fields
	configItems   []configItem // TTS config items for display
	configIdx     int          // cursor in config list
	configEditBuf string       // input buffer for editing

	// block-by-block detail fields
	detailBlocks   []tui.MarkdownBlock
	detailBlockIdx int
}

// activeTheme is set when the TUI starts; used by standalone renderers.
var activeTheme *Theme
