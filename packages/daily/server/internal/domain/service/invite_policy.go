package service

import "daily/internal/domain/entity"

const (
	maxActiveMemberInvitesPerAdmin int64 = 20
	maxActiveAdminInvitesPerAdmin  int64 = 5
)

func MaxActiveInvitesPerCreator(role entity.UserRole) int64 {
	if role == entity.UserRoleAdmin {
		return maxActiveAdminInvitesPerAdmin
	}
	return maxActiveMemberInvitesPerAdmin
}

func IsInviteQuotaExceeded(role entity.UserRole, activeCount int64) bool {
	return activeCount >= MaxActiveInvitesPerCreator(role)
}
