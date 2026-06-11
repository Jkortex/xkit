package tui

import (
	"time"

	"daily/internal/infrastructure/config"
	"github.com/charmbracelet/lipgloss"
)

type UIMemo struct {
	UUID      string
	Content   string
	CreatedAt string
	Expanded  bool
	Tags      []string
	RowStatus string
	ExpiresAt *time.Time
}

type TagStat struct {
	Name  string
	Count int
}

type ConfigItem struct {
	Label    string
	Key      string
	Value    string
	Type     string
	Options  []string
	Source   config.ConfigSource
	ReadOnly bool
}

type InlineState struct {
	InBold   bool
	InItalic bool
	InCode   bool
	InStrike bool
}

type MarkdownStyle struct {
	H1Style, H2Style, H3Style, H4Style lipgloss.Style
	BoldStyle, ItalicStyle, InlineCodeStyle, StrikeStyle lipgloss.Style
	LinkStyle, URLStyle, TagStyle lipgloss.Style
	CodeBlockStyle lipgloss.Style
	QuoteStyle, QuoteBorder lipgloss.Style
	BulletStyle, NumStyle lipgloss.Style
	TaskDoneStyle, TaskTodoStyle lipgloss.Style
	HR lipgloss.Style
}
