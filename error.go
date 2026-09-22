// Package errors provides structured application-specific error handling with integer error codes,
// error chaining, and seamless integration with zerolog structured logging.
package errors

import (
	"errors"
	"fmt"
)

type AppError interface {
	error

	Code() int
	// Msg returns the first-layer abstract message from NewAppError,
	// without With/WithCause chain details. Suitable for API responses.
	Msg() string

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
	msg  string // stable public message; not changed by With/WithCause
}

// Error implements the error interface. Includes the full With/WithCause chain
// and ", code=N" — use Msg() for API responses instead.
func (e *AppCommonError) Error() string {
	// Be defensive for nil receivers
	if e == nil {
		return "<nil>"
	}

	// Special handling for success code
	if e.code == 0 {
		if e.err == nil {
			return ""
		}
		// For success with additional context, return only the context
		return e.err.Error()
	}

	baseMsg := ""
	if e.err != nil {
		baseMsg = e.err.Error()
	}

	if baseMsg == "" {
		return fmt.Sprintf("code=%d", e.code)
	}

	return fmt.Sprintf("%s, code=%d", baseMsg, e.code)
}

// Msg returns the first-layer abstract message defined at NewAppError time.
func (e *AppCommonError) Msg() string {
	if e == nil {
		return ""
	}
	return e.msg
}

// Code returns the application-specific error code.
func (e *AppCommonError) Code() int {
	if e == nil {
		return -1
	}
	return e.code
}

// Is reports whether this AppError matches the target AppError code.
func (e *AppCommonError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}

	var tgt *AppCommonError
	if errors.As(target, &tgt) {
		return e.code == tgt.code
	}

	return false
}

// Unwrap returns the next error in the error chain.
func (e *AppCommonError) Unwrap() error {
	if e == nil {
		return nil
	}

	// Return the stored underlying error. Do not unwrap one level here; let callers
	// use errors.Unwrap/Is/As to traverse further if needed.
	return e.err
}

// WithCause creates a new AppError instance based on the current AppError template (e)
func (e *AppCommonError) WithCause(cause error) AppError {
	if e == nil {
		if cause == nil {
			return nil
		}
		return &AppCommonError{code: -1, err: cause}
	}

	if cause == nil {
		// Return a copy of the current error
		return &AppCommonError{
			code: e.code,
			err:  e.err,
			msg:  e.msg,
		}
	}

	if e.err == nil {
		// No base message to prepend; use the cause directly.
		return &AppCommonError{
			code: e.code,
			err:  cause,
			msg:  e.msg,
		}
	}

	wrappedErr := fmt.Errorf("%s: %w", e.err.Error(), cause)

	return &AppCommonError{
		code: e.code,
		err:  wrappedErr,
		msg:  e.msg,
	}
}

// With adds formatted context to the current AppError's underlying error chain.
// Returns a new AppError instance to avoid modifying the original.
func (e *AppCommonError) With(format string, args ...any) AppError {
	if e == nil {
		// If receiver is nil, create a new error with code -1
		return &AppCommonError{code: -1, err: fmt.Errorf(format, args...)}
	}

	contextMsg := fmt.Sprintf(format, args...)
	var wrappedErr error
	if e.err == nil {
		// No underlying error to wrap; return a simple error with only context
		wrappedErr = errors.New(contextMsg)
	} else {
		wrappedErr = fmt.Errorf("%s: %w", contextMsg, e.err)
	}

	return &AppCommonError{
		code: e.code,
		err:  wrappedErr,
		msg:  e.msg,
	}
}

// LogPanic logs the AppError at panic level using the default logger and panics.
func (e *AppCommonError) LogPanic() {
	if e == nil {
		// For nil receiver, do not add a code field; just log the error (nil) to avoid panics
		Panic().Err(nil).Caller(1).Send()
		return
	}
	Panic().Int("code", e.code).Err(e.err).Caller(1).Send()
}

// LogFatal logs the AppError at fatal level using the default logger and exits.
func (e *AppCommonError) LogFatal() {
	if e == nil {
		Fatal().Err(nil).Caller(1).Send()
		return
	}
	Fatal().Int("code", e.code).Err(e.err).Caller(1).Send()
}

// LogInfo logs the AppError at info level using the default logger.
func (e *AppCommonError) LogInfo() AppError {
	if e == nil {
		Info().Err(nil).Caller(1).Send()
		return e
	}
	Info().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogWarn logs the AppError at warn level using the default logger.
func (e *AppCommonError) LogWarn() AppError {
	if e == nil {
		Warn().Err(nil).Caller(1).Send()
		return e
	}
	Warn().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogError logs the AppError at error level using the default logger.
func (e *AppCommonError) LogError() AppError {
	if e == nil {
		Error().Err(nil).Caller(1).Send()
		return e
	}
	Error().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogDebug logs the AppError at debug level using the default logger.
func (e *AppCommonError) LogDebug() AppError {
	if e == nil {
		Debug().Err(nil).Caller(1).Send()
		return e
	}
	Debug().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// LogTrace logs the AppError at trace level using the default logger.
func (e *AppCommonError) LogTrace() AppError {
	if e == nil {
		Trace().Err(nil).Caller(1).Send()
		return e
	}
	Trace().Int("code", e.code).Err(e.err).Caller(1).Send()
	return e
}

// NewAppError creates a new basic AppCommonError instance with a code and a base message.
func NewAppError(code int, msg string) AppError {
	base := errors.New(msg)
	return &AppCommonError{
		code: code,
		err:  base,
		msg:  msg,
	}
}

// Predefined common errors
var (
	Success = &AppCommonError{code: 0, err: nil}
)

var (
	Unknown   = NewAppError(-1, "unknown error")
	ErrSystem = NewAppError(999999, "system error")

	// Common application errors
	ErrInvalidInput       = NewAppError(400, "invalid input")
	ErrUnauthorized       = NewAppError(401, "unauthorized")
	ErrForbidden          = NewAppError(403, "forbidden")
	ErrNotFound           = NewAppError(404, "not found")
	ErrConflict           = NewAppError(409, "conflict")
	ErrInternalError      = NewAppError(500, "internal server error")
	ErrServiceUnavailable = NewAppError(503, "service unavailable")
)

// IsAppError checks if an error is an AppError and returns it
func IsAppError(err error) (*AppCommonError, bool) {
	var appErr *AppCommonError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetErrorCode extracts the error code from an error, returns -1 if not an AppError
func GetErrorCode(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.Code()
	}
	return -1
}
