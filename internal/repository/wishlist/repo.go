package wishlist

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qwersedzxc/wishlist-backend/internal/definitions"
	"github.com/qwersedzxc/wishlist-backend/internal/dto"
	"github.com/qwersedzxc/wishlist-backend/internal/entity"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create создаёт новый вишлист
func (r *Repository) Create(ctx context.Context, input dto.CreateWishlistInput) (entity.Wishlist, error) {
	sql, args, err := buildInsertWishlist(input).ToSql()
	if err != nil {
		return entity.Wishlist{}, err
	}

	var dbWish dbWishlist
	if err := pgxscan.Get(ctx, r.db, &dbWish, sql, args...); err != nil {
		// Логируем ошибку SQL
		return entity.Wishlist{}, fmt.Errorf("failed to create wishlist: %w (SQL: %s, Args: %+v)", err, sql, args)
	}

	return dbWish.toEntity(), nil
}

// GetByID возвращает вишлист по ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (entity.Wishlist, error) {
	sql, args, err := buildSelectWishlistByID(id).ToSql()
	if err != nil {
		return entity.Wishlist{}, err
	}

	var dbWish dbWishlist
	if err := pgxscan.Get(ctx, r.db, &dbWish, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Wishlist{}, definitions.ErrNotFound
		}
		return entity.Wishlist{}, err
	}

	return dbWish.toEntity(), nil
}

// List возвращает список вишлистов с пагинацией
func (r *Repository) List(ctx context.Context, filter dto.WishlistFilter) ([]entity.Wishlist, int, error) {
	// Используем window function COUNT(*) OVER() для получения total в одном запросе
	baseQuery := buildSelectWishlists(filter)
	
	// Добавляем COUNT(*) OVER() для получения total count
	sql, args, err := baseQuery.
		Column("COUNT(*) OVER() as total_count").
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var wishlists []entity.Wishlist
	var total int

	for rows.Next() {
		var dbWish dbWishlist
		var totalCount int
		
		err := rows.Scan(
			&dbWish.ID, &dbWish.UserID, &dbWish.Title, &dbWish.Description,
			&dbWish.EventName, &dbWish.EventDate, &dbWish.ImageURL,
			&dbWish.IsPublic, &dbWish.PrivacyLevel, &dbWish.ShareToken,
			&dbWish.CreatedAt, &dbWish.UpdatedAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}
		
		wishlists = append(wishlists, dbWish.toEntity())
		total = totalCount
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return wishlists, total, nil
}

// Update обновляет вишлист
func (r *Repository) Update(ctx context.Context, id uuid.UUID, input dto.UpdateWishlistInput) (entity.Wishlist, error) {
	sql, args, err := buildUpdateWishlist(id, input).ToSql()
	if err != nil {
		return entity.Wishlist{}, err
	}

	var dbWish dbWishlist
	if err := pgxscan.Get(ctx, r.db, &dbWish, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Wishlist{}, definitions.ErrNotFound
		}
		return entity.Wishlist{}, err
	}

	return dbWish.toEntity(), nil
}

// Delete удаляет вишлист
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := buildDeleteWishlist(id).ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return definitions.ErrNotFound
	}

	return nil
}

// ItemRepository репозиторий для элементов вишлиста
type ItemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

// Create создаёт элемент вишлиста
func (r *ItemRepository) Create(ctx context.Context, input dto.CreateWishlistItemInput) (entity.WishlistItem, error) {
	sql, args, err := buildInsertWishlistItem(input).ToSql()
	if err != nil {
		return entity.WishlistItem{}, err
	}

	var dbItem dbWishlistItem
	if err := pgxscan.Get(ctx, r.db, &dbItem, sql, args...); err != nil {
		return entity.WishlistItem{}, err
	}

	return dbItem.toEntity(), nil
}

// GetByID возвращает элемент по ID
func (r *ItemRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.WishlistItem, error) {
	sql, args, err := buildSelectWishlistItemByID(id).ToSql()
	if err != nil {
		return entity.WishlistItem{}, err
	}

	var dbItem dbWishlistItem
	if err := pgxscan.Get(ctx, r.db, &dbItem, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.WishlistItem{}, definitions.ErrNotFound
		}
		return entity.WishlistItem{}, err
	}

	return dbItem.toEntity(), nil
}

// List возвращает список элементов вишлиста
func (r *ItemRepository) List(ctx context.Context, filter dto.WishlistItemFilter) ([]entity.WishlistItem, int, error) {
	// Используем window function COUNT(*) OVER() для получения total в одном запросе
	baseQuery := buildSelectWishlistItems(filter)
	
	// Добавляем COUNT(*) OVER() для получения total count
	sql, args, err := baseQuery.
		Column("COUNT(*) OVER() as total_count").
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []entity.WishlistItem
	var total int

	for rows.Next() {
		var dbItem dbWishlistItem
		var totalCount int
		
		err := rows.Scan(
			&dbItem.ID, &dbItem.WishlistID, &dbItem.Title, &dbItem.Description,
			&dbItem.URL, &dbItem.ImageURL, &dbItem.Price, &dbItem.Priority,
			&dbItem.Category, &dbItem.IsPurchased, &dbItem.ReservedBy, &dbItem.ReservedAt,
			&dbItem.IsIncognitoReservation, &dbItem.CreatedAt, &dbItem.UpdatedAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}
		
		items = append(items, dbItem.toEntity())
		total = totalCount
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Update обновляет элемент вишлиста
func (r *ItemRepository) Update(ctx context.Context, id uuid.UUID, input dto.UpdateWishlistItemInput) (entity.WishlistItem, error) {
	sql, args, err := buildUpdateWishlistItem(id, input).ToSql()
	if err != nil {
		return entity.WishlistItem{}, err
	}

	var dbItem dbWishlistItem
	if err := pgxscan.Get(ctx, r.db, &dbItem, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.WishlistItem{}, definitions.ErrNotFound
		}
		return entity.WishlistItem{}, err
	}

	return dbItem.toEntity(), nil
}

// Delete удаляет элемент вишлиста
func (r *ItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := buildDeleteWishlistItem(id).ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return definitions.ErrNotFound
	}

	return nil
}

// Reserve бронирует элемент вишлиста
func (r *ItemRepository) Reserve(ctx context.Context, itemID, userID uuid.UUID, isIncognito bool) error {
	query := `
		UPDATE wishlist_items
		SET reserved_by = $1, reserved_at = NOW(), is_incognito_reservation = $2, updated_at = NOW()
		WHERE id = $3 AND reserved_by IS NULL
		RETURNING id
	`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, userID, isIncognito, itemID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("item already reserved or not found")
		}
		return err
	}
	return nil
}

// Unreserve снимает бронирование с элемента
func (r *ItemRepository) Unreserve(ctx context.Context, itemID, userID uuid.UUID) error {
	query := `
		UPDATE wishlist_items
		SET reserved_by = NULL, reserved_at = NULL, is_incognito_reservation = FALSE, updated_at = NOW()
		WHERE id = $1 AND reserved_by = $2
		RETURNING id
	`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, itemID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("item not reserved by this user or not found")
		}
		return err
	}
	return nil
}
