# Abacus

A powerful terminal UI for visualizing and navigating [Beads](https://github.com/beadscli/beads) issue tracking projects.

## Overview

Abacus transforms your Beads issue database into an interactive, hierarchical tree view right in your terminal. It provides an intuitive interface for exploring complex dependency graphs, viewing issue details, and understanding project structure at a glance.

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
  - Parent/child relationships
  - Blocking dependencies
  - Comments with timestamps
- **Live Search**: Filter issues by title with instant results
- **Dual-Pane Interface**: Navigate the tree while viewing detailed information
- **Smart Layout**: Responsive design with text wrapping and viewport management
- **Statistics Dashboard**: Real-time counts of total, in-progress, ready, blocked, and closed issues

## Installation

### Prerequisites

- Go 1.25.3 or later
- [Beads CLI](https://github.com/beadscli/beads) installed and initialized in your project

### Build from Source

```bash
git clone https://github.com/yourusername/abacus.git
cd abacus
go build
```

This creates an `abacus` binary in the current directory.

### Install

```bash
# Install to your Go bin directory
go install

# Or move the binary to your PATH
sudo mv abacus /usr/local/bin/
```

## Usage

Navigate to any directory containing a Beads project and run:

```bash
abacus
```

The application will automatically load all issues from your `.beads/` database.

## Keyboard Shortcuts

### Navigation
- `↑/k` - Move cursor up
- `↓/j` - Move cursor down
- `←/h` - Collapse current node
- `→/l` or `space` - Toggle expand/collapse current node

### Views
- `enter` - Toggle detail panel
- `tab` - Switch focus between tree and detail panel
- `/` - Enter search mode
- `esc` - Clear search filter

### Detail Panel (when focused)
- `↑/↓` or `j/k` - Scroll through issue details
- `PgUp/PgDn` or `ctrl+f`/`ctrl+b` - Page through details
- `Home/End` or `g/G` - Jump to top or bottom of the detail pane

### General
- `q` or `ctrl+c` - Quit

## How It Works

Abacus interfaces with the Beads CLI to:

1. Load all issues from your project with `bd list --json`
2. Fetch full details for each issue with `bd show`
3. Build a dependency graph based on parent-child and blocking relationships
4. Render an interactive TUI using [Bubble Tea](https://github.com/charmbracelet/bubbletea)

The graph automatically identifies root nodes (issues with no parents or deepest parents in the hierarchy) and organizes the tree to minimize visual depth while accurately representing all relationships.

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

## Contributing

Contributions are welcome! Areas for improvement:

- Additional filtering options (by status, priority, labels)
- Export views (to markdown, JSON, etc.)
- Bulk operations on selected issues
- Integration with git for change tracking
- Performance optimizations for very large issue sets

## License

This project is licensed under the [MIT License](./LICENSE).

## Acknowledgments

Built with excellent TUI libraries from [Charm](https://github.com/charmbracelet).
Designed for use with [Beads](https://github.com/beadscli/beads).
