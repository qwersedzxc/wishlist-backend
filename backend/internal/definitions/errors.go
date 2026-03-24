package definitions

import "errors"

// Sentinel-ошибки доменного слоя.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrForbidden     = errors.New("forbidden")
	ErrBadRequest    = errors.New("bad request")
)
