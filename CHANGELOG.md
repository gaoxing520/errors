# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Comprehensive test suite with unit tests, benchmarks, and examples
- New predefined HTTP status code errors (400, 401, 403, 404, 409, 500, 503)
- Utility functions `IsAppError()` and `GetErrorCode()`
- Improved logger configuration with `SetLogger()` and `SetLogOutput()`
- Automated CI/CD pipeline with GitHub Actions
- Release automation workflow
- Development tools (Makefile, release script)

### Changed

- Replaced `github.com/pkg/errors` dependency with Go standard library
- Fixed memory safety issue in `With()` method - now returns new instance instead of modifying original
- Improved `WithCause()` method with better null handling
- Fixed `Success` error to return empty string correctly
- Enhanced error message formatting

### Fixed

- Dependency issues with undeclared packages
- Potential memory leaks from shared error instances
- Success error formatting bug

### Security

- Removed external dependency to reduce attack surface

## [v0.1.0] - Initial Release

### Added

- Basic `AppError` interface and `AppCommonError` implementation
- Error codes and chaining support
- Integration with zerolog for structured logging
- Predefined common errors (`Success`, `Unknown`, `ErrSystem`)
- Method chaining for error operations
- Standard Go error interface compatibility
