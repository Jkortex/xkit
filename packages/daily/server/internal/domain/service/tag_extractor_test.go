package service

import (
	"reflect"
	"testing"
)

func TestTagExtractor_Extract(t *testing.T) {
	extractor := NewTagExtractor()

	tests := []struct {
		name        string
		content     string
		wantTags    []string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "Pure tag lines at the end",
			content:     "Main content\n#Tag1 #Tag2\n#Tag3",
			wantTags:    []string{"Tag1", "Tag2", "Tag3"},
			wantContent: "Main content",
			wantErr:     false,
		},
		{
			name:        "Tags with trailing spaces",
			content:     "Hello world\n#Tag1 \n\n",
			wantTags:    []string{"Tag1"},
			wantContent: "Hello world",
			wantErr:     false,
		},
		{
			name:        "Inline tags are IGNORED and kept in content",
			content:     "I am #working now.\n#Tag1",
			wantTags:    []string{"Tag1"},
			wantContent: "I am #working now.",
			wantErr:     false,
		},
		{
			name:        "Multiple lines of tags",
			content:     "Line 1\n#T1\n#T2",
			wantTags:    []string{"T1", "T2"},
			wantContent: "Line 1",
			wantErr:     false,
		},
		{
			name:        "Only tags",
			content:     "#T1\n#T2",
			wantTags:    []string{"T1", "T2"},
			wantContent: "",
			wantErr:     false,
		},
		{
			name:        "Invalid tag in tag block causes error",
			content:     "Content\n#!!!",
			wantTags:    nil,
			wantContent: "",
			wantErr:     true,
		},
		{
			name:        "Tag too long causes error",
			content:     "Content\n#ThisTagIsWayTooLongForOurSystemToAccept",
			wantTags:    nil,
			wantContent: "",
			wantErr:     true,
		},
		{
			name:        "Mixed content and tags in one line (not a formal tag line)",
			content:     "Note with #tag",
			wantTags:    nil,
			wantContent: "Note with #tag",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTags, gotContent, err := extractor.Extract(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTags, tt.wantTags) {
				t.Errorf("Extract() gotTags = %v, wantTags %v", gotTags, tt.wantTags)
			}
			if gotContent != tt.wantContent {
				t.Errorf("Extract() gotContent = %q, wantContent %q", gotContent, tt.wantContent)
			}
		})
	}
}
