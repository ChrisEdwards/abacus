#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DRY_RUN=true
DO_PUSH=true
VERSION=""

# Script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
    cat <<EOF
Usage: $0 [options]

Automated release script that orchestrates the complete release process.

Options:
  --execute         Actually perform the release (default is dry-run)
  --dry-run         Show what would happen without making changes (default)
  --no-push         Create commit and tag but don't push to remote
  -h, --help        Show this help message

Examples:
  $0                      # Dry run, show what would happen
  $0 --execute            # Execute the full release
  $0 --execute --no-push  # Execute but don't push (for testing)

Prerequisites:
  1. Run '/update-changelog' to prepare CHANGELOG.md and next-version.txt
  2. Review and commit those changes
  3. Ensure all tests pass
  4. Ensure working directory is clean

The release process:
  1. Read version from next-version.txt
  2. Run pre-flight checks (tests, clean git, etc.)
  3. Finalize CHANGELOG.md ([Unreleased] → [X.Y.Z] - date)
  4. Create release commit and tag
  5. Push to trigger CI/CD pipeline
  6. Auto-bump next-version.txt to X.Y.(Z+1)
  7. Commit and push the version bump

Safety:
  - By default, runs in dry-run mode
  - Checks that working directory is clean
  - Validates all prerequisites before making changes
  - Confirms with user before pushing
EOF
}

log_info() {
    echo -e "${BLUE}ℹ${NC} $*"
}

log_success() {
    echo -e "${GREEN}✓${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $*"
}

log_error() {
    echo -e "${RED}✗${NC} $*" >&2
}

validate_version() {
    local version="$1"
    if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format: $version"
        log_error "Version must be in X.Y.Z format (e.g., 0.1.0)"
        return 1
    fi
    return 0
}

check_git_clean() {
    if [ "$DRY_RUN" = true ]; then
        log_info "Would check git working directory is clean"
        return 0
    fi

    if ! git diff-index --quiet HEAD -- 2>/dev/null; then
        log_error "Working directory is not clean"
        log_error "Please commit or stash your changes first"
        git status --short
        return 1
    fi
    return 0
}

check_main_branch() {
    local current_branch=$(git branch --show-current)
    if [ "$current_branch" != "main" ]; then
        log_warning "You are not on the 'main' branch (current: $current_branch)"
        if [ "$DRY_RUN" = false ]; then
            read -p "Continue anyway? (y/n) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                return 1
            fi
        fi
    fi
    return 0
}

check_tag_exists() {
    local version="$1"
    local tag="v$version"

    if git rev-parse "$tag" >/dev/null 2>&1; then
        log_error "Tag $tag already exists"
        return 1
    fi
    return 0
}

check_unreleased_content() {
    if ! grep -A 5 "## \[Unreleased\]" "$REPO_ROOT/CHANGELOG.md" | grep -q "^### "; then
        log_error "CHANGELOG.md [Unreleased] section appears empty"
        log_error ""
        log_error "Please run '/update-changelog' to generate changelog entries,"
        log_error "or manually add entries to the [Unreleased] section."
        return 1
    fi
    return 0
}

run_tests() {
    if [ "$DRY_RUN" = true ]; then
        log_info "Would run: go test ./..."
        return 0
    fi

    log_info "Running tests..."
    if go test ./... > /tmp/abacus-test-output.log 2>&1; then
        log_success "All tests passed"
        return 0
    else
        log_error "Tests failed"
        cat /tmp/abacus-test-output.log
        return 1
    fi
}

run_build() {
    if [ "$DRY_RUN" = true ]; then
        log_info "Would run: go build ./cmd/abacus"
        return 0
    fi

    log_info "Building..."
    if go build ./cmd/abacus > /tmp/abacus-build-output.log 2>&1; then
        log_success "Build successful"
        return 0
    else
        log_error "Build failed"
        cat /tmp/abacus-build-output.log
        return 1
    fi
}

bump_patch_version() {
    local version="$1"
    local major minor patch
    IFS='.' read -r major minor patch <<< "$version"
    patch=$((patch + 1))
    echo "$major.$minor.$patch"
}

finalize_changelog() {
    local version="$1"
    local date=$(date +%Y-%m-%d)

    if [ "$DRY_RUN" = true ]; then
        log_info "Would update CHANGELOG.md:"
        echo "  - Replace [Unreleased] with [Unreleased] + [$version] - $date"
        echo "  - Update version comparison links"
        return 0
    fi

    log_info "Finalizing CHANGELOG.md..."

    # Update CHANGELOG.md - Add new Unreleased section and version
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|## \[Unreleased\]|## [Unreleased]\n\n## [$version] - $date|" "$REPO_ROOT/CHANGELOG.md"
    else
        # Linux
        sed -i "s|## \[Unreleased\]|## [Unreleased]\n\n## [$version] - $date|" "$REPO_ROOT/CHANGELOG.md"
    fi

    # Update version comparison links
    if grep -q "\[Unreleased\]: " "$REPO_ROOT/CHANGELOG.md"; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|\[Unreleased\]: .*|\[Unreleased\]: https://github.com/ChrisEdwards/abacus/compare/v$version...HEAD\n[$version]: https://github.com/ChrisEdwards/abacus/releases/tag/v$version|" "$REPO_ROOT/CHANGELOG.md"
        else
            sed -i "s|\[Unreleased\]: .*|\[Unreleased\]: https://github.com/ChrisEdwards/abacus/compare/v$version...HEAD\n[$version]: https://github.com/ChrisEdwards/abacus/releases/tag/v$version|" "$REPO_ROOT/CHANGELOG.md"
        fi
    fi

    log_success "CHANGELOG.md finalized"
}

create_release_commit() {
    local version="$1"

    if [ "$DRY_RUN" = true ]; then
        log_info "Would create git commit:"
        echo "  Message: Release v$version"
        echo "  Files: CHANGELOG.md, next-version.txt"
        return 0
    fi

    log_info "Creating release commit..."
    git add CHANGELOG.md next-version.txt
    git commit -m "Release v$version

Prepare release v$version with finalized changelog."
    log_success "Created release commit"
}

create_tag() {
    local version="$1"
    local tag="v$version"

    if [ "$DRY_RUN" = true ]; then
        log_info "Would create annotated git tag: $tag"
        return 0
    fi

    log_info "Creating annotated git tag: $tag..."
    git tag -a "$tag" -m "Release $tag"
    log_success "Created tag $tag"
}

push_release() {
    local version="$1"
    local tag="v$version"

    if [ "$DRY_RUN" = true ]; then
        log_info "Would push commit and tag to remote"
        echo "  git push origin main $tag"
        return 0
    fi

    if [ "$DO_PUSH" = false ]; then
        log_warning "Skipping push (--no-push specified)"
        return 0
    fi

    log_info "Ready to push release to remote..."
    echo ""
    log_warning "This will:"
    echo "  1. Push the release commit to main"
    echo "  2. Push tag $tag"
    echo "  3. Trigger the GitHub Actions release workflow"
    echo ""
    read -p "Continue? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Push cancelled by user"
        log_info "To push manually later:"
        echo "  git push origin main $tag"
        return 1
    fi

    log_info "Pushing commit and tag to remote..."
    if git push origin main "$tag"; then
        log_success "Pushed changes to remote"
        return 0
    else
        log_error "Push failed!"
        log_error ""
        log_error "To retry:"
        echo "  git push origin main $tag"
        log_error ""
        log_error "To rollback (if needed):"
        echo "  git tag -d $tag"
        echo "  git reset --hard HEAD~1"
        return 1
    fi
}

update_next_version() {
    local current_version="$1"
    local next_version=$(bump_patch_version "$current_version")

    if [ "$DRY_RUN" = true ]; then
        log_info "Would update next-version.txt:"
        echo "  $current_version → $next_version"
        return 0
    fi

    log_info "Updating next-version.txt to $next_version..."
    echo "$next_version" > "$REPO_ROOT/next-version.txt"

    git add next-version.txt
    git commit -m "Prepare for next release ($next_version)"

    if [ "$DO_PUSH" = true ]; then
        git push origin main
    fi

    log_success "Updated next-version.txt to $next_version"
}

main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --execute)
                DRY_RUN=false
                shift
                ;;
            --no-push)
                DO_PUSH=false
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    cd "$REPO_ROOT"

    # Read version from next-version.txt
    if [ ! -f "next-version.txt" ]; then
        log_error "next-version.txt not found"
        log_error "Please run '/update-changelog' first to prepare the release"
        exit 1
    fi

    VERSION=$(cat next-version.txt | tr -d '[:space:]')

    # Validate version format
    if ! validate_version "$VERSION"; then
        exit 1
    fi

    # Header
    echo ""
    echo "═══════════════════════════════════════"
    echo "  Abacus Release"
    echo "═══════════════════════════════════════"
    echo ""
    echo "Version:  $VERSION"
    echo "Mode:     $([ "$DRY_RUN" = true ] && echo "DRY RUN" || echo "EXECUTE")"
    echo "Push:     $([ "$DO_PUSH" = true ] && echo "YES" || echo "NO")"
    echo ""
    echo "═══════════════════════════════════════"
    echo ""

    # Pre-flight checks
    log_info "Running pre-flight checks..."
    echo ""

    if ! check_main_branch; then
        exit 1
    fi

    if ! check_git_clean; then
        exit 1
    fi

    if ! check_tag_exists "$VERSION"; then
        exit 1
    fi

    if ! check_unreleased_content; then
        exit 1
    fi

    if ! run_tests; then
        exit 1
    fi

    if ! run_build; then
        exit 1
    fi

    log_success "All pre-flight checks passed"
    echo ""

    # Release process
    log_info "Executing release process..."
    echo ""

    finalize_changelog "$VERSION"
    create_release_commit "$VERSION"
    create_tag "$VERSION"

    if ! push_release "$VERSION"; then
        exit 1
    fi

    # Post-release
    echo ""
    log_info "Preparing for next release..."
    echo ""

    update_next_version "$VERSION"

    # Summary
    echo ""
    echo "═══════════════════════════════════════"
    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN - No changes were made"
        echo ""
        log_info "To execute the release, run:"
        echo "  $0 --execute"
    else
        log_success "Release complete!"
        echo ""
        log_info "Release workflow triggered:"
        echo "  https://github.com/ChrisEdwards/abacus/actions"
        echo ""
        log_info "Release will be available at:"
        echo "  https://github.com/ChrisEdwards/abacus/releases/tag/v$VERSION"
        echo ""
        log_info "Next version prepared: $(cat next-version.txt)"
    fi
    echo "═══════════════════════════════════════"
    echo ""
}

main "$@"
