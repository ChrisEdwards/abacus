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

### Basic Movement

| Key | Action |
|-----|--------|
| `↓` or `j` | Move cursor down |
| `↑` or `k` | Move cursor up |
| `→` or `l` | Expand current node |
| `←` or `h` | Collapse current node |
| `Space` | Toggle expand/collapse |

### Tree Navigation

The cursor (indicated by highlighting) shows your current position. Navigate through the tree to explore issue relationships.

**Tips:**
- Collapsed nodes show a count of hidden children: `[+3]`
- The tree automatically scrolls to keep the cursor visible
- Expanding a node reveals its children and dependencies

## Tree View

The tree view shows issues organized by their relationships.

### Node Structure

Each node displays:
- **Status icon** - Visual indicator of issue state
- **Issue ID** - Unique identifier (e.g., `beads-123`)
- **Title** - Brief description

Example:
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

### Tree Organization

Issues are organized to show:

1. **Root Issues** - Issues with no parents appear at the top level
2. **Parent-Child Relationships** - Children are indented under their parents
3. **Dependencies** - Blocked issues appear under their blockers
4. **Smart Sorting** - Within each level:
   - In-progress issues first
   - Ready (unblocked, open) issues second
   - Other issues follow

### Expanding and Collapsing

- **Expand** a node to see its children and dependents
- **Collapse** a node to hide its children
- Collapsed nodes show `[+N]` where N is the number of hidden descendants

Use this to focus on specific parts of your project.

## Detail Panel

The detail panel shows comprehensive information about the selected issue.

### Opening and Closing

- Press `Enter` to toggle the detail panel
- When closed, the tree view uses the full width
- When open, the view splits into tree and detail panes

### Focusing the Detail Panel

- Press `Tab` to switch focus between tree view and detail panel
- When focused, the detail panel has a highlighted border
- Navigation keys scroll the detail content

### Detail Panel Content

The detail panel includes:

#### Metadata Section

```
beads-123

Status: in_progress
Type: feature
Priority: 2 (medium)
Labels: auth, security
Created: 2024-03-15 10:30
Updated: 2024-03-20 14:45
```

#### Description

Full issue description with markdown rendering:
- **Headers**, *emphasis*, `code`
- Lists and nested lists
- Code blocks with syntax highlighting
- Links

#### Relationships

**Parent:**
```
beads-100: User Management System
```

**Children:**
```
- beads-124: Design authentication flow
- beads-125: Implement JWT tokens
```

**Blocked By:**
```
- beads-110: Database schema update
```

**Blocks:**
```
- beads-130: User profile page
```

#### Comments

```
Comment by @alice (2024-03-18 09:15)
Let's use OAuth2 instead of custom auth

Comment by @bob (2024-03-18 11:30)
Agreed, I'll update the design
```

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

## Search Functionality

Search allows you to quickly find issues by title.

### Activating Search

Press `/` to enter search mode. The search bar appears at the bottom:

```
Search: |
```

### Searching

- Type your search query
- Results update instantly as you type
- Search is case-insensitive
- Searches issue titles only

Example:
```
Search: /auth
```

Shows all issues with "auth" in their title.

### Search Behavior

- The tree view updates to show only matching issues
- Parent issues are shown if any descendants match
- Issue count updates to reflect filtered results
- The cursor automatically moves to the first match

### Clearing Search

- Press `Esc` to clear the search filter
- All issues are shown again
- The cursor returns to your previous position

### Search Tips

- Use partial words: `/feat` matches "Feature X"
- Search by ID: `/beads-123`
- Combine terms: `/auth user` (matches both words)

## Statistics Dashboard

The statistics bar at the top shows project metrics:

```
Total: 45  In Progress: 3  Ready: 12  Blocked: 5  Closed: 25
```

### Metrics Explained

- **Total** - All issues in the database
- **In Progress** - Issues currently being worked on
- **Ready** - Open issues with no blockers (ready to start)
- **Blocked** - Issues waiting on dependencies
- **Closed** - Completed issues

### When Search is Active

Statistics update to reflect filtered results:

```
Total: 8  In Progress: 1  Ready: 3  Blocked: 1  Closed: 3  (filtered)
```

Use this to understand the scope of your search results.

## Auto-Refresh

Abacus can automatically refresh the issue list to show changes made outside the application.

### Enabling Auto-Refresh

Auto-refresh is enabled by default. Disable it with:

```bash
abacus --no-auto-refresh
```

Or configure it (see [Configuration](configuration.md)):

```yaml
auto-refresh: true
refresh-interval: 3s
```

### How It Works

When enabled, Abacus:
1. Periodically reloads the issue database (default: every 3 seconds)
2. Rebuilds the tree with updated data
3. Maintains your cursor position
4. Preserves expanded/collapsed state
5. Keeps your search filter active

### Manual Refresh

Even with auto-refresh disabled, you can manually refresh by:
- Quitting (`q`) and restarting
- Or use `Ctrl+L` (if implemented)

### Refresh Interval

Adjust the polling interval:

```bash
abacus --refresh-interval 5s
```

Supported formats:
- `500ms` - Milliseconds
- `2s` - Seconds
- `1m` - Minutes

## Command-Line Options

Abacus supports several command-line flags to customize behavior.

### Basic Usage

```bash
abacus [options]
```

### Available Options

#### --db-path

Specify a custom database path:

```bash
abacus --db-path /path/to/.beads/beads.db
```

**Default:** Automatically searches for `.beads/` in current and parent directories.

#### --auto-refresh

Enable automatic background refresh:

```bash
abacus --auto-refresh
```

**Default:** Enabled

#### --no-auto-refresh

Disable automatic background refresh:

```bash
abacus --no-auto-refresh
```

#### --refresh-interval

Set the refresh polling interval:

```bash
abacus --refresh-interval 5s
```

**Default:** `3s`

**Formats:** `500ms`, `2s`, `1m`, etc.

#### --output-format

Set the markdown rendering style for the detail panel:

```bash
abacus --output-format rich
```

**Options:**
- `rich` - Full color and styling (default)
- `light` - Simplified styling
- `plain` - No styling, plain text

#### --json-output

Print issue data as JSON and exit (for scripting):

```bash
abacus --json-output > issues.json
```

This loads all issues and outputs JSON without starting the UI.

### Examples

**Minimal resource usage:**
```bash
abacus --no-auto-refresh --output-format plain
```

**Fast refresh for active development:**
```bash
abacus --refresh-interval 1s
```

**Custom database location:**
```bash
abacus --db-path ~/projects/myapp/.beads/beads.db
```

## Working with Issues

While Abacus is primarily for visualization, you can work with issues using Beads CLI alongside it.

### Typical Workflow

1. **Browse in Abacus** - Find the issue you want to work on
2. **Note the Issue ID** - e.g., `beads-123`
3. **Update via Beads CLI** - Use `bd` commands in another terminal
4. **See Updates in Abacus** - Auto-refresh shows the changes

### Example: Starting Work

1. In Abacus, find a ready issue:
   ```
   ○ beads-123: Implement feature X
   ```

2. In another terminal:
   ```bash
   bd update beads-123 --status in_progress
   ```

3. In Abacus (auto-refresh), see:
   ```
   ◐ beads-123: Implement feature X
   ```

### Example: Creating a Dependency

1. Notice in Abacus that you need to create a subtask

2. In terminal:
   ```bash
   bd create --title "Subtask A" --type task
   bd dep beads-123 beads-124  # beads-123 blocks beads-124
   ```

3. Abacus shows the new relationship:
   ```
   ◐ beads-123: Implement feature X
     ⛔ beads-124: Subtask A (blocked)
   ```

### Quick Reference: Beads CLI Commands

While viewing issues in Abacus, use these commands in another terminal:

```bash
# Create issue
bd create --title "Issue title" --type task

# Update status
bd update beads-123 --status in_progress

# Add dependency (A blocks B)
bd dep beads-123 beads-124

# Close issue
bd close beads-123

# View detailed info
bd show beads-123
```

## Keyboard Shortcuts Reference

See the [Keyboard Shortcuts](keyboard-shortcuts.md) page for a complete reference.

## Tips and Tricks

### Finding Blocked Issues

Issues with the `⛔` icon are blocked. Expand their parent nodes to see what's blocking them.

### Focus on Ready Work

Look for `○` icons with white color - these are ready to work on with no blockers.

### Understanding Dependencies

When you select a blocked issue, the detail panel shows what's blocking it in the "Blocked By" section.

### Large Projects

For projects with many issues:
- Keep non-essential branches collapsed
- Use search (`/`) to filter
- Focus on "In Progress" and "Ready" issues

### Terminal Sizing

Abacus adapts to your terminal size:
- **Minimum recommended:** 80x24
- **Comfortable:** 120x40
- **Detail panel** automatically wraps text

### Performance

If Abacus feels slow with many issues:
- Disable auto-refresh: `--no-auto-refresh`
- Increase refresh interval: `--refresh-interval 10s`
- Use plain output: `--output-format plain`

## Next Steps

- Customize Abacus with [Configuration](configuration.md)
- Learn all shortcuts in [Keyboard Shortcuts](keyboard-shortcuts.md)
- Troubleshoot issues in [Troubleshooting](troubleshooting.md)
