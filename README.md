# Abacus

A powerful terminal UI for visualizing and navigating Steve Yegge's awesome [Beads](https://github.com/steveyegge/beads) issue tracking project.

[![Latest Release](https://img.shields.io/github/v/release/ChrisEdwards/abacus)](https://github.com/ChrisEdwards/abacus/releases)
[![Go Version](https://img.shields.io/badge/go-1.25.3%2B-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)

## Overview

Abacus transforms your Beads issue database into an interactive, hierarchical tree view right in your terminal. It provides an intuitive interface for exploring complex dependency graphs, viewing issue details, and understanding project structure at a glance.

## Preview

![Abacus Terminal UI](assets/abacus-preview.png)

*Abacus showing a hierarchical tree view of Beads issues with the detail panel displaying comprehensive issue information.*

## Features

- **Hierarchical Tree View**: Visualize parent-child relationships and dependencies in an expandable tree structure
- **Smart Sorting**: Automatically prioritizes in-progress and ready-to-work issues
- **Status Indicators**: Color-coded icons show issue status at a glance
  - `◐` In Progress (cyan)
  - `○` Open/Ready (white)
  - `✔` Closed (gray)
  - `⛔` Blocked (red)
- **Rich Detail Panel**: View comprehensive issue information including:
  - Metadata (status, type, priority, labels, timestamps)
  - Full description with markdown rendering
  - Notes section with implementation details
  - Relationship sections (see below)
  - Comments with timestamps
- **Multi-Parent Support**: Tasks can belong to multiple parent epics:
  - Tasks appear under ALL their parent epics in the tree
  - `*` suffix indicates an item has multiple parents
  - Cross-highlighting: selecting one instance highlights all duplicates
  - Expansion state is shared across all instances
- **Live Search**: Filter issues by title with instant results
- **Dual-Pane Interface**: Navigate the tree while viewing detailed information
- **Smart Layout**: Responsive design with text wrapping and viewport management
- **Statistics Dashboard**: Real-time counts of total, in-progress, ready, blocked, and closed issues

## Quick Start

### Prerequisites

- [Beads CLI](https://github.com/steveyegge/beads) v0.24.0 or later installed and initialized in your project (`bd --version`)
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
  --db-path string            Path to the Beads database file
  --auto-refresh-seconds int  Auto-refresh interval in seconds (0 disables; default: 3)
  --output-format string      Detail panel style: rich, light, plain (default: "rich")
  --skip-version-check        Skip Beads CLI version validation (or set AB_SKIP_VERSION_CHECK=true)
```

Key workflows are summarized below—run `abacus --help` anytime for the full flag list.

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

## Keyboard Shortcuts

| Action | Keys | Description |
|--------|------|-------------|
| Navigate | `↑/k` `↓/j` | Move cursor up/down |
| Expand/Collapse | `→/l` `←/h` or `Space` | Expand/collapse nodes |
| Detail Panel | `Enter` | Toggle detail panel |
| Switch Focus | `Tab` | Switch between tree and detail |
| Search | `/` | Enter search mode |
| Clear Search | `Esc` | Clear search filter |
| Copy ID | `c` | Copy bead ID to clipboard |
| Help | `?` | Show keyboard shortcuts overlay |
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
output:
  json: false
  format: rich
database:
  path: .beads/beads.db
skip-version-check: false
```

Set `output.json: true`, `AB_OUTPUT_JSON=true`, or pass `--json-output` to print all issues as JSON (useful for scripting) and skip the TUI entirely. In JSON mode, startup animations are disabled so output stays machine-friendly.

## How It Works

Abacus interfaces with the Beads CLI to:

1. Load all issues from your project with `bd list --json`
2. Fetch full details for each issue with `bd show`
3. Build a dependency graph based on parent-child and blocking relationships
4. Render an interactive TUI using [Bubble Tea](https://github.com/charmbracelet/bubbletea)
5. Display a short-lived, witty spinner summarizing progress while the data loads

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
- `internal/beads/`: Thin wrapper over `bd list/show`

## License

This project is licensed under the [MIT License](./LICENSE).

## Acknowledgments

Built with excellent TUI libraries from [Charm](https://github.com/charmbracelet).
Designed for use with [Beads](https://github.com/beadscli/beads).
