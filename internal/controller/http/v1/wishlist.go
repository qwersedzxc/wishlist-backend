package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/qwersedzxc/wishlist-backend/internal/controller/http/v1/request"
	"github.com/qwersedzxc/wishlist-backend/internal/controller/http/v1/response"
	"github.com/qwersedzxc/wishlist-backend/internal/definitions"
	"github.com/qwersedzxc/wishlist-backend/internal/dto"
	"github.com/qwersedzxc/wishlist-backend/internal/helpers"
	"github.com/qwersedzxc/wishlist-backend/internal/usecase"
)

type WishlistHandler struct {
	uc  usecase.WishlistUseCase
	log *slog.Logger
}

func newWishlistHandler(uc usecase.WishlistUseCase, log *slog.Logger) *WishlistHandler {
	return &WishlistHandler{
		uc:  uc,
		log: log,
	}
}

// CreateWishlist создаёт новый вишлист
// @Summary     Создать вишлист
// @Tags        wishlists
// @Accept      json
// @Produce     json
// @Param       body body request.CreateWishlistRequest true "Данные вишлиста"
// @Success     201 {object} response.WishlistResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /wishlists [post]
func (h *WishlistHandler) CreateWishlist(w http.ResponseWriter, r *http.Request) {
	var req request.CreateWishlistRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	if err := helpers.Validate(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем userID из контекста авторизации
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(fmt.Errorf("unauthorized")))
		return
	}

	input := dto.CreateWishlistInput{
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		EventName:    req.EventName,
		EventDate:    req.EventDate,
		ImageURL:     req.ImageURL,
		IsPublic:     req.IsPublic,
		PrivacyLevel: req.PrivacyLevel,
	}

	if input.PrivacyLevel == "" {
		if input.IsPublic {
			input.PrivacyLevel = "public"
		} else {
			input.PrivacyLevel = "friends_only"
		}
	}

	h.log.Info("creating wishlist", "input", fmt.Sprintf("%+v", input))

	wishlist, err := h.uc.CreateWishlist(r.Context(), input)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response.NewWishlistResponse(wishlist))
}

// GetWishlist возвращает вишлист по ID
// @Summary     Получить вишлист
// @Tags        wishlists
// @Produce     json
// @Param       id path string true "ID вишлиста"
// @Success     200 {object} response.WishlistResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /wishlists/{id} [get]
func (h *WishlistHandler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid wishlist ID")))
		return
	}

	wishlist, err := h.uc.GetWishlist(r.Context(), id)
	if err != nil {
		if errors.Is(err, definitions.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.JSON(w, r, response.NewWishlistResponse(wishlist))
}

// ListWishlists возвращает список вишлистов
// @Summary     Список вишлистов
// @Tags        wishlists
// @Produce     json
// @Param       page query int false "Номер страницы"
// @Param       per_page query int false "Элементов на странице"
// @Param       user_id query string false "ID пользователя"
// @Param       is_public query bool false "Публичные вишлисты"
// @Success     200 {object} response.WishlistListResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /wishlists [get]
func (h *WishlistHandler) ListWishlists(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	filter := dto.WishlistFilter{
		Page:    page,
		PerPage: perPage,
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &userID
		}
	}

	if isPublicStr := r.URL.Query().Get("is_public"); isPublicStr != "" {
		isPublic := isPublicStr == "true"
		filter.IsPublic = &isPublic
	}

	wishlists, total, err := h.uc.ListWishlists(r.Context(), filter)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.JSON(w, r, response.NewWishlistListResponse(wishlists, total, filter.Page))
}

// UpdateWishlist обновляет вишлист
// @Summary     Обновить вишлист
// @Tags        wishlists
// @Accept      json
// @Produce     json
// @Param       id path string true "ID вишлиста"
// @Param       body body request.UpdateWishlistRequest true "Данные для обновления"
// @Success     200 {object} response.WishlistResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /wishlists/{id} [patch]
func (h *WishlistHandler) UpdateWishlist(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid wishlist ID")))
		return
	}

	var req request.UpdateWishlistRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	if err := helpers.Validate(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем ID текущего пользователя
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(errors.New("unauthorized")))
		return
	}

	input := dto.UpdateWishlistInput{
		Title:       req.Title,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	wishlist, err := h.uc.UpdateWishlist(r.Context(), id, userID, input)
	if err != nil {
		if errors.Is(err, definitions.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
		} else if errors.Is(err, definitions.ErrForbidden) {
			render.Status(r, http.StatusForbidden)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.JSON(w, r, response.NewWishlistResponse(wishlist))
}

// DeleteWishlist удаляет вишлист
// @Summary     Удалить вишлист
// @Tags        wishlists
// @Param       id path string true "ID вишлиста"
// @Success     204
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /wishlists/{id} [delete]
func (h *WishlistHandler) DeleteWishlist(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid wishlist ID")))
		return
	}

	// Получаем ID текущего пользователя
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(errors.New("unauthorized")))
		return
	}

	if err := h.uc.DeleteWishlist(r.Context(), id, userID); err != nil {
		if errors.Is(err, definitions.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
		} else if errors.Is(err, definitions.ErrForbidden) {
			render.Status(r, http.StatusForbidden)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateItem создаёт элемент вишлиста
// @Summary     Создать элемент вишлиста
// @Tags        wishlist-items
// @Accept      json
// @Produce     json
// @Param       wishlist_id path string true "ID вишлиста"
// @Param       body body request.CreateWishlistItemRequest true "Данные элемента"
// @Success     201 {object} response.WishlistItemResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Router      /wishlists/{wishlist_id}/items [post]
func (h *WishlistHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	wishlistIDStr := chi.URLParam(r, "wishlist_id")
	wishlistID, err := uuid.Parse(wishlistIDStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid wishlist ID")))
		return
	}

	var req request.CreateWishlistItemRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	if err := helpers.Validate(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем ID текущего пользователя
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(errors.New("unauthorized")))
		return
	}

	input := dto.CreateWishlistItemInput{
		WishlistID:  wishlistID,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Priority:    req.Priority,
		Category:    req.Category,
	}

	item, err := h.uc.CreateItem(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, definitions.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
		} else if errors.Is(err, definitions.ErrForbidden) {
			render.Status(r, http.StatusForbidden)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем информацию о вишлисте для проверки владельца
	wishlist, err := h.uc.GetWishlist(r.Context(), wishlistID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	currentUserID := helpers.GetUserIDFromCtxOptional(r.Context())
	isOwner := currentUserID != nil && *currentUserID == wishlist.UserID

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response.NewWishlistItemResponse(item, currentUserID, isOwner))
}

// GetItem возвращает элемент вишлиста
// @Summary     Получить элемент вишлиста
// @Tags        wishlist-items
// @Produce     json
// @Param       wishlist_id path string true "ID вишлиста"
// @Param       id path string true "ID элемента"
// @Success     200 {object} response.WishlistItemResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /wishlists/{wishlist_id}/items/{id} [get]
func (h *WishlistHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid item ID")))
		return
	}

	item, err := h.uc.GetItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, definitions.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем информацию о вишлисте для проверки владельца
	wishlist, err := h.uc.GetWishlist(r.Context(), item.WishlistID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	currentUserID := helpers.GetUserIDFromCtxOptional(r.Context())
	isOwner := currentUserID != nil && *currentUserID == wishlist.UserID

	render.JSON(w, r, response.NewWishlistItemResponse(item, currentUserID, isOwner))
}

// ListItems возвращает список элементов вишлиста
// @Summary     Список элементов вишлиста
// @Tags        wishlist-items
// @Produce     json
// @Param       wishlist_id path string true "ID вишлиста"
// @Param       page query int false "Номер страницы"
// @Param       per_page query int false "Элементов на странице"
// @Param       is_purchased query bool false "Фильтр по статусу покупки"
// @Success     200 {object} response.WishlistItemListResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /wishlists/{wishlist_id}/items [get]
func (h *WishlistHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	wishlistIDStr := chi.URLParam(r, "wishlist_id")
	wishlistID, err := uuid.Parse(wishlistIDStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid wishlist ID")))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	filter := dto.WishlistItemFilter{
		WishlistID: wishlistID,
		Page:       page,
		PerPage:    perPage,
	}

	if isPurchasedStr := r.URL.Query().Get("is_purchased"); isPurchasedStr != "" {
		isPurchased := isPurchasedStr == "true"
		filter.IsPurchased = &isPurchased
	}

	items, total, err := h.uc.ListItems(r.Context(), filter)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем информацию о вишлисте для проверки владельца
	wishlist, err := h.uc.GetWishlist(r.Context(), wishlistID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем текущего пользователя (может быть nil если не авторизован)
	currentUserID := helpers.GetUserIDFromCtxOptional(r.Context())

	// Проверяем является ли текущий пользователь владельцем
	isOwner := currentUserID != nil && *currentUserID == wishlist.UserID

	render.JSON(w, r, response.NewWishlistItemListResponse(items, total, filter.Page, currentUserID, isOwner))
}

// UpdateItem обновляет элемент вишлиста
// @Summary     Обновить элемент вишлиста
// @Tags        wishlist-items
// @Accept      json
// @Produce     json
// @Param       wishlist_id path string true "ID вишлиста"
// @Param       id path string true "ID элемента"
// @Param       body body request.UpdateWishlistItemRequest true "Данные для обновления"
// @Success     200 {object} response.WishlistItemResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /wishlists/{wishlist_id}/items/{id} [patch]
func (h *WishlistHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid item ID")))
		return
	}

	var req request.UpdateWishlistItemRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	if err := helpers.Validate(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем ID текущего пользователя
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(errors.New("unauthorized")))
		return
	}

	input := dto.UpdateWishlistItemInput{
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Price:       req.Price,
		Priority:    req.Priority,
		IsPurchased: req.IsPurchased,
	}

	item, err := h.uc.UpdateItem(r.Context(), id, userID, input)
	if err != nil {
		if errors.Is(err, definitions.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
		} else if errors.Is(err, definitions.ErrForbidden) {
			render.Status(r, http.StatusForbidden)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Получаем информацию о вишлисте для проверки владельца
	wishlist, err := h.uc.GetWishlist(r.Context(), item.WishlistID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	currentUserID := helpers.GetUserIDFromCtxOptional(r.Context())
	isOwner := currentUserID != nil && *currentUserID == wishlist.UserID

	render.JSON(w, r, response.NewWishlistItemResponse(item, currentUserID, isOwner))
}

// DeleteItem удаляет элемент вишлиста
// @Summary     Удалить элемент вишлиста
// @Tags        wishlist-items
// @Param       wishlist_id path string true "ID вишлиста"
// @Param       id path string true "ID элемента"
// @Success     204
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /wishlists/{wishlist_id}/items/{id} [delete]
func (h *WishlistHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid item ID")))
		return
	}

	// Получаем ID текущего пользователя
	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(errors.New("unauthorized")))
		return
	}

	if err := h.uc.DeleteItem(r.Context(), id, userID); err != nil {
		if errors.Is(err, definitions.ErrNotFound) {
			render.Status(r, http.StatusNotFound)
		} else if errors.Is(err, definitions.ErrForbidden) {
			render.Status(r, http.StatusForbidden)
		} else {
			render.Status(r, http.StatusInternalServerError)
		}
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReserveItem бронирует элемент вишлиста
func (h *WishlistHandler) ReserveItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid item ID")))
		return
	}

	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	// Читаем параметр isIncognito из body
	var req struct {
		IsIncognito bool `json:"isIncognito"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Если body пустой, используем false по умолчанию
		req.IsIncognito = false
	}

	if err := h.uc.ReserveItem(r.Context(), itemID, userID, req.IsIncognito); err != nil {
		h.log.Error("failed to reserve item", "error", err, "itemID", itemID, "userID", userID)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "item reserved"})
}

// UnreserveItem снимает бронирование с элемента
func (h *WishlistHandler) UnreserveItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.NewErrorResponse(errors.New("invalid item ID")))
		return
	}

	userID, err := helpers.GetUserIDFromCtx(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	if err := h.uc.UnreserveItem(r.Context(), itemID, userID); err != nil {
		h.log.Error("failed to unreserve item", "error", err, "itemID", itemID, "userID", userID)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.NewErrorResponse(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "item unreserved"})
}

