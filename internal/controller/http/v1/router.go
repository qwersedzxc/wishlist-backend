package v1

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KaoriEl/golang-boilerplate/internal/config"
	"github.com/KaoriEl/golang-boilerplate/internal/controller/http/middleware"
	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/KaoriEl/golang-boilerplate/internal/oauth"
	"github.com/KaoriEl/golang-boilerplate/internal/types"
	"github.com/KaoriEl/golang-boilerplate/internal/usecase"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter собирает chi-роутер с маршрутами API v1 и Swagger UI.
func NewRouter(wishlistUC usecase.WishlistUseCase, authUC AuthUseCase, friendshipUC FriendshipUseCase, roleRepo RoleRepository, provider oauth.Provider, providerName string, s3cfg config.S3Cfg, emailService EmailService, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(middleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)

	// CORS для работы с frontend
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	wishlistHandler := newWishlistHandler(wishlistUC, log)
	authHandler := newAuthHandler(provider, providerName, authUC, log)
	uploadHandler := newUploadHandler(log, S3Config{
		Endpoint:        s3cfg.Endpoint,
		AccessKeyID:     s3cfg.AccessKeyID,
		SecretAccessKey: s3cfg.SecretAccessKey,
		Bucket:          s3cfg.Bucket,
		BaseURL:         s3cfg.BaseURL,
		Region:          s3cfg.Region,
	})
	friendshipHandler := newFriendshipHandler(friendshipUC, emailService, log)
	roleHandler := NewRoleController(roleRepo)
	roleMiddleware := middleware.NewRoleMiddleware(roleRepo)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.JSONContentType)
		r.Route("/wishlists", func(r chi.Router) {
			r.Get("/", wishlistHandler.ListWishlists)
			r.With(middleware.Auth(authUC, log)).Post("/", wishlistHandler.CreateWishlist)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", wishlistHandler.GetWishlist)
				r.With(middleware.Auth(authUC, log)).Patch("/", wishlistHandler.UpdateWishlist)
				r.With(middleware.Auth(authUC, log)).Delete("/", wishlistHandler.DeleteWishlist)
			})

			r.Route("/{wishlist_id}/items", func(r chi.Router) {
				r.With(middleware.OptionalAuth(authUC, log)).Get("/", wishlistHandler.ListItems)
				r.With(middleware.Auth(authUC, log)).Post("/", wishlistHandler.CreateItem)

				r.Route("/{id}", func(r chi.Router) {
					r.With(middleware.OptionalAuth(authUC, log)).Get("/", wishlistHandler.GetItem)
					r.With(middleware.Auth(authUC, log)).Patch("/", wishlistHandler.UpdateItem)
					r.With(middleware.Auth(authUC, log)).Delete("/", wishlistHandler.DeleteItem)
					r.With(middleware.Auth(authUC, log)).Post("/reserve", wishlistHandler.ReserveItem)
					r.With(middleware.Auth(authUC, log)).Delete("/reserve", wishlistHandler.UnreserveItem)
				})
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.LoginEmail)
			r.With(middleware.Auth(authUC, log)).Get("/me", authHandler.Me)
			r.With(middleware.Auth(authUC, log)).Patch("/profile", authHandler.UpdateProfile)
			r.Get("/oauth/login", authHandler.Login)
			r.Get("/oauth/callback", authHandler.Callback)
		})

		r.Route("/upload", func(r chi.Router) {
			r.Use(middleware.Auth(authUC, log))
			r.Post("/image", uploadHandler.UploadImage)
		})

		// Поиск пользователей (публичный)
		r.Get("/users/search", friendshipHandler.SearchUsers)
		r.With(middleware.Auth(authUC, log)).Get("/users/{id}", friendshipHandler.GetUserProfile)

		// Дружба (требует авторизации)
		r.Route("/friends", func(r chi.Router) {
			r.Use(middleware.Auth(authUC, log))
			r.Get("/", friendshipHandler.GetFriends)
			r.Post("/request", friendshipHandler.SendRequest)
			r.Get("/requests", friendshipHandler.GetPendingRequests)
			r.Post("/requests/{id}/accept", friendshipHandler.AcceptRequest)
			r.Post("/requests/{id}/reject", friendshipHandler.RejectRequest)
			r.Delete("/{friendId}", friendshipHandler.RemoveFriend)
		})

		// Роли и разрешения
		r.Route("/roles", func(r chi.Router) {
			r.Use(middleware.Auth(authUC, log))
			r.Use(roleMiddleware.LoadUserRoles())
			
			// Получение информации о ролях (доступно всем авторизованным)
			r.Get("/my", roleHandler.GetMyRoles)
			r.Get("/user/{userId}", roleHandler.GetUserRoles)
			
			// Управление ролями (только для админов)
			r.With(roleMiddleware.RequireAdmin()).Get("/", roleHandler.GetAllRoles)
			r.With(roleMiddleware.RequireAdmin()).Post("/", roleHandler.CreateRole)
			r.With(roleMiddleware.RequireAdmin()).Get("/{id}", roleHandler.GetRole)
			r.With(roleMiddleware.RequireAdmin()).Post("/assign", roleHandler.AssignRole)
			r.With(roleMiddleware.RequireAdmin()).Post("/remove", roleHandler.RemoveRole)
		})
	})

	return r
}


// AuthUseCase интерфейс для use case аутентификации
type AuthUseCase interface {
	Register(ctx context.Context, input dto.UserRegisterInput) (*dto.AuthResponse, error)
	Login(ctx context.Context, input dto.UserLoginInput) (*dto.AuthResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserOutput, error)
	ValidateToken(tokenString string) (uuid.UUID, error)
	FindOrCreateByOAuth(ctx context.Context, provider, providerID, email, name, avatarURL string) (*dto.AuthResponse, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, input dto.UpdateProfileInput) (*dto.UserOutput, error)
}

// EmailService интерфейс для отправки email уведомлений
type EmailService interface {
	SendFriendRequest(to string, data types.FriendRequestData) error
	SendBirthdayReminder(to string, data types.BirthdayReminderData) error
}

// RoleRepository интерфейс для работы с ролями
type RoleRepository interface {
	GetAllRoles(ctx context.Context) ([]entity.Role, error)
	GetRoleByID(ctx context.Context, id int) (*entity.Role, error)
	GetRoleByName(ctx context.Context, name string) (*entity.Role, error)
	CreateRole(ctx context.Context, role *entity.Role) error
	UpdateRole(ctx context.Context, role *entity.Role) error
	DeleteRole(ctx context.Context, id int) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.Role, error)
	GetUserWithRoles(ctx context.Context, userID uuid.UUID) (*entity.UserWithRoles, error)
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID int, grantedBy *uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleID int) error
	GetUsersWithRole(ctx context.Context, roleName string) ([]entity.UserWithRoles, error)
}
