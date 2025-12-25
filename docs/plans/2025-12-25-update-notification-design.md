# Update Notification & Auto-Update Design

**Bead:** ab-lanr
**Date:** 2025-12-25
**Status:** Approved

## Overview

Show users when an abacus update is available and help them upgrade. The notification appears as a toast during startup (10 seconds), with an optional hotkey to auto-update for direct download users.

## Architecture

```
Startup → GitHub API → Compare versions → If newer:
  ├─ Homebrew user → Toast with "brew upgrade abacus" instructions
  └─ Direct download → Toast with [U] hotkey to auto-update
```

### Components

| Component | Location | Purpose |
|-----------|----------|---------|
| Version Checker | `internal/update/checker.go` | Query GitHub API, compare versions, detect Homebrew |
| Auto-Updater | `internal/update/updater.go` | Download binary, verify, atomic replace |
| Update Toast | `internal/ui/view_toasts.go` | 10s notification with `[U]` hotkey |
| Startup Integration | `cmd/abacus/main.go` | Async check, pass to UI |

## Version Checker

**Package:** `internal/update/checker.go`

### Types

```go
type UpdateInfo struct {
    Available      bool      // Whether an update exists
    CurrentVersion string    // e.g., "0.6.1"
    LatestVersion  string    // e.g., "0.6.2"
    ReleaseURL     string    // GitHub release page URL
    DownloadURL    string    // Direct binary download URL
    IsHomebrew     bool      // Detected via binary path
}

type Checker struct {
    httpClient *http.Client
    repoOwner  string  // "ChrisEdwards"
    repoName   string  // "abacus"
}
```

### Homebrew Detection

```go
func isHomebrewInstall() bool {
    exe, err := os.Executable()
    if err != nil {
        return false
    }
    return strings.Contains(exe, "/opt/homebrew/") ||
           strings.Contains(exe, "/usr/local/Cellar/") ||
           strings.Contains(exe, "/home/linuxbrew/")
}
```

### Timeout

5 seconds - fail silently if network unavailable.

## Auto-Updater

**Package:** `internal/update/updater.go`

### Update Flow

```go
func (u *Updater) Update(info UpdateInfo) error {
    // 1. Find current binary path
    exe, _ := os.Executable()

    // 2. Download new binary to temp file
    tmpFile := exe + ".new"
    download(info.DownloadURL, tmpFile)

    // 3. Verify checksum (optional but recommended)
    verifyChecksum(tmpFile, info.ChecksumURL)

    // 4. Atomic replace: rename old → .old, new → current
    os.Rename(exe, exe + ".old")
    os.Rename(tmpFile, exe)
    os.Remove(exe + ".old")

    // 5. Return success - user should restart app
    return nil
}
```

### Asset Selection

Match `runtime.GOOS` and `runtime.GOARCH` to find correct tarball (e.g., `abacus_0.6.2_darwin_arm64.tar.gz`).

### Post-Update

Show message "Update complete. Please restart abacus." (don't auto-restart).

## Update Toast UI

### New App Fields

```go
// Update toast state
updateToastVisible   bool
updateToastStart     time.Time
updateInfo           *update.UpdateInfo  // nil if no update available
updateInProgress     bool                // true while downloading
```

### Toast Appearance

Homebrew users:
```
 ⬆ Update available: v0.6.2
   Run: brew upgrade abacus     [10s]
```

Direct download users:
```
 ⬆ Update available: v0.6.2
   Press [U] to update           [10s]
```

### Hotkey Handling

- `U` key only active when `updateToastVisible && !updateInfo.IsHomebrew`
- Triggers async update, sets `updateInProgress = true`
- On completion: show success message or fallback instructions

### Styling

Use `styleInfoToast()` - blue/cyan tint to differentiate from success/error.

## Startup Integration

### Main Flow

```go
// In main(), after version check for beads CLI:

// Start update check async (non-blocking)
updateCh := make(chan *update.UpdateInfo, 1)
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    checker := update.NewChecker()
    info, _ := checker.Check(ctx, Version)  // Errors silently ignored
    updateCh <- info
}()

// Continue with normal startup...
cfg := ui.Config{
    // ... existing fields ...
    UpdateChan: updateCh,
}
```

### UI App Initialization

- Receive from `UpdateChan` in a non-blocking select
- If update available, set `updateToastVisible = true` and `updateToastStart = time.Now()`

### Config Flag

Add `--skip-update-check` flag (and `AB_SKIP_UPDATE_CHECK` env var) for opt-out.

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Network error during version check | Silently ignored, app starts normally |
| Network error during auto-update | Show fallback: "Run: curl ... \| bash" |
| Permission denied | Show: "Permission denied. Run install script." |
| Binary in use (Windows) | Rename to `.old`, write new, suggest restart |
| GitHub rate limited | Silently skip |
| Dev build (`Version = "dev"`) | Skip update check entirely |

## Behavior Matrix

| Install Method | Toast Shows | Hotkey | Action |
|----------------|-------------|--------|--------|
| Homebrew | `brew upgrade abacus` | None | Manual only |
| Direct Download | `Press [U] to update` | `U` | Auto-download + replace |
| Dev build | No toast | N/A | Skip check entirely |

## Implementation Order

1. Create `internal/update/` package with checker
2. Add auto-updater with atomic binary replacement
3. Add update toast to UI layer
4. Integrate async check into startup
5. Add `--skip-update-check` flag
6. Write tests for each component
