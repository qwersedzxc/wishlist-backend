package helpers

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

// WithUserID помещает userID в контекст.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// GetUserIDFromCtx извлекает userID из контекста.
// Возвращает ошибку, если ключ отсутствует.
func GetUserIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return uuid.Nil, errors.New("userID not found in context")
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("userID has invalid type in context")
	}

	return id, nil
}

// GetUserIDFromCtxOptional извлекает userID из контекста.
// Возвращает nil если userID отсутствует (пользователь не авторизован).
func GetUserIDFromCtxOptional(ctx context.Context) *uuid.UUID {
	v := ctx.Value(userIDKey)
	if v == nil {
		return nil
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
