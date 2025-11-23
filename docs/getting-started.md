# Getting Started with Abacus

This guide will help you get Abacus up and running in just a few minutes.

## Prerequisites

Before you begin, make sure you have:

1. **Go 1.25.3 or later** - Check your version with `go version`
2. **Beads CLI installed** - Install from [github.com/beadscli/beads](https://github.com/beadscli/beads)
3. **A Beads project** - You need a directory with a `.beads/` folder containing your issue database

## Installation

### Option 1: Install via Go (Recommended)

```bash
go install github.com/yourusername/abacus/cmd/abacus@latest
```

This installs the `abacus` binary to your `$GOPATH/bin` directory. Make sure this directory is in your `PATH`.

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/abacus.git
cd abacus

# Build the binary
go build -o abacus ./cmd/abacus

# Optionally, move it to your PATH
sudo mv abacus /usr/local/bin/
```

### Verify Installation

```bash
abacus --help
```

If you see the help output, you're ready to go!

## First Run

Navigate to a directory that contains a Beads project:

```bash
cd /path/to/your/beads/project
```

The directory should have a `.beads/` folder with your issue database. Then simply run:

```bash
abacus
```

Abacus will automatically:
1. Locate your `.beads/` database
2. Load all issues
3. Build the dependency graph
4. Display the interactive tree view

## Your First Session

Once Abacus launches, you'll see a tree view of your issues. Here's what to try:

### Navigate the Tree

- Press `↓` or `j` to move down
- Press `↑` or `k` to move up
- Press `→` or `l` to expand a node
- Press `←` or `h` to collapse a node

### View Issue Details

- Press `Enter` to open the detail panel for the selected issue
- The detail panel shows:
  - Issue metadata (ID, status, type, priority)
  - Full description with markdown formatting
  - Parent/child relationships
  - Blocking dependencies
  - Comments

### Search for Issues

- Press `/` to enter search mode
- Type part of an issue title
- Results update instantly as you type
- Press `Esc` to clear the search and show all issues

### Switch Between Panes

- Press `Tab` to switch focus between the tree view and detail panel
- When the detail panel is focused, you can scroll through long descriptions

### Exit

- Press `q` or `Ctrl+C` to quit

## Understanding the Display

### Status Icons

Issues are marked with icons indicating their status:

- `◐` **In Progress** (cyan) - Someone is actively working on this
- `○` **Open** (white) - Ready to be worked on
- `✔` **Closed** (gray) - Completed
- `⛔` **Blocked** (red) - Waiting on dependencies

### Tree Structure

The tree shows parent-child and dependency relationships:

```
○ beads-123: Parent Feature
  ◐ beads-124: Subtask 1
  ⛔ beads-125: Subtask 2 (blocked by beads-124)
```

### Statistics Bar

At the top of the screen, you'll see project statistics:

```
Total: 45  In Progress: 3  Ready: 12  Blocked: 5  Closed: 25
```

## Next Steps

Now that you're familiar with the basics:

- Read the [User Guide](user-guide.md) for detailed feature documentation
- Learn about [Configuration](configuration.md) to customize Abacus
- Check out the [Keyboard Shortcuts](keyboard-shortcuts.md) reference
- Enable auto-refresh to see changes in real-time

## Common First-Time Issues

### "No .beads directory found"

Make sure you're in a directory with a Beads project. The `.beads/` folder should exist in the current directory or a parent directory.

### "No issues found"

Your Beads database might be empty. Create some issues using:

```bash
bd create --title="My first issue" --type=task
```

### Terminal Colors Look Wrong

Ensure your terminal supports 256 colors. Most modern terminals do, but you may need to set:

```bash
export TERM=xterm-256color
```

For more help, see the [Troubleshooting](troubleshooting.md) guide.
