package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"

	"golang.org/x/crypto/bcrypt"
)

type IdentityService struct {
	userRepo port.UserRepository
	nowFunc  func() time.Time
}

func NewIdentityService(
	userRepo port.UserRepository,
) *IdentityService {
	return &IdentityService{
		userRepo: userRepo,
		nowFunc:  time.Now,
	}
}

func (s *IdentityService) GetDefaultAdmin(ctx context.Context, username string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *IdentityService) EnsureBootstrapAdmin(
	ctx context.Context,
	username string,
	password string,
) error {
	if username == "" || password == "" {
		return nil
	}

	_, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return nil
	}

	if !errors.Is(err, apperr.ErrNotFound) {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap admin password: %w", err)
	}

	admin := &entity.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		Role:         entity.UserRoleAdmin,
		Status:       entity.UserStatusActive,
	}
	return s.userRepo.Create(ctx, admin)
}
