package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	
	"github.com/qwersedzxc/wishlist-backend/internal/helpers"
)

// AuthUseCase интерфейс для валидации токенов
type AuthUseCase interface {
	ValidateToken(tokenString string) (uuid.UUID, error)
}

// Auth middleware для проверки JWT токенов
func Auth(authUC AuthUseCase, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			
			if authHeader == "" {
				log.Warn("missing authorization header", "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Проверяем формат "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn("invalid authorization header format", "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Валидируем токен
			userID, err := authUC.ValidateToken(token)
			if err != nil {
				log.Warn("invalid token", "error", err, "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			log.Info("Auth successful", "userID", userID, "path", r.URL.Path)

			// Добавляем userID в контекст
			ctx := helpers.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware для опциональной авторизации
// Если токен присутствует и валиден - добавляет userID в контекст
// Если токена нет или он невалиден - продолжает без авторизации
func OptionalAuth(authUC AuthUseCase, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Нет токена - продолжаем без авторизации
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Неправильный формат - продолжаем без авторизации
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]
			userID, err := authUC.ValidateToken(token)
			if err != nil {
				// Невалидный токен - продолжаем без авторизации
				next.ServeHTTP(w, r)
				return
			}

			// Токен валиден - добавляем userID в контекст
			ctx := helpers.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
