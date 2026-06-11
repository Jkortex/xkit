package api_presenter

import (
	"errors"
	"testing"

	"daily/internal/application/apperr"
)

func TestErrorCodeFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "invalid input", err: apperr.ErrInvalidInput, want: "INVALID_INPUT"},
		{name: "unauthorized", err: apperr.ErrUnauthorized, want: "UNAUTHORIZED"},
		{name: "forbidden", err: apperr.ErrForbidden, want: "FORBIDDEN"},
		{name: "not found", err: apperr.ErrNotFound, want: "NOT_FOUND"},
		{name: "conflict", err: apperr.ErrConflict, want: "CONFLICT"},
		{name: "unknown", err: errors.New("boom"), want: "INTERNAL_ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorCodeFromError(tt.err)
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}
