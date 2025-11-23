# Getting Started

Get Abacus running in your Beads project.

## Prerequisites

- [Beads CLI](https://github.com/steveyegge/beads) v0.24.0+ (`bd --version`)
- A Beads project (directory with `.beads/` folder)

## Installation

See the [Installation Guide](installation.md) for platform-specific instructions, or:

```bash
brew install abacus
```

## First Run

```bash
cd /path/to/your/beads/project
abacus
```

Abacus automatically finds your `.beads/` database and displays your issues.

## Basic Navigation

| Key | Action |
|-----|--------|
| `↓/j` | Move down |
| `↑/k` | Move up |
| `→/l` or `Space` | Expand node |
| `←/h` | Collapse node |
| `Enter` | Toggle detail panel |
| `/` | Search |
| `Esc` | Clear search |
| `q` | Quit |

## Status Icons

- `◐` In Progress (cyan) - Being worked on
- `○` Open (white) - Ready to work on
- `✔` Closed (gray) - Completed
- `⛔` Blocked (red) - Waiting on dependencies

## Next Steps

- [User Guide](user-guide.md) - Complete feature documentation
- [Configuration](configuration.md) - Customize Abacus
- [Troubleshooting](troubleshooting.md) - If you run into issues
