package cmd

import (
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Palette ──

// Palette defines semantic colors for a theme.
type Palette struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Muted      lipgloss.Color
	DarkMuted  lipgloss.Color
	Success    lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Foreground lipgloss.Color
	SelBg      lipgloss.Color
}

// ── Theme ──

// Theme holds a complete set of derived styles from a Palette.
type Theme struct {
	Palette

	TopBar       lipgloss.Style
	Count        lipgloss.Style
	Sel          lipgloss.Style
	SelBg        lipgloss.Style
	Dim          lipgloss.Style
	DimBold      lipgloss.Style
	Accent       lipgloss.Style
	Bullet       lipgloss.Style
	Cursor       lipgloss.Style
	ExpIcon      lipgloss.Style
	Filter       lipgloss.Style
	Status       lipgloss.Style
	Error        lipgloss.Style
	HelpKey      lipgloss.Style
	HelpSep      lipgloss.Style
	HR           lipgloss.Style
	HRRule       lipgloss.Style
	DetailHR     lipgloss.Style
	Create       lipgloss.Style
	CheckOn      lipgloss.Style
	CheckOff     lipgloss.Style
	ListBorder   lipgloss.Style
	DetailBorder lipgloss.Style

	// Tab navigation
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	// Markdown header styles
	H1Style lipgloss.Style
	H2Style lipgloss.Style
	H3Style lipgloss.Style
	H4Style lipgloss.Style

	// Markdown block styles
	CodeBlockStyle lipgloss.Style
	QuoteStyle     lipgloss.Style
	QuoteBorder    lipgloss.Style
	BulletStyle    lipgloss.Style
	NumStyle       lipgloss.Style
	TaskDoneStyle  lipgloss.Style
	TaskTodoStyle  lipgloss.Style

	// Inline styles
	LinkStyle       lipgloss.Style
	URLStyle        lipgloss.Style
	TagStyle        lipgloss.Style
	InlineCodeStyle lipgloss.Style
	BoldStyle       lipgloss.Style
	ItalicStyle     lipgloss.Style
	StrikeStyle     lipgloss.Style

	// Status badge
	StatusActive   lipgloss.Style
	StatusArchived lipgloss.Style

	// Detail block highlight
	BlockHighlight lipgloss.Style

	// Stats border
	StatsBorder lipgloss.Style
	DetailBox   lipgloss.Style
}

// NewTheme builds a Theme from a Palette.
func NewTheme(p Palette) *Theme {
	t := &Theme{Palette: p}

	t.TopBar = lipgloss.NewStyle().Bold(true).Foreground(p.Primary).MarginLeft(1)
	t.Count = lipgloss.NewStyle().Foreground(p.Muted)
	t.Sel = lipgloss.NewStyle().Bold(true).Foreground(p.Primary)
	t.SelBg = lipgloss.NewStyle().Bold(true).Foreground(p.Primary).Background(p.SelBg)
	t.Dim = lipgloss.NewStyle().Foreground(p.Muted)
	t.DimBold = lipgloss.NewStyle().Foreground(p.Muted).Bold(true)
	t.Accent = lipgloss.NewStyle().Foreground(p.Accent)
	t.Bullet = t.DimBold.Copy().SetString("◯")
	t.Cursor = t.Sel.Copy().SetString("▸")
	t.ExpIcon = t.Dim.Copy().SetString("⤵")
	t.Filter = lipgloss.NewStyle().Foreground(p.Secondary).Bold(true)
	t.Status = lipgloss.NewStyle().Foreground(p.Warning)
	t.Error = lipgloss.NewStyle().Foreground(p.Error)
	t.HelpKey = lipgloss.NewStyle().Foreground(p.Secondary).Bold(true)
	t.HelpSep = t.Dim.Copy().SetString(" · ")
	t.HR = lipgloss.NewStyle().Foreground(p.DarkMuted)
	t.HRRule = t.HR.Copy().SetString("──────────────────────────────────────────────────────────────")
	t.DetailHR = lipgloss.NewStyle().Foreground(p.DarkMuted).SetString("  ──")
	t.Create = lipgloss.NewStyle().Foreground(p.Success).Bold(true)
	t.CheckOn = lipgloss.NewStyle().Foreground(p.Success).Bold(true).SetString("☑ ")
	t.CheckOff = lipgloss.NewStyle().Foreground(p.Muted).SetString("☐ ")
	t.ListBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Secondary).Padding(0, 1)
	t.DetailBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Primary).Padding(0, 1)

	// Tab styles
	t.TabActive = lipgloss.NewStyle().Bold(true).Foreground(p.Foreground).Background(p.Primary).Padding(0, 2).MarginRight(2)
	t.TabInactive = lipgloss.NewStyle().Foreground(p.Muted).Background(p.DarkMuted).Padding(0, 2).MarginRight(2)

	// Markdown header styles
	t.H1Style = lipgloss.NewStyle().Bold(true).Foreground(p.Primary)
	t.H2Style = lipgloss.NewStyle().Bold(true).Foreground(p.Secondary)
	t.H3Style = lipgloss.NewStyle().Bold(true).Foreground(p.Warning)
	t.H4Style = lipgloss.NewStyle().Bold(true).Foreground(p.Muted)

	// Markdown block styles
	t.CodeBlockStyle = lipgloss.NewStyle().Foreground(p.Foreground).Background(p.DarkMuted)
	t.QuoteStyle = lipgloss.NewStyle().Italic(true).Foreground(p.Muted)
	t.QuoteBorder = lipgloss.NewStyle().Foreground(p.Secondary).SetString("│ ")
	t.BulletStyle = lipgloss.NewStyle().Foreground(p.Success).Bold(true).SetString("• ")
	t.NumStyle = lipgloss.NewStyle().Foreground(p.Success).Bold(true)
	t.TaskDoneStyle = lipgloss.NewStyle().Foreground(p.Success).Bold(true).SetString("☑ ")
	t.TaskTodoStyle = lipgloss.NewStyle().Foreground(p.Muted).SetString("☐ ")

	// Inline styles
	t.LinkStyle = lipgloss.NewStyle().Underline(true).Foreground(p.Primary)
	t.URLStyle = lipgloss.NewStyle().Foreground(p.Muted)
	t.TagStyle = lipgloss.NewStyle().Foreground(p.Secondary).Bold(true)
	t.InlineCodeStyle = lipgloss.NewStyle().Foreground(p.Warning).Background(p.DarkMuted).Padding(0, 1)
	t.BoldStyle = lipgloss.NewStyle().Bold(true)
	t.ItalicStyle = lipgloss.NewStyle().Italic(true)
	t.StrikeStyle = lipgloss.NewStyle().Strikethrough(true).Foreground(p.Muted)

	// Status badge
	t.StatusActive = lipgloss.NewStyle().Foreground(p.Success).Bold(true).SetString("[Active]")
	t.StatusArchived = lipgloss.NewStyle().Foreground(p.Muted).Bold(true).SetString("[Archived]")

	// Detail block highlight
	t.BlockHighlight = lipgloss.NewStyle().Background(lipgloss.Color("#331a33")).Foreground(lipgloss.Color("#e5a4ff")).Bold(true)

	// Stats border
	t.StatsBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Secondary).Padding(1, 2).Width(45)
	t.DetailBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Primary).Padding(1, 2)

	return t
}

// ── Built-in Palettes ──

var paletteEmerald = Palette{
	Primary:    lipgloss.Color("#22c55e"),
	Secondary:  lipgloss.Color("#10b981"),
	Accent:     lipgloss.Color("#34d399"),
	Muted:      lipgloss.Color("#6b7280"),
	DarkMuted:  lipgloss.Color("#333333"),
	Success:    lipgloss.Color("#22c55e"),
	Error:      lipgloss.Color("#ef4444"),
	Warning:    lipgloss.Color("#f59e0b"),
	Foreground: lipgloss.Color("#e0e0e0"),
	SelBg:      lipgloss.Color("#0d3320"),
}

var paletteOcean = Palette{
	Primary:    lipgloss.Color("#00d4ff"),
	Secondary:  lipgloss.Color("#a855f7"),
	Accent:     lipgloss.Color("#38bdf8"),
	Muted:      lipgloss.Color("#6b7280"),
	DarkMuted:  lipgloss.Color("#333333"),
	Success:    lipgloss.Color("#22c55e"),
	Error:      lipgloss.Color("#ef4444"),
	Warning:    lipgloss.Color("#f59e0b"),
	Foreground: lipgloss.Color("#e0e0e0"),
	SelBg:      lipgloss.Color("#0a2540"),
}

var paletteAmethyst = Palette{
	Primary:    lipgloss.Color("#a855f7"),
	Secondary:  lipgloss.Color("#7c3aed"),
	Accent:     lipgloss.Color("#c084fc"),
	Muted:      lipgloss.Color("#6b7280"),
	DarkMuted:  lipgloss.Color("#333333"),
	Success:    lipgloss.Color("#22c55e"),
	Error:      lipgloss.Color("#ef4444"),
	Warning:    lipgloss.Color("#f59e0b"),
	Foreground: lipgloss.Color("#e0e0e0"),
	SelBg:      lipgloss.Color("#1a0d33"),
}

var paletteSunset = Palette{
	Primary:    lipgloss.Color("#f97316"),
	Secondary:  lipgloss.Color("#f59e0b"),
	Accent:     lipgloss.Color("#fb923c"),
	Muted:      lipgloss.Color("#6b7280"),
	DarkMuted:  lipgloss.Color("#333333"),
	Success:    lipgloss.Color("#22c55e"),
	Error:      lipgloss.Color("#ef4444"),
	Warning:    lipgloss.Color("#fbbf24"),
	Foreground: lipgloss.Color("#e0e0e0"),
	SelBg:      lipgloss.Color("#331a0a"),
}

var paletteMono = Palette{
	Primary:    lipgloss.Color("#d4d4d4"),
	Secondary:  lipgloss.Color("#a3a3a3"),
	Accent:     lipgloss.Color("#e5e5e5"),
	Muted:      lipgloss.Color("#737373"),
	DarkMuted:  lipgloss.Color("#404040"),
	Success:    lipgloss.Color("#a3a3a3"),
	Error:      lipgloss.Color("#ef4444"),
	Warning:    lipgloss.Color("#d4d4d4"),
	Foreground: lipgloss.Color("#e0e0e0"),
	SelBg:      lipgloss.Color("#262626"),
}

// ── Theme Registry ──

var builtinThemes = map[string]Palette{
	"emerald":  paletteEmerald,
	"ocean":    paletteOcean,
	"amethyst": paletteAmethyst,
	"sunset":   paletteSunset,
	"mono":     paletteMono,
}

// GetTheme returns a Theme by name. Falls back to "ocean" if not found.
func GetTheme(name string) *Theme {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		name = "ocean"
	}
	if p, ok := builtinThemes[name]; ok {
		return NewTheme(p)
	}
	return NewTheme(paletteOcean)
}

// ThemeNames returns sorted list of built-in theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(builtinThemes))
	for k := range builtinThemes {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// LoadTheme resolves theme name from config → env → default.
func LoadTheme(configTheme string) *Theme {
	name := configTheme
	if name == "" {
		name = os.Getenv("TUI_THEME")
	}
	return GetTheme(name)
}
