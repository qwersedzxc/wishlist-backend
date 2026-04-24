package v1

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/qwersedzxc/wishlist-backend/internal/config"
	"github.com/qwersedzxc/wishlist-backend/internal/controller/http/middleware"
	"github.com/qwersedzxc/wishlist-backend/internal/dto"
	"github.com/qwersedzxc/wishlist-backend/internal/entity"
	"github.com/qwersedzxc/wishlist-backend/internal/oauth"
	"github.com/qwersedzxc/wishlist-backend/internal/types"
	"github.com/qwersedzxc/wishlist-backend/internal/usecase"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter собирает chi-роутер с маршрутами API v1 и Swagger UI.
func NewRouter(
	wishlistUC usecase.WishlistUseCase,
	authUC AuthUseCase,
	friendshipUC FriendshipUseCase,
	roleRepo RoleRepository,
	provider oauth.Provider,
	providerName string,
	s3cfg config.S3Cfg,
	emailService EmailService,
	log *slog.Logger,
	cfg *config.Config,
) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("🔍 REQUEST RECEIVED",
				"method", r.Method,
				"path", r.URL.Path,
				"full_url", r.URL.String())
			next.ServeHTTP(w, r)
		})
	})

	// Глобальные middleware
	r.Use(middleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)

	// CORS для работы с frontend
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

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

	frontendURL := cfg.FrontendURL

	wishlistHandler := newWishlistHandler(wishlistUC, log)
	authHandler := newAuthHandler(provider, providerName, authUC, log, frontendURL)
	uploadHandler := newUploadHandler(log, S3Config{
		Endpoint:        s3cfg.Endpoint,
		AccessKeyID:     s3cfg.AccessKeyID,
		SecretAccessKey: s3cfg.SecretAccessKey,
		Bucket:          s3cfg.Bucket,
		BaseURL:         s3cfg.BaseURL,
		Region:          s3cfg.Region,
	})
	friendshipHandler := newFriendshipHandler(friendshipUC, emailService, log)
	roleHandler := NewRoleController(roleRepo, log)
	roleMiddleware := middleware.NewRoleMiddleware(roleRepo)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.JSONContentType)
		r.Route("/", func(r chi.Router) {
			r.Get("/", wishlistHandler.ListWishlists)
			r.With(middleware.Auth(authUC, log)).Post("/", wishlistHandler.CreateWishlist)

			r.Route("/", func(r chi.Router) {
				// Wishlist endpoints
				r.Get("/wishlists", wishlistHandler.ListWishlists)
				r.With(middleware.Auth(authUC, log)).Post("/wishlists", wishlistHandler.CreateWishlist)

				r.Route("/wishlists/{wishlist_id}", func(r chi.Router) {
					r.Get("/", wishlistHandler.GetWishlist)
					r.With(middleware.Auth(authUC, log)).Patch("/", wishlistHandler.UpdateWishlist)
					r.With(middleware.Auth(authUC, log)).Delete("/", wishlistHandler.DeleteWishlist)

					// Item endpoints
					r.Route("/items", func(r chi.Router) {
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
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.LoginEmail)
			r.Post("/logout", authHandler.Logout)
			r.With(middleware.Auth(authUC, log)).Get("/me", authHandler.Me)
			r.With(middleware.Auth(authUC, log)).Patch("/profile", authHandler.UpdateProfile)
			r.Get("/oauth/login", authHandler.Login)
			r.Get("/oauth/callback", authHandler.Callback)
		})

		r.Route("/upload", func(r chi.Router) {
			r.Use(middleware.Auth(authUC, log))
			r.Post("/image", uploadHandler.UploadImage)
		})

		// Проксирование изображений (публичный доступ для обхода CORS)
		r.Get("/proxy/image", uploadHandler.ProxyImage)

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
