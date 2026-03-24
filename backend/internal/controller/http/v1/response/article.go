package response

import (
	"time"

	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/google/uuid"
)

// ArticleResponse JSON-представление статьи в ответе API.
type ArticleResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FromEntity конвертирует доменную сущность в DTO ответа.
func FromEntity(a entity.Article) ArticleResponse {
	return ArticleResponse{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// ArticleListResponse JSON-представление постраничного списка статей.
type ArticleListResponse struct {
	Items []ArticleResponse `json:"items"`
	Total int               `json:"total"`
}
