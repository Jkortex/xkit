package tui

import (
	"testing"
)

func TestSpeakWithCacheEmptyText(t *testing.T) {
	SpeakWithCache("", "", nil, "T", nil)
}

func TestShowCacheStatusInvalidUUID(t *testing.T) {
	result := ShowCacheStatus("nonexistent-uuid", "T", nil)
	if result == "" {
		t.Fatal("expected non-empty status string")
	}
}
