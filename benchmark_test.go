package errors

import (
	"fmt"
	"testing"
)

func BenchmarkNewAppError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewAppError(1000, "benchmark error")
	}
}

func BenchmarkWithCause(b *testing.B) {
	baseErr := NewAppError(1001, "base error")
	cause := fmt.Errorf("underlying cause")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseErr.WithCause(cause)
	}
}

func BenchmarkWith(b *testing.B) {
	baseErr := NewAppError(1002, "base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseErr.With("context for iteration %d", i)
	}
}

func BenchmarkErrorString(b *testing.B) {
	err := NewAppError(1003, "benchmark error").
		WithCause(fmt.Errorf("underlying error")).
		With("additional context")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkIsAppError(b *testing.B) {
	err := NewAppError(1004, "test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = IsAppError(err)
	}
}

func BenchmarkGetErrorCode(b *testing.B) {
	err := NewAppError(1005, "test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetErrorCode(err)
	}
}

func BenchmarkChaining(b *testing.B) {
	originalErr := fmt.Errorf("original error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAppError(1006, "base error").
			WithCause(originalErr).
			With("context %d", i)
	}
}
