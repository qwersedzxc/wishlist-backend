package entity

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы
type User struct {
	ID                uuid.UUID
	Email             string
	Username          string
	PasswordHash      *string
	Provider          *string
	ProviderID        *string
	AvatarURL         *string
	FullName          *string
	BirthDate         *time.Time
	Bio               *string
	Phone             *string
	City              *string
	NotificationEmail *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Friendship представляет дружескую связь между пользователями
type Friendship struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FriendID  uuid.UUID
	Status    string // pending, accepted, rejected
	CreatedAt time.Time
	UpdatedAt time.Time
}
