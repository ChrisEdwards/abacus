# Troubleshooting

Quick fixes for common Abacus issues.

## Installation

- **Beads backend missing**: Run `br --version` or `bd --version`. Install one of:
  - [beads_rust (br)](https://github.com/Dicklesworthstone/beads_rust) — Recommended
  - [beads (bd)](https://github.com/steveyegge/beads) — Supported at v0.38.0
- **Binary not in PATH**: Add the install directory (Homebrew prefix, `~/.local/bin`, `~/.cargo/bin`, etc.) to your shell PATH.

## Backend Detection

- **"Both bd and br found" error in CI**: Non-interactive environments require explicit backend selection:
  ```bash
  abacus --backend br
  # Or pre-configure .abacus/config.yaml
  ```
- **"Stored backend not found" message**: Your configured backend binary is no longer on PATH. Abacus will prompt you to switch backends.
- **Version check failure**: Ensure your backend meets minimum version requirements:
  - `br`: v0.1.7 or later
  - `bd`: v0.30.0 to v0.38.0
- **BD version > 0.38.0 warning**: This is informational only. Abacus officially supports BD up to v0.38.0; newer versions may work but are not guaranteed.

## Runtime Problems

- **No data shown**: Confirm you are inside a directory with a `.beads/` database or pass `--db-path /path/to/.beads/beads.db`.
- **Wrong backend being used**: Check the status bar indicator (`[bd]` or `[br]`). Override with:
  ```bash
  abacus --backend br   # or bd
  ```
- **Terminal rendering glitches**: Ensure `$TERM` supports 256 colors (e.g., `xterm-256color`) and use a modern terminal emulator.
- **Tree doesn't refresh**: Resize the terminal or toggle auto refresh:
  ```bash
  abacus --auto-refresh-seconds 0   # disable
  abacus --auto-refresh-seconds 5   # slower cadence
  ```

## Performance

- Disable auto-refresh (`--auto-refresh-seconds 0`) if your `.beads` DB is huge.
- Use `--output-format plain` to reduce rendering cost.
- Collapse branches you are not actively inspecting.

## Database Issues

- **Database locked**: Close other instances of Abacus or Beads (`bd`/`br`) that are using the same `.beads` files.
- **Wrong project**: Use `abacus --db-path /path/to/project/.beads/beads.db` to point at a specific repo.

## Configuration

- **Backend preference not saved**: Backend selection is stored per-project in `.abacus/config.yaml`. If this directory doesn't exist, Abacus creates it on first run.
- **Config file location**: Project config is at `.abacus/config.yaml` in your project root. User config is at `~/.abacus/config.yaml`.

## Getting Help

Still stuck? File an issue at [github.com/ChrisEdwards/abacus/issues](https://github.com/ChrisEdwards/abacus/issues) with:
- Your OS and terminal emulator
- Backend version (`br --version` or `bd --version`)
- Abacus version (`abacus --version`)
- Steps to reproduce the issue
