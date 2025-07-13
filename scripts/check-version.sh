#!/bin/bash

# Version checking utility for errors library
# Usage: ./scripts/check-version.sh [version]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to compare date-based versions (YY.MM.DD)
version_compare() {
    local v1=$1
    local v2=$2

    # Split versions into arrays
    IFS='.' read -ra V1 <<< "$v1"
    IFS='.' read -ra V2 <<< "$v2"

    # Compare year, month, day
    for i in {0..2}; do
        local num1=${V1[i]:-0}
        local num2=${V2[i]:-0}

        # Remove leading zeros for comparison
        num1=$((10#$num1))
        num2=$((10#$num2))

        if (( num1 > num2 )); then
            return 1  # v1 > v2
        elif (( num1 < num2 )); then
            return 2  # v1 < v2
        fi
    done

    return 0  # v1 == v2
}

# Get current version info
print_info "Checking version information..."

# Get latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
print_info "Latest tag: $LATEST_TAG"

# Get current branch
CURRENT_BRANCH=$(git branch --show-current)
print_info "Current branch: $CURRENT_BRANCH"

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    print_warning "Working directory has uncommitted changes"
else
    print_success "Working directory is clean"
fi

# Count commits since last tag
if [ "$LATEST_TAG" != "none" ]; then
    COMMITS_SINCE=$(git rev-list ${LATEST_TAG}..HEAD --count)
    print_info "Commits since $LATEST_TAG: $COMMITS_SINCE"
    
    if [ "$COMMITS_SINCE" -gt 0 ]; then
        print_info "Recent commits:"
        git log --oneline ${LATEST_TAG}..HEAD | head -5
    fi
else
    TOTAL_COMMITS=$(git rev-list --count HEAD)
    print_info "Total commits: $TOTAL_COMMITS (no tags found)"
fi

# If version is provided, validate it
if [ $# -gt 0 ]; then
    NEW_VERSION=$1
    print_info "Validating proposed version: $NEW_VERSION"
    
    # Validate format (YY.MM.DD)
    if [[ ! $NEW_VERSION =~ ^[0-9]{2}\.[0-9]{2}\.[0-9]{2}(-[a-zA-Z0-9]+)?$ ]]; then
        print_error "Invalid version format: $NEW_VERSION"
        echo "Expected format: YY.MM.DD or YY.MM.DD-suffix"
        exit 1
    fi
    
    # Check if tag already exists
    if git tag -l | grep -q "^$NEW_VERSION$"; then
        print_error "Tag $NEW_VERSION already exists"
        exit 1
    fi
    
    # Compare with latest version
    if [ "$LATEST_TAG" != "none" ]; then
        version_compare "$NEW_VERSION" "$LATEST_TAG"
        case $? in
            0)
                print_warning "Version $NEW_VERSION is the same as latest tag $LATEST_TAG"
                ;;
            1)
                print_success "Version $NEW_VERSION is newer than $LATEST_TAG"
                ;;
            2)
                print_error "Version $NEW_VERSION is older than latest tag $LATEST_TAG"
                exit 1
                ;;
        esac
    fi
    
    print_success "Version $NEW_VERSION is valid"
fi

# Suggest next version based on current date
if [ "$LATEST_TAG" != "none" ]; then
    # Parse current version (YY.MM.DD)
    IFS='.' read -ra VERSION_PARTS <<< "$LATEST_TAG"

    YEAR=${VERSION_PARTS[0]}
    MONTH=${VERSION_PARTS[1]}
    DAY=${VERSION_PARTS[2]}

    # Get current date
    CURRENT_DATE=$(date +"%y.%m.%d")
    TODAY_YEAR=$(date +"%y")
    TODAY_MONTH=$(date +"%m")
    TODAY_DAY=$(date +"%d")

    # Suggest next versions
    echo ""
    print_info "Suggested next versions:"
    echo "  Today: $CURRENT_DATE"

    # Suggest next day if today is same as latest
    if [ "$CURRENT_DATE" == "$LATEST_TAG" ]; then
        NEXT_DAY=$(date -j -v+1d -f "%y.%m.%d" "$LATEST_TAG" +"%y.%m.%d" 2>/dev/null || echo "Next day calculation failed")
        if [ "$NEXT_DAY" != "Next day calculation failed" ]; then
            echo "  Next day: $NEXT_DAY"
        fi
    fi

    # Show some manual increment examples
    echo "  Manual increment examples:"
    echo "    Same day patch: $YEAR.$MONTH.$DAY-patch1"
    echo "    Next day: $TODAY_YEAR.$TODAY_MONTH.$TODAY_DAY"
fi

# Check Go module version
print_info "Checking Go module..."
go list -m github.com/gaoxing520/errors 2>/dev/null || print_warning "Module not in current directory"

print_success "Version check completed"
