package entity

import (
	"time"

	"github.com/google/uuid"
)

// Wishlist представляет список желаний пользователя
type Wishlist struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Title        string
	Description  *string
	EventName    *string
	EventDate    *time.Time
	ImageURL     *string
	IsPublic     bool
	PrivacyLevel string // public, friends_only, link_only
	ShareToken   *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// Данные автора (опционально, из JOIN)
	AuthorUsername  *string
	AuthorFullName  *string
	AuthorAvatarURL *string
	AuthorBio       *string
	AuthorCity      *string
	AuthorPhone     *string
	AuthorBirthDate *time.Time
}

// WishlistItem представляет элемент в списке желаний
type WishlistItem struct {
	ID                     uuid.UUID
	WishlistID             uuid.UUID
	Title                  string
	Description            *string
	URL                    *string
	ImageURL               *string
	Price                  *float64
	Priority               int
	Category               *string
	IsPurchased            bool
	ReservedBy             *uuid.UUID
	ReservedAt             *time.Time
	IsIncognitoReservation bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
	// Данные пользователя который забронировал (опционально, из JOIN)
	ReservedByUsername  *string
	ReservedByFullName  *string
	ReservedByAvatarURL *string
}
