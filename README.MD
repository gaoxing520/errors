# Application Errors with Integrated Logging

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
    * Provides convenience methods (`Log`, `LogError`, `LogWarn`, etc.) to log `AppError` instances
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

## License

This project is licensed under the [MIT LICENSE] - see the [LICENSE](LICENSE) file for details.