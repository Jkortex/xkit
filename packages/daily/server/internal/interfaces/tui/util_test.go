package tui

import (
	"testing"
	"time"
)

func TestSimplifyTime(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "ISO 8601 Z",
			in:   "2026-06-09T14:30:00Z",
			want: time.Date(2026, 6, 9, 14, 30, 0, 0, time.UTC).Local().Format("01-02 15:04"),
		},
		{
			name: "ISO 8601 with millis",
			in:   "2026-06-09T14:30:00.000Z",
			want: time.Date(2026, 6, 9, 14, 30, 0, 0, time.UTC).Local().Format("01-02 15:04"),
		},
		{
			name: "SQL datetime",
			in:   "2026-06-09 14:30:00",
			want: time.Date(2026, 6, 9, 14, 30, 0, 0, time.UTC).Local().Format("01-02 15:04"),
		},
		{
			name: "already short",
			in:   "06-09 14:30",
			want: "06-09 14:30",
		},
		{
			name: "empty",
			in:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SimplifyTime(tt.in)
			if got != tt.want {
				t.Errorf("SimplifyTime(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSplitTags(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "single", in: "work", want: []string{"work"}},
		{name: "comma separated", in: "work,personal", want: []string{"work", "personal"}},
		{name: "with spaces", in: "  work , personal ", want: []string{"work", "personal"}},
		{name: "empty", in: "", want: nil},
		{name: "only commas", in: ",,,", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitTags(tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("SplitTags(%q) = %v (len=%d), want %v (len=%d)", tt.in, got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SplitTags(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsChineseText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "pure Chinese", in: "你好世界", want: true},
		{name: "mixed mostly Chinese", in: "你好世界hello", want: true},
		{name: "pure English", in: "hello world", want: false},
		{name: "mixed mostly English", in: "hello world 你好世界", want: true}, // 4/17 > 0.15
		{name: "empty", in: "", want: false},
		{name: "numbers only", in: "12345", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsChineseText(tt.in)
			if got != tt.want {
				t.Errorf("IsChineseText(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsAlphaNum(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "lowercase", r: 'a', want: true},
		{name: "uppercase", r: 'Z', want: true},
		{name: "digit", r: '5', want: true},
		{name: "unicode CJK", r: '你', want: true},
		{name: "space", r: ' ', want: false},
		{name: "hyphen", r: '-', want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAlphaNum(tt.r)
			if got != tt.want {
				t.Errorf("IsAlphaNum(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestIsTagChar(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "lowercase", r: 'a', want: true},
		{name: "digit", r: '5', want: true},
		{name: "hyphen", r: '-', want: true},
		{name: "underscore", r: '_', want: true},
		{name: "slash", r: '/', want: true},
		{name: "unicode CJK", r: '你', want: true},
		{name: "space", r: ' ', want: false},
		{name: "dot", r: '.', want: false},
		{name: "at", r: '@', want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTagChar(tt.r)
			if got != tt.want {
				t.Errorf("IsTagChar(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestCleanMarkdownForTTS(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bold and italic",
			in:   "**hello** *world*",
			want: "hello world",
		},
		{
			name: "code blocks",
			in:   "```go\nfmt.Println()\n```",
			want: "fmt.Println()",
		},
		{
			name: "headers",
			in:   "## Title\ncontent",
			want: "Title\ncontent",
		},
		{
			name: "links",
			in:   "[text](url)",
			want: "text",
		},
		{
			name: "task lists",
			in:   "- [ ] todo\n- [x] done",
			want: "todo\ndone",
		},
		{
			name: "bullet lists",
			in:   "- item\n* item",
			want: "item\nitem",
		},
		{
			name: "empty",
			in:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanMarkdownForTTS(tt.in)
			if got != tt.want {
				t.Errorf("CleanMarkdownForTTS(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractTranslation(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "[T] marker",
			in:   "Hello\n[T]Bonjour\n[V]vocab",
			want: "Bonjour",
		},
		{
			name: "## Translation marker",
			in:   "Hello\n## Translation\nBonjour\n## Vocabulary\nwords",
			want: "Bonjour",
		},
		{
			name: "no translation marker",
			in:   "Hello world",
			want: "",
		},
		{
			name: "## 翻译 marker",
			in:   "Hello\n## 翻译\n你好\n## 词汇\nword",
			want: "你好",
		},
		{
			name: "empty content",
			in:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTranslation(tt.in)
			if got != tt.want {
				t.Errorf("ExtractTranslation(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGetShortSummary(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "first meaningful line",
			in:   "## Translation\n\nThis is the content",
			want: "This is the content",
		},
		{
			name: "skip structure markers",
			in:   "[T]\n[V]\nReal content here",
			want: "Real content here",
		},
		{
			name: "skip header lines",
			in:   "# Translation\n# Vocabulary\nActual memo content",
			want: "Actual memo content",
		},
		{
			name: "empty content",
			in:   "",
			want: "Empty Memo",
		},
		{
			name: "only markers",
			in:   "[T]\n[V]\n",
			want: "Empty Memo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetShortSummary(tt.in)
			if got != tt.want {
				t.Errorf("GetShortSummary(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseMarkdownBlocks(t *testing.T) {
	type testCase struct {
		name    string
		in      string
		wantLen int
		check   func(t *testing.T, blocks []MarkdownBlock)
	}
	tests := []testCase{
		{
			name:    "empty",
			in:      "",
			wantLen: 0,
		},
		{
			name:    "single header",
			in:      "# Hello",
			wantLen: 1,
		},
		{
			name:    "header and paragraph",
			in:      "# Hello\n\nWorld content",
			wantLen: 2,
		},
		{
			name:    "code block",
			in:      "```go\nfmt.Println()\n```",
			wantLen: 1,
			check:   func(t *testing.T, blocks []MarkdownBlock) { t.Helper(); if blocks[0].Type != "code" { t.Errorf("block type = %q, want %q", blocks[0].Type, "code") } },
		},
		{
			name:    "list items",
			in:      "- item1\n- item2",
			wantLen: 2,
		},
		{
			name:    "mixed: header + paragraph + list",
			in:      "# Title\n\ncontent\n\n- a\n- b",
			wantLen: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMarkdownBlocks(tt.in)
			if len(got) != tt.wantLen {
				t.Errorf("ParseMarkdownBlocks(%q) returned %d blocks, want %d", tt.in, len(got), tt.wantLen)
				for i, b := range got {
					t.Logf("  block[%d]: type=%q content=%q", i, b.Type, b.Content)
				}
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestAnySlice(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{name: "ints", in: []int{1, 2, 3}, want: 3},
		{name: "empty", in: []int{}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnySlice(tt.in)
			if len(got) != tt.want {
				t.Errorf("AnySlice(%v) len = %d, want %d", tt.in, len(got), tt.want)
			}
		})
	}
}

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		maxWidth int
		want     string
	}{
		{name: "ascii fits", in: "hello", maxWidth: 10, want: "hello"},
		{name: "ascii truncate", in: "hello world", maxWidth: 5, want: "hello"},
		{name: "cjk fits", in: "你好", maxWidth: 4, want: "你好"},
		{name: "cjk truncate", in: "你好世界", maxWidth: 4, want: "你好"},
		{name: "empty", in: "", maxWidth: 10, want: ""},
		{name: "zero max", in: "hello", maxWidth: 0, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateToWidth(tt.in, tt.maxWidth)
			if got != tt.want {
				t.Errorf("TruncateToWidth(%q, %d) = %q, want %q", tt.in, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestWrapLine(t *testing.T) {
	tests := []struct {
		name string
		in   string
		max  int
		want int // expected number of wrapped lines
	}{
		{name: "fits in one", in: "hello world", max: 50, want: 1},
		{name: "wraps", in: "hello world foo bar", max: 10, want: 3},
		{name: "empty", in: "", max: 50, want: 1},
		{name: "zero max", in: "hello", max: 0, want: 1},
		{name: "long word breaks", in: "superlongword", max: 5, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapLine(tt.in, tt.max)
			if len(got) != tt.want {
				t.Errorf("WrapLine(%q, %d) returned %d lines, want %d", tt.in, tt.max, len(got), tt.want)
				for i, l := range got {
					t.Logf("  line[%d]: %q", i, l)
				}
			}
		})
	}
}
