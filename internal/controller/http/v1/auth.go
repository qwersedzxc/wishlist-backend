package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"golang.org/x/oauth2"

	"github.com/qwersedzxc/wishlist-backend/internal/controller/http/v1/request"
	"github.com/qwersedzxc/wishlist-backend/internal/controller/http/v1/response"
	"github.com/qwersedzxc/wishlist-backend/internal/dto"
	"github.com/qwersedzxc/wishlist-backend/internal/helpers"
	"github.com/qwersedzxc/wishlist-backend/internal/oauth"
)

// AuthHandler обрабатывает HTTP-запросы авторизации.
type AuthHandler struct {
	provider     oauth.Provider
	providerName string
	stateStore   *oauth.StateStore
	authUC       AuthUseCase
	log          *slog.Logger
	frontendURL  string
}

func newAuthHandler(provider oauth.Provider, providerName string, authUC AuthUseCase, log *slog.Logger, frontendURL string) *AuthHandler {
	return &AuthHandler{
		provider:     provider,
		providerName: providerName,
		stateStore:   oauth.NewStateStore(),
		authUC:       authUC,
		log:          log,
		frontendURL:  frontendURL,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.log.Info("OAuth login request received", "provider", h.providerName)

	state := h.stateStore.Generate()
	h.log.Info("Generated OAuth state", "state", state)

	opts := []oauth2.AuthCodeOption{oauth2.AccessTypeOnline}

	// Если пользователь явно вышел — принудительно показываем экран выбора аккаунта
	if r.URL.Query().Get("prompt") == "true" {
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "select_account"))
	}

	url := h.provider.Config().AuthCodeURL(state, opts...)
	h.log.Info("Redirecting to OAuth provider", "url", url)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	h.log.Info("OAuth callback received", "query", r.URL.RawQuery)

	state := r.URL.Query().Get("state")
	if !h.stateStore.Validate(state) {
		h.log.Error("Invalid OAuth state", "state", state)
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		h.log.Error("Missing OAuth code")
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	h.log.Info("Exchanging OAuth code for token", "code", code[:10]+"...")

	token, err := h.provider.Config().Exchange(r.Context(), code)
	if err != nil {
		h.log.Error("oauth exchange failed", "error", err)
		http.Error(w, "oauth exchange failed", http.StatusInternalServerError)
		return
	}

	h.log.Info("OAuth token received, getting user info")

	userInfo, err := h.provider.GetUserInfo(r.Context(), token)
	if err != nil {
		h.log.Error("get user info failed", "error", err)
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}

	h.log.Info("User info received", "external_id", userInfo.ExternalID, "email", userInfo.Email, "name", userInfo.Name)

	authResp, err := h.authUC.FindOrCreateByOAuth(r.Context(), h.providerName, userInfo.ExternalID, userInfo.Email, userInfo.Name, userInfo.AvatarURL)
	if err != nil {
		h.log.Error("find or create user failed", "error", err)
		http.Error(w, "failed to authenticate user", http.StatusInternalServerError)
		return
	}

	h.log.Info("User authenticated successfully", "user_id", authResp.User.ID)

	// Редиректим на фронтенд без токена
	h.setTokenCookie(w, authResp.Token)

	h.log.Info("Redirecting to frontend", "url", h.frontendURL)
	http.Redirect(w, r, h.frontendURL, http.StatusFound)
}

func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,                 // Запрещаем доступ из JavaScript
		SameSite: http.SameSiteLaxMode, // Защита от CSRF
		MaxAge:   7 * 24 * 60 * 60,     // 7 дней
	}
	http.SetCookie(w, cookie)
	h.log.Debug("Token cookie set")
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req request.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("invalid request body"))
		return
	}

	authResp, err := h.authUC.Register(r.Context(), dto.UserRegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error(err.Error()))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response.AuthResponseFromDTO(authResp))
}

func (h *AuthHandler) LoginEmail(w http.ResponseWriter, r *http.Request) {
	var req request.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("invalid request body"))
		return
	}

	authResp, err := h.authUC.Login(r.Context(), dto.UserLoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("invalid credentials"))
		return
	}

	render.JSON(w, r, response.AuthResponseFromDTO(authResp))
}

// Me возвращает данные текущего авторизованного пользователя
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("unauthorized"))
		return
	}

	u, err := h.authUC.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, fmt.Errorf("user not found")) {
			render.Status(r, http.StatusNotFound)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.Error(err.Error()))
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"user": u,
	})
}

// UpdateProfile обновляет профиль текущего пользователя
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.Error("unauthorized"))
		return
	}

	var input dto.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Error("invalid request body"))
		return
	}

	u, err := h.authUC.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Error(err.Error()))
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"user": u,
	})
}

// Logout удаляет cookie с токеном и завершает сессию
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Logout request received")
	
	// Удаляем cookie с токеном
	cookie := &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Удаляем cookie
	}
	http.SetCookie(w, cookie)
	
	h.log.Info("Token cookie deleted, user logged out")
	
	render.JSON(w, r, map[string]string{
		"message": "logged out successfully",
	})
}
