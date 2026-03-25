package wishlist

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/KaoriEl/golang-boilerplate/internal/entity"
)

// dbWishlist структура для сканирования из БД
type dbWishlist struct {
	ID           uuid.UUID   `db:"id"`
	UserID       uuid.UUID   `db:"user_id"`
	Title        string      `db:"title"`
	Description  *string     `db:"description"`
	EventName    *string     `db:"event_name"`
	EventDate    pgtype.Date `db:"event_date"`
	ImageURL     *string     `db:"image_url"`
	IsPublic     bool        `db:"is_public"`
	PrivacyLevel string      `db:"privacy_level"`
	ShareToken   *string     `db:"share_token"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
	// Поля автора (из JOIN с users)
	AuthorUsername  *string `db:"author_username"`
	AuthorFullName  *string `db:"author_full_name"`
	AuthorAvatarURL *string `db:"author_avatar_url"`
	AuthorBio       *string `db:"author_bio"`
	AuthorCity      *string `db:"author_city"`
	AuthorPhone     *string `db:"author_phone"`
	AuthorBirthDate *string `db:"author_birth_date"`
}

func (db dbWishlist) toEntity() entity.Wishlist {
	var eventDate *time.Time
	if db.EventDate.Valid {
		t := db.EventDate.Time
		eventDate = &t
	}

	return entity.Wishlist{
		ID:              db.ID,
		UserID:          db.UserID,
		Title:           db.Title,
		Description:     db.Description,
		EventName:       db.EventName,
		EventDate:       eventDate,
		ImageURL:        db.ImageURL,
		IsPublic:        db.IsPublic,
		PrivacyLevel:    db.PrivacyLevel,
		ShareToken:      db.ShareToken,
		CreatedAt:       db.CreatedAt,
		UpdatedAt:       db.UpdatedAt,
		AuthorUsername:  db.AuthorUsername,
		AuthorFullName:  db.AuthorFullName,
		AuthorAvatarURL: db.AuthorAvatarURL,
		AuthorBio:       db.AuthorBio,
		AuthorCity:      db.AuthorCity,
		AuthorPhone:     db.AuthorPhone,
		AuthorBirthDate: db.AuthorBirthDate,
	}
}

// dbWishlistItem структура для сканирования элементов из БД
type dbWishlistItem struct {
	ID                     uuid.UUID        `db:"id"`
	WishlistID             uuid.UUID        `db:"wishlist_id"`
	Title                  string           `db:"title"`
	Description            *string          `db:"description"`
	URL                    *string          `db:"url"`
	ImageURL               *string          `db:"image_url"`
	Price                  *float64         `db:"price"`
	Priority               int              `db:"priority"`
	Category               *string          `db:"category"`
	IsPurchased            bool             `db:"is_purchased"`
	ReservedBy             *uuid.UUID       `db:"reserved_by"`
	ReservedAt             pgtype.Timestamp `db:"reserved_at"`
	IsIncognitoReservation bool             `db:"is_incognito_reservation"`
	CreatedAt              time.Time        `db:"created_at"`
	UpdatedAt              time.Time        `db:"updated_at"`
	// Поля пользователя который забронировал (из JOIN с users)
	ReservedByUsername  *string `db:"reserved_by_username"`
	ReservedByFullName  *string `db:"reserved_by_full_name"`
	ReservedByAvatarURL *string `db:"reserved_by_avatar_url"`
}

func (db dbWishlistItem) toEntity() entity.WishlistItem {
	var reservedAt *time.Time
	if db.ReservedAt.Valid {
		t := db.ReservedAt.Time
		reservedAt = &t
	}

	return entity.WishlistItem{
		ID:                     db.ID,
		WishlistID:             db.WishlistID,
		Title:                  db.Title,
		Description:            db.Description,
		URL:                    db.URL,
		ImageURL:               db.ImageURL,
		Price:                  db.Price,
		Priority:               db.Priority,
		Category:               db.Category,
		IsPurchased:            db.IsPurchased,
		ReservedBy:             db.ReservedBy,
		ReservedAt:             reservedAt,
		IsIncognitoReservation: db.IsIncognitoReservation,
		CreatedAt:              db.CreatedAt,
		UpdatedAt:              db.UpdatedAt,
		ReservedByUsername:     db.ReservedByUsername,
		ReservedByFullName:     db.ReservedByFullName,
		ReservedByAvatarURL:    db.ReservedByAvatarURL,
	}
}
