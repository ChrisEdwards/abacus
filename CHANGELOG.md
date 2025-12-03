# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Exit summary displayed when quitting: shows session duration and bead stats with change deltas highlighted (ab-0hc)
- Surface layering regression guardrails: Dracula/Solarized/Nord golden snapshots, App.View reset integration test, and documented perf benchmark notes (ab-smg0)

## [0.3.0] - 2025-11-28

### Added
- Help screen overlay with `?` key showing all keyboard shortcuts (ab-0nv)
- Copy bead ID to clipboard with `c` key, shows success toast with 5-second countdown (ab-ftk)

### Changed
- Footer redesigned with pill-style key hints, context-sensitive keys, and cleaner layout (ab-jbb)
- Refactored all key handling to use idiomatic `key.Matches()` pattern (ab-zbn)

### Fixed
- Pressing ESC to clear search filter now preserves current selection instead of jumping to first item (ab-7pt)

## [0.2.0] - 2025-11-27

### Added
- Multi-parent tree display: issues with multiple parents now appear under all parent epics (ab-k2o)
- Notes section in detail pane showing implementation notes (ab-k7a)
- Related and Discovered-From relationship sections in detail pane (ab-749)
- Error toast overlay when background refresh fails (ab-9sl)
- Startup progress indicators with helpful status messages (ab-cbf)

### Changed
- **Requires Beads CLI 0.25.0+**: needed for dependency_type field to correctly display parent/child relationships and other relationship types (ab-e0v)
- Detail pane relationship sections renamed for clarity (Dependencies → Blocked By, Dependents split by type)
- Blocked items now use lighter red (203) for better visibility
- Sibling highlight for multi-parent nodes is now more visible (ab-8ld)

### Fixed
- Expanding/collapsing a multi-parent node now only affects the selected instance, not all instances (ab-vue)
- Detail pane now shows all blockers, not just open ones
- Dependency and Dependent JSON parsing now correctly maps API response fields (ab-0g1)
- Dependents are now filtered by type to prevent incorrect parent relationships

## [0.1.0] - 2025-11-23

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
