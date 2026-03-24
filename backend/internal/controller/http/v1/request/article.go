package request

import "github.com/google/uuid"

// CreateArticleRequest тело запроса на создание статьи.
type CreateArticleRequest struct {
	Title string `json:"title" validate:"required,min=3,max=255"`
	Body  string `json:"body"  validate:"required"`
}

// UpdateArticleRequest тело запроса на обновление статьи.
type UpdateArticleRequest struct {
	Title *string `json:"title" validate:"omitempty,min=3,max=255"`
	Body  *string `json:"body"  validate:"omitempty"`
}

// ArticleListRequest параметры запроса списка статей (query-string).
type ArticleListRequest struct {
	Page    int `form:"page"     validate:"min=1"`
	PerPage int `form:"per_page" validate:"min=1,max=100"`
}

// IDFromPath вспомогательный тип для UUID из URL-параметра.
type IDFromPath struct {
	ID uuid.UUID
}
