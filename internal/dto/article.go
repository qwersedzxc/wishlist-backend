package dto

import "github.com/google/uuid"

// CreateArticleInput входные данные для создания статьи.
type CreateArticleInput struct {
	Title    string    `json:"title"     validate:"required,min=3,max=255"`
	Body     string    `json:"body"      validate:"required"`
	AuthorID uuid.UUID `json:"authorId" validate:"required"`
}

// UpdateArticleInput входные данные для обновления статьи.
type UpdateArticleInput struct {
	ID    uuid.UUID
	Title *string `json:"title" validate:"omitempty,min=3,max=255"`
	Body  *string `json:"body"  validate:"omitempty"`
}

// ArticleFilter параметры фильтрации/пагинации списка статей.
type ArticleFilter struct {
	Page    int `form:"page"     validate:"min=1"`
	PerPage int `form:"per_page" validate:"min=1,max=100"`
}
