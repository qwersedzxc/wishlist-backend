package response

import (
	"time"

	"github.com/google/uuid"

	"github.com/qwersedzxc/wishlist-backend/internal/entity"
)

// WishlistAuthor данные автора вишлиста
type WishlistAuthor struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  *string   `json:"fullName,omitempty"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Bio       *string   `json:"bio,omitempty"`
	City      *string   `json:"city,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	BirthDate *string   `json:"birthDate,omitempty"`
}

// WishlistResponse ответ с данными вишлиста
type WishlistResponse struct {
	ID           uuid.UUID       `json:"id"`
	UserID       uuid.UUID       `json:"userId"`
	Title        string          `json:"title"`
	Description  *string         `json:"description"`
	EventName    *string         `json:"eventName"`
	EventDate    *time.Time      `json:"eventDate"`
	ImageURL     *string         `json:"imageUrl"`
	IsPublic     bool            `json:"isPublic"`
	PrivacyLevel string          `json:"privacyLevel"`
	ShareToken   *string         `json:"shareToken"`
	Author       *WishlistAuthor `json:"author,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

func NewWishlistResponse(w entity.Wishlist) WishlistResponse {
	resp := WishlistResponse{
		ID:           w.ID,
		UserID:       w.UserID,
		Title:        w.Title,
		Description:  w.Description,
		EventName:    w.EventName,
		EventDate:    w.EventDate,
		ImageURL:     w.ImageURL,
		IsPublic:     w.IsPublic,
		PrivacyLevel: w.PrivacyLevel,
		ShareToken:   w.ShareToken,
		CreatedAt:    w.CreatedAt,
		UpdatedAt:    w.UpdatedAt,
	}
	// TODO: Author information should be fetched separately or via DTO
	return resp
}

// WishlistListResponse ответ со списком вишлистов
type WishlistListResponse struct {
	Items []WishlistResponse `json:"items"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
}

func NewWishlistListResponse(wishlists []entity.Wishlist, total, page int) WishlistListResponse {
	items := make([]WishlistResponse, len(wishlists))
	for i, w := range wishlists {
		items[i] = NewWishlistResponse(w)
	}

	return WishlistListResponse{
		Items: items,
		Total: total,
		Page:  page,
	}
}

// WishlistItemResponse ответ с данными элемента вишлиста
type WishlistItemResponse struct {
	ID                     uuid.UUID  `json:"id"`
	WishlistID             uuid.UUID  `json:"wishlistId"`
	Title                  string     `json:"title"`
	Description            *string    `json:"description"`
	URL                    *string    `json:"url"`
	ImageURL               *string    `json:"imageUrl"`
	Price                  *float64   `json:"price"`
	Priority               int        `json:"priority"`
	Category               *string    `json:"category"`
	IsPurchased            bool       `json:"isPurchased"`
	IsReserved             bool       `json:"isReserved"`
	ReservedByMe           bool       `json:"reservedByMe"`
	IsIncognitoReservation bool       `json:"isIncognitoReservation"`
	ReservedBy             *UserInfo  `json:"reservedBy,omitempty"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}

// UserInfo краткая информация о пользователе
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  *string   `json:"fullName,omitempty"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
}

func NewWishlistItemResponse(item entity.WishlistItem, currentUserID *uuid.UUID, isOwner bool) WishlistItemResponse {
	isReserved := item.ReservedBy != nil
	reservedByMe := false
	showIncognito := false

	if currentUserID != nil && item.ReservedBy != nil {
		reservedByMe = *item.ReservedBy == *currentUserID
	}

	// Владелец не видит инкогнито бронирование
	if isOwner && item.IsIncognitoReservation {
		isReserved = false
	} else if isReserved {
		showIncognito = item.IsIncognitoReservation
	}

	resp := WishlistItemResponse{
		ID:                     item.ID,
		WishlistID:             item.WishlistID,
		Title:                  item.Title,
		Description:            item.Description,
		URL:                    item.URL,
		ImageURL:               item.ImageURL,
		Price:                  item.Price,
		Priority:               item.Priority,
		Category:               item.Category,
		IsPurchased:            item.IsPurchased,
		IsReserved:             isReserved,
		ReservedByMe:           reservedByMe,
		IsIncognitoReservation: showIncognito,
		CreatedAt:              item.CreatedAt,
		UpdatedAt:              item.UpdatedAt,
	}

	// Добавляем информацию о пользователе который забронировал
	// TODO: ReservedBy information should be fetched separately or via DTO
	// Currently entity doesn't contain JOIN fields

	return resp
}

// WishlistItemListResponse ответ со списком элементов вишлиста
type WishlistItemListResponse struct {
	Items []WishlistItemResponse `json:"items"`
	Total int                    `json:"total"`
	Page  int                    `json:"page"`
}

func NewWishlistItemListResponse(items []entity.WishlistItem, total, page int, currentUserID *uuid.UUID, isOwner bool) WishlistItemListResponse {
	responseItems := make([]WishlistItemResponse, len(items))
	for i, item := range items {
		responseItems[i] = NewWishlistItemResponse(item, currentUserID, isOwner)
	}

	return WishlistItemListResponse{
		Items: responseItems,
		Total: total,
		Page:  page,
	}
}
