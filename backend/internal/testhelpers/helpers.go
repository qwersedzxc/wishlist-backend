package testhelpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// NewTestContext создаёт context.Background для использования в тестах.
func NewTestContext(t *testing.T) context.Context {
	t.Helper()

	return context.Background()
}

// AssertError проверяет, что err соответствует target через errors.Is.
func AssertError(t *testing.T, err error, target error) {
	t.Helper()
	require.ErrorIs(t, err, target)
}

// AssertNoError проверяет отсутствие ошибки.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}
