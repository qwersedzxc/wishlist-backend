package response

import (
	"errors"
	"net/http"

	"github.com/KaoriEl/golang-boilerplate/internal/definitions"
	"github.com/go-chi/render"
)

// ErrorResponse унифицированный формат ошибки в JSON-ответе.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// NewErrorResponse создаёт ErrorResponse из ошибки
func NewErrorResponse(err error) ErrorResponse {
	code := http.StatusInternalServerError
	message := err.Error()

	switch {
	case errors.Is(err, definitions.ErrNotFound):
		code = http.StatusNotFound
		message = "not found"
	case errors.Is(err, definitions.ErrBadRequest):
		code = http.StatusBadRequest
	case errors.Is(err, definitions.ErrForbidden):
		code = http.StatusForbidden
		message = "forbidden"
	}

	return ErrorResponse{
		Error: message,
		Code:  code,
	}
}

// ErrNotFound отвечает с 404 Not Found.
func ErrNotFound(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNotFound)
	render.JSON(w, r, ErrorResponse{Error: "not found", Code: http.StatusNotFound})
}

// ErrBadRequest отвечает с 400 Bad Request.
func ErrBadRequest(w http.ResponseWriter, r *http.Request, msg string) {
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, ErrorResponse{Error: msg, Code: http.StatusBadRequest})
}

// ErrInternal отвечает с 500 Internal Server Error.
func ErrInternal(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, ErrorResponse{Error: "internal server error", Code: http.StatusInternalServerError})
}

// MapError переводит доменную ошибку в соответствующий HTTP-ответ.
func MapError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, definitions.ErrNotFound):
		ErrNotFound(w, r)
	case errors.Is(err, definitions.ErrBadRequest):
		ErrBadRequest(w, r, err.Error())
	case errors.Is(err, definitions.ErrForbidden):
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, ErrorResponse{Error: "forbidden", Code: http.StatusForbidden})
	default:
		ErrInternal(w, r)
	}
}


// Error создает простой ErrorResponse с сообщением
func Error(message string) ErrorResponse {
	return ErrorResponse{
		Error: message,
		Code:  http.StatusBadRequest,
	}
}
