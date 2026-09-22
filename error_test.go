package errors

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(1001, "test error")

	if err.Code() != 1001 {
		t.Errorf("Expected code 1001, got %d", err.Code())
	}

	if err.Msg() != "test error" {
		t.Errorf("Expected Msg 'test error', got '%s'", err.Msg())
	}

	expected := "test error, code=1001"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestSuccessError(t *testing.T) {
	if Success.Code() != 0 {
		t.Errorf("Expected success code 0, got %d", Success.Code())
	}

	if Success.Error() != "" {
		t.Errorf("Expected empty string for success, got '%s'", Success.Error())
	}

	if Success.Msg() != "" {
		t.Errorf("Expected empty Msg for success, got '%s'", Success.Msg())
	}
}

func TestMsgStableAcrossWithAndCause(t *testing.T) {
	err := NewAppError(1003, "query database error").
		With("load group failed").
		WithCause(fmt.Errorf("record not found"))

	if got := err.Msg(); got != "query database error" {
		t.Errorf("Expected Msg 'query database error', got '%s'", got)
	}

	expectedError := "load group failed: query database error: record not found, code=1003"
	if err.Error() != expectedError {
		t.Errorf("Expected Error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestWithCause(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	appErr := NewAppError(1002, "app error")

	wrappedErr := appErr.WithCause(originalErr)

	if wrappedErr.Code() != 1002 {
		t.Errorf("Expected code 1002, got %d", wrappedErr.Code())
	}

	expected := "app error: original error, code=1002"
	if wrappedErr.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, wrappedErr.Error())
	}

	// Test that original error is preserved
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Original error should be preserved in the chain")
	}
}

func TestWithCauseNil(t *testing.T) {
	appErr := NewAppError(1003, "app error")
	wrappedErr := appErr.WithCause(nil)

	if wrappedErr.Code() != 1003 {
		t.Errorf("Expected code 1003, got %d", wrappedErr.Code())
	}

	expected := "app error, code=1003"
	if wrappedErr.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, wrappedErr.Error())
	}
}

func TestWith(t *testing.T) {
	appErr := NewAppError(1004, "base error")
	contextErr := appErr.With("context for user %s", "john")

	if contextErr.Code() != 1004 {
		t.Errorf("Expected code 1004, got %d", contextErr.Code())
	}

	expected := "context for user john: base error, code=1004"
	if contextErr.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, contextErr.Error())
	}

	// Test that original error is not modified
	originalExpected := "base error, code=1004"
	if appErr.Error() != originalExpected {
		t.Errorf("Original error should not be modified. Expected '%s', got '%s'", originalExpected, appErr.Error())
	}
}

func TestIs(t *testing.T) {
	err1 := NewAppError(1005, "error 1")
	err2 := NewAppError(1005, "error 2")
	err3 := NewAppError(1006, "error 3")

	if !err1.Is(err2) {
		t.Error("Errors with same code should be equal")
	}

	if err1.Is(err3) {
		t.Error("Errors with different codes should not be equal")
	}

	if err1.Is(nil) {
		t.Error("Error should not equal nil")
	}
}

func TestIsAppError(t *testing.T) {
	appErr := NewAppError(1007, "app error")
	stdErr := fmt.Errorf("standard error")

	// Test with AppError
	if extracted, ok := IsAppError(appErr); !ok {
		t.Error("Should recognize AppError")
	} else if extracted.Code() != 1007 {
		t.Errorf("Expected code 1007, got %d", extracted.Code())
	}

	// Test with standard error
	if _, ok := IsAppError(stdErr); ok {
		t.Error("Should not recognize standard error as AppError")
	}
}

func TestGetErrorCode(t *testing.T) {
	appErr := NewAppError(1008, "app error")
	stdErr := fmt.Errorf("standard error")

	if code := GetErrorCode(appErr); code != 1008 {
		t.Errorf("Expected code 1008, got %d", code)
	}

	if code := GetErrorCode(stdErr); code != -1 {
		t.Errorf("Expected code -1 for standard error, got %d", code)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  AppError
		code int
	}{
		{"Success", Success, 0},
		{"Unknown", Unknown, -1},
		{"ErrSystem", ErrSystem, 999999},
		{"ErrInvalidInput", ErrInvalidInput, 400},
		{"ErrUnauthorized", ErrUnauthorized, 401},
		{"ErrForbidden", ErrForbidden, 403},
		{"ErrNotFound", ErrNotFound, 404},
		{"ErrConflict", ErrConflict, 409},
		{"ErrInternalError", ErrInternalError, 500},
		{"ErrServiceUnavailable", ErrServiceUnavailable, 503},
	}

	for _, test := range tests {
		if test.err.Code() != test.code {
			t.Errorf("%s: expected code %d, got %d", test.name, test.code, test.err.Code())
		}
	}
}

func TestChaining(t *testing.T) {
	originalErr := fmt.Errorf("database connection failed")

	err := NewAppError(1009, "operation failed").
		WithCause(originalErr).
		With("failed to process user %s", "alice")

	if err.Code() != 1009 {
		t.Errorf("Expected code 1009, got %d", err.Code())
	}

	// Check that the error chain is preserved
	if !errors.Is(err, originalErr) {
		t.Error("Original error should be in the chain")
	}

	expected := "failed to process user alice: operation failed: database connection failed, code=1009"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNilReceiverLoggingDoesNotPanicAndOmitsCode(t *testing.T) {
	var buf bytes.Buffer
	// ensure all levels enabled
	SetLogLevel(zerolog.TraceLevel)
	SetLogOutput(&buf)

	var e *AppCommonError = nil

	// Call several logging methods; these should not panic and should produce output
	buf.Reset()
	e.LogInfo()
	out := buf.String()
	if out == "" {
		t.Errorf("expected some log output for LogInfo(), got empty")
	}
	if strings.Contains(out, "\"code\"") {
		t.Errorf("expected code field to be omitted for nil receiver, but found: %s", out)
	}

	buf.Reset()
	e.LogWarn()
	out = buf.String()
	if out == "" {
		t.Errorf("expected some log output for LogWarn(), got empty")
	}
	if strings.Contains(out, "\"code\"") {
		t.Errorf("expected code field to be omitted for nil receiver, but found: %s", out)
	}

	buf.Reset()
	e.LogError()
	out = buf.String()
	if out == "" {
		t.Errorf("expected some log output for LogError(), got empty")
	}
	if strings.Contains(out, "\"code\"") {
		t.Errorf("expected code field to be omitted for nil receiver, but found: %s", out)
	}

	buf.Reset()
	e.LogDebug()
	out = buf.String()
	if out == "" {
		t.Errorf("expected some log output for LogDebug(), got empty")
	}
	if strings.Contains(out, "\"code\"") {
		t.Errorf("expected code field to be omitted for nil receiver, but found: %s", out)
	}

	buf.Reset()
	e.LogTrace()
	out = buf.String()
	if out == "" {
		t.Errorf("expected some log output for LogTrace(), got empty")
	}
	if strings.Contains(out, "\"code\"") {
		t.Errorf("expected code field to be omitted for nil receiver, but found: %s", out)
	}
}
