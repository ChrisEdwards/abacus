# Release Process Guide

This document describes the process for creating and publishing new releases of abacus.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Release Checklist](#release-checklist)
- [Step-by-Step Release Process](#step-by-step-release-process)
- [Post-Release Verification](#post-release-verification)
- [Hotfix Releases](#hotfix-releases)
- [Rollback Procedure](#rollback-procedure)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

- **Git**: Version control
- **Go**: 1.23 or later (for local testing)
- **GoReleaser**: Installed locally for testing (optional but recommended)
  ```bash
  brew install --cask goreleaser
  ```

### Required Access

1. **GitHub Repository**: Write access to `ChrisEdwards/abacus`
2. **Homebrew Tap**: Write access configured via `HOMEBREW_TAP_TOKEN` secret
3. **Git Configuration**: Properly configured user name and email

### Verify Setup

```bash
# Check git config
git config user.name
git config user.email

# Verify you're on main branch and up to date
git checkout main
git pull origin main

# Verify working directory is clean
git status

# Check GoReleaser (optional)
goreleaser check
```

## Release Checklist

Before starting a release, ensure:

- [ ] All intended features/fixes are merged to `main`
- [ ] All tests pass: `go test ./...`
- [ ] All builds succeed: `make build`
- [ ] CHANGELOG.md `[Unreleased]` section contains all changes
- [ ] Working directory is clean (no uncommitted changes)
- [ ] You have reviewed the changes since last release
- [ ] You have determined the correct version number (following [Semantic Versioning](https://semver.org/))

## Step-by-Step Release Process

### 1. Update CHANGELOG

Edit `CHANGELOG.md` to prepare for release:

```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD

### Added
- New feature descriptions

### Changed
- Changed behavior descriptions

### Fixed
- Bug fix descriptions
```

**Important**: The bump-version script will handle moving the Unreleased header and updating version links, but you should ensure all changes are documented first.

### 2. Run Version Bump Script

The version bump script automates version updates, commit creation, tagging, and pushing.

**Preview changes (dry run - default):**
```bash
./scripts/bump-version.sh X.Y.Z
```

**Execute the release:**
```bash
# Option 1: Just update files and commit
./scripts/bump-version.sh X.Y.Z --execute --commit

# Option 2: Update, commit, and tag
./scripts/bump-version.sh X.Y.Z --execute --tag

# Option 3: Full release (update, commit, tag, and push)
./scripts/bump-version.sh X.Y.Z --execute --push
```

**What the script does:**
- Updates CHANGELOG.md with release date and version links
- Creates a git commit with version changes
- (Optional) Creates an annotated git tag `vX.Y.Z`
- (Optional) Pushes commit and tag to GitHub

### 3. Monitor GitHub Actions

Once the tag is pushed, the release workflow automatically triggers:

1. **Go to GitHub Actions**:
   https://github.com/ChrisEdwards/abacus/actions

2. **Watch the `Release` workflow**:
   - **GoReleaser job**: Builds binaries for all platforms
   - **Update Homebrew job**: Updates the Homebrew formula

3. **Expected duration**: 3-5 minutes

### 4. Verify GitHub Release

After the workflow completes:

1. **Check the release page**:
   https://github.com/ChrisEdwards/abacus/releases

2. **Verify release assets**:
   - `abacus_X.Y.Z_darwin_amd64.tar.gz`
   - `abacus_X.Y.Z_darwin_arm64.tar.gz`
   - `abacus_X.Y.Z_linux_amd64.tar.gz`
   - `abacus_X.Y.Z_linux_arm64.tar.gz`
   - `abacus_X.Y.Z_windows_amd64.tar.gz`
   - `checksums.txt`

3. **Verify release notes**: Should include CHANGELOG entries

### 5. Verify Homebrew Formula

The GitHub Actions workflow automatically updates the Homebrew formula:

1. **Check the tap repository**:
   https://github.com/ChrisEdwards/homebrew-tap

2. **Verify Formula/abacus.rb** was updated:
   - Version number matches
   - SHA256 checksums are present (not placeholders)
   - URLs point to correct release

3. **Test Homebrew installation** (optional but recommended):
   ```bash
   # Uninstall old version if present
   brew uninstall ChrisEdwards/tap/abacus 2>/dev/null || true

   # Install new version
   brew install ChrisEdwards/tap/abacus

   # Verify version
   abacus --version
   ```

## Post-Release Verification

### Test Installation Methods

#### 1. Homebrew (macOS/Linux)
```bash
brew uninstall ChrisEdwards/tap/abacus 2>/dev/null || true
brew install ChrisEdwards/tap/abacus
abacus --version
```

#### 2. Direct Download
```bash
# Download from GitHub Releases
curl -LO https://github.com/ChrisEdwards/abacus/releases/download/vX.Y.Z/abacus_X.Y.Z_darwin_arm64.tar.gz
tar -xzf abacus_X.Y.Z_darwin_arm64.tar.gz
./abacus --version
```

#### 3. Go Install
```bash
go install github.com/ChrisEdwards/abacus/cmd/abacus@vX.Y.Z
abacus --version
```

### Verify Version Output

The `--version` flag should display:
```
abacus version X.Y.Z (build: <commit-sha>) [<timestamp>]
Go version: go1.25.3
OS/Arch: darwin/arm64
```

## Hotfix Releases

For critical bug fixes that need immediate release:

### 1. Create Hotfix Branch (Optional)
```bash
git checkout -b hotfix/X.Y.Z
```

### 2. Make and Test Fix
```bash
# Make your changes
# Run tests
go test ./...
make build
```

### 3. Update CHANGELOG

Add hotfix entry to CHANGELOG.md:
```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD

### Fixed
- Critical bug description
```

### 4. Commit and Merge
```bash
git add .
git commit -m "Fix: critical bug description"
git checkout main
git merge hotfix/X.Y.Z
git push origin main
```

### 5. Release as Normal
Follow the standard release process with the hotfix version number.

## Rollback Procedure

If a release has critical issues:

### Option 1: Delete Release and Tag (Immediate)

**⚠️ Use with caution - this removes the release from public view**

```bash
# Delete remote tag
git push --delete origin vX.Y.Z

# Delete local tag
git tag -d vX.Y.Z

# Delete GitHub Release
# Go to: https://github.com/ChrisEdwards/abacus/releases
# Click on the release → Delete
```

### Option 2: Publish Hotfix Release (Recommended)

Instead of rolling back, publish a new patch release:

```bash
# Example: If v0.2.0 has issues, release v0.2.1
./scripts/bump-version.sh 0.2.1 --execute --push
```

### Revert Homebrew Formula (If Needed)

If the Homebrew formula needs to be reverted:

```bash
# Clone tap
git clone https://github.com/ChrisEdwards/homebrew-tap.git
cd homebrew-tap

# Revert to previous version
git revert HEAD
git push origin main
```

## Troubleshooting

### Release Workflow Fails

**Problem**: GitHub Actions workflow fails during release

**Solutions**:
1. **Check workflow logs**: https://github.com/ChrisEdwards/abacus/actions
2. **Common causes**:
   - GoReleaser configuration error → Fix `.goreleaser.yml`
   - Missing HOMEBREW_TAP_TOKEN → Verify GitHub secret
   - Build failures → Run `go test ./...` and `make build` locally
   - Network issues → Re-run workflow

### Homebrew Formula Not Updated

**Problem**: Formula in tap repository wasn't updated

**Solutions**:
1. **Verify workflow completed**: Check GitHub Actions
2. **Check HOMEBREW_TAP_TOKEN**: Verify secret exists and has correct permissions
3. **Manual update** (if needed):
   ```bash
   # Clone tap
   git clone https://github.com/ChrisEdwards/homebrew-tap.git
   cd homebrew-tap

   # Edit Formula/abacus.rb manually
   # Update version and checksums

   # Commit and push
   git add Formula/abacus.rb
   git commit -m "Update abacus to vX.Y.Z"
   git push origin main
   ```

### Wrong Version in Binaries

**Problem**: Built binaries show wrong version with `--version`

**Solutions**:
1. **Check GoReleaser config**: Verify ldflags in `.goreleaser.yml`
2. **Verify tag format**: Must be `vX.Y.Z` (with 'v' prefix)
3. **Check main.go**: Ensure version variables exist

### Checksums Don't Match

**Problem**: SHA256 checksums in formula don't match downloads

**Solutions**:
1. **Re-download assets**: GitHub may have updated them
2. **Regenerate checksums**:
   ```bash
   # Download the archive
   curl -LO https://github.com/ChrisEdwards/abacus/releases/download/vX.Y.Z/abacus_X.Y.Z_darwin_arm64.tar.gz

   # Calculate SHA256
   shasum -a 256 abacus_X.Y.Z_darwin_arm64.tar.gz
   ```
3. **Update formula manually** if needed

### Homebrew Install Fails

**Problem**: `brew install` fails with error

**Solutions**:
1. **Update Homebrew**: `brew update`
2. **Check formula syntax**:
   ```bash
   brew install --formula Formula/abacus.rb
   ```
3. **Verify URLs are accessible**:
   ```bash
   curl -I https://github.com/ChrisEdwards/abacus/releases/download/vX.Y.Z/abacus_X.Y.Z_darwin_arm64.tar.gz
   ```
4. **Clear Homebrew cache**: `brew cleanup`

### Tag Already Exists

**Problem**: Cannot create tag because it already exists

**Solutions**:
1. **Use a different version number**: Increment patch version
2. **Delete existing tag** (if it was a mistake):
   ```bash
   git push --delete origin vX.Y.Z
   git tag -d vX.Y.Z
   ```

## Support

For additional help:
- **GitHub Issues**: https://github.com/ChrisEdwards/abacus/issues
- **GitHub Actions Logs**: https://github.com/ChrisEdwards/abacus/actions
- **Homebrew Tap**: https://github.com/ChrisEdwards/homebrew-tap

## Version History

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.
