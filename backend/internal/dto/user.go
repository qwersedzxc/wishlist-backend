package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserRegisterInput - данные для регистрации пользователя
type UserRegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserLoginInput - данные для входа пользователя
type UserLoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserOutput - данные пользователя для ответа
type UserOutput struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	AvatarURL *string    `json:"avatarUrl,omitempty"`
	FullName  *string    `json:"fullName,omitempty"`
	BirthDate *time.Time `json:"birthDate,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	City      *string    `json:"city,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// UpdateProfileInput - данные для обновления профиля
type UpdateProfileInput struct {
	FullName  *string `json:"fullName"`
	BirthDate *string `json:"birthDate"`
	Bio       *string `json:"bio"`
	Phone     *string `json:"phone"`
	City      *string `json:"city"`
}

// AuthResponse - ответ при успешной аутентификации
type AuthResponse struct {
	User  UserOutput `json:"user"`
	Token string     `json:"token"`
}

// FriendshipInput - данные для создания дружбы
type FriendshipInput struct {
	FriendID uuid.UUID `json:"friend_id" validate:"required"`
}

// FriendshipOutput - данные о дружбе для ответа
type FriendshipOutput struct {
	ID        uuid.UUID  `json:"id"`
	User      UserOutput `json:"user"`
	Friend    UserOutput `json:"friend"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}
