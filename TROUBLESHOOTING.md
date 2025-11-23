# Troubleshooting

Quick fixes for common Abacus issues.

## Installation

- **Beads missing**: run `bd --version`. Install from https://github.com/steveyegge/beads if it fails.
- **Binary not in PATH**: add the install directory (Homebrew prefix, `~/.local/bin`, etc.) to your shell PATH.

## Runtime Problems

- **No data shown**: confirm you are inside a directory with a `.beads/` database or pass `--db-path /path/to/.beads/beads.db`.
- **Terminal rendering glitches**: ensure `$TERM` supports 256 colors (e.g., `xterm-256color`) and use a modern terminal emulator.
- **Tree doesn't refresh**: resize the terminal or toggle auto refresh:
  ```bash
  abacus --auto-refresh-seconds 0   # disable
  abacus --auto-refresh-seconds 5   # slower cadence
  ```

## Performance

- Disable auto-refresh (`--auto-refresh-seconds 0`) if your `.beads` DB is huge.
- Use `--output-format plain` to reduce rendering cost.
- Collapse branches you are not actively inspecting.

## Database Issues

- **Database locked**: close other instances of Abacus or Beads that are using the same `.beads` files.
- **Wrong project**: use `abacus --db-path /path/to/project/.beads/beads.db` to point at a specific repo.

Still stuck? File an issue with details about your OS, terminal, Beads version, and the steps to reproduce.
