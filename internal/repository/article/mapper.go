package article

import (
	"time"

	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/google/uuid"
)

// articleRow модель строки таблицы articles из базы данных.
type articleRow struct {
	ID          uuid.UUID  `db:"id"`
	Title       string     `db:"title"`
	Body        string     `db:"body"`
	AuthorID    uuid.UUID  `db:"author_id"`
	PublishedAt *time.Time `db:"published_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// toEntity преобразует строку БД в доменную сущность.
func toEntity(row articleRow) entity.Article {
	return entity.Article{
		ID:          row.ID,
		Title:       row.Title,
		Body:        row.Body,
		AuthorID:    row.AuthorID,
		PublishedAt: row.PublishedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// toEntitySlice преобразует срез строк БД в срез доменных сущностей.
func toEntitySlice(rows []articleRow) []entity.Article {
	articles := make([]entity.Article, 0, len(rows))
	for _, r := range rows {
		articles = append(articles, toEntity(r))
	}

	return articles
}
