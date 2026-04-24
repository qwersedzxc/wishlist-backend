package role

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qwersedzxc/wishlist-backend/internal/entity"
)

// Repository интерфейс для работы с ролями
type Repository interface {
	// Роли
	GetAllRoles(ctx context.Context) ([]entity.Role, error)
	GetRoleByID(ctx context.Context, id int) (*entity.Role, error)
	GetRoleByName(ctx context.Context, name string) (*entity.Role, error)
	CreateRole(ctx context.Context, role *entity.Role) error
	UpdateRole(ctx context.Context, role *entity.Role) error
	DeleteRole(ctx context.Context, id int) error

	// Пользователи и роли
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.Role, error)
	GetUserWithRoles(ctx context.Context, userID uuid.UUID) (*entity.UserWithRoles, error)
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID int, grantedBy *uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleID int) error
	GetUsersWithRole(ctx context.Context, roleName string) ([]entity.UserWithRoles, error)
}

// repository реализация Repository
type repository struct {
	db  *pgxpool.Pool
	qb  squirrel.StatementBuilderType
	log *slog.Logger
}

// New создает новый экземпляр репозитория ролей
func New(db *pgxpool.Pool, log *slog.Logger) Repository {
	return &repository{
		db:  db,
		qb:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		log: log,
	}
}

// GetAllRoles получает все роли
func (r *repository) GetAllRoles(ctx context.Context) ([]entity.Role, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM roles
		ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []entity.Role
	for rows.Next() {
		var role entity.Role
		var permissionsJSON []byte

		err := rows.Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(permissionsJSON, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetRoleByID получает роль по ID
func (r *repository) GetRoleByID(ctx context.Context, id int) (*entity.Role, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM roles
		WHERE id = $1`

	var role entity.Role
	var permissionsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&permissionsJSON,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.log.Error("GetRoleByID scan failed", "error", err)
		return nil, err
	}
	if err := json.Unmarshal(permissionsJSON, &role.Permissions); err != nil {
		r.log.Error("Failed to unmarshal permissions", "error", err, "raw", string(permissionsJSON))
		return nil, err
	}
	return &role, nil
}

// GetRoleByName получает роль по названию
func (r *repository) GetRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM roles
		WHERE name = $1`

	var role entity.Role
	err := r.db.QueryRow(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.Permissions, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// CreateRole создает новую роль
func (r *repository) CreateRole(ctx context.Context, role *entity.Role) error {
	query := `
		INSERT INTO roles (name, description, permissions)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query, role.Name, role.Description, role.Permissions).
		Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
}

// UpdateRole обновляет роль
func (r *repository) UpdateRole(ctx context.Context, role *entity.Role) error {
	query := `
		UPDATE roles
		SET name = $2, description = $3, permissions = $4, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRow(ctx, query, role.ID, role.Name, role.Description, role.Permissions).
		Scan(&role.UpdatedAt)
}

// DeleteRole удаляет роль
func (r *repository) DeleteRole(ctx context.Context, id int) error {
	query := `DELETE FROM roles WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role with id %d not found", id)
	}

	return nil
}

// GetUserRoles получает роли пользователя
func (r *repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.Role, error) {
	query := `
        SELECT r.id, r.name, r.description, r.permissions, r.created_at, r.updated_at
        FROM roles r
        INNER JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1 AND ur.is_active = true
        AND (ur.expires_at IS NULL OR ur.expires_at > now())
        ORDER BY r.name`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []entity.Role
	for rows.Next() {
		var role entity.Role
		var id int
		var name string
		var description *string
		var permissionsJSON []byte           // ← сканируем как []byte
		var createdAt, updatedAt interface{} // временно

		err := rows.Scan(&id, &name, &description, &permissionsJSON, &createdAt, &updatedAt)
		if err != nil {
			r.log.Error("Scan failed", "error", err, "userID", userID)
			return nil, err
		}

		// Заполняем роль
		role.ID = id
		role.Name = name
		role.Description = description

		// Парсим JSON
		if err := json.Unmarshal(permissionsJSON, &role.Permissions); err != nil {
			r.log.Error("JSON unmarshal failed", "error", err, "raw", string(permissionsJSON))
			return nil, err
		}

		// Конвертируем время
		if t, ok := createdAt.(time.Time); ok {
			role.CreatedAt = t
		}
		if t, ok := updatedAt.(time.Time); ok {
			role.UpdatedAt = t
		}

		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetUserWithRoles получает пользователя с его ролями
func (r *repository) GetUserWithRoles(ctx context.Context, userID uuid.UUID) (*entity.UserWithRoles, error) {
	// Добавьте лог
	r.log.Info("GetUserWithRoles called", "userID", userID)

	userQuery := `
        SELECT id, email, username, password_hash, provider, provider_id, 
               avatar_url, created_at, updated_at
        FROM users
        WHERE id = $1`

	var user entity.User
	err := r.db.QueryRow(ctx, userQuery, userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.Provider, &user.ProviderID, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.log.Error("QueryRow failed", "error", err)
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	r.log.Info("User found", "userID", user.ID, "email", user.Email)

	// Получаем роли пользователя
	roles, err := r.GetUserRoles(ctx, userID)
	if err != nil {
		r.log.Error("GetUserRoles failed", "error", err)
		return nil, err
	}

	r.log.Info("Roles found", "count", len(roles))

	return &entity.UserWithRoles{
		User:  user,
		Roles: roles,
	}, nil
}

// AssignRoleToUser назначает роль пользователю
func (r *repository) AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID int, grantedBy *uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) WHERE is_active = true
		DO UPDATE SET granted_by = $3, granted_at = now()`

	_, err := r.db.Exec(ctx, query, userID, roleID, grantedBy)
	return err
}

// RemoveRoleFromUser удаляет роль у пользователя
func (r *repository) RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleID int) error {
	query := `
		UPDATE user_roles
		SET is_active = false
		WHERE user_id = $1 AND role_id = $2 AND is_active = true`

	result, err := r.db.Exec(ctx, query, userID, roleID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user role not found or already inactive")
	}

	return nil
}

// GetUsersWithRole получает всех пользователей с определенной ролью
func (r *repository) GetUsersWithRole(ctx context.Context, roleName string) ([]entity.UserWithRoles, error) {
	query := `
		SELECT u.id, u.email, u.username, u.password_hash, u.provider, u.provider_id,
		       u.avatar_url, u.created_at, u.updated_at,
		       r.id as role_id, r.name as role_name, r.description as role_description,
		       r.permissions as role_permissions, r.created_at as role_created_at, r.updated_at as role_updated_at
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name = $1 AND ur.is_active = true
		AND (ur.expires_at IS NULL OR ur.expires_at > now())
		ORDER BY u.username`

	rows, err := r.db.Query(ctx, query, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.UserWithRoles
	userMap := make(map[uuid.UUID]*entity.UserWithRoles)

	for rows.Next() {
		var user entity.User
		var role entity.Role

		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.PasswordHash,
			&user.Provider, &user.ProviderID, &user.AvatarURL,
			&user.CreatedAt, &user.UpdatedAt,
			&role.ID, &role.Name, &role.Description, &role.Permissions,
			&role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userWithRoles, exists := userMap[user.ID]; exists {
			userWithRoles.Roles = append(userWithRoles.Roles, role)
		} else {
			userWithRoles := &entity.UserWithRoles{
				User:  user,
				Roles: []entity.Role{role},
			}
			userMap[user.ID] = userWithRoles
			users = append(users, *userWithRoles)
		}
	}

	return users, rows.Err()
}
