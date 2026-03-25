package article

import (
	"context"
	"fmt"

	"github.com/KaoriEl/golang-boilerplate/internal/database"
	"github.com/KaoriEl/golang-boilerplate/internal/definitions"
	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
)

// Repository реализует ArticleRepository поверх PostgreSQL.
type Repository struct {
	db *database.Database
}

// New создаёт новый Repository.
func New(db *database.Database) *Repository {
	return &Repository{db: db}
}

// GetByID возвращает статью по ID или definitions.ErrNotFound.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (entity.Article, error) {
	query, args, err := buildSelectByID(r.db.Sq, id).ToSql()
	if err != nil {
		return entity.Article{}, fmt.Errorf("repository.GetByID build query: %w", err)
	}

	var row articleRow
	if err := pgxscan.Get(ctx, r.db.Pool, &row, query, args...); err != nil {
		if pgxscan.NotFound(err) {
			return entity.Article{}, definitions.ErrNotFound
		}

		return entity.Article{}, fmt.Errorf("repository.GetByID: %w", err)
	}

	return toEntity(row), nil
}

// List возвращает постраничный список статей и общее количество.
func (r *Repository) List(ctx context.Context, filter dto.ArticleFilter) ([]entity.Article, int, error) {
	// Получаем общее количество
	countQuery, countArgs, err := buildCount(r.db.Sq, filter).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("repository.List build count query: %w", err)
	}

	var total int
	if err := r.db.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository.List count: %w", err)
	}

	// Получаем страницу
	query, args, err := buildSelectList(r.db.Sq, filter).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("repository.List build query: %w", err)
	}

	var rows []articleRow
	if err := pgxscan.Select(ctx, r.db.Pool, &rows, query, args...); err != nil {
		return nil, 0, fmt.Errorf("repository.List: %w", err)
	}

	return toEntitySlice(rows), total, nil
}

// Create сохраняет новую статью и возвращает сохранённую сущность.
func (r *Repository) Create(ctx context.Context, input dto.CreateArticleInput) (entity.Article, error) {
	row := articleRow{
		Title:    input.Title,
		Body:     input.Body,
		AuthorID: input.AuthorID,
	}

	query, args, err := buildInsert(r.db.Sq, row).ToSql()
	if err != nil {
		return entity.Article{}, fmt.Errorf("repository.Create build query: %w", err)
	}

	var result articleRow
	if err := pgxscan.Get(ctx, r.db.Pool, &result, query, args...); err != nil {
		return entity.Article{}, fmt.Errorf("repository.Create: %w", err)
	}

	return toEntity(result), nil
}

// Update обновляет поля статьи и возвращает обновлённую сущность.
func (r *Repository) Update(ctx context.Context, input dto.UpdateArticleInput) (entity.Article, error) {
	query, args, err := buildUpdate(r.db.Sq, input).ToSql()
	if err != nil {
		return entity.Article{}, fmt.Errorf("repository.Update build query: %w", err)
	}

	var row articleRow
	if err := pgxscan.Get(ctx, r.db.Pool, &row, query, args...); err != nil {
		if pgxscan.NotFound(err) {
			return entity.Article{}, definitions.ErrNotFound
		}

		return entity.Article{}, fmt.Errorf("repository.Update: %w", err)
	}

	return toEntity(row), nil
}

// Delete удаляет статью по ID.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := buildDelete(r.db.Sq, id).ToSql()
	if err != nil {
		return fmt.Errorf("repository.Delete build query: %w", err)
	}

	ct, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return definitions.ErrNotFound
	}

	return nil
}
