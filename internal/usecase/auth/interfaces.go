package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/qwersedzxc/wishlist-backend/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByProviderID(ctx context.Context, provider, providerID string) (*entity.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, fullName, bio, phone, city *string, birthDate *string) (*entity.User, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]*entity.User, error)
}
