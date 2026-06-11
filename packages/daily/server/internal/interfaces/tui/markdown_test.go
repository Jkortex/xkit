package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func testStyle() *MarkdownStyle {
	return &MarkdownStyle{
		H1Style:         lipgloss.NewStyle(),
		H2Style:         lipgloss.NewStyle(),
		H3Style:         lipgloss.NewStyle(),
		H4Style:         lipgloss.NewStyle(),
		BoldStyle:       lipgloss.NewStyle().Bold(true),
		ItalicStyle:     lipgloss.NewStyle().Italic(true),
		InlineCodeStyle: lipgloss.NewStyle(),
		StrikeStyle:     lipgloss.NewStyle(),
		LinkStyle:       lipgloss.NewStyle(),
		URLStyle:        lipgloss.NewStyle(),
		TagStyle:        lipgloss.NewStyle(),
		CodeBlockStyle:  lipgloss.NewStyle(),
		QuoteStyle:      lipgloss.NewStyle(),
		QuoteBorder:     lipgloss.NewStyle(),
		BulletStyle:     lipgloss.NewStyle(),
		NumStyle:        lipgloss.NewStyle(),
		TaskDoneStyle:   lipgloss.NewStyle(),
		TaskTodoStyle:   lipgloss.NewStyle(),
		HR:              lipgloss.NewStyle(),
	}
}

func TestRenderInline(t *testing.T) {
	s := testStyle()

	tests := []struct {
		name string
		in   string
	}{
		{name: "plain text", in: "hello world"},
		{name: "bold", in: "hello **world**"},
		{name: "italic", in: "hello *world*"},
		{name: "code", in: "hello `world`"},
		{name: "strikethrough", in: "hello ~~world~~"},
		{name: "link", in: "hello [world](https://example.com)"},
		{name: "empty", in: ""},
		{name: "tag", in: "hello #world"},
		{name: "mixed", in: "**bold** and *italic* and `code`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderInline(tt.in, &InlineState{}, s)
			if tt.in == "" && got != "" {
				t.Errorf("expected empty, got %q", got)
			}
			if tt.in != "" && got == "" {
				t.Errorf("expected non-empty, got empty")
			}
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	s := testStyle()

	tests := []struct {
		name    string
		in      string
		width   int
		wantMin int
	}{
		{
			name:    "empty",
			in:      "",
			width:   80,
			wantMin: 0,
		},
		{
			name:    "header",
			in:      "# Hello",
			width:   80,
			wantMin: 1,
		},
		{
			name:    "header and paragraph",
			in:      "# Hello\n\nWorld content here",
			width:   80,
			wantMin: 2,
		},
		{
			name:    "code block",
			in:      "```\nfmt.Println()\n```",
			width:   80,
			wantMin: 1,
		},
		{
			name:    "bullet list",
			in:      "- item1\n- item2\n- item3",
			width:   80,
			wantMin: 3,
		},
		{
			name:    "numbered list",
			in:      "1. first\n2. second",
			width:   80,
			wantMin: 2,
		},
		{
			name:    "task list",
			in:      "- [ ] todo\n- [x] done",
			width:   80,
			wantMin: 2,
		},
		{
			name:    "blockquote",
			in:      "> quoted text",
			width:   80,
			wantMin: 1,
		},
		{
			name:    "horizontal rule",
			in:      "---\n***\n___",
			width:   80,
			wantMin: 3,
		},
		{
			name:    "complex markdown",
			in:      "# Title\n\nParagraph with **bold** and *italic*.\n\n- list item\n- another item\n\n> A quote",
			width:   80,
			wantMin: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderMarkdown(tt.in, tt.width, s)
			if len(got) < tt.wantMin {
				t.Errorf("RenderMarkdown returned %d lines, want >= %d", len(got), tt.wantMin)
				for i, l := range got {
					t.Logf("  line[%d]: %q", i, l)
				}
			}
		})
	}
}

func TestRenderMarkdownWithTheme(t *testing.T) {
	// Verify that a themed render produces output containing the original content
	content := "# Hello\n\nThis is a **test** paragraph."
	width := 80

	s := testStyle()
	result := RenderMarkdown(content, width, s)

	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	full := strings.Join(result, "\n")
	if !strings.Contains(full, "Hello") || !strings.Contains(full, "test") {
		t.Errorf("rendered output should contain original text: %q", full)
	}
}

func TestRenderInlineStateTracking(t *testing.T) {
	s := testStyle()
	state := &InlineState{}

	// Bold text should track state
	result1 := RenderInline("**hello", state, s)
	if !state.InBold {
		t.Error("expected InBold = true after opening **")
	}
	_ = result1

	result2 := RenderInline("world**", state, s)
	if state.InBold {
		t.Error("expected InBold = false after closing **")
	}
	_ = result2

	// Empty state
	empty := RenderInline("", state, s)
	if empty != "" {
		t.Errorf("expected empty, got %q", empty)
	}
}
