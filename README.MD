# Application Errors with Integrated Logging

[English](README.MD) | [中文文档](docs/README_zh.md)

`errors` is a Go package providing a structured approach to application-specific errors, incorporating integer error
codes, error chaining, and seamless integration with the `zeralog` structured logging library.

It aims to standardize error handling across your application, making it easier to:

* Identify specific error conditions using predefined codes.
* Maintain context and cause using error wrapping.
* Log errors consistently with relevant structured data (like the error code and chain).
* Unify successful and erroneous API responses using a common structure.

## features

* Custom Error Type (`AppError`)
* Defines a standard structure for application errors.
* Integer Error Codes:
    * Each `AppError` includes a distinct integer code, useful for API responses or internal logic branches.
* (Error Chaining:
    * Supports wrapping underlying errors (`WithCause`) and adding contextual messages (`With`),
      preserving the error chain compatible with `errors.Is` and `errors.As`.
* `zevalog` Integration:
    * Provides convenience methods (`LogError`, `LogWarn`, `LogInfo`, etc.) to log `AppError` instances
      directly, automatically including the code and the full error chain.
* Predefined Common Errors:
    * Includes a set of standard error codes and messages for common scenarios.
* Unified Response Structure:
    * The `Error()` method is formatted to be suitable for direct use in API response messages alongside the code.
      Special
      handling for the `Success` code ensures a clean output.

## Installation

Install the package using `go get`:

```bash
go get github.com/gaoxing520/errors
go mod tidy
```

## Usage

### Basic Error Creation

Create application-specific errors with integer codes:

```go
package main

import (
	"fmt"
	"github.com/gaoxing520/errors"
)

func main() {
	// Create a custom error with code and message
	err := errors.NewAppError(1001, "user not found")
	fmt.Printf("Error: %s, Code: %d\n", err.Error(), err.Code())
	// Output: Error: user not found, code=1001, Code: 1001
}
```

### Using Predefined Errors

The library provides several predefined error types:

```go
// Success case (code: 0)
successErr := errors.Success
fmt.Println(successErr.Error()) // Output: "" (empty string for success)

// Unknown error (code: -1)
unknownErr := errors.Unknown
fmt.Println(unknownErr.Error()) // Output: unknown error, code=-1

// System error (code: 999999)
systemErr := errors.ErrSystem
fmt.Println(systemErr.Error()) // Output: system error, code=999999
```

### Error Chaining and Context

Add context and wrap underlying errors:

```go
// Add context to an error
err := errors.NewAppError(1002, "database operation failed")
contextErr := err.With("failed to update user %s", "john_doe")
fmt.Println(contextErr.Error())
// Output: failed to update user john_doe: database operation failed, code=1002

// Wrap an underlying error
originalErr := fmt.Errorf("connection timeout")
wrappedErr := errors.NewAppError(1003, "service unavailable").WithCause(originalErr)
fmt.Println(wrappedErr.Error())
// Output: service unavailable: connection timeout, code=1003
```

### Logging Integration

The library integrates seamlessly with zerolog for structured logging:

```go
// Error logging at different levels
err := errors.NewAppError(1004, "validation failed")

// Log at different levels
err.LogInfo() // Log at INFO level
err.LogWarn()  // Log at WARN level
err.LogError() // Log at ERROR level
err.LogDebug() // Log at DEBUG level
err.LogTrace() // Log at TRACE level

// Fatal and panic logging
err.LogFatal() // Logs and exits the application
err.LogPanic() // Logs and panics
```

Note about nil receivers
------------------------

To make the logging helpers safe to call even when an `AppCommonError` pointer is `nil`, the Log* methods
omit the `code` field when the receiver is `nil`. This prevents nil-pointer dereference panics in call sites
that may have an absent error value. For example:

```go
var e *errors.AppCommonError = nil
e.LogInfo() // safe: will log an info-level event but will not include a `code` field
```

This design choice favors robustness and avoids introducing a magic numeric default in logs. If you prefer a
different behavior (for example: always including a default code), consider wrapping the call site or
constructing a non-nil `AppCommonError` with the desired default code.

### Chaining Operations

You can chain error operations for concise code:

```go
err := errors.NewAppError(1005, "authentication failed").
With("user attempted login with invalid credentials").
WithCause(fmt.Errorf("password mismatch")).
LogWarn() // Chain logging

// The error is logged and can still be used
return err
```

### Error Comparison

Use standard Go error comparison methods:

```go
err1 := errors.NewAppError(1006, "not found")
err2 := errors.NewAppError(1006, "not found")

// Compare errors by code
if errors.Is(err1, err2) {
fmt.Println("Errors have the same code")
}

// Check specific error types
var appErr *errors.AppCommonError
if errors.As(err1, &appErr) {
fmt.Printf("Error code: %d\n", appErr.Code())
}
```

### Custom Logger Configuration

Configure the default logger used by the error logging methods:

```go
import (
"os"
"github.com/rs/zerolog"
"github.com/gaoxing520/errors"
)

// Set custom output
errors.DefaultLogOutput = os.Stdout

// Configure the default logger
errors.DefaultLogger = zerolog.New(os.Stdout).
With().
Timestamp().
Caller().
Logger().
Level(zerolog.InfoLevel)
```

### Complete Example

Here's a complete example showing typical usage in an application:

```go
package main

import (
	"fmt"
	"github.com/gaoxing520/errors"
)

// Define application-specific errors
var (
	ErrUserNotFound    = errors.NewAppError(2001, "user not found")
	ErrInvalidPassword = errors.NewAppError(2002, "invalid password")
	ErrDatabaseError   = errors.NewAppError(2003, "database error")
)

func authenticateUser(username, password string) error {
	// Simulate user lookup
	if username == "" {
		return ErrUserNotFound.With("username is empty").LogWarn()
	}

	// Simulate password validation
	if password != "secret" {
		return ErrInvalidPassword.
			With("authentication failed for user %s", username).
			LogError()
	}

	// Simulate database error
	if username == "dbfail" {
		dbErr := fmt.Errorf("connection lost")
		return ErrDatabaseError.
			WithCause(dbErr).
			With("failed to verify user %s", username).
			LogError()
	}

	// Success case
	errors.Success.With("user %s authenticated successfully", username).LogInfo()
	return nil
}

func main() {
	// Test different scenarios
	scenarios := []struct {
		username, password string
	}{
		{"", "secret"},       // Empty username
		{"john", "wrong"},    // Wrong password
		{"dbfail", "secret"}, // Database error
		{"john", "secret"},   // Success
	}

	for _, scenario := range scenarios {
		err := authenticateUser(scenario.username, scenario.password)
		if err != nil {
			fmt.Printf("Authentication failed: %s\n", err.Error())
		} else {
			fmt.Println("Authentication successful")
		}
	}
}
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make cover

# Run benchmarks
make bench

# Run all checks (format, vet, test, bench)
make all
```

### Project Structure

```
├── error.go           # Core error types and functions
├── logger.go          # Logging integration
├── *_test.go         # Test files
├── docs/             # Documentation
│   └── README_zh.md  # Chinese documentation
└── .github/workflows/ # CI/CD workflows
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run the test suite (`make all`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is licensed under the [MIT LICENSE] - see the [LICENSE](LICENSE) file for details.