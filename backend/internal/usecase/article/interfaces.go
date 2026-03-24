package article

import (
	"context"

	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/google/uuid"
)

// Repository описывает контракт слоя хранилища для статей.
type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.Article, error)
	List(ctx context.Context, filter dto.ArticleFilter) ([]entity.Article, int, error)
	Create(ctx context.Context, input dto.CreateArticleInput) (entity.Article, error)
	Update(ctx context.Context, input dto.UpdateArticleInput) (entity.Article, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
