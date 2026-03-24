package helpers

import "github.com/samber/lo"

// Map преобразует срез типа T в срез типа R с помощью функции fn.
func Map[T, R any](s []T, fn func(T, int) R) []R {
	return lo.Map(s, fn)
}

// Filter возвращает элементы среза, удовлетворяющие предикату fn.
func Filter[T any](s []T, fn func(T, int) bool) []T {
	return lo.Filter(s, fn)
}

// Contains возвращает true, если элемент присутствует в срезе.
func Contains[T comparable](s []T, elem T) bool {
	return lo.Contains(s, elem)
}
