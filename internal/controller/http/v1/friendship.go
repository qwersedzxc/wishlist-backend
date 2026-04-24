package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log/slog"

	"github.com/qwersedzxc/wishlist-backend/internal/controller/http/v1/response"
	"github.com/qwersedzxc/wishlist-backend/internal/entity"
	"github.com/qwersedzxc/wishlist-backend/internal/helpers"
	"github.com/qwersedzxc/wishlist-backend/internal/types"
)

type FriendshipUseCase interface {
	SendRequest(ctx context.Context, userID, friendID uuid.UUID) (*entity.Friendship, error)
	AcceptRequest(ctx context.Context, friendshipID uuid.UUID) error
	RejectRequest(ctx context.Context, friendshipID uuid.UUID) error
	RemoveFriend(ctx context.Context, userID, friendID uuid.UUID) error
	GetFriends(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error)
	GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error)
	SearchUsers(ctx context.Context, query string) ([]*entity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

type FriendshipHandler struct {
	uc           FriendshipUseCase
	emailService EmailService
	log          *slog.Logger
}

func newFriendshipHandler(uc FriendshipUseCase, emailService EmailService, log *slog.Logger) *FriendshipHandler {
	return &FriendshipHandler{
		uc:           uc,
		emailService: emailService,
		log:          log,
	}
}

// SearchUsers ищет пользователей по username/email
func (h *FriendshipHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		render.JSON(w, r, map[string]interface{}{"users": []interface{}{}})
		return
	}

	users, err := h.uc.SearchUsers(r.Context(), q)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		result = append(result, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
		})
	}
	render.JSON(w, r, map[string]interface{}{"users": result})
}

// SendRequest отправляет запрос на дружбу
func (h *FriendshipHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	var body struct {
		FriendID uuid.UUID `json:"friendId"`
	}
	if err := render.DecodeJSON(r.Body, &body); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	f, err := h.uc.SendRequest(r.Context(), userID, body.FriendID)
	if err != nil {
		h.log.Error("failed to send friend request", 
			"error", err, 
			"userID", userID, 
			"friendID", body.FriendID)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Отправляем email уведомление
	go func() {
		// Получаем данные отправителя и получателя
		sender, err := h.uc.GetUserByID(context.Background(), userID)
		if err != nil {
			h.log.Error("failed to get sender user for email", "error", err, "userID", userID)
			return
		}

		recipient, err := h.uc.GetUserByID(context.Background(), body.FriendID)
		if err != nil {
			h.log.Error("failed to get recipient user for email", "error", err, "friendID", body.FriendID)
			return
		}

		// Отправляем уведомление
		emailData := types.FriendRequestData{
			RecipientName: recipient.Username,
			SenderName:    sender.Username,
			SenderEmail:   sender.Email,
			AppURL:        "http://localhost:3000", // TODO: получать из конфига
		}

		if err := h.emailService.SendFriendRequest(recipient.Email, emailData); err != nil {
			h.log.Error("failed to send friend request email", "error", err, "to", recipient.Email)
		} else {
			h.log.Info("friend request email sent", "to", recipient.Email, "from", sender.Username)
		}
	}()

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, f)
}

// AcceptRequest принимает запрос на дружбу
func (h *FriendshipHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid id")))
		return
	}

	if err := h.uc.AcceptRequest(r.Context(), id); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "accepted"})
}

// RejectRequest отклоняет запрос на дружбу
func (h *FriendshipHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid id")))
		return
	}

	if err := h.uc.RejectRequest(r.Context(), id); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "rejected"})
}

// RemoveFriend удаляет друга
func (h *FriendshipHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserIDFromCtx(r.Context())
if err != nil {
render.Status(r, http.StatusUnauthorized)
render.JSON(w, r, response.NewErrorResponse(err))
return
}

	friendIDStr := chi.URLParam(r, "friendId")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid friendId")))
		return
	}

	if err := h.uc.RemoveFriend(r.Context(), userID, friendID); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetFriends возвращает список друзей
func (h *FriendshipHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserIDFromCtx(r.Context())
if err != nil {
render.Status(r, http.StatusUnauthorized)
render.JSON(w, r, response.NewErrorResponse(err))
return
}

	friendships, err := h.uc.GetFriends(r.Context(), userID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Обогащаем данными пользователей
	result := make([]map[string]interface{}, 0, len(friendships))
	for _, f := range friendships {
		friendUserID := f.FriendID
		if f.UserID != userID {
			friendUserID = f.UserID
		}
		u, err := h.uc.GetUserByID(r.Context(), friendUserID)
		if err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"friendshipId": f.ID,
			"id":           u.ID,
			"username":     u.Username,
			"email":        u.Email,
			"birthDate":    u.BirthDate,
			"addedAt":      f.CreatedAt,
		})
	}

	render.JSON(w, r, map[string]interface{}{"friends": result})
}

// GetPendingRequests возвращает входящие запросы на дружбу
func (h *FriendshipHandler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserIDFromCtx(r.Context())
if err != nil {
render.Status(r, http.StatusUnauthorized)
render.JSON(w, r, response.NewErrorResponse(err))
return
}

	requests, err := h.uc.GetPendingRequests(r.Context(), userID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	result := make([]map[string]interface{}, 0, len(requests))
	for _, f := range requests {
		u, err := h.uc.GetUserByID(r.Context(), f.UserID)
		if err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"friendshipId": f.ID,
			"id":           u.ID,
			"username":     u.Username,
			"email":        u.Email,
		})
	}

	render.JSON(w, r, map[string]interface{}{"requests": result})
}

// GetUserProfile возвращает публичный профиль пользователя по ID
func (h *FriendshipHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid id")))
		return
	}

	u, err := h.uc.GetUserByID(r.Context(), id)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"id":        u.ID,
		"username":  u.Username,
		"fullName":  u.FullName,
		"avatarUrl": u.AvatarURL,
		"bio":       u.Bio,
		"city":      u.City,
		"phone":     u.Phone,
		"birthDate": u.BirthDate,
	})
}


