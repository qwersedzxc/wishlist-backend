package helpers

import (
	"net/http"
	"sync"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

var (
	once     sync.Once
	validate *validator.Validate
	decoder  *form.Decoder
)

func getValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
		decoder = form.NewDecoder()
	})

	return validate
}

func getDecoder() *form.Decoder {
	once.Do(func() {
		validate = validator.New()
		decoder = form.NewDecoder()
	})

	return decoder
}

// ValidateStruct валидирует структуру по тегам validate.
func ValidateStruct(s any) error {
	return getValidator().Struct(s)
}

// Validate алиас для ValidateStruct
func Validate(s any) error {
	return ValidateStruct(s)
}

// DecodeForm декодирует query-параметры запроса в структуру dst.
func DecodeForm(r *http.Request, dst any) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	return getDecoder().Decode(dst, r.Form)
}
