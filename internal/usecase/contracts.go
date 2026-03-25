package usecase

import (
	"context"

	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/google/uuid"
)

// WishlistUseCase описывает бизнес-операции над вишлистами
type WishlistUseCase interface {
	CreateWishlist(ctx context.Context, input dto.CreateWishlistInput) (entity.Wishlist, error)
	GetWishlist(ctx context.Context, id uuid.UUID) (entity.Wishlist, error)
	ListWishlists(ctx context.Context, filter dto.WishlistFilter) ([]entity.Wishlist, int, error)
	UpdateWishlist(ctx context.Context, id uuid.UUID, input dto.UpdateWishlistInput) (entity.Wishlist, error)
	DeleteWishlist(ctx context.Context, id uuid.UUID) error

	CreateItem(ctx context.Context, input dto.CreateWishlistItemInput) (entity.WishlistItem, error)
	GetItem(ctx context.Context, id uuid.UUID) (entity.WishlistItem, error)
	ListItems(ctx context.Context, filter dto.WishlistItemFilter) ([]entity.WishlistItem, int, error)
	UpdateItem(ctx context.Context, id uuid.UUID, input dto.UpdateWishlistItemInput) (entity.WishlistItem, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	ReserveItem(ctx context.Context, itemID, userID uuid.UUID, isIncognito bool) error
	UnreserveItem(ctx context.Context, itemID, userID uuid.UUID) error
}
