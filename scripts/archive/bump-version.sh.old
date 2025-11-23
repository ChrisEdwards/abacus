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
DO_COMMIT=false
DO_TAG=false
DO_PUSH=false
VERSION=""

# Script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
    cat <<EOF
Usage: $0 <version> [options]

Bump version across all relevant files in the repository.

Arguments:
  <version>         Version number in X.Y.Z format (e.g., 0.1.0)

Options:
  --commit          Create a git commit with the changes
  --tag             Create an annotated git tag (implies --commit)
  --push            Push commit and tag to remote (implies --tag --commit)
  --dry-run         Show what would change without modifying files (default)
  --execute         Actually perform the changes (opposite of --dry-run)
  -h, --help        Show this help message

Examples:
  $0 0.1.0                           # Dry run, show what would change
  $0 0.1.0 --execute                 # Update files only
  $0 0.1.0 --execute --commit        # Update files and commit
  $0 0.1.0 --execute --tag           # Update files, commit, and tag
  $0 0.1.0 --execute --push          # Update files, commit, tag, and push

Safety:
  - By default, runs in dry-run mode
  - Checks that working directory is clean before making changes
  - Validates version format
  - Prevents duplicate versions (checks for existing tags)
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

check_tag_exists() {
    local version="$1"
    local tag="v$version"

    if git rev-parse "$tag" >/dev/null 2>&1; then
        log_error "Tag $tag already exists"
        return 1
    fi
    return 0
}

update_file() {
    local file="$1"
    local pattern="$2"
    local replacement="$3"
    local description="$4"

    if [ ! -f "$file" ]; then
        log_warning "File not found: $file (skipping)"
        return 0
    fi

    if [ "$DRY_RUN" = true ]; then
        log_info "Would update $file: $description"
        # Show what would change
        if grep -q "$pattern" "$file" 2>/dev/null; then
            echo "  Current: $(grep "$pattern" "$file" | head -1)"
            echo "  New:     $replacement"
        fi
    else
        log_info "Updating $file: $description"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            sed -i '' "s|$pattern|$replacement|g" "$file"
        else
            # Linux
            sed -i "s|$pattern|$replacement|g" "$file"
        fi
        log_success "Updated $file"
    fi
}

bump_version() {
    local version="$1"

    cd "$REPO_ROOT"

    log_info "Bumping version to $version"
    echo ""

    # Update CHANGELOG.md - Update Unreleased section
    update_file "CHANGELOG.md" \
        "## \[Unreleased\]" \
        "## [Unreleased]\\\n\\\n## [$version] - $(date +%Y-%m-%d)" \
        "Add release entry"

    # Update CHANGELOG.md - Update version links
    if [ "$DRY_RUN" = false ]; then
        if grep -q "\[Unreleased\]: " CHANGELOG.md; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s|\[Unreleased\]: .*|\[Unreleased\]: https://github.com/ChrisEdwards/abacus/compare/v$version...HEAD\n[$version]: https://github.com/ChrisEdwards/abacus/releases/tag/v$version|" CHANGELOG.md
            else
                sed -i "s|\[Unreleased\]: .*|\[Unreleased\]: https://github.com/ChrisEdwards/abacus/compare/v$version...HEAD\n[$version]: https://github.com/ChrisEdwards/abacus/releases/tag/v$version|" CHANGELOG.md
            fi
        fi
    fi

    echo ""
    log_success "Version bump to $version complete"
}

create_commit() {
    local version="$1"

    if [ "$DRY_RUN" = true ]; then
        log_info "Would create git commit"
        echo "  Message: Bump version to $version"
        echo "  Files: CHANGELOG.md"
        return 0
    fi

    log_info "Creating git commit..."
    git add CHANGELOG.md
    git commit -m "Bump version to $version

Prepare release v$version"
    log_success "Created commit"
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

push_changes() {
    local version="$1"
    local tag="v$version"

    if [ "$DRY_RUN" = true ]; then
        log_info "Would push commit and tag to remote"
        return 0
    fi

    log_info "Pushing commit and tag to remote..."
    git push origin main
    git push origin "$tag"
    log_success "Pushed changes to remote"
}

main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            --commit)
                DO_COMMIT=true
                shift
                ;;
            --tag)
                DO_TAG=true
                DO_COMMIT=true
                shift
                ;;
            --push)
                DO_PUSH=true
                DO_TAG=true
                DO_COMMIT=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --execute)
                DRY_RUN=false
                shift
                ;;
            *)
                if [ -z "$VERSION" ]; then
                    VERSION="$1"
                else
                    log_error "Unknown option: $1"
                    usage
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate version argument
    if [ -z "$VERSION" ]; then
        log_error "Version argument is required"
        usage
        exit 1
    fi

    # Validate version format
    if ! validate_version "$VERSION"; then
        exit 1
    fi

    # Header
    echo ""
    echo "═══════════════════════════════════════"
    echo "  Abacus Version Bump"
    echo "═══════════════════════════════════════"
    echo ""
    echo "Version:  $VERSION"
    echo "Mode:     $([ "$DRY_RUN" = true ] && echo "DRY RUN" || echo "EXECUTE")"
    echo "Commit:   $([ "$DO_COMMIT" = true ] && echo "YES" || echo "NO")"
    echo "Tag:      $([ "$DO_TAG" = true ] && echo "YES" || echo "NO")"
    echo "Push:     $([ "$DO_PUSH" = true ] && echo "YES" || echo "NO")"
    echo ""
    echo "═══════════════════════════════════════"
    echo ""

    # Safety checks
    if ! check_git_clean; then
        exit 1
    fi

    if ! check_tag_exists "$VERSION"; then
        exit 1
    fi

    # Bump version in files
    bump_version "$VERSION"

    # Git operations
    if [ "$DO_COMMIT" = true ]; then
        echo ""
        create_commit "$VERSION"
    fi

    if [ "$DO_TAG" = true ]; then
        echo ""
        create_tag "$VERSION"
    fi

    if [ "$DO_PUSH" = true ]; then
        echo ""
        push_changes "$VERSION"
    fi

    # Summary
    echo ""
    echo "═══════════════════════════════════════"
    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN - No changes were made"
        echo ""
        log_info "To execute these changes, run:"
        echo "  $0 $VERSION --execute"
    else
        log_success "Version bump complete!"
        if [ "$DO_PUSH" = true ]; then
            echo ""
            log_info "Release workflow triggered"
            log_info "Check: https://github.com/ChrisEdwards/abacus/actions"
        elif [ "$DO_TAG" = true ]; then
            echo ""
            log_info "To trigger the release, push the tag:"
            echo "  git push origin v$VERSION"
        fi
    fi
    echo "═══════════════════════════════════════"
    echo ""
}

main "$@"
