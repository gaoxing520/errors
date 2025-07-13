package errors_test

import (
	"fmt"

	"github.com/gaoxing520/errors"
)

func ExampleNewAppError() {
	err := errors.NewAppError(1001, "user not found")
	fmt.Printf("Error: %s, Code: %d\n", err.Error(), err.Code())
	// Output: Error: user not found, code=1001, Code: 1001
}

func ExampleAppCommonError_WithCause() {
	originalErr := fmt.Errorf("connection timeout")
	appErr := errors.NewAppError(1002, "service unavailable")
	wrappedErr := appErr.WithCause(originalErr)

	fmt.Println(wrappedErr.Error())
	// Output: service unavailable: connection timeout, code=1002
}

func ExampleAppCommonError_With() {
	err := errors.NewAppError(1003, "validation failed")
	contextErr := err.With("invalid email format for user %s", "john@")

	fmt.Println(contextErr.Error())
	// Output: invalid email format for user john@: validation failed, code=1003
}

func ExampleIsAppError() {
	appErr := errors.NewAppError(1004, "application error")
	stdErr := fmt.Errorf("standard error")

	if extracted, ok := errors.IsAppError(appErr); ok {
		fmt.Printf("AppError with code: %d\n", extracted.Code())
	}

	if _, ok := errors.IsAppError(stdErr); !ok {
		fmt.Println("Not an AppError")
	}
	// Output: AppError with code: 1004
	// Not an AppError
}

func ExampleGetErrorCode() {
	appErr := errors.NewAppError(1005, "application error")
	stdErr := fmt.Errorf("standard error")

	fmt.Printf("AppError code: %d\n", errors.GetErrorCode(appErr))
	fmt.Printf("Standard error code: %d\n", errors.GetErrorCode(stdErr))
	// Output: AppError code: 1005
	// Standard error code: -1
}

func Example_chaining() {
	// Simulate a complex error scenario
	dbErr := fmt.Errorf("connection lost")

	err := errors.NewAppError(2001, "user operation failed").
		WithCause(dbErr).
		With("failed to update profile for user %s", "alice")

	fmt.Printf("Final error: %s\n", err.Error())
	fmt.Printf("Error code: %d\n", err.Code())
	// Output: Final error: failed to update profile for user alice: user operation failed: connection lost, code=2001
	// Error code: 2001
}

func Example_predefinedErrors() {
	// Using predefined errors
	fmt.Printf("Success: '%s' (code: %d)\n", errors.Success.Error(), errors.Success.Code())
	fmt.Printf("Not Found: '%s' (code: %d)\n", errors.ErrNotFound.Error(), errors.ErrNotFound.Code())
	fmt.Printf("Unauthorized: '%s' (code: %d)\n", errors.ErrUnauthorized.Error(), errors.ErrUnauthorized.Code())
	// Output: Success: '' (code: 0)
	// Not Found: 'not found, code=404' (code: 404)
	// Unauthorized: 'unauthorized, code=401' (code: 401)
}
