# Configuration Guide

Abacus supports flexible configuration through multiple sources with clear precedence rules. This guide covers all configuration options and how to set them.

## Configuration Precedence

Configuration is loaded in the following order (later sources override earlier ones):

1. **Built-in defaults**
2. **User configuration** (`~/.config/abacus/config.yaml`)
3. **Project configuration** (`.abacus/config.yaml` in project or parent directories)
4. **Environment variables** (prefixed with `AB_`)
5. **Command-line flags** (highest priority)

## Configuration Files

### User Configuration

Global settings that apply to all projects.

**Location:** `~/.config/abacus/config.yaml`

**Example:**
```yaml
# Global user preferences
auto-refresh: true
refresh-interval: 3s
output:
  format: rich
```

### Project Configuration

Project-specific settings that override user configuration.

**Location:** `.abacus/config.yaml` (in project root or parent directories)

**Example:**
```yaml
# Project-specific settings
auto-refresh: false  # This project changes rarely
refresh-interval: 10s
database:
  path: .beads/beads.db
```

### Creating Configuration Files

#### User Configuration

```bash
# Create the directory
mkdir -p ~/.config/abacus

# Create the config file
cat > ~/.config/abacus/config.yaml << EOF
auto-refresh: true
refresh-interval: 3s
output:
  format: rich
EOF
```

#### Project Configuration

```bash
# In your project root
mkdir -p .abacus

# Create the config file
cat > .abacus/config.yaml << EOF
auto-refresh: true
database:
  path: .beads/beads.db
EOF
```

## Configuration Options

### Auto-Refresh Settings

#### auto-refresh

Enable or disable automatic background refresh.

**Type:** Boolean
**Default:** `true`
**Config file:**
```yaml
auto-refresh: true
```

**Environment variable:**
```bash
export AB_AUTO_REFRESH=true
```

**Command-line flag:**
```bash
abacus --auto-refresh
```

#### no-auto-refresh

Explicitly disable auto-refresh (overrides `auto-refresh`).

**Type:** Boolean
**Default:** `false`
**Config file:**
```yaml
no-auto-refresh: true
```

**Environment variable:**
```bash
export AB_NO_AUTO_REFRESH=true
```

**Command-line flag:**
```bash
abacus --no-auto-refresh
```

#### refresh-interval

Set the interval between automatic refreshes.

**Type:** Duration
**Default:** `3s`
**Formats:** `500ms`, `2s`, `1m`, etc.
**Config file:**
```yaml
refresh-interval: 5s
```

**Environment variable:**
```bash
export AB_REFRESH_INTERVAL=5s
```

**Command-line flag:**
```bash
abacus --refresh-interval 5s
```

### Database Settings

#### database.path

Specify the path to the Beads database file.

**Type:** String
**Default:** Auto-detected (searches for `.beads/beads.db`)
**Config file:**
```yaml
database:
  path: /path/to/.beads/beads.db
```

**Environment variable:**
```bash
export AB_DATABASE_PATH=/path/to/.beads/beads.db
```

**Command-line flag:**
```bash
abacus --db-path /path/to/.beads/beads.db
```

### Output Settings

#### output.format

Set the markdown rendering style for the detail panel.

**Type:** String
**Default:** `rich`
**Options:** `rich`, `light`, `plain`
**Config file:**
```yaml
output:
  format: rich
```

**Environment variable:**
```bash
export AB_OUTPUT_FORMAT=rich
```

**Command-line flag:**
```bash
abacus --output-format rich
```

**Option details:**
- `rich` - Full color, styling, and formatting
- `light` - Simplified styling for readability
- `plain` - Plain text with no styling

#### output.json

Print issue data as JSON and exit (for scripting).

**Type:** Boolean
**Default:** `false`
**Config file:**
```yaml
output:
  json: true
```

**Environment variable:**
```bash
export AB_OUTPUT_JSON=true
```

**Command-line flag:**
```bash
abacus --json-output
```

## Environment Variables

All configuration options can be set via environment variables using the `AB_` prefix.

### Naming Convention

Convert config keys to environment variables:
1. Add `AB_` prefix
2. Convert to uppercase
3. Replace `.` and `-` with `_`

**Examples:**

| Config Key | Environment Variable |
|------------|---------------------|
| `auto-refresh` | `AB_AUTO_REFRESH` |
| `refresh-interval` | `AB_REFRESH_INTERVAL` |
| `database.path` | `AB_DATABASE_PATH` |
| `output.format` | `AB_OUTPUT_FORMAT` |
| `output.json` | `AB_OUTPUT_JSON` |

### Setting Environment Variables

**Bash/Zsh:**
```bash
export AB_AUTO_REFRESH=true
export AB_REFRESH_INTERVAL=5s
```

**Fish:**
```fish
set -gx AB_AUTO_REFRESH true
set -gx AB_REFRESH_INTERVAL 5s
```

**Per-command:**
```bash
AB_AUTO_REFRESH=false abacus
```

## Configuration Examples

### Example 1: Fast-Paced Development

For active development with frequent changes:

```yaml
# ~/.config/abacus/config.yaml
auto-refresh: true
refresh-interval: 1s
output:
  format: rich
```

### Example 2: Low-Resource Environment

For systems with limited resources:

```yaml
# ~/.config/abacus/config.yaml
auto-refresh: false
output:
  format: plain
```

### Example 3: Multiple Projects

**User config:** `~/.config/abacus/config.yaml`
```yaml
# Defaults for all projects
auto-refresh: true
refresh-interval: 3s
output:
  format: rich
```

**Project A:** `~/projects/projectA/.abacus/config.yaml`
```yaml
# Fast refresh for active project
refresh-interval: 1s
```

**Project B:** `~/projects/projectB/.abacus/config.yaml`
```yaml
# Slow refresh for stable project
refresh-interval: 10s
```

### Example 4: CI/CD Integration

For scripts and automation:

```bash
#!/bin/bash
# Export issues as JSON for processing
abacus --json-output --db-path .beads/beads.db > issues.json

# Process JSON with jq
cat issues.json | jq '.[] | select(.status == "blocked")'
```

Or with config:

```yaml
# .abacus/config.yaml
output:
  json: true
database:
  path: .beads/beads.db
```

```bash
abacus > issues.json
```

## Verifying Configuration

To see which configuration is being used:

### Check Command-Line Help

```bash
abacus --help
```

Shows all available flags and their defaults.

### Test Configuration

Create a test config and verify behavior:

```bash
# Test with explicit flags
abacus --refresh-interval 10s --no-auto-refresh

# Test with environment variables
AB_REFRESH_INTERVAL=10s AB_NO_AUTO_REFRESH=true abacus

# Test with config file (after creating .abacus/config.yaml)
abacus
```

### Debug Configuration

If configuration doesn't seem to work:

1. **Check file locations:**
   ```bash
   ls -la ~/.config/abacus/config.yaml
   ls -la .abacus/config.yaml
   ```

2. **Check file syntax:**
   ```bash
   cat ~/.config/abacus/config.yaml
   ```
   Ensure proper YAML formatting (2-space indentation, no tabs)

3. **Check environment variables:**
   ```bash
   env | grep AB_
   ```

4. **Check command-line flags:**
   ```bash
   abacus --help
   ```

## Configuration Tips

### Tip 1: Start with User Config

Set your personal preferences in `~/.config/abacus/config.yaml` and override per-project as needed.

### Tip 2: Use Project Config for Team Settings

Check `.abacus/config.yaml` into version control so the team shares project settings.

### Tip 3: Environment Variables for CI/CD

Use environment variables in CI/CD pipelines rather than config files.

### Tip 4: Command-Line Flags for Experiments

Use flags for one-off changes without modifying config files.

### Tip 5: Document Project Settings

Add comments to `.abacus/config.yaml` explaining why settings differ from defaults.

## YAML Syntax Reference

Configuration files use YAML format.

### Basic Syntax

```yaml
# Comment
key: value
another-key: another value

# Nested values
parent:
  child: value
  another-child: value

# Boolean values
enabled: true
disabled: false

# Numbers
count: 42
priority: 2

# Durations (as strings)
interval: 3s
timeout: 500ms
```

### Common Mistakes

**❌ Wrong:**
```yaml
# Tabs (YAML doesn't support tabs)
output:
	format: rich

# Missing space after colon
key:value

# Inconsistent indentation
parent:
  child: value
   another-child: value  # Wrong indent
```

**✅ Correct:**
```yaml
# Two spaces for indentation
output:
  format: rich

# Space after colon
key: value

# Consistent indentation
parent:
  child: value
  another-child: value
```

## Troubleshooting Configuration

### Configuration Not Loading

**Problem:** Changes to config files don't take effect.

**Solutions:**
- Check file location: `~/.config/abacus/config.yaml` or `.abacus/config.yaml`
- Verify YAML syntax: no tabs, proper indentation
- Restart Abacus (config is loaded at startup)

### Wrong Configuration Applied

**Problem:** Unexpected configuration values are used.

**Solutions:**
- Remember precedence: CLI flags > env vars > project config > user config > defaults
- Check for environment variables: `env | grep AB_`
- Check for project config in parent directories

### Cannot Find Database

**Problem:** Abacus can't locate the `.beads/` directory.

**Solutions:**
- Run from a directory containing `.beads/`
- Or specify explicitly: `--db-path /path/to/.beads/beads.db`
- Or set in config: `database.path: /path/to/.beads/beads.db`

## Next Steps

- Return to the [User Guide](user-guide.md) to learn features
- Check [Troubleshooting](troubleshooting.md) for solutions to common issues
- See [Keyboard Shortcuts](keyboard-shortcuts.md) for quick reference
