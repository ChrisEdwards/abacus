# Abacus Documentation

Welcome to the Abacus documentation! Abacus is a powerful terminal UI for visualizing and navigating [Beads](https://github.com/steveyegge/beads) issue tracking projects.

## What is Abacus?

Abacus transforms your Beads issue database into an interactive, hierarchical tree view directly in your terminal. It provides an intuitive interface for exploring complex dependency graphs, viewing issue details, and understanding project structure at a glance.

## Documentation Structure

- **[Getting Started](getting-started.md)** - Quick start guide to get up and running in minutes
- **[Installation](installation.md)** - Detailed installation instructions for various platforms
- **[User Guide](user-guide.md)** - Comprehensive guide to using Abacus features
- **[Configuration](configuration.md)** - How to configure Abacus for your workflow
- **[Keyboard Shortcuts](keyboard-shortcuts.md)** - Complete keyboard shortcut reference
- **[Troubleshooting](troubleshooting.md)** - Solutions to common issues
- **[Architecture](architecture.md)** - Technical overview and design decisions

## Key Features

### Hierarchical Tree View
Navigate your issues through an expandable tree structure that shows parent-child relationships and dependencies clearly.

### Smart Sorting
Issues are automatically prioritized with in-progress and ready-to-work items appearing first, helping you focus on what matters.

### Status Indicators
Color-coded visual indicators show issue status at a glance:
- `◐` In Progress (cyan)
- `○` Open/Ready (white)
- `✔` Closed (gray)
- `⛔` Blocked (red)

### Rich Detail Panel
View comprehensive issue information including metadata, full descriptions with markdown rendering, relationships, and comments.

### Live Search
Filter issues by title with instant results, making it easy to find specific issues in large projects.

### Dual-Pane Interface
Navigate the tree while viewing detailed information side-by-side.

### Smart Layout
Responsive design with text wrapping and viewport management adapts to your terminal size.

### Statistics Dashboard
Real-time counts of total, in-progress, ready, blocked, and closed issues help you understand project health.

## Quick Start

```bash
# Install Abacus
go install github.com/yourusername/abacus/cmd/abacus@latest

# Navigate to your Beads project
cd /path/to/your/project

# Run Abacus
abacus
```

## System Requirements

- Go 1.25.3 or later (for building from source)
- [Beads CLI](https://github.com/steveyegge/beads) v0.24.0 or later installed and initialized in your project (`bd --version`)
- Terminal with 256-color support recommended

## Getting Help

- Check the [User Guide](user-guide.md) for detailed usage instructions
- See [Troubleshooting](troubleshooting.md) for common issues
- Report bugs or request features on [GitHub Issues](https://github.com/yourusername/abacus/issues)

## License

Abacus is open source software licensed under the [MIT License](../LICENSE).
