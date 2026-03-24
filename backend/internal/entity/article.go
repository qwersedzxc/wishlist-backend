package entity

import (
	"time"

	"github.com/google/uuid"
)

// Article представляет новостную статью в доменном слое.
type Article struct {
	ID          uuid.UUID
	Title       string
	Body        string
	AuthorID    uuid.UUID
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
