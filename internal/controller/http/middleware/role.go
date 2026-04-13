package middleware

import (
	"context"
	"net/http"

	"github.com/qwersedzxc/wishlist-backend/internal/entity"
	"github.com/qwersedzxc/wishlist-backend/internal/helpers"
	"github.com/qwersedzxc/wishlist-backend/internal/repository/role"
)

// RoleMiddleware middleware для проверки ролей и разрешений
type RoleMiddleware struct {
	roleRepo role.Repository
}

type contextKey string

const UserWithRolesKey contextKey = "user_with_roles"

// NewRoleMiddleware создает новый middleware для ролей
func NewRoleMiddleware(roleRepo role.Repository) *RoleMiddleware {
	return &RoleMiddleware{
		roleRepo: roleRepo,
	}
}

// RequireRole проверяет что у пользователя есть определенная роль
func (m *RoleMiddleware) RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Используем уже загруженные роли из контекста
			userWithRoles := GetUserWithRoles(r)

			if userWithRoles == nil {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			if !userWithRoles.HasRole(roleName) {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission проверяет что у пользователя есть определенное разрешение
func (m *RoleMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := helpers.GetUserIDFromCtx(r.Context())
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userWithRoles, err := m.roleRepo.GetUserWithRoles(r.Context(), userID)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if userWithRoles == nil || !userWithRoles.HasPermission(permission) {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			// Добавляем пользователя с ролями в контекст
			ctx := context.WithValue(r.Context(), UserWithRolesKey, userWithRoles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin проверяет что пользователь является администратором
func (m *RoleMiddleware) RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userWithRoles := GetUserWithRoles(r)

			if userWithRoles == nil || !userWithRoles.IsAdmin() {
				http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoadUserRoles загружает роли пользователя в контекст (не блокирует доступ)
func (m *RoleMiddleware) LoadUserRoles() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := helpers.GetUserIDFromCtx(r.Context())
			if err == nil {
				userWithRoles, err := m.roleRepo.GetUserWithRoles(r.Context(), userID)
				if err == nil && userWithRoles != nil {
					ctx := context.WithValue(r.Context(), "user_with_roles", userWithRoles)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserWithRoles получает пользователя с ролями из контекста
func GetUserWithRoles(r *http.Request) *entity.UserWithRoles {
	// Проверяем, что значение существует в контексте
	val := r.Context().Value("user_with_roles")
	if val == nil {
		return nil
	}

	// Проверяем, что значение имеет правильный тип
	userWithRoles, ok := val.(*entity.UserWithRoles)
	if !ok {
		return nil
	}

	return userWithRoles

}
