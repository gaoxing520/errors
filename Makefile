.PHONY: test bench cover lint fmt vet clean help release pre-release check-release generate-version

# Default target
help:
	@echo "Available targets:"
	@echo "  test         - Run all tests"
	@echo "  bench        - Run benchmarks"
	@echo "  cover        - Run tests with coverage"
	@echo "  lint         - Run golangci-lint"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  clean        - Clean build artifacts"
	@echo "  all          - Run fmt, vet, test, and bench"
	@echo "  pre-release    - Run all checks before release"
	@echo "  check-release  - Check if ready for release"
	@echo "  generate-version - Generate date-based version"
	@echo "  release        - Create a new release (requires VERSION=YY.MM.DD)"

# Run tests
test:
	go test -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run tests with coverage
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run golangci-lint (requires golangci-lint to be installed)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html

# Run all checks
all: fmt vet test bench

# Install development dependencies
deps:
	go mod download
	go mod tidy

# Pre-release checks
pre-release: fmt vet test bench
	@echo "All pre-release checks passed!"

# Check if ready for release
check-release:
	@echo "Checking release readiness..."
	@git diff --quiet || (echo "❌ Working directory is not clean" && exit 1)
	@git branch --show-current | grep -q "master" || (echo "❌ Not on master branch" && exit 1)
	@go mod verify || (echo "❌ go.mod verification failed" && exit 1)
	@echo "✅ Ready for release!"

# Generate date-based version
generate-version:
	@./scripts/generate-version.sh

# Create a new release
release: check-release
ifndef VERSION
	@echo "❌ VERSION is required. Usage: make release VERSION=25.07.10"
	@exit 1
endif
	@echo "Creating release $(VERSION)..."
	@./scripts/release.sh $(VERSION)
