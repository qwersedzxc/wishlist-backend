package wishlist

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/KaoriEl/golang-boilerplate/internal/definitions"
	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
)

type UseCase struct {
	repo     Repository
	itemRepo ItemRepository
	log      *slog.Logger
}

func New(repo Repository, itemRepo ItemRepository, log *slog.Logger) *UseCase {
	return &UseCase{
		repo:     repo,
		itemRepo: itemRepo,
		log:      log,
	}
}

// CreateWishlist создаёт новый вишлист
func (uc *UseCase) CreateWishlist(ctx context.Context, input dto.CreateWishlistInput) (entity.Wishlist, error) {
	wishlist, err := uc.repo.Create(ctx, input)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to create wishlist", "error", err)
		return entity.Wishlist{}, err
	}

	uc.log.InfoContext(ctx, "wishlist created", "id", wishlist.ID, "imageURL", wishlist.ImageURL, "eventName", wishlist.EventName, "privacyLevel", wishlist.PrivacyLevel)
	return wishlist, nil
}

// GetWishlist возвращает вишлист по ID
func (uc *UseCase) GetWishlist(ctx context.Context, id uuid.UUID) (entity.Wishlist, error) {
	wishlist, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to get wishlist", "id", id, "error", err)
		return entity.Wishlist{}, err
	}

	return wishlist, nil
}

// ListWishlists возвращает список вишлистов с пагинацией
func (uc *UseCase) ListWishlists(ctx context.Context, filter dto.WishlistFilter) ([]entity.Wishlist, int, error) {
	if filter.Page < 1 {
		filter.Page = definitions.DefaultPage
	}
	if filter.PerPage < 1 || filter.PerPage > definitions.MaxPerPage {
		filter.PerPage = definitions.DefaultPerPage
	}

	wishlists, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to list wishlists", "error", err)
		return nil, 0, err
	}

	return wishlists, total, nil
}

// UpdateWishlist обновляет вишлист
func (uc *UseCase) UpdateWishlist(ctx context.Context, id uuid.UUID, input dto.UpdateWishlistInput) (entity.Wishlist, error) {
	wishlist, err := uc.repo.Update(ctx, id, input)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to update wishlist", "id", id, "error", err)
		return entity.Wishlist{}, err
	}

	uc.log.InfoContext(ctx, "wishlist updated", "id", id)
	return wishlist, nil
}

// DeleteWishlist удаляет вишлист
func (uc *UseCase) DeleteWishlist(ctx context.Context, id uuid.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.log.ErrorContext(ctx, "failed to delete wishlist", "id", id, "error", err)
		return err
	}

	uc.log.InfoContext(ctx, "wishlist deleted", "id", id)
	return nil
}

// CreateItem создаёт элемент вишлиста
func (uc *UseCase) CreateItem(ctx context.Context, input dto.CreateWishlistItemInput) (entity.WishlistItem, error) {
	// Проверяем существование вишлиста
	if _, err := uc.repo.GetByID(ctx, input.WishlistID); err != nil {
		return entity.WishlistItem{}, err
	}

	item, err := uc.itemRepo.Create(ctx, input)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to create wishlist item", "error", err)
		return entity.WishlistItem{}, err
	}

	uc.log.InfoContext(ctx, "wishlist item created", "id", item.ID)
	return item, nil
}

// GetItem возвращает элемент вишлиста по ID
func (uc *UseCase) GetItem(ctx context.Context, id uuid.UUID) (entity.WishlistItem, error) {
	item, err := uc.itemRepo.GetByID(ctx, id)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to get wishlist item", "id", id, "error", err)
		return entity.WishlistItem{}, err
	}

	return item, nil
}

// ListItems возвращает список элементов вишлиста
func (uc *UseCase) ListItems(ctx context.Context, filter dto.WishlistItemFilter) ([]entity.WishlistItem, int, error) {
	if filter.Page < 1 {
		filter.Page = definitions.DefaultPage
	}
	if filter.PerPage < 1 || filter.PerPage > definitions.MaxPerPage {
		filter.PerPage = definitions.DefaultPerPage
	}

	items, total, err := uc.itemRepo.List(ctx, filter)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to list wishlist items", "error", err)
		return nil, 0, err
	}

	return items, total, nil
}

// UpdateItem обновляет элемент вишлиста
func (uc *UseCase) UpdateItem(ctx context.Context, id uuid.UUID, input dto.UpdateWishlistItemInput) (entity.WishlistItem, error) {
	item, err := uc.itemRepo.Update(ctx, id, input)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to update wishlist item", "id", id, "error", err)
		return entity.WishlistItem{}, err
	}

	uc.log.InfoContext(ctx, "wishlist item updated", "id", id)
	return item, nil
}

// DeleteItem удаляет элемент вишлиста
func (uc *UseCase) DeleteItem(ctx context.Context, id uuid.UUID) error {
	if err := uc.itemRepo.Delete(ctx, id); err != nil {
		uc.log.ErrorContext(ctx, "failed to delete wishlist item", "id", id, "error", err)
		return err
	}

	uc.log.InfoContext(ctx, "wishlist item deleted", "id", id)
	return nil
}

// ReserveItem бронирует элемент вишлиста
func (uc *UseCase) ReserveItem(ctx context.Context, itemID, userID uuid.UUID, isIncognito bool) error {
	// Проверяем что элемент существует
	item, err := uc.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to get item", "itemID", itemID, "error", err)
		return err
	}

	// Проверяем что пользователь не владелец вишлиста
	wishlist, err := uc.repo.GetByID(ctx, item.WishlistID)
	if err != nil {
		uc.log.ErrorContext(ctx, "failed to get wishlist", "wishlistID", item.WishlistID, "error", err)
		return err
	}

	if wishlist.UserID == userID {
		return definitions.ErrForbidden
	}

	// Бронируем
	if err := uc.itemRepo.Reserve(ctx, itemID, userID, isIncognito); err != nil {
		uc.log.ErrorContext(ctx, "failed to reserve item", "itemID", itemID, "userID", userID, "error", err)
		return err
	}

	uc.log.InfoContext(ctx, "item reserved", "itemID", itemID, "userID", userID, "isIncognito", isIncognito)
	return nil
}

// UnreserveItem снимает бронирование с элемента
func (uc *UseCase) UnreserveItem(ctx context.Context, itemID, userID uuid.UUID) error {
	if err := uc.itemRepo.Unreserve(ctx, itemID, userID); err != nil {
		uc.log.ErrorContext(ctx, "failed to unreserve item", "itemID", itemID, "userID", userID, "error", err)
		return err
	}

	uc.log.InfoContext(ctx, "item unreserved", "itemID", itemID, "userID", userID)
	return nil
}
