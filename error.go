package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type AppError interface {
	error

	Code() int

	WithCause(cause error) AppError
	With(format string, args ...any) AppError
	Is(target error) bool
	Unwrap() error

	LogError() AppError
	LogWarn() AppError
	LogInfo() AppError
	LogDebug() AppError
	LogTrace() AppError
	LogFatal()
	LogPanic()
}

// AppCommonError represents an error with a specific code and an underlying cause.
type AppCommonError struct {
	err  error
	code int
}

// Error implements the error interface.
func (e *AppCommonError) Error() string {
	if e.Is(Success) {
		if e.err == nil {
			return ""
		}
		return e.err.Error()
	}

	baseMsg := ""
	if e.err != nil {
		baseMsg = e.err.Error()
	}

	return fmt.Sprintf("%s, code=%d", baseMsg, e.code)
}

// Code returns the application-specific error code.
func (e *AppCommonError) Code() int {
	return e.code
}

// Is reports whether this AppError matches the target AppError code.
func (e *AppCommonError) Is(target error) bool {
	// Using standard errors.Is from Go 1.13+ is generally preferred for comparing errors.
	if target == nil {
		return false
	}

	var tgt *AppCommonError
	if ok := errors.As(target, &tgt); !ok {
		return false
	}

	// Compare codes
	return e.code == tgt.code
}

// Unwrap returns the next error in the error chain.
func (e *AppCommonError) Unwrap() error {
	return errors.Unwrap(e.err)
}

// WithCause creates a new AppError instance based on the current AppError template (e)
func (e *AppCommonError) WithCause(cause error) AppError {
	if cause == nil {
		return NewAppError(e.code, e.err.Error())
	}
	wrappedErr := fmt.Errorf("%s: %w", e.err.Error(), cause)

	return &AppCommonError{
		code: e.code,
		err:  wrappedErr, // Store the wrapped error chain
	}
}

// With adds formatted context to the current AppError's underlying error chain.
func (e *AppCommonError) With(format string, args ...any) AppError {
	e.err = errors.Wrapf(e.err, format, args...)
	return e
}

// LogPanic logs the AppError at panic level using the default logger and panics.
func (e *AppCommonError) LogPanic() {
	Panic().Int("code", e.code).Err(e.err).Caller(1).Send()
}

// LogFatal logs the AppError at fatal level using the default logger and exits.
func (e *AppCommonError) LogFatal() {
	Fatal().Int("code", e.code).Err(e.err).Caller(1).Send()
}

// LogInfo logs the AppError at info level using the default logger.
func (e *AppCommonError) LogInfo() AppError {
	Info().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogWarn logs the AppError at warn level using the default logger.
func (e *AppCommonError) LogWarn() AppError {
	Warn().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogError logs the AppError at error level using the default logger.
func (e *AppCommonError) LogError() AppError {
	Error().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogDebug logs the AppError at debug level using the default logger.
func (e *AppCommonError) LogDebug() AppError {
	Debug().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogTrace logs the AppError at trace level using the default logger.
func (e *AppCommonError) LogTrace() AppError {
	Trace().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// NewAppError creates a new basic AppCommonError instance with a code and a base message.
func NewAppError(code int, msg string) AppError {
	base := errors.New(msg)
	return &AppCommonError{
		code: code,
		err:  base,
	}
}

var (
	Success = NewAppError(0, "success")
)

var (
	Unknown   = NewAppError(-1, "unknown error")
	ErrSystem = NewAppError(999999, "system error")
)
