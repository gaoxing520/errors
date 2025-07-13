# Release Guide

This document describes how to create releases for the errors library.

## Release Process

### Prerequisites

1. Ensure you're on the `master` branch
2. Working directory is clean (no uncommitted changes)
3. All tests pass
4. Go modules are verified

### Automated Release (Recommended)

Use the provided release script for a streamlined process:

```bash
# Make sure the script is executable
chmod +x scripts/release.sh

# Create a release
./scripts/release.sh 25.07.10
```

Or use the Makefile:

```bash
make release VERSION=25.07.10
```

### Manual Release

If you prefer to create releases manually:

1. **Run pre-release checks:**
   ```bash
   make pre-release
   ```

2. **Update CHANGELOG.md:**
    - Move items from `[Unreleased]` to a new version section
    - Add release date
    - Create new `[Unreleased]` section

3. **Create and push tag:**
   ```bash
   git tag -a 25.07.10 -m "Release 25.07.10"
   git push origin 25.07.10
   ```

4. **GitHub Actions will automatically:**
    - Run tests
    - Create GitHub release
    - Generate release notes
    - Verify module can be imported

## Version Numbering

This project uses **date-based versioning** in the format `YY.MM.DD`:

- **YY** - Two-digit year (e.g., 25 for 2025)
- **MM** - Two-digit month (01-12)
- **DD** - Two-digit day (01-31)

### Examples:

- `25.07.10` - Release on July 10, 2025
- `25.12.31` - Release on December 31, 2025
- `26.01.15` - Release on January 15, 2026
- `25.07.10-beta1` - Pre-release on July 10, 2025

## Release Types

### Stable Release

```bash
./scripts/release.sh 25.07.10
```

### Pre-release

```bash
./scripts/release.sh 25.07.10-beta1
./scripts/release.sh 25.07.10-rc1
```

## What Happens During Release

1. **Validation:**
    - Check branch is master
    - Verify working directory is clean
    - Validate version format
    - Check if tag already exists

2. **Testing:**
    - Run full test suite
    - Run benchmarks
    - Verify go.mod

3. **Tagging:**
    - Create annotated git tag
    - Push tag to origin

4. **GitHub Actions (Automatic):**
    - Run CI tests
    - Generate changelog from commits
    - Create GitHub release
    - Test module import

## Troubleshooting

### Common Issues

**"Working directory is not clean"**

```bash
git status
git add .
git commit -m "Prepare for release"
```

**"Not on master branch"**

```bash
git checkout master
git pull origin master
```

**"Tag already exists"**

```bash
# Delete local tag
git tag -d 25.07.10
# Delete remote tag (if needed)
git push origin :refs/tags/25.07.10
```

**"Tests failed"**

```bash
go test -v ./...
# Fix failing tests and try again
```

### Manual Cleanup

If something goes wrong during release:

```bash
# Delete local tag
git tag -d 25.07.10

# Delete remote tag
git push origin :refs/tags/25.07.10

# Delete GitHub release (manual via web interface)
```

## Post-Release

After a successful release:

1. **Verify the release:**
    - Check GitHub releases page
    - Test importing the new version: `go get github.com/gaoxing520/errors@25.07.10`

2. **Update documentation:**
    - Update README.md if needed
    - Update examples with new version

3. **Announce:**
    - Update any dependent projects
    - Announce in relevant channels

## Release Checklist

- [ ] All tests pass
- [ ] Benchmarks run successfully
- [ ] CHANGELOG.md is updated
- [ ] Version follows semantic versioning
- [ ] On master branch
- [ ] Working directory is clean
- [ ] Tag doesn't already exist
- [ ] GitHub Actions workflow completes successfully
- [ ] Release appears on GitHub releases page
- [ ] Module can be imported with new version
