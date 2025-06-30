package repository

import (
	"context"

	"github.com/iots1/mingkwan-api/internal/user/domain"
)

type AuthRepository interface {
	RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error)
	Login(ctx context.Context, email string, password string) (*domain.User, error)
	RefreshToken(ctx context.Context, userID string, refreshToken string) (*domain.User, error)
	GetProfile(ctx context.Context, email string) (*domain.User, error)
}
