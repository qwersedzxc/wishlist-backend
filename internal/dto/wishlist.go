package dto

import "github.com/google/uuid"

// CreateWishlistInput данные для создания вишлиста
type CreateWishlistInput struct {
	UserID       uuid.UUID  `json:"userId"       validate:"required"`
	Title        string     `json:"title"        validate:"required,min=1,max=255"`
	Description  *string    `json:"description"  validate:"omitempty,max=1000"`
	EventName    *string    `json:"eventName"    validate:"omitempty,max=255"`
	EventDate    *string    `json:"eventDate"    validate:"omitempty"`
	ImageURL     *string    `json:"imageUrl"     validate:"omitempty"`
	IsPublic     bool       `json:"isPublic"`
	PrivacyLevel string     `json:"privacyLevel" validate:"omitempty,oneof=public friends_only link_only"`
}

// UpdateWishlistInput данные для обновления вишлиста
type UpdateWishlistInput struct {
	Title        *string `json:"title"        validate:"omitempty,min=1,max=255"`
	Description  *string `json:"description"  validate:"omitempty,max=1000"`
	EventName    *string `json:"eventName"    validate:"omitempty,max=255"`
	EventDate    *string `json:"eventDate"    validate:"omitempty"`
	ImageURL     *string `json:"imageUrl"     validate:"omitempty"`
	IsPublic     *bool   `json:"isPublic"`
	PrivacyLevel *string `json:"privacyLevel" validate:"omitempty,oneof=public friends_only link_only"`
}

// WishlistFilter фильтры для списка вишлистов
type WishlistFilter struct {
	UserID   *uuid.UUID
	IsPublic *bool
	Page     int
	PerPage  int
}

// CreateWishlistItemInput данные для создания элемента вишлиста
type CreateWishlistItemInput struct {
	WishlistID  uuid.UUID `json:"wishlistId"  validate:"required"`
	Title       string    `json:"title"       validate:"required,min=1,max=255"`
	Description *string   `json:"description" validate:"omitempty,max=1000"`
	URL         *string   `json:"url"         validate:"omitempty,url"`
	ImageURL    *string   `json:"imageUrl"    validate:"omitempty,url"`
	Price       *float64  `json:"price"       validate:"omitempty,gte=0"`
	Priority    int       `json:"priority"    validate:"gte=0,lte=10"`
	Category    *string   `json:"category"    validate:"omitempty,max=100"`
}

// UpdateWishlistItemInput данные для обновления элемента вишлиста
type UpdateWishlistItemInput struct {
	Title       *string  `json:"title"       validate:"omitempty,min=1,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=1000"`
	URL         *string  `json:"url"         validate:"omitempty,url"`
	ImageURL    *string  `json:"imageUrl"    validate:"omitempty,url"`
	Price       *float64 `json:"price"       validate:"omitempty,gte=0"`
	Priority    *int     `json:"priority"    validate:"omitempty,gte=0,lte=10"`
	Category    *string  `json:"category"    validate:"omitempty,max=100"`
	IsPurchased *bool    `json:"isPurchased"`
}

// WishlistItemFilter фильтры для элементов вишлиста
type WishlistItemFilter struct {
	WishlistID  uuid.UUID
	IsPurchased *bool
	Page        int
	PerPage     int
}

// WishlistWithAuthor представление вишлиста с данными автора (для JOIN запросов)
type WishlistWithAuthor struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"userId"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	EventName       *string    `json:"eventName"`
	EventDate       *string    `json:"eventDate"`
	ImageURL        *string    `json:"imageUrl"`
	IsPublic        bool       `json:"isPublic"`
	PrivacyLevel    string     `json:"privacyLevel"`
	ShareToken      *string    `json:"shareToken"`
	CreatedAt       string     `json:"createdAt"`
	UpdatedAt       string     `json:"updatedAt"`
	AuthorUsername  *string    `json:"authorUsername,omitempty"`
	AuthorFullName  *string    `json:"authorFullName,omitempty"`
	AuthorAvatarURL *string    `json:"authorAvatarUrl,omitempty"`
	AuthorBio       *string    `json:"authorBio,omitempty"`
	AuthorCity      *string    `json:"authorCity,omitempty"`
	AuthorPhone     *string    `json:"authorPhone,omitempty"`
	AuthorBirthDate *string    `json:"authorBirthDate,omitempty"`
}

// WishlistItemWithReserver представление элемента вишлиста с данными забронировавшего (для JOIN запросов)
type WishlistItemWithReserver struct {
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
	ReservedBy             *uuid.UUID `json:"reservedBy"`
	ReservedAt             *string    `json:"reservedAt"`
	IsIncognitoReservation bool       `json:"isIncognitoReservation"`
	CreatedAt              string     `json:"createdAt"`
	UpdatedAt              string     `json:"updatedAt"`
	ReservedByUsername     *string    `json:"reservedByUsername,omitempty"`
	ReservedByFullName     *string    `json:"reservedByFullName,omitempty"`
	ReservedByAvatarURL    *string    `json:"reservedByAvatarUrl,omitempty"`
}
