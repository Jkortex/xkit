package auth

import (
	"daily/internal/application/dto"
	"daily/internal/domain/entity"
)

func toUserResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		Status:   string(user.Status),
	}
}
