package response

import (
	"github.com/qwersedzxc/wishlist-backend/internal/dto"
	"github.com/google/uuid"
)

// AuthResponse ответ с данными пользователя после авторизации.
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse - данные пользователя в ответе
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
}

// AuthResponseFromDTO преобразует DTO в response
func AuthResponseFromDTO(authDTO *dto.AuthResponse) AuthResponse {
	return AuthResponse{
		User: UserResponse{
			ID:        authDTO.User.ID,
			Email:     authDTO.User.Email,
			Username:  authDTO.User.Username,
			AvatarURL: authDTO.User.AvatarURL,
		},
		Token: authDTO.Token,
	}
}

