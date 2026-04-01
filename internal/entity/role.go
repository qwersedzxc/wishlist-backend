package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Permissions представляет массив разрешений
type Permissions []string

// Value реализует интерфейс driver.Valuer для сохранения в БД
func (p Permissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan реализует интерфейс sql.Scanner для чтения из БД
func (p *Permissions) Scan(value interface{}) error {
	if value == nil {
		*p = Permissions{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into Permissions", value)
	}

	return json.Unmarshal(bytes, p)
}

// Role представляет роль в системе
type Role struct {
	ID          int         `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description *string     `json:"description" db:"description"`
	Permissions Permissions `json:"permissions" db:"permissions"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// UserRole представляет связь пользователя с ролью
type UserRole struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	RoleID    int        `json:"role_id" db:"role_id"`
	GrantedBy *uuid.UUID `json:"granted_by" db:"granted_by"`
	GrantedAt time.Time  `json:"granted_at" db:"granted_at"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	IsActive  bool       `json:"is_active" db:"is_active"`
}

// UserWithRoles представляет пользователя с его ролями
type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}

// HasPermission проверяет есть ли у пользователя определенное разрешение
func (u *UserWithRoles) HasPermission(permission string) bool {
	for _, role := range u.Roles {
		// Админы имеют все разрешения
		for _, perm := range role.Permissions {
			if perm == "*" || perm == permission {
				return true
			}
		}
	}
	return false
}

// HasRole проверяет есть ли у пользователя определенная роль
func (u *UserWithRoles) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// IsAdmin проверяет является ли пользователь администратором
func (u *UserWithRoles) IsAdmin() bool {
	return u.HasRole("admin")
}

// GetRoleNames возвращает список названий ролей пользователя
func (u *UserWithRoles) GetRoleNames() []string {
	names := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		names[i] = role.Name
	}
	return names
}