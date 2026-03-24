package request

// CreateWishlistRequest запрос на создание вишлиста
type CreateWishlistRequest struct {
	Title        string  `json:"title"        validate:"required,min=1,max=255"`
	Description  *string `json:"description"  validate:"omitempty,max=1000"`
	EventName    *string `json:"eventName"    validate:"omitempty,max=255"`
	EventDate    *string `json:"eventDate"    validate:"omitempty"`
	ImageURL     *string `json:"imageUrl"     validate:"omitempty"`
	IsPublic     bool    `json:"isPublic"`
	PrivacyLevel string  `json:"privacyLevel" validate:"omitempty,oneof=public friends_only link_only"`
}

// UpdateWishlistRequest запрос на обновление вишлиста
type UpdateWishlistRequest struct {
	Title        *string `json:"title"        validate:"omitempty,min=1,max=255"`
	Description  *string `json:"description"  validate:"omitempty,max=1000"`
	EventName    *string `json:"eventName"    validate:"omitempty,max=255"`
	EventDate    *string `json:"eventDate"    validate:"omitempty"`
	ImageURL     *string `json:"imageUrl"     validate:"omitempty"`
	IsPublic     *bool   `json:"isPublic"`
	PrivacyLevel *string `json:"privacyLevel" validate:"omitempty,oneof=public friends_only link_only"`
}

// CreateWishlistItemRequest запрос на создание элемента вишлиста
type CreateWishlistItemRequest struct {
	Title       string   `json:"title"       validate:"required,min=1,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=1000"`
	URL         *string  `json:"url"         validate:"omitempty,url"`
	ImageURL    *string  `json:"imageUrl"    validate:"omitempty"`
	Price       *float64 `json:"price"       validate:"omitempty,gte=0"`
	Priority    int      `json:"priority"    validate:"gte=0,lte=10"`
	Category    *string  `json:"category"    validate:"omitempty,max=100"`
}

// UpdateWishlistItemRequest запрос на обновление элемента вишлиста
type UpdateWishlistItemRequest struct {
	Title       *string  `json:"title"       validate:"omitempty,min=1,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=1000"`
	URL         *string  `json:"url"         validate:"omitempty,url"`
	ImageURL    *string  `json:"imageUrl"    validate:"omitempty"`
	Price       *float64 `json:"price"       validate:"omitempty,gte=0"`
	Priority    *int     `json:"priority"    validate:"omitempty,gte=0,lte=10"`
	Category    *string  `json:"category"    validate:"omitempty,max=100"`
	IsPurchased *bool    `json:"isPurchased"`
}
