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
			var token string

			log.Info("Auth debug",
				"path", r.URL.Path,
				"authHeader", r.Header.Get("Authorization"),
				"cookie", r.Header.Get("Cookie"),
			)

			// 1. Пробуем получить токен из Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
					log.Info("Token from Authorization header")
				}
			}

			// 2. Если нет в header, пробуем из cookie
			if token == "" {
				cookie, err := r.Cookie("token")
				if err == nil {
					token = cookie.Value
					log.Info("Token from cookie", "token_preview", token[:20]+"...")
				} else {
					log.Info("No token cookie found", "error", err)
				}
			}

			if token == "" {
				log.Warn("missing token", "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

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
func OptionalAuth(authUC AuthUseCase, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			// 1. Пробуем из header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}

			// 2. Пробуем из cookie
			if token == "" {
				cookie, err := r.Cookie("token")
				if err == nil {
					token = cookie.Value
				}
			}

			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := authUC.ValidateToken(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := helpers.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
