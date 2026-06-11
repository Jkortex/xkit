package memo

import (
	"daily/internal/domain/entity"
	"testing"
)

func TestParseSearchDSL(t *testing.T) {
	tests := []struct {
		input       string
		wantTags    []string
		wantExclude []string
		wantHasRes  *bool
		wantStatus  *entity.RowStatus
		wantSearch  string
	}{
		{
			input:      "hello world",
			wantTags:   []string{},
			wantSearch: "hello world",
		},
		{
			input:      "tag:Work meeting",
			wantTags:   []string{"Work"},
			wantSearch: "meeting",
		},
		{
			input:      "tag:Work tag:Meeting 会议",
			wantTags:   []string{"Work", "Meeting"},
			wantSearch: "会议",
		},
		{
			input:       "tag:Work -tag:Draft 总结",
			wantTags:    []string{"Work"},
			wantExclude: []string{"Draft"},
			wantSearch:  "总结",
		},
		{
			input:      "has:attachment tag:Project",
			wantTags:   []string{"Project"},
			wantHasRes: boolPtr(true),
			wantSearch: "",
		},
		{
			input:      "after:2026-01-01 before:2026-05-01 test",
			wantSearch: "test",
		},
		{
			input:      "from:2026-01-01 to:2026-05-01 test",
			wantSearch: "test",
		},
		{
			input:      "is:normal tag:Work",
			wantTags:   []string{"Work"},
			wantStatus: rowStatusPtr(entity.RowStatusNormal),
			wantSearch: "",
		},
		{
			input:      "is:archived tag:Work",
			wantTags:   []string{"Work"},
			wantStatus: rowStatusPtr(entity.RowStatusArchived),
			wantSearch: "",
		},
	}

	for _, tt := range tests {
		got := ParseSearchDSL(tt.input)
		if len(got.Tags) != len(tt.wantTags) || got.Search != tt.wantSearch {
			t.Errorf("ParseSearchDSL(%q) mismatch. got=%+v, wantSearch=%q", tt.input, got, tt.wantSearch)
			continue
		}
		if len(got.TagsExclude) != len(tt.wantExclude) {
			t.Errorf("ParseSearchDSL(%q) TagsExclude mismatch. got=%v, want=%v", tt.input, got.TagsExclude, tt.wantExclude)
		}
		if tt.wantHasRes != nil && (got.HasResource == nil || *got.HasResource != *tt.wantHasRes) {
			t.Errorf("ParseSearchDSL(%q) HasResource mismatch. got=%v, want=%v", tt.input, got.HasResource, tt.wantHasRes)
		}
		if tt.wantStatus != nil && (got.RowStatus == nil || *got.RowStatus != *tt.wantStatus) {
			t.Errorf("ParseSearchDSL(%q) RowStatus mismatch. got=%v, want=%v", tt.input, got.RowStatus, tt.wantStatus)
		}
	}
}

func boolPtr(b bool) *bool                              { return &b }
func rowStatusPtr(s entity.RowStatus) *entity.RowStatus { return &s }
