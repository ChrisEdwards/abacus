# Abacus

A powerful terminal UI for visualizing and navigating [Beads](https://github.com/steveyegge/beads) issue tracking databases.

[![Latest Release](https://img.shields.io/github/v/release/ChrisEdwards/abacus)](https://github.com/ChrisEdwards/abacus/releases)
[![Go Version](https://img.shields.io/badge/go-1.25.3%2B-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)

## Overview

Abacus transforms your Beads issue database into an interactive, hierarchical tree view right in your terminal. It provides an intuitive interface for exploring complex dependency graphs, viewing issue details, and understanding project structure at a glance.

## Beads Backends: BD vs BR

Abacus supports two Beads backends:

| Backend | Binary | Repository | Status |
|---------|--------|------------|--------|
| **beads** (Go) | `bd` | [steveyegge/beads](https://github.com/steveyegge/beads) | Frozen at v0.38.0 |
| **beads_rust** | `br` | [Dicklesworthstone/beads_rust](https://github.com/Dicklesworthstone/beads_rust) | Active development |

### Why Two Backends?

**beads (bd)** is Steve Yegge's original Go-based issue tracker. As Steve continues evolving beads toward [GasTown](https://github.com/steveyegge/gastown) and beyond — with Dolt integration, daemon-based RPC, and advanced features — the architecture has naturally diverged from the simpler SQLite + JSONL model that many tools were built around.

**beads_rust (br)** was created by Jeffrey Emanuel as a Rust port that freezes the "classic beads" architecture. In his words:

> *"Rather than ask Steve to maintain a legacy mode for my niche use case, I created this Rust port that freezes the 'classic beads' architecture I depend on... This isn't a criticism of beads; Steve's taking it in exciting directions. It's simply that my tooling needs a stable snapshot of the architecture I built around, and maintaining my own fork is the right solution for that."*

**Steve Yegge has given his full endorsement of the beads_rust project.**

Key differences:

| Aspect | br (Rust) | bd (Go) |
|--------|-----------|---------|
| Git operations | **Never** (explicit only) | Auto-commit, hooks |
| Background daemon | **No** | Yes |
| Storage | SQLite + JSONL | Dolt/SQLite |
| Binary size | ~5-8 MB | ~30+ MB |
| Philosophy | Minimal, non-invasive | Feature-rich |

### Which Backend Should I Use?

- **New projects**: We recommend **br** ([beads_rust](https://github.com/Dicklesworthstone/beads_rust)) for its simplicity and active development
- **Existing bd projects**: Continue using **bd** — Abacus fully supports it at v0.38.0
- **Mixed usage**: Different repositories can use different backends — Abacus auto-detects per project

### Support Policy

- **BD support is frozen at version 0.38.0** — We will continue supporting BD indefinitely at this version, but will not add support for newer BD versions or features
- **BR is the path forward** — New features and improvements will be developed for the BR backend
- **100% JSONL compatible** — Both backends use the same `.beads/issues.jsonl` format, so your data is portable

## Preview

![Abacus Terminal UI](assets/abacus-preview.png)

*Abacus showing a hierarchical tree view of Beads issues with the detail panel displaying comprehensive issue information.*

## Features

### Tree View & Navigation
- **Hierarchical Tree View**: Visualize parent-child relationships and dependencies in an expandable tree structure
- **Smart Sorting**: Automatically prioritizes in-progress and ready-to-work issues
- **Status Indicators**: Color-coded icons show issue status at a glance
  - `◐` In Progress (cyan)
  - `○` Open/Ready (white)
  - `✔` Closed (gray)
  - `⛔` Blocked (red)
- **Multi-Parent Support**: Tasks can belong to multiple parent epics:
  - Tasks appear under ALL their parent epics in the tree
  - `*` suffix indicates an item has multiple parents
  - Cross-highlighting: selecting one instance highlights all duplicates
  - Expansion state is shared across all instances
- **View Mode Filtering**: Press `v` to cycle between All/Active/Ready views to hide closed issues
- **Live Search**: Filter issues by title with instant results

### Bead Management
- **Create Beads**: Press `n` for root beads or `N` for child beads with a streamlined modal
- **Edit Beads**: Press `e` to edit existing beads with pre-populated values
- **Quick Status Changes**: Press `s` to open the status overlay with single-key selection
- **Label Management**: Press `L` to add/remove labels with chip-based UI and autocomplete
- **Delete with Confirmation**: Press `Del` to delete beads with a safety confirmation dialog
- **Bulk Entry Mode**: Press `Ctrl+Enter` in the create modal to add multiple beads quickly
- **Type Auto-Inference**: The create modal suggests bead type based on title keywords

### Theming
- **20+ Built-in Themes**: Including TokyoNight (default), Dracula, Nord, Solarized, Catppuccin, Kanagawa, Gruvbox, One Dark, Rose Pine, GitHub, and more
- **Easy Theme Cycling**: Press `t` to cycle forward, `T` to cycle backward
- **Theme Persistence**: Your selected theme is saved and restored automatically

### Detail Panel
- **Rich Detail Panel**: View comprehensive issue information including:
  - Metadata (status, type, priority, labels, timestamps)
  - Full description with markdown rendering
  - Notes section with implementation details
  - Relationship sections (see below)
  - Comments with timestamps

### Interface
- **Dual-Pane Interface**: Navigate the tree while viewing detailed information
- **Smart Layout**: Responsive design with text wrapping and viewport management
- **Statistics Dashboard**: Real-time counts of total, in-progress, ready, blocked, and closed issues
- **Exit Summary**: See session duration and bead statistics when you quit

## Quick Start

### Prerequisites

- **Beads backend**: At least one of the following must be installed and on your PATH:
  - [**beads_rust (br)**](https://github.com/Dicklesworthstone/beads_rust) — Recommended for new projects
  - [**beads (bd)**](https://github.com/steveyegge/beads) — Supported at v0.38.0 for existing projects
  - Abacus auto-detects which backend is available and prompts if both are present
- Go 1.25.3 or later (only required for `go install` or building from source)

### Installation

**Option 1: Homebrew (macOS/Linux) - Recommended**
```bash
brew tap ChrisEdwards/tap
brew install abacus
```

**Option 2: Install Script (Unix/macOS/Linux)**
```bash
curl -fsSL https://raw.githubusercontent.com/ChrisEdwards/abacus/main/scripts/install.sh | bash
```

**Option 3: Install Script (Windows PowerShell)**
```powershell
irm https://raw.githubusercontent.com/ChrisEdwards/abacus/main/install.ps1 | iex
```

**Option 4: Go Install**
```bash
go install github.com/ChrisEdwards/abacus/cmd/abacus@latest
```

**Option 5: Download Binary**

Download the latest release for your platform from [GitHub Releases](https://github.com/ChrisEdwards/abacus/releases).

**Option 6: Build from Source**
```bash
git clone https://github.com/ChrisEdwards/abacus.git
cd abacus
make build
```

Prefer prebuilt binaries? Use the release assets, Brew formula, or install script.

## Usage

Navigate to any directory containing a Beads project and run:

```bash
abacus
```

The application will automatically load all issues from your `.beads/` database.

### Command-Line Options

```bash
abacus [options]

Options:
  --backend string            Backend to use: bd or br (default: auto-detect)
  --db-path string            Path to the Beads database file
  --auto-refresh-seconds int  Auto-refresh interval in seconds (0 disables; default: 3)
  --output-format string      Detail panel style: rich, light, plain (default: "rich")
  --skip-version-check        Skip Beads CLI version validation (or set AB_SKIP_VERSION_CHECK=true)
  --skip-update-check         Skip checking for updates at startup (or set AB_SKIP_UPDATE_CHECK=true)
```

Key workflows are summarized below—run `abacus --help` anytime for the full flag list.

### Backend Selection

Abacus supports both **beads (bd)** and **beads_rust (br)** backends. See [Beads Backends: BD vs BR](#beads-backends-bd-vs-br) for details on choosing between them.

**Auto-detection:** On startup, abacus checks which binaries are available:
- Only `br` on PATH → uses br automatically
- Only `bd` on PATH → uses bd automatically
- Both available → prompts you to choose (selection saved to `.abacus/config.yaml`)
- Neither available → shows error with installation instructions

**Manual override:**
```bash
# Use --backend flag (overrides config, not saved)
abacus --backend br
abacus --backend bd

# Or configure in .abacus/config.yaml
beads:
  backend: br  # or bd
```

**Status bar indicator:** The current backend is always shown in the status bar as `[bd]` or `[br]`.

**CI/Non-interactive environments:** Use `--backend` flag or pre-configure `.abacus/config.yaml` since the selection prompt requires an interactive terminal.

**Version requirements:**
- **br**: v0.1.7 or later
- **bd**: v0.30.0 to v0.38.0 (versions > 0.38.0 may work but are not officially supported)

**Note on BD version support:** If you're using BD version > 0.38.0, Abacus will display a one-time informational notice. The software may still work, but we cannot guarantee compatibility with newer BD features or breaking changes. For the best experience, we recommend migrating to BR for new projects.

### Detail Panel Relationship Sections

The detail panel shows different types of relationships:

| Section | Meaning | Description |
|---------|---------|-------------|
| **Part Of** | Parent epics | Epics/tasks this issue belongs to |
| **Subtasks** | Child tasks | Work items underneath this issue |
| **Must Complete First** | Blockers | Issues that block this one from starting |
| **Will Unblock** | Downstream | Issues waiting on this one to complete |
| **Related** | Soft links | Issues related but not blocking |
| **Discovered From** | Origin | Issues that led to discovering this one |

Items within each section are sorted intelligently:
- **Subtasks**: In-progress → ready (high-impact first) → blocked (closest to ready) → closed
- **Blockers**: Items you can work on now appear first
- **Will Unblock**: Items that become ready first appear first

### Search & Filtering

- Press `/` to search; results update live while you type. `Esc` clears the filter.
- Collapsed nodes show `[+N]` to indicate the number of hidden children.
- The statistics bar (top row) always reflects the currently visible issues.

### Auto-Refresh

- Enabled by default at 3 seconds; change with `--auto-refresh-seconds N`.
- Set `0` to disable background refresh if you want to control reloads manually.
- Auto-refresh preserves cursor, expanded nodes, and search filters.
- If a refresh fails, an error toast appears briefly in the bottom-right corner.

### Update Notifications

Abacus checks for updates when you launch it. If a new version is available,
you'll see a toast notification in the bottom-right corner:

- **Homebrew users**: The toast shows `brew upgrade abacus` command
- **Direct download users**: Press `U` to auto-update, or download from [releases](https://github.com/ChrisEdwards/abacus/releases)

To disable update checks:
```bash
abacus --skip-update-check
# or
export AB_SKIP_UPDATE_CHECK=true
```

## Keyboard Shortcuts

### Navigation
| Action | Keys | Description |
|--------|------|-------------|
| Navigate | `↑/k` `↓/j` | Move cursor up/down |
| Expand/Collapse | `→/l` `←/h` or `Space` | Expand/collapse nodes |
| Jump | `Home/End` | Jump to first/last item |
| Page | `PgUp/PgDn` | Page up/down in tree |
| Detail Panel | `Enter` | Toggle detail panel |
| Switch Focus | `Tab` | Switch between tree and detail |

### Bead Actions
| Action | Keys | Description |
|--------|------|-------------|
| New Root Bead | `n` | Create a new root-level bead |
| New Child Bead | `N` | Create bead under selected parent |
| Edit Bead | `e` | Edit selected bead |
| Change Status | `s` | Open status overlay |
| Manage Labels | `L` | Open labels overlay |
| Delete Bead | `Del` | Delete bead (with confirmation) |
| Copy ID | `c` | Copy bead ID to clipboard |

### Display
| Action | Keys | Description |
|--------|------|-------------|
| Cycle Theme | `t/T` | Cycle through themes (forward/backward) |
| Cycle View | `v/V` | Cycle view modes (All/Active/Ready) |
| Refresh | `r` | Manual refresh |
| Help | `?` | Show keyboard shortcuts overlay |

### Search & Other
| Action | Keys | Description |
|--------|------|-------------|
| Search | `/` | Enter search mode |
| Clear/Cancel | `Esc` | Clear search or close overlay |
| Quit | `q` or `Ctrl+C` | Exit application |

Detail panel focused shortcuts: `↑/↓` or `j/k` scroll, `Ctrl+F/B` or `PgDn/Up` page, `g/G` or `Home/End` jump.

## Configuration

Abacus can be configured via:
- Configuration files (`~/.abacus/config.yaml` or `.abacus/config.yaml`)
- Environment variables (prefixed with `AB_`)
- Command-line flags

**Example configuration:**
```yaml
auto-refresh-seconds: 3
beads:
  backend: br  # or bd (auto-detected if not set)
output:
  format: rich
database:
  path: .beads/beads.db
skip-version-check: false
```

Set `output.json: true`, `AB_OUTPUT_JSON=true`, or pass `--json-output` to print all issues as JSON (useful for scripting) and skip the TUI entirely. In JSON mode, startup animations are disabled so output stays machine-friendly.

## How It Works

Abacus interfaces with the Beads backend (bd or br) to:

1. Detect available backend and load preference
2. Load all issues from your project via SQLite (fast read path)
3. Perform mutations via CLI commands (create, update, delete, etc.)
4. Build a dependency graph based on parent-child and blocking relationships
5. Render an interactive TUI using [Bubble Tea](https://github.com/charmbracelet/bubbletea)
6. Display a short-lived, witty spinner summarizing progress while the data loads

The graph automatically identifies root nodes (issues with no parents or deepest parents in the hierarchy) and organizes the tree to minimize visual depth while accurately representing all relationships.

Internally the app follows Bubble Tea's Elm-inspired update/view cycle with domain, graph, and config layers separated for testability.

## Architecture

Abacus is built with:

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: The Elm Architecture for Go TUIs
- **[Bubbles](https://github.com/charmbracelet/bubbles)**: TUI components (viewport, text input)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Style definitions and layout
- **[Glamour](https://github.com/charmbracelet/glamour)**: Markdown rendering for descriptions

The codebase is organized into logical sections:
- Style definitions (colors, text styles, layout styles)
- Data structures (Issue, Node, Stats)
- Graph building logic (dependency resolution, tree construction)
- TUI logic (Bubble Tea Model/View/Update pattern)
- Rendering utilities (text wrapping, formatting, viewport management)

## Why Abacus?

While the Beads CLI is powerful for managing issues, complex projects with many dependencies can be difficult to visualize. Abacus solves this by:

- Showing the full project structure at a glance
- Making dependencies and blockers immediately visible
- Providing context-aware navigation
- Offering rich, formatted issue views without leaving the terminal

## Troubleshooting

Having issues? See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for quick fixes.

- Installation issues
- Database connectivity
- Display problems
- Performance tuning
- Terminal compatibility

## Contributing

Contributions are welcome! Areas for improvement:

- Additional filtering options (by status, priority, labels)
- Export views (to markdown, JSON, etc.)
- Bulk operations on selected issues
- Integration with git for change tracking
- Performance optimizations for very large issue sets
- Automated dependency management improvements (Dependabot is configured via `.github/dependabot.yml` and PRs welcome for additional ecosystems)

See the references below for a quick map of the codebase.

### Development

```bash
# Clone the repository
git clone https://github.com/yourusername/abacus.git
cd abacus

# Run tests
make test

# Run linter
make lint

# Build
make build
```

For information about creating releases, see **[RELEASING.md](RELEASING.md)**.

## References

- `cmd/abacus/`: CLI entrypoint and flag parsing
- `internal/ui/`: Bubble Tea models, tree/detail rendering, search, auto-refresh
- `internal/config/`: Viper-backed configuration (env, files, overrides)
- `internal/graph/`: Dependency graph construction and sorting
- `internal/beads/`: Backend abstraction (bd/br detection, SQLite reads, CLI writes)

## License

This project is licensed under the [MIT License](./LICENSE).

## Acknowledgments

Built with excellent TUI libraries from [Charm](https://github.com/charmbracelet).

Designed for use with:
- [**beads_rust (br)**](https://github.com/Dicklesworthstone/beads_rust) — Recommended backend, actively developed
- [**beads (bd)**](https://github.com/steveyegge/beads) — Original Go implementation by Steve Yegge
