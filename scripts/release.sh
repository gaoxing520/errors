#!/bin/bash

# Release script for errors library
# Usage: ./scripts/release.sh [version]
# Example: ./scripts/release.sh v1.0.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version is provided
if [ $# -eq 0 ]; then
    print_error "Version is required"
    echo "Usage: $0 <version>"
    echo "Example: $0 25.07.10"
    exit 1
fi

VERSION=$1

# Validate version format (YY.MM.DD or YY.MM.DD-suffix)
if [[ ! $VERSION =~ ^[0-9]{2}\.[0-9]{2}\.[0-9]{2}(-[a-zA-Z0-9]+)?$ ]]; then
    print_error "Invalid version format. Expected format: YY.MM.DD or YY.MM.DD-suffix"
    echo "Examples: 25.07.10, 25.12.31, 26.01.15-beta1"
    exit 1
fi

print_info "Starting release process for version: $VERSION"

# Check if we're on master branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "master" ]; then
    print_warning "You are not on master branch (current: $CURRENT_BRANCH)"
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Release cancelled"
        exit 0
    fi
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    print_error "Working directory is not clean. Please commit or stash your changes."
    git status --short
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^$VERSION$"; then
    print_error "Tag $VERSION already exists"
    exit 1
fi

# Pull latest changes
print_info "Pulling latest changes..."
git pull origin master

# Run tests
print_info "Running tests..."
go test -v ./...
if [ $? -ne 0 ]; then
    print_error "Tests failed"
    exit 1
fi

# Run benchmarks
print_info "Running benchmarks..."
go test -bench=. -benchmem ./...

# Check go mod
print_info "Verifying go.mod..."
go mod tidy
go mod verify

# Check if there are any changes after go mod tidy
if [ -n "$(git status --porcelain)" ]; then
    print_warning "go mod tidy made changes. Please review and commit them first."
    git status --short
    exit 1
fi

# Create and push tag
print_info "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

print_info "Pushing tag to origin..."
git push origin "$VERSION"

print_success "Release $VERSION has been created and pushed!"
print_info "GitHub Actions will automatically create the release."
print_info "You can monitor the progress at: https://github.com/gaoxing520/errors/actions"

# Optional: Open the releases page
if command -v open >/dev/null 2>&1; then
    read -p "Open releases page in browser? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open "https://github.com/gaoxing520/errors/releases"
    fi
elif command -v xdg-open >/dev/null 2>&1; then
    read -p "Open releases page in browser? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        xdg-open "https://github.com/gaoxing520/errors/releases"
    fi
fi

print_success "Release process completed!"
