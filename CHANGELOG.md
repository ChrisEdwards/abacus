# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Initial release of abacus - a TUI viewer for Beads issue tracking.

### Added
- Interactive TUI for browsing Beads issues with tree view and detail panel
- Issue list with filtering and sorting capabilities
- Hierarchical child sorting with cascading status and date prioritization
- Status icons and colors for beads in detail pane lists
- Detail panel showing full issue information including:
  - Design section with implementation notes
  - Acceptance criteria section
  - Dependencies and relationships
- Pre-TUI loading spinner with witty status messaging
- Prefetch all comments at startup to reduce navigation lag
- Auto-refresh capability with configurable intervals
- Manual refresh with 'r' key
- Version management infrastructure with `--version` flag
- Search functionality with ability to filter by bead ID
- Configuration file support with Viper integration
- JSON output mode (later removed in favor of TUI focus)
- Beads CLI version validation with user-friendly error messages
- GoReleaser configuration for multi-platform builds (Linux, macOS, Windows)
- Release automation pipeline with GitHub Actions
- Homebrew tap and formula for easy installation
- Comprehensive user documentation
- LICENSE file (MIT)
- CI/CD pipeline with automated testing
- golangci-lint configuration and Makefile with standard build targets
- Dependabot configuration for automated dependency updates

### Changed
- Restructured codebase into well-architected Go packages (cmd/, internal/ui, internal/graph, internal/config)
- Simplified auto-refresh CLI flags to single `--auto-refresh-seconds` flag
- Consolidated documentation from docs/ folder into README
- Streamlined user documentation for clarity

### Fixed
- Detail pane header no longer starts scrolled off after changing selection
- Detail pane title wrapping for long bead IDs
- Tree scrolling when selection goes off screen
- Word wrapping throughout the UI
- Detail pane spacing and indentation consistency
- Search filter behavior with tree expand/collapse
- ESC key now properly clears search criteria
- Tab key properly switches keyboard focus to detail pane
- Bead count and filter highlight accuracy when searching by ID
- Tree End key no longer panics on empty list
- `--db-path` flag now properly honored
- Cursor panic prevention after filtering
- Startup errors now shown before clearing screen
- Comment loading with retry after fetch errors
- Viewport dimension clamping to prevent rendering issues

### Removed
- Unused `--json-output` CLI flag (consolidated into main JSON mode)
- docs/ folder (consolidated into README)
