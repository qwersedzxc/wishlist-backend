package wishlist

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/qwersedzxc/wishlist-backend/internal/dto"
)

const (
	wishlistsTable     = "wishlists"
	wishlistItemsTable = "wishlist_items"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// buildInsertWishlist создаёт INSERT-запрос для вишлиста
func buildInsertWishlist(input dto.CreateWishlistInput) sq.InsertBuilder {
	return psql.
		Insert(wishlistsTable).
		Columns("user_id", "title", "description", "event_name", "event_date", "image_url", "is_public", "privacy_level", "share_token").
		Values(input.UserID, input.Title, input.Description, input.EventName, input.EventDate, input.ImageURL, input.IsPublic, input.PrivacyLevel, nil).
		Suffix("RETURNING id, user_id, title, description, event_name, event_date, image_url, is_public, privacy_level, share_token, created_at, updated_at")
}

// buildSelectWishlistByID создаёт SELECT-запрос для получения вишлиста по ID
func buildSelectWishlistByID(id uuid.UUID) sq.SelectBuilder {
	return psql.
		Select(
			"w.id", "w.user_id", "w.title", "w.description", "w.event_name", "w.event_date",
			"w.image_url", "w.is_public", "w.privacy_level", "w.share_token", "w.created_at", "w.updated_at",
			"u.username AS author_username", "u.full_name AS author_full_name", "u.avatar_url AS author_avatar_url",
			"u.bio AS author_bio", "u.city AS author_city", "u.phone AS author_phone",
			"u.birth_date AS author_birth_date",
		).
		From(wishlistsTable + " w").
		LeftJoin("users u ON u.id = w.user_id").
		Where(sq.Eq{"w.id": id})
}

// buildSelectWishlists создаёт SELECT-запрос для списка вишлистов
func buildSelectWishlists(filter dto.WishlistFilter) sq.SelectBuilder {
	q := psql.
		Select(
			"w.id", "w.user_id", "w.title", "w.description", "w.event_name", "w.event_date",
			"w.image_url", "w.is_public", "w.privacy_level", "w.share_token", "w.created_at", "w.updated_at",
			"u.username AS author_username", "u.full_name AS author_full_name", "u.avatar_url AS author_avatar_url",
			"u.bio AS author_bio", "u.city AS author_city", "u.phone AS author_phone",
			"u.birth_date AS author_birth_date",
		).
		From(wishlistsTable + " w").
		LeftJoin("users u ON u.id = w.user_id")

	if filter.UserID != nil {
		q = q.Where(sq.Eq{"w.user_id": *filter.UserID})
	}
	if filter.IsPublic != nil {
		q = q.Where(sq.Eq{"w.is_public": *filter.IsPublic})
	}

	q = q.OrderBy("w.created_at DESC")

	offset := (filter.Page - 1) * filter.PerPage
	q = q.Limit(uint64(filter.PerPage)).Offset(uint64(offset))

	return q
}

// buildCountWishlists создаёт COUNT-запрос для вишлистов
func buildCountWishlists(filter dto.WishlistFilter) sq.SelectBuilder {
	q := psql.Select("COUNT(*)").From(wishlistsTable + " w")

	if filter.UserID != nil {
		q = q.Where(sq.Eq{"w.user_id": *filter.UserID})
	}
	if filter.IsPublic != nil {
		q = q.Where(sq.Eq{"w.is_public": *filter.IsPublic})
	}

	return q
}

// buildUpdateWishlist создаёт UPDATE-запрос для вишлиста
func buildUpdateWishlist(id uuid.UUID, input dto.UpdateWishlistInput) sq.UpdateBuilder {
	q := psql.Update(wishlistsTable).Where(sq.Eq{"id": id})

	if input.Title != nil {
		q = q.Set("title", *input.Title)
	}
	if input.Description != nil {
		q = q.Set("description", *input.Description)
	}
	if input.EventName != nil {
		q = q.Set("event_name", *input.EventName)
	}
	if input.EventDate != nil {
		q = q.Set("event_date", *input.EventDate)
	}
	if input.ImageURL != nil {
		q = q.Set("image_url", *input.ImageURL)
	}
	if input.IsPublic != nil {
		q = q.Set("is_public", *input.IsPublic)
	}
	if input.PrivacyLevel != nil {
		q = q.Set("privacy_level", *input.PrivacyLevel)
	}

	q = q.Set("updated_at", sq.Expr("NOW()"))
	q = q.Suffix("RETURNING id, user_id, title, description, event_name, event_date, image_url, is_public, privacy_level, share_token, created_at, updated_at")

	return q
}

// buildDeleteWishlist создаёт DELETE-запрос для вишлиста
func buildDeleteWishlist(id uuid.UUID) sq.DeleteBuilder {
	return psql.Delete(wishlistsTable).Where(sq.Eq{"id": id})
}

// buildInsertWishlistItem создаёт INSERT-запрос для элемента вишлиста
func buildInsertWishlistItem(input dto.CreateWishlistItemInput) sq.InsertBuilder {
	return psql.
		Insert(wishlistItemsTable).
		Columns("wishlist_id", "title", "description", "url", "image_url", "price", "priority", "category").
		Values(input.WishlistID, input.Title, input.Description, input.URL, input.ImageURL, input.Price, input.Priority, input.Category).
		Suffix("RETURNING id, wishlist_id, title, description, url, image_url, price, priority, category, is_purchased, created_at, updated_at")
}

// buildSelectWishlistItemByID создаёт SELECT-запрос для элемента по ID
func buildSelectWishlistItemByID(id uuid.UUID) sq.SelectBuilder {
	return psql.
		Select(
			"wi.id", "wi.wishlist_id", "wi.title", "wi.description", "wi.url", "wi.image_url",
			"wi.price", "wi.priority", "wi.category", "wi.is_purchased",
			"wi.reserved_by", "wi.reserved_at", "wi.is_incognito_reservation",
			"wi.created_at", "wi.updated_at",
			"u.username AS reserved_by_username", "u.full_name AS reserved_by_full_name", "u.avatar_url AS reserved_by_avatar_url",
		).
		From(wishlistItemsTable + " wi").
		LeftJoin("users u ON u.id = wi.reserved_by").
		Where(sq.Eq{"wi.id": id})
}

// buildSelectWishlistItems создаёт SELECT-запрос для списка элементов
func buildSelectWishlistItems(filter dto.WishlistItemFilter) sq.SelectBuilder {
	q := psql.
		Select(
			"wi.id", "wi.wishlist_id", "wi.title", "wi.description", "wi.url", "wi.image_url",
			"wi.price", "wi.priority", "wi.category", "wi.is_purchased",
			"wi.reserved_by", "wi.reserved_at", "wi.is_incognito_reservation",
			"wi.created_at", "wi.updated_at",
			"u.username AS reserved_by_username", "u.full_name AS reserved_by_full_name", "u.avatar_url AS reserved_by_avatar_url",
		).
		From(wishlistItemsTable + " wi").
		LeftJoin("users u ON u.id = wi.reserved_by").
		Where(sq.Eq{"wi.wishlist_id": filter.WishlistID})

	if filter.IsPurchased != nil {
		q = q.Where(sq.Eq{"wi.is_purchased": *filter.IsPurchased})
	}

	q = q.OrderBy("wi.priority DESC", "wi.created_at DESC")

	offset := (filter.Page - 1) * filter.PerPage
	q = q.Limit(uint64(filter.PerPage)).Offset(uint64(offset))

	return q
}

// buildCountWishlistItems создаёт COUNT-запрос для элементов
func buildCountWishlistItems(filter dto.WishlistItemFilter) sq.SelectBuilder {
	q := psql.Select("COUNT(*)").From(wishlistItemsTable).Where(sq.Eq{"wishlist_id": filter.WishlistID})

	if filter.IsPurchased != nil {
		q = q.Where(sq.Eq{"is_purchased": *filter.IsPurchased})
	}

	return q
}

// buildUpdateWishlistItem создаёт UPDATE-запрос для элемента
func buildUpdateWishlistItem(id uuid.UUID, input dto.UpdateWishlistItemInput) sq.UpdateBuilder {
	q := psql.Update(wishlistItemsTable).Where(sq.Eq{"id": id})

	if input.Title != nil {
		q = q.Set("title", *input.Title)
	}
	if input.Description != nil {
		q = q.Set("description", *input.Description)
	}
	if input.URL != nil {
		q = q.Set("url", *input.URL)
	}
	if input.ImageURL != nil {
		q = q.Set("image_url", *input.ImageURL)
	}
	if input.Price != nil {
		q = q.Set("price", *input.Price)
	}
	if input.Priority != nil {
		q = q.Set("priority", *input.Priority)
	}
	if input.Category != nil {
		q = q.Set("category", *input.Category)
	}
	if input.IsPurchased != nil {
		q = q.Set("is_purchased", *input.IsPurchased)
	}

	q = q.Set("updated_at", sq.Expr("NOW()"))
	q = q.Suffix("RETURNING id, wishlist_id, title, description, url, image_url, price, priority, category, is_purchased, created_at, updated_at")

	return q
}

// buildDeleteWishlistItem создаёт DELETE-запрос для элемента
func buildDeleteWishlistItem(id uuid.UUID) sq.DeleteBuilder {
	return psql.Delete(wishlistItemsTable).Where(sq.Eq{"id": id})
}
