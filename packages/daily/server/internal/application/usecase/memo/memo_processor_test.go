package memo

import (
	"context"
	"strings"
	"testing"
	"time"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

// MockTagRepo for MemoProcessor tests
type mockTagRepo struct {
	port.TagRepository
	aliases map[string]string
}

func (m *mockTagRepo) ResolveCanonicalTag(ctx context.Context, tag string) (string, error) {
	if canonical, ok := m.aliases[tag]; ok {
		return canonical, nil
	}
	return tag, nil
}

func TestMemoProcessor_Process(t *testing.T) {
	tagRepo := &mockTagRepo{
		aliases: map[string]string{
			"SRE": "Ops",
		},
	}
	tokenizer := &MockTokenizer{} // From create_memo_test.go
	processor := NewMemoProcessor(tagRepo, tokenizer)

	ctx := context.Background()

	t.Run("Normal memo with tags and filenames", func(t *testing.T) {
		content := "Meeting notes\n#Work #SRE"
		resourceNames := []string{"plan.pdf", "image.png"}

		tags, cleanedContent, expiresAt, si, err := processor.Process(ctx, content, nil, "", resourceNames)

		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		if cleanedContent != "Meeting notes" {
			t.Errorf("Expected cleaned content 'Meeting notes', got %q", cleanedContent)
		}

		// 1. Verify Tags (including canonical resolution)
		expectedTags := []string{"Work", "Ops"}
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %v", tags)
		}
		for _, tag := range expectedTags {
			found := false
			for _, got := range tags {
				if got == tag {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Tag %s not found in results %v", tag, tags)
			}
		}

		// 2. Verify Expiration
		if expiresAt != nil {
			t.Error("Expected no expiration for normal memo")
		}

		// 3. Verify SearchIndex fields
		if !strings.Contains(si.TagsTokens, "Work") || !strings.Contains(si.TagsTokens, "Ops") {
			t.Errorf("Indexed tags missing: %s", si.TagsTokens)
		}
		if !strings.Contains(si.FilesTokens, "plan.pdf") || !strings.Contains(si.FilesTokens, "image.png") {
			t.Errorf("Indexed filenames missing: %s", si.FilesTokens)
		}
		if !strings.Contains(si.BodyTokens, "Meeting notes") {
			t.Errorf("Indexed body missing original content: %s", si.BodyTokens)
		}
		if si.IsEphemeral {
			t.Error("Normal memo should not be ephemeral")
		}
	})

	t.Run("Ephemeral memo with temp tag", func(t *testing.T) {
		content := "Temporary thought\n#temp"

		tags, cleanedContent, expiresAt, si, err := processor.Process(ctx, content, nil, "", nil)

		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		if cleanedContent != "Temporary thought" {
			t.Errorf("Expected cleaned content 'Temporary thought', got %q", cleanedContent)
		}
		if expiresAt == nil {
			t.Fatal("Expected expiration for #temp memo")
		}

		duration := time.Until(*expiresAt)
		if duration < 71*time.Hour || duration > 73*time.Hour {
			t.Errorf("Expected ~72h TTL, got %v", duration)
		}

		// 2. Verify SearchIndex (should be ephemeral with empty tokens)
		if !si.IsEphemeral {
			t.Error("Expected IsEphemeral for #temp memo")
		}
		if si.BodyTokens != "" || si.TagsTokens != "" || si.FilesTokens != "" {
			t.Errorf("Expected empty tokens for ephemeral memo, got %+v", si)
		}

		foundTemp := false
		for _, tag := range tags {
			if tag == entity.EphemeralTag {
				foundTemp = true
				break
			}
		}
		if !foundTemp {
			t.Error("Tag #temp not found in processed tags")
		}
	})

	t.Run("Custom TTL", func(t *testing.T) {
		content := "Quick note"
		ttl := "12h"

		_, _, expiresAt, _, err := processor.Process(ctx, content, nil, ttl, nil)

		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		if expiresAt == nil {
			t.Fatal("Expected expiration for custom TTL")
		}

		duration := time.Until(*expiresAt)
		if duration < 11*time.Hour || duration > 13*time.Hour {
			t.Errorf("Expected ~12h TTL, got %v", duration)
		}
	})

	t.Run("Merge explicit and extracted tags", func(t *testing.T) {
		content := "Content\n#Extracted"
		explicit := []string{"Explicit", "SRE"}

		tags, _, _, _, err := processor.Process(ctx, content, explicit, "", nil)
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		// Extracted + Explicit + SRE(Ops) = 3 tags
		expected := []string{"Extracted", "Explicit", "Ops"}
		if len(tags) != 3 {
			t.Errorf("Expected 3 tags, got %v", tags)
		}
		for _, e := range expected {
			found := false
			for _, g := range tags {
				if g == e {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Tag %s missing in %v", e, tags)
			}
		}
	})
}
