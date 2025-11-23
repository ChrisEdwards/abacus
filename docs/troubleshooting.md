# Troubleshooting Guide

Solutions to common issues and problems you might encounter with Abacus.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Database Issues](#database-issues)
- [Display Issues](#display-issues)
- [Performance Issues](#performance-issues)
- [Configuration Issues](#configuration-issues)
- [Terminal Compatibility](#terminal-compatibility)
- [Beads Integration](#beads-integration)
- [Getting Help](#getting-help)

## Installation Issues

### "go: command not found"

**Problem:** Go is not installed or not in PATH.

**Solution:**

1. Install Go from [golang.org/dl](https://golang.org/dl)
2. Verify installation:
   ```bash
   go version
   ```
3. Ensure Go bin directory is in PATH:
   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```

### "cannot find package"

**Problem:** GOPATH not configured correctly.

**Solution:**

1. Check GOPATH:
   ```bash
   go env GOPATH
   ```
2. Should output something like `/home/username/go`
3. If empty or wrong, set it:
   ```bash
   export GOPATH=$HOME/go
   export PATH="$PATH:$GOPATH/bin"
   ```

### "permission denied" when installing

**Problem:** Insufficient permissions for installation.

**Solution:**

Don't use `sudo` with `go install`. Instead:

```bash
# Ensure GOPATH/bin is writable
mkdir -p $(go env GOPATH)/bin
chmod 755 $(go env GOPATH)/bin

# Install without sudo
go install github.com/yourusername/abacus/cmd/abacus@latest
```

### Build fails with "undefined: X"

**Problem:** Go version too old or dependencies not downloaded.

**Solution:**

1. Check Go version (need 1.25.3+):
   ```bash
   go version
   ```
2. Update Go if needed from [golang.org/dl](https://golang.org/dl)
3. Clear module cache and rebuild:
   ```bash
   go clean -modcache
   go build ./cmd/abacus
   ```

## Database Issues

### "No .beads directory found"

**Problem:** Abacus can't locate the Beads database.

**Solution:**

1. Ensure you're in a Beads project:
   ```bash
   ls -la .beads
   ```
2. Or initialize Beads:
   ```bash
   bd init
   ```
3. Or specify database path:
   ```bash
   abacus --db-path /path/to/.beads/beads.db
   ```

### "Failed to load database"

**Problem:** Database file is corrupted or inaccessible.

**Solution:**

1. Check file exists and is readable:
   ```bash
   ls -l .beads/beads.db
   ```
2. Check permissions:
   ```bash
   chmod 644 .beads/beads.db
   ```
3. Try running Beads CLI to verify database:
   ```bash
   bd list
   ```
4. If database is corrupted, restore from backup or reinitialize

### "No issues found"

**Problem:** Database is empty or query failed.

**Solution:**

1. Verify issues exist:
   ```bash
   bd list
   ```
2. If no issues, create some:
   ```bash
   bd create --title "Test issue" --type task
   ```
3. If Beads shows issues but Abacus doesn't, try:
   ```bash
   # Force reload
   abacus --db-path .beads/beads.db
   ```

### "Database locked"

**Problem:** Another process has exclusive lock on database.

**Solution:**

1. Check for other Beads/Abacus processes:
   ```bash
   ps aux | grep -E 'bd|abacus'
   ```
2. Close other instances
3. If persistent, check for stale lock files:
   ```bash
   ls -la .beads/
   # Look for .lock files
   ```

## Display Issues

### Colors don't appear correctly

**Problem:** Terminal doesn't support 256 colors.

**Solution:**

1. Set TERM environment variable:
   ```bash
   export TERM=xterm-256color
   ```
2. Or use plain output:
   ```bash
   abacus --output-format plain
   ```
3. Check terminal color support:
   ```bash
   echo $TERM
   tput colors  # Should output 256
   ```

### Unicode characters appear as boxes or question marks

**Problem:** Terminal encoding is not UTF-8.

**Solution:**

1. Set locale to UTF-8:
   ```bash
   export LANG=en_US.UTF-8
   export LC_ALL=en_US.UTF-8
   ```
2. Configure terminal to use UTF-8 encoding
3. Use ASCII-compatible terminal emulator

### Text wrapping is broken

**Problem:** Terminal width not detected correctly.

**Solution:**

1. Resize terminal window
2. Restart Abacus
3. Ensure terminal supports size queries:
   ```bash
   stty size  # Should output rows and columns
   ```

### Detail panel content is garbled

**Problem:** Markdown rendering issues or terminal compatibility.

**Solution:**

1. Use simpler output format:
   ```bash
   abacus --output-format light
   ```
   or
   ```bash
   abacus --output-format plain
   ```
2. Update terminal emulator to latest version
3. Check terminal supports required features

### Screen doesn't refresh

**Problem:** Display update issues.

**Solution:**

1. Force redraw by resizing terminal
2. Restart Abacus
3. Disable auto-refresh:
   ```bash
   abacus --auto-refresh-seconds 0
   ```
4. Check terminal emulator settings

## Performance Issues

### Slow startup

**Problem:** Large issue database or slow disk.

**Solution:**

1. Disable auto-refresh:
   ```bash
   abacus --auto-refresh-seconds 0
   ```
2. Use simpler output:
   ```bash
   abacus --output-format plain
   ```
3. Increase refresh interval:
   ```bash
   abacus --auto-refresh-seconds 10
   ```

### High CPU usage

**Problem:** Frequent refreshes or rendering issues.

**Solution:**

1. Increase refresh interval:
   ```bash
   abacus --auto-refresh-seconds 5
   ```
2. Disable auto-refresh:
   ```bash
   abacus --auto-refresh-seconds 0
   ```
3. Use plain output:
   ```bash
   abacus --output-format plain
   ```

### Sluggish navigation

**Problem:** Large tree with many nodes.

**Solution:**

1. Keep branches collapsed when not needed
2. Use search to filter issues:
   ```
   /your-query
   ```
3. Disable auto-refresh:
   ```bash
   abacus --auto-refresh-seconds 0
   ```

### Memory usage grows over time

**Problem:** Memory leak or cache buildup.

**Solution:**

1. Restart Abacus periodically
2. Disable auto-refresh if not needed
3. Report the issue on GitHub with details

## Configuration Issues

### Config file not loading

**Problem:** Configuration changes have no effect.

**Solution:**

1. Check file location:
   ```bash
   ls -la ~/.config/abacus/config.yaml
   ls -la .abacus/config.yaml
   ```
2. Verify YAML syntax:
   ```bash
   cat ~/.config/abacus/config.yaml
   ```
3. Ensure no tabs (YAML requires spaces)
4. Check for parsing errors

### Environment variables not working

**Problem:** Environment variables ignored.

**Solution:**

1. Verify variable name format:
   ```bash
   env | grep AB_
   ```
2. Use correct prefix and format:
   ```bash
   export AB_AUTO_REFRESH_SECONDS=5
   ```
3. Remember: CLI flags override env vars

### Command-line flags don't work

**Problem:** Flags have no effect.

**Solution:**

1. Check flag syntax:
   ```bash
   abacus --help
   ```
2. Ensure no typos:
   ```bash
   abacus --auto-refresh-seconds 5  # Correct
   abacus --auto_refresh_seconds 5   # Wrong (underscore)
   ```
3. Quote values with spaces:
   ```bash
   abacus --db-path "/path with spaces/beads.db"
   ```

## Terminal Compatibility

### Issues in tmux

**Problem:** Display or key binding issues in tmux.

**Solution:**

1. Enable 256 colors in tmux config (`~/.tmux.conf`):
   ```
   set -g default-terminal "screen-256color"
   ```
2. If `Ctrl+B` doesn't work (conflicts with tmux prefix):
   - Use `PgUp`/`PgDn` instead
   - Or change tmux prefix to something else
3. Reload tmux config:
   ```bash
   tmux source-file ~/.tmux.conf
   ```

### Issues in GNU Screen

**Problem:** Display rendering issues.

**Solution:**

1. Enable 256 colors in screen config (`~/.screenrc`):
   ```
   term screen-256color
   ```
2. Restart screen session

### Issues on Windows

**Problem:** Various compatibility issues on Windows.

**Solution:**

1. Use Windows Terminal (best compatibility)
2. Or use WSL2 with Linux terminal
3. Ensure terminal supports UTF-8:
   - Windows Terminal: Built-in
   - cmd.exe: `chcp 65001`
4. Use PowerShell or WSL bash

### Issues over SSH

**Problem:** Display issues when running over SSH.

**Solution:**

1. Ensure SSH forwards terminal properly:
   ```bash
   ssh -t user@host
   ```
2. Set TERM on remote:
   ```bash
   export TERM=xterm-256color
   ```
3. Enable X11 forwarding if needed:
   ```bash
   ssh -X user@host
   ```

## Beads Integration

### Abacus doesn't show recent changes

**Problem:** Auto-refresh not working or disabled.

**Solution:**

1. Enable auto-refresh:
   ```bash
   abacus --auto-refresh-seconds 3
   ```
2. Reduce refresh interval:
   ```bash
   abacus --auto-refresh-seconds 1
   ```
3. Or manually restart Abacus

### Can't find Beads CLI

**Problem:** `bd` command not in PATH.

**Solution:**

1. Install Beads CLI (v0.24.0+):
   ```bash
   go install github.com/steveyegge/beads/cmd/bd@latest
   ```
2. Add to PATH:
   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```
3. Verify installation:
   ```bash
   bd --version
   ```

### Dependency graph looks wrong

**Problem:** Incorrect parent-child or blocking relationships.

**Solution:**

1. Verify dependencies in Beads:
   ```bash
   bd show beads-123
   ```
2. Check for circular dependencies:
   ```bash
   bd blocked
   ```
3. Rebuild relationships:
   ```bash
   bd dep beads-123 beads-124
   ```

## Getting Help

### Check Documentation

- [User Guide](user-guide.md) - Feature documentation
- [Configuration](configuration.md) - Configuration help
- [Installation](installation.md) - Installation help
- [Keyboard Shortcuts](keyboard-shortcuts.md) - Shortcut reference

### Enable Debug Mode

If available in your version:

```bash
# Run with verbose output
abacus --debug

# Or with environment variable
AB_DEBUG=true abacus
```

### Collect Debug Information

When reporting issues, include:

1. **Version information:**
   ```bash
   abacus --version
   go version
   bd --version
   ```

2. **Environment:**
   ```bash
   echo $TERM
   echo $SHELL
   tput colors
   ```

3. **Configuration:**
   ```bash
   cat ~/.config/abacus/config.yaml
   env | grep AB_
   ```

4. **Database info:**
   ```bash
   ls -l .beads/
   bd list --json | jq '. | length'
   ```

### Report Issues

If you've tried the solutions above and still have problems:

1. Check existing issues: [GitHub Issues](https://github.com/yourusername/abacus/issues)
2. Create a new issue with:
   - Clear description of problem
   - Steps to reproduce
   - Debug information (see above)
   - Screenshots if applicable
   - Terminal emulator and OS

### Community Support

- GitHub Discussions: Ask questions and share tips
- IRC: #beads on Libera.Chat
- Matrix: #beads:matrix.org

## Common Error Messages

### "exec: \"bd\": executable file not found in $PATH"

**Cause:** Beads CLI not installed or not in PATH.

**Fix:** Install Beads and add to PATH (see [Beads Integration](#beads-integration)).

### "failed to parse config: yaml: unmarshal errors"

**Cause:** Syntax error in config file.

**Fix:** Check YAML syntax, ensure spaces (not tabs), proper indentation.

### "failed to open database: database is locked"

**Cause:** Another process has exclusive lock.

**Fix:** Close other Beads/Abacus instances (see [Database Issues](#database-issues)).

### "failed to initialize UI: terminal too small"

**Cause:** Terminal window is too small.

**Fix:** Resize terminal to at least 80x24 characters.

## Performance Tuning

For best performance:

```yaml
# ~/.config/abacus/config.yaml
auto-refresh-seconds: 0    # Or set a higher number to reduce refreshes
output:
  format: plain            # Simplify rendering
```

Or command-line:

```bash
abacus --auto-refresh-seconds 0 --output-format plain
```

## Still Having Issues?

If nothing here helps:

1. Check the [Architecture](architecture.md) documentation for technical details
2. Search [GitHub Issues](https://github.com/yourusername/abacus/issues)
3. Create a new issue with full details
4. Ask on community channels

We're here to help!
