package controller

import (
	"context"
	"daily/internal/application/dto"
	authuc "daily/internal/application/usecase/auth"
	"strings"
)

type AuthController struct {
	identitySvc *authuc.IdentityService
}

func NewAuthController(
	identitySvc *authuc.IdentityService,
) *AuthController {
	return &AuthController{
		identitySvc: identitySvc,
	}
}

func (ctrl *AuthController) GetDefaultAdmin(ctx context.Context, username string) (*dto.UserResponse, error) {
	return ctrl.identitySvc.GetDefaultAdmin(ctx, username)
}

func (ctrl *AuthController) EnsureBootstrapAdmin(
	ctx context.Context,
	username string,
	password string,
) error {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	return ctrl.identitySvc.EnsureBootstrapAdmin(ctx, username, password)
}
