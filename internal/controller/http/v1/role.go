package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/qwersedzxc/wishlist-backend/internal/controller/http/middleware"
	"github.com/qwersedzxc/wishlist-backend/internal/entity"
	"github.com/qwersedzxc/wishlist-backend/internal/repository/role"
)

// RoleController контроллер для управления ролями
type RoleController struct {
	roleRepo role.Repository
	log      *slog.Logger
}

// NewRoleController создает новый контроллер ролей
func NewRoleController(roleRepo role.Repository, log *slog.Logger) *RoleController {
	return &RoleController{
		roleRepo: roleRepo,
		log:      log,
	}
}

// CreateRoleRequest запрос на создание роли
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=50"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
}

// AssignRoleRequest запрос на назначение роли
type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID int       `json:"role_id" validate:"required"`
}

// RoleResponse ответ с информацией о роли
type RoleResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// UserRoleResponse ответ с информацией о пользователе и его ролях
type UserRoleResponse struct {
	ID       uuid.UUID      `json:"id"`
	Email    string         `json:"email"`
	Username string         `json:"username"`
	Roles    []RoleResponse `json:"roles"`
}

// GetAllRoles получает все роли (только для админов)
func (c *RoleController) GetAllRoles(w http.ResponseWriter, r *http.Request) {
	c.log.Info("🔥 GetAllRoles: START", "path", r.URL.Path)

	roles, err := c.roleRepo.GetAllRoles(r.Context())
	if err != nil {
		c.log.Error("🔥 GetAllRoles: GetAlLRoles error", "error", err)
		http.Error(w, "Failed to get roles", http.StatusInternalServerError)
		return
	}

	c.log.Info("🔥 GetAllRoles: got roles from DB", "count", len(roles))

	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Permissions: role.Permissions,
			CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.log.Info("🔥 GetAllRoles: success, sending response")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateRole создает новую роль (только для админов)
func (c *RoleController) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := &entity.Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: entity.Permissions(req.Permissions),
	}

	if err := c.roleRepo.CreateRole(r.Context(), role); err != nil {
		http.Error(w, "Failed to create role", http.StatusInternalServerError)
		return
	}

	response := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: role.Permissions,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetRole получает роль по ID
func (c *RoleController) GetRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	role, err := c.roleRepo.GetRoleByID(r.Context(), roleID)
	if err != nil {
		http.Error(w, "Failed to get role", http.StatusInternalServerError)
		return
	}

	if role == nil {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}

	response := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: role.Permissions,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AssignRole назначает роль пользователю (только для админов)
func (c *RoleController) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем ID администратора из контекста
	userWithRoles := middleware.GetUserWithRoles(r)
	var grantedBy *uuid.UUID
	if userWithRoles != nil {
		grantedBy = &userWithRoles.ID
	}

	if err := c.roleRepo.AssignRoleToUser(r.Context(), req.UserID, req.RoleID, grantedBy); err != nil {
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Role assigned successfully"})
}

// RemoveRole удаляет роль у пользователя (только для админов)
func (c *RoleController) RemoveRole(w http.ResponseWriter, r *http.Request) {
	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.roleRepo.RemoveRoleFromUser(r.Context(), req.UserID, req.RoleID); err != nil {
		http.Error(w, "Failed to remove role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Role removed successfully"})
}

// GetUserRoles получает роли пользователя
func (c *RoleController) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	userWithRoles, err := c.roleRepo.GetUserWithRoles(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user roles", http.StatusInternalServerError)
		return
	}

	if userWithRoles == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	roles := make([]RoleResponse, len(userWithRoles.Roles))
	for i, role := range userWithRoles.Roles {
		roles[i] = RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Permissions: role.Permissions,
			CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := UserRoleResponse{
		ID:       userWithRoles.ID,
		Email:    userWithRoles.Email,
		Username: userWithRoles.Username,
		Roles:    roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMyRoles получает роли текущего пользователя
func (c *RoleController) GetMyRoles(w http.ResponseWriter, r *http.Request) {
	userWithRoles := middleware.GetUserWithRoles(r)
	if userWithRoles == nil {
		// Ошибка на стороне клиента: его данные не загружены
		http.Error(w, "User roles not found. Make sure you are authenticated and the LoadUserRoles middleware is applied.", http.StatusForbidden)
		return
	}

	roles := make([]RoleResponse, len(userWithRoles.Roles))
	for i, role := range userWithRoles.Roles {
		roles[i] = RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Permissions: role.Permissions,
			CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := UserRoleResponse{
		ID:       userWithRoles.ID,
		Email:    userWithRoles.Email,
		Username: userWithRoles.Username,
		Roles:    roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
