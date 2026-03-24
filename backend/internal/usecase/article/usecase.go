package article

import (
	"context"
	"log/slog"

	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/google/uuid"
)

// UseCase реализует бизнес-логику для статей.
type UseCase struct {
	repo Repository
	log  *slog.Logger
}

// New создаёт новый экземпляр UseCase.
func New(repo Repository, log *slog.Logger) *UseCase {
	return &UseCase{repo: repo, log: log}
}

// GetByID возвращает статью по идентификатору.
func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (entity.Article, error) {
	uc.log.Info("usecase.GetByID", "id", id)

	return uc.repo.GetByID(ctx, id)
}

// List возвращает список статей с пагинацией.
func (uc *UseCase) List(ctx context.Context, filter dto.ArticleFilter) ([]entity.Article, int, error) {
	uc.log.Info("usecase.List", "page", filter.Page, "per_page", filter.PerPage)

	return uc.repo.List(ctx, filter)
}

// Create создаёт новую статью.
func (uc *UseCase) Create(ctx context.Context, input dto.CreateArticleInput) (entity.Article, error) {
	uc.log.Info("usecase.Create", "title", input.Title)

	return uc.repo.Create(ctx, input)
}

// Update обновляет существующую статью.
func (uc *UseCase) Update(ctx context.Context, input dto.UpdateArticleInput) (entity.Article, error) {
	uc.log.Info("usecase.Update", "id", input.ID)

	return uc.repo.Update(ctx, input)
}

// Delete удаляет статью по идентификатору.
func (uc *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	uc.log.Info("usecase.Delete", "id", id)

	return uc.repo.Delete(ctx, id)
}
