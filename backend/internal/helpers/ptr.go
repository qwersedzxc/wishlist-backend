package helpers

// ToPtr возвращает указатель на переданное значение.
func ToPtr[T any](v T) *T {
	return &v
}

// FromPtr разыменовывает указатель; если он nil возвращает значение по умолчанию def.
func FromPtr[T any](p *T, def T) T {
	if p == nil {
		return def
	}

	return *p
}
