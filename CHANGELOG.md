# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.8.0] - 2026-01-21

### Added
- **beads_rust (br) backend support**: Abacus now supports both `bd` (Go) and `br` (Rust) beads backends with automatic detection and user selection prompt (ab-pccw)
- **Backend CLI flag**: Use `--backend bd` or `--backend br` to override detection without prompts
- **Backend indicator**: Status bar shows `[bd]` or `[br]` to indicate active backend
- **Backend selection prompt**: First-run prompts users to choose preferred backend when both are available, saved to config
- **Integration test separation**: Tests now use Go build tags to separate unit and integration tests (ab-64oh)
- **Conformance tests**: Verify bd/br output compatibility for consistent TUI behavior

### Changed
- **Client architecture**: Internal refactor splits Client interface into Reader + Writer for cleaner multi-backend support (ab-ynml, ab-nwke, ab-9srf)
- **bd backend frozen**: bd client frozen at v0.38.0 compatibility; br backend recommended for new features

### Fixed
- **Assignee clearing**: Pass `--assignee` to CLI even when empty to allow clearing assignments (ab-e11f)
- **Edit overlay**: Allow converting unknown issue type to 'task' from edit overlay

### Removed
- **--json-output flag**: Removed deprecated flag (ab-obak)

## [0.7.1] - 2026-01-09

### Fixed
- **Description truncation**: Edit bead popup no longer truncates descriptions to 2000 characters; limit increased to 50,000 (ab-0dts)

### Changed
- **Dependencies**: Bump modernc.org/sqlite from 1.41.0 to 1.42.2

## [0.7.0] - 2025-12-26

### Added
- **Auto-update system**: Check for updates at startup with non-blocking GitHub API check, toast notification when updates available, and `U` hotkey for one-key self-update (ab-lanr)
- **Blocked and deferred statuses**: Full support for blocked/deferred in tree view with icons, detail view, status overlay, sorting order, and view filters—blocked/deferred items excluded from Active and Ready views (ab-9sbt, ab-lqny, ab-meoq, ab-gpqt)
- **Tree columns**: New column zone showing priority, last updated time, and comment count; toggle with `Shift+C`; responsive hiding on narrow terminals (ab-flie, ab-bzms, ab-opn0, ab-z0ai)
- **Graph link support**: Display duplicate_of and superseded_by relationships in detail view; handle relates-to dependency type (ab-qdh0)
- **Debug logging**: Add `--debug` flag for troubleshooting (ab-ur7f)
- **Stripe-style time formatting**: Human-friendly "2h ago" notation with consistent styling (ab-xcyg, ab-unj4)

### Changed
- **Textarea submit shortcut**: Comment and description textareas now use `Ctrl+S` to submit, while `Enter` inserts newlines (standard cross-terminal behavior)
- **Slack-style multiline input**: Description field uses Enter for newlines, Shift+Enter to submit (consistent with modern chat apps) (ab-aues, ab-qf7y)
- **Remove bulk create mode**: Removed `Ctrl+Enter` bulk entry mode from create overlay (ab-6yr0)

### Fixed
- **Updater download URL**: Updater now uses DownloadURL discovered by version checker instead of constructing URLs (ab-uld7)
- **Tree sort order**: Blocked and deferred statuses now correctly sort after in_progress but before closed (ab-ndns)
- **Delete overlay formatting**: Fixed bead ID and title formatting in delete confirmation (ab-53rg)
- **Empty database handling**: Graceful error message when no beads database found (ab-3zw.7)
- **Window resize**: Added debounce to prevent excessive redraws during resize (ab-mhto)
- **Status transitions**: Fixed status label formatting and double error toast rendering
- **Debug logging**: debug.Close() now properly disables logging

## [0.6.1] - 2025-12-24

### Changed
- Require Beads CLI 0.30.0 or later (previously 0.25.0) for improved compatibility (ab-pa68)

## [0.6.0] - 2025-12-21

### Added
- **Add comments**: Press `m` to add comments to beads directly from the TUI with a multi-line textarea (ab-d03)
- **SQLite direct read**: Abacus now reads beads directly from the SQLite database instead of spawning CLI processes, significantly reducing overhead (ab-c6o9)
- **Background comment loading**: Comments load asynchronously on startup for faster TUI launch (ab-fkyz)

### Changed
- **Instantaneous tree navigation**: Removed blocking comment fetch so keystrokes never queue up (ab-o0fm)
- **Unified overlay framework**: All overlays now share consistent styling with automatic bright theme support (ab-kfms)
- **Auto-refresh improvements**: Reduced update frequency and improved refresh indicator UX

### Fixed
- **Refresh placeholder styling**: Fixed grey background artifact when refresh indicator appears
- **Comment flicker**: Comments now preserved during refresh to prevent visual flicker
- **Border overflow**: Corrected border width overflow in input components
- **Comment textarea submission**: Comment textarea submission key handling improved

## [0.5.0] - 2025-12-11

### Added
- **Edit bead**: Press `e` to edit existing beads directly from the TUI with pre-populated values (ab-jve)
- **Colorized labels**: Labels now display with theme-aware Info() color for better visibility

### Changed
- **Responsive dialogs**: Create/edit dialogs now adapt to terminal width (44-120 chars based on 70% of terminal width) (ab-11wd)
- **Standardized overlays**: Delete, status, and label overlays now have consistent width and styling (ab-nr58)

### Fixed
- **Epic parent validation**: Epics can now only be children of other epics, with clear error messaging (ab-jve)
- **Auto-refresh**: Now correctly detects WAL file changes and continues refreshing when overlays are open
- **Tab selection**: Tab key in combo boxes now selects the highlighted item before moving focus
- **Combobox highlight**: Dropdown highlight alignment fixed for consistent appearance (ab-3us9)
- **Flaky CI tests**: Fixed timing issues in CLI tests for more reliable CI builds

## [0.4.0] - 2025-12-06

### Added
- **Theme system**: 20+ themes including TokyoNight (now default), Dracula, Nord, Solarized, Catppuccin, Kanagawa, Gruvbox, One Dark, Rose Pine, and more (ab-4a9p, ab-i19v)
- **Theme cycling**: Press `t` to cycle forward, `T` (Shift+t) to cycle backward through themes
- **Theme persistence**: Selected theme is saved to `~/.abacus/config.yaml` and restored on startup
- **View mode filter**: Press `v`/`V` to cycle between All/Active/Ready views, hiding closed issues (ab-gmw4)
- **New bead creation**: Press `n` to create a root bead, `N` (Shift+n) to create a child under the selected bead (ab-ifnc)
- **Status overlay**: Press `s` to quickly change bead status with single-key selection (ab-6s4)
- **Labels overlay**: Press `L` to manage labels with chip-based UI, autocomplete, and inline creation
- **Delete bead**: Press `Del` to delete a bead with confirmation dialog showing C/D hotkeys (ab-6vs)
- **New bead modal redesign**: 5-zone HUD architecture with editable parent, properties grid, labels chips, and assignee autocomplete (ab-3dn)
- **Bulk entry mode**: Press `Ctrl+Enter` in new bead modal to create and add another
- **Type auto-inference**: Modal automatically suggests type based on title keywords (e.g., "Fix..." → Bug)
- **Instant tree injection**: New beads appear in tree in <50ms without full reload
- **Exit summary**: Shows session duration and bead stats with change deltas on quit (ab-0hc)
- Surface layering regression guardrails: Dracula/Solarized/Nord golden snapshots, App.View reset integration test (ab-smg0)

### Changed
- **Default theme**: Changed from Dracula to TokyoNight (ab-i19v)
- **Config location**: Moved from `~/.config/abacus/` to `~/.abacus/` (ab-3d7u)
- **Bead creation hotkeys**: Swapped `n`/`N` - lowercase now creates root, uppercase creates child (ab-ifnc)

### Fixed
- Duplicate bead creation when pressing Enter multiple times quickly (ab-ip2p)
- Labels combo box now selects exact matches over partial matches (ab-qa72)
- Label not being added on Enter in create bead modal (ab-mod2)
- Flaky TestCLIClient_CreateFull_OptionalParameters test (ab-ofmz)
- Auto-refresh now works correctly with modal overlays open (ab-mlg2)
- Backend errors now show as toast over modal instead of inline (ab-orte)
- Tree immediately updates after bead changes instead of waiting for refresh

### Removed
- **Redundant hotkeys**: Removed `i` (start work) and `x` (close bead) shortcuts - use `s` to open status menu instead (ab-3zw.9)

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
