// Package errors provides structured application-specific error handling with integer error codes,
// error chaining, and seamless integration with zerolog structured logging.
package errors

import (
	"errors"
	"fmt"
	"strings"
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
// Pointer fields come first so the garbage collector scans a smaller prefix.
type AppCommonError struct {
	err  error
	msg  string // stable public message; not changed by With/WithCause
	code int
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
	if errors.As(target, &tgt) && tgt != nil {
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

// WithCause returns a new AppError that records cause after the current chain.
// The previous chain and cause both stay visible to errors.Is and errors.As.
// A nested AppError contributes its chain text only, so Error appends this
// value's ", code=N" suffix once.
func (e *AppCommonError) WithCause(cause error) AppError {
	if e == nil {
		if cause == nil {
			return nil
		}
		return &AppCommonError{code: -1, err: wrapCause(-1, nil, cause)}
	}

	if cause == nil {
		// Return a copy of the current error
		return &AppCommonError{
			code: e.code,
			err:  e.err,
			msg:  e.msg,
		}
	}

	return &AppCommonError{
		code: e.code,
		err:  wrapCause(e.code, e.err, cause),
		msg:  e.msg,
	}
}

// linkedError keeps the previous chain and the new cause together.
// Unwrap exposes both so errors.Is and errors.As can match either one.
// Pointer fields come first so the garbage collector scans a smaller prefix.
type linkedError struct {
	prev  error
	cause error
	msg   string
}

func (e *linkedError) Error() string {
	if e == nil {
		return ""
	}
	return e.msg
}

func (e *linkedError) Unwrap() []error {
	if e == nil {
		return nil
	}
	out := make([]error, 0, 2)
	if e.prev != nil {
		out = append(out, e.prev)
	}
	if e.cause != nil {
		out = append(out, e.cause)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// chainText is the error text without this value's ", code=N" suffix.
func (e *AppCommonError) chainText() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

// messageWithoutCode returns text suitable for nesting under another AppError.
// AppError values contribute chain text so the outer Error method is the only
// place that appends ", code=N".
func messageWithoutCode(err error) string {
	if err == nil {
		return ""
	}
	if app, ok := err.(*AppCommonError); ok {
		return app.chainText()
	}

	var app *AppCommonError
	if errors.As(err, &app) && app != nil {
		full := err.Error()
		coded := app.Error()
		if coded != "" && strings.HasSuffix(full, coded) {
			return full[:len(full)-len(coded)] + app.chainText()
		}
	}
	return err.Error()
}

// wrapCause links prev and cause. code 0 prints the underlying error as-is,
// so its text is preserved. Any other code appends ", code=N" in Error and
// must not repeat a nested AppError code.
func wrapCause(code int, prev, cause error) error {
	if cause == nil {
		return prev
	}

	suffix := cause.Error()
	if code != 0 {
		suffix = messageWithoutCode(cause)
	}
	if prev == nil {
		if code == 0 {
			return cause
		}
		return &linkedError{msg: suffix, cause: cause}
	}
	return &linkedError{
		msg:   prev.Error() + ": " + suffix,
		prev:  prev,
		cause: cause,
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
		// For nil receiver, do not add a code field; just log the error (nil) to avoid panics.
		// CallerSkipFrame adjusts the logger's caller hook so the field points at the caller
		// without writing a second caller key.
		Panic().Err(nil).CallerSkipFrame(1).Send()
		return
	}
	Panic().Int("code", e.code).Err(e.err).CallerSkipFrame(1).Send()
}

// LogFatal logs the AppError at fatal level using the default logger and exits.
func (e *AppCommonError) LogFatal() {
	if e == nil {
		Fatal().Err(nil).CallerSkipFrame(1).Send()
		return
	}
	Fatal().Int("code", e.code).Err(e.err).CallerSkipFrame(1).Send()
}

// LogInfo logs the AppError at info level using the default logger.
// A nil receiver returns a nil AppError.
func (e *AppCommonError) LogInfo() AppError {
	if e == nil {
		Info().Err(nil).CallerSkipFrame(1).Send()
		return nil
	}
	Info().Int("code", e.code).Err(e.err).CallerSkipFrame(1).Send()
	return e
}

// LogWarn logs the AppError at warn level using the default logger.
// A nil receiver returns a nil AppError.
func (e *AppCommonError) LogWarn() AppError {
	if e == nil {
		Warn().Err(nil).CallerSkipFrame(1).Send()
		return nil
	}
	Warn().Int("code", e.code).Err(e.err).CallerSkipFrame(1).Send()
	return e
}

// LogError logs the AppError at error level using the default logger.
// A nil receiver returns a nil AppError.
func (e *AppCommonError) LogError() AppError {
	if e == nil {
		Error().Err(nil).CallerSkipFrame(1).Send()
		return nil
	}
	Error().Int("code", e.code).Err(e.err).CallerSkipFrame(1).Send()
	return e
}

// LogDebug logs the AppError at debug level using the default logger.
// A nil receiver returns a nil AppError.
func (e *AppCommonError) LogDebug() AppError {
	if e == nil {
		Debug().Err(nil).CallerSkipFrame(1).Send()
		return nil
	}
	Debug().Int("code", e.code).Err(e.err).CallerSkipFrame(1).Send()
	return e
}

// LogTrace logs the AppError at trace level using the default logger.
// A nil receiver returns a nil AppError.
func (e *AppCommonError) LogTrace() AppError {
	if e == nil {
		Trace().Err(nil).CallerSkipFrame(1).Send()
		return nil
	}
	Trace().Int("code", e.code).Err(e.err).CallerSkipFrame(1).Send()
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

// IsAppError reports whether err is an AppError and returns it.
// A nil error, and a nil *AppCommonError stored in a non-nil error interface, report false.
func IsAppError(err error) (*AppCommonError, bool) {
	if err == nil {
		return nil, false
	}
	var appErr *AppCommonError
	if errors.As(err, &appErr) && appErr != nil {
		return appErr, true
	}
	return nil, false
}

// ErrorCode reports the integer code carried by err.
// The boolean is false when err is nil or not an AppError.
// Unknown is an AppError, so it returns (-1, true). That distinguishes it from
// a nil error or a plain error, both of which return (-1, false).
// When the boolean is false the returned code is -1 and should be ignored.
func ErrorCode(err error) (int, bool) {
	appErr, ok := IsAppError(err)
	if !ok {
		return -1, false
	}
	return appErr.code, true
}

// GetErrorCode returns the code carried by err.
// It returns -1 when err is nil or not an AppError.
// Unknown's code is also -1; use ErrorCode when those cases must be told apart.
func GetErrorCode(err error) int {
	code, _ := ErrorCode(err)
	return code
}
