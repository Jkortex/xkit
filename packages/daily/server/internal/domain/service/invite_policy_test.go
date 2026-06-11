package service

import (
	"testing"

	"daily/internal/domain/entity"
)

func TestMaxActiveInvitesPerCreator(t *testing.T) {
	if got := MaxActiveInvitesPerCreator(entity.UserRoleAdmin); got != 5 {
		t.Fatalf("expected admin limit 5, got %d", got)
	}
	if got := MaxActiveInvitesPerCreator(entity.UserRoleMember); got != 20 {
		t.Fatalf("expected member limit 20, got %d", got)
	}
}

func TestIsInviteQuotaExceeded(t *testing.T) {
	tests := []struct {
		name        string
		role        entity.UserRole
		activeCount int64
		want        bool
	}{
		{name: "admin below limit", role: entity.UserRoleAdmin, activeCount: 4, want: false},
		{name: "admin at limit", role: entity.UserRoleAdmin, activeCount: 5, want: true},
		{name: "member below limit", role: entity.UserRoleMember, activeCount: 19, want: false},
		{name: "member at limit", role: entity.UserRoleMember, activeCount: 20, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInviteQuotaExceeded(tt.role, tt.activeCount)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
