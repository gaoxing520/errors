#!/bin/bash

# Generate date-based version for errors library
# Usage: ./scripts/generate-version.sh [date]
# If no date provided, uses current date

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
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

# Function to validate date format
validate_date() {
    local date_str=$1
    
    # Check if date matches YY.MM.DD format
    if [[ ! $date_str =~ ^[0-9]{2}\.[0-9]{2}\.[0-9]{2}$ ]]; then
        return 1
    fi
    
    # Extract components
    local year=${date_str:0:2}
    local month=${date_str:3:2}
    local day=${date_str:6:2}
    
    # Validate ranges
    if (( month < 1 || month > 12 )); then
        return 1
    fi
    
    if (( day < 1 || day > 31 )); then
        return 1
    fi
    
    return 0
}

# Get date input
if [ $# -eq 0 ]; then
    # Use current date
    VERSION=$(date +"%y.%m.%d")
    print_info "Using current date: $VERSION"
else
    # Use provided date
    INPUT_DATE=$1
    
    # If input is in YYYY-MM-DD format, convert to YY.MM.DD
    if [[ $INPUT_DATE =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
        YEAR=${INPUT_DATE:2:2}
        MONTH=${INPUT_DATE:5:2}
        DAY=${INPUT_DATE:8:2}
        VERSION="$YEAR.$MONTH.$DAY"
        print_info "Converted $INPUT_DATE to $VERSION"
    elif [[ $INPUT_DATE =~ ^[0-9]{2}\.[0-9]{2}\.[0-9]{2}$ ]]; then
        VERSION=$INPUT_DATE
        print_info "Using provided date: $VERSION"
    else
        echo "Error: Invalid date format"
        echo "Supported formats:"
        echo "  YY.MM.DD (e.g., 25.07.10)"
        echo "  YYYY-MM-DD (e.g., 2025-07-10)"
        exit 1
    fi
fi

# Validate the date
if ! validate_date "$VERSION"; then
    echo "Error: Invalid date: $VERSION"
    echo "Please check month (01-12) and day (01-31) values"
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^$VERSION$"; then
    print_warning "Tag $VERSION already exists"
    
    # Suggest alternatives
    echo ""
    echo "Alternatives:"
    echo "  $VERSION-patch1"
    echo "  $VERSION-hotfix"
    echo "  $VERSION-beta1"
    
    read -p "Use $VERSION-patch1? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        VERSION="$VERSION-patch1"
    else
        echo "Please choose a different version or suffix"
        exit 1
    fi
fi

# Display version info
print_success "Generated version: $VERSION"

# Show what this version represents
YEAR_FULL="20${VERSION:0:2}"
MONTH_NAME=$(date -j -f "%m" "${VERSION:3:2}" +"%B" 2>/dev/null || echo "Month ${VERSION:3:2}")
DAY_NUM="${VERSION:6:2}"

echo ""
print_info "Version details:"
echo "  Date: $MONTH_NAME $DAY_NUM, $YEAR_FULL"
echo "  Format: YY.MM.DD"

# Offer to create release
echo ""
read -p "Create release with this version? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_info "Creating release $VERSION..."
    exec ./scripts/release.sh "$VERSION"
else
    print_info "Version generated. To create release later, run:"
    echo "  ./scripts/release.sh $VERSION"
fi
