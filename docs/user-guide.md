# User Guide

This comprehensive guide covers all features and functionality of Abacus.

## Table of Contents

- [Overview](#overview)
- [Interface Layout](#interface-layout)
- [Navigation](#navigation)
- [Tree View](#tree-view)
- [Detail Panel](#detail-panel)
- [Search Functionality](#search-functionality)
- [Statistics Dashboard](#statistics-dashboard)
- [Auto-Refresh](#auto-refresh)
- [Command-Line Options](#command-line-options)
- [Working with Issues](#working-with-issues)

## Overview

Abacus provides a terminal-based interface for visualizing and navigating Beads issue databases. It displays issues in a hierarchical tree structure, showing relationships and dependencies clearly.

## Interface Layout

The Abacus interface consists of several components:

```
┌─────────────────────────────────────────────────────────────┐
│ Total: 45  In Progress: 3  Ready: 12  Blocked: 5  Closed: 25│ ← Statistics
├─────────────────────────────────────────────────────────────┤
│ Tree View                    │ Detail Panel                 │
│                              │                              │
│ ○ beads-123: Feature X       │ beads-123                   │
│   ◐ beads-124: Subtask A     │ Status: open                │
│   ○ beads-125: Subtask B     │ Type: feature               │
│ ○ beads-126: Feature Y       │ Priority: 2                 │
│   ✔ beads-127: Subtask C     │                             │
│                              │ Full description with       │
│ Search: /feature             │ markdown formatting...      │
└─────────────────────────────────────────────────────────────┘
```

### Components

1. **Statistics Bar** (top) - Shows project metrics
2. **Tree View** (left) - Hierarchical issue list
3. **Detail Panel** (right) - Detailed issue information (toggleable)
4. **Search Bar** (bottom) - Active when searching

## Navigation

| Key | Action |
|-----|--------|
| `↓` or `j` | Move cursor down |
| `↑` or `k` | Move cursor up |
| `→` or `l` | Expand current node |
| `←` or `h` | Collapse current node |
| `Space` | Toggle expand/collapse |

Collapsed nodes show `[+N]` with count of hidden children.

## Tree View

Each node displays status icon, issue ID, and title:
```
◐ beads-123: Implement user authentication
```

### Status Icons

| Icon | Status | Color | Meaning |
|------|--------|-------|---------|
| `◐` | In Progress | Cyan | Actively being worked on |
| `○` | Open | White | Available to work on |
| `✔` | Closed | Gray | Completed |
| `⛔` | Blocked | Red | Waiting on dependencies |

### Organization

Issues are organized hierarchically:
- Root issues (no parents) at top level
- Children indented under parents
- Blocked issues under their blockers
- Smart sorted: in-progress first, then ready, then others

## Detail Panel

Press `Enter` to toggle, `Tab` to switch focus between tree and detail.

### Content

Shows metadata (status, type, priority, labels, dates), full description with markdown rendering, parent/child relationships, blocking dependencies, and comments.

### Scrolling the Detail Panel

When the detail panel is focused (press `Tab`):

| Key | Action |
|-----|--------|
| `↓` or `j` | Scroll down one line |
| `↑` or `k` | Scroll up one line |
| `Ctrl+F` or `PgDn` | Scroll down one page |
| `Ctrl+B` or `PgUp` | Scroll up one page |
| `g` or `Home` | Jump to top |
| `G` or `End` | Jump to bottom |

## Search

Press `/` to enter search mode. Type to filter issues by title (case-insensitive). Results update instantly. Press `Esc` to clear.

## Statistics

The top bar shows: Total, In Progress, Ready (unblocked), Blocked, and Closed counts. Updates when filtering.

## Auto-Refresh

Enabled by default, reloads issues every 3 seconds while preserving cursor position and tree state.

```bash
abacus --auto-refresh-seconds 0       # Disable
abacus --auto-refresh-seconds 5       # Change interval
```

## Command-Line Options

```bash
abacus [options]
```

| Option | Description |
|--------|-------------|
| `--db-path` | Custom database path |
| `--auto-refresh-seconds` | Interval in seconds (0 disables; default: 3) |
| `--output-format` | Detail style: rich, light, plain |
| `--skip-version-check` | Skip Beads version check |

## Working with Issues

Use Abacus for visualization, Beads CLI for updates:

1. Browse in Abacus, note the issue ID
2. Update via `bd` in another terminal
3. See changes with auto-refresh

Common commands:
```bash
bd update beads-123 --status in_progress
bd dep beads-123 beads-124    # 123 blocks 124
bd close beads-123
```

## Tips

- `⛔` icons show blocked issues - check detail panel for blockers
- Keep non-essential branches collapsed in large projects
- Use search `/` to filter
- Minimum recommended terminal: 80x24

For customization see [Configuration](configuration.md), for issues see [Troubleshooting](troubleshooting.md).
