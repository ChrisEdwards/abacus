package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FindBeadsDB locates the beads database file.
// It checks in order:
// 1. .beads/beads.db walking up from the current directory
// 2. ~/.beads/default.db as a fallback
// Returns the path, modification time, and any error.
func FindBeadsDB() (string, time.Time, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("get working directory: %w", err)
	}
	if dbPath, modTime, err := findBeadsDBFromDir(wd); err == nil {
		return dbPath, modTime, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("locate beads db: %w", err)
	}
	fallback := filepath.Join(homeDir, ".beads", "default.db")
	info, err := os.Stat(fallback)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("no beads database found. Run 'beads init' to create one, or install beads from https://github.com/steveyegge/beads")
	}
	if info.IsDir() {
		return "", time.Time{}, fmt.Errorf("default beads db is a directory: %s", fallback)
	}
	modTime := info.ModTime()
	if latest, err := latestModTimeForDB(fallback); err == nil {
		modTime = latest
	}
	return fallback, modTime, nil
}

func findBeadsDBFromDir(startDir string) (string, time.Time, error) {
	if strings.TrimSpace(startDir) == "" {
		return "", time.Time{}, fmt.Errorf("start directory is required")
	}
	dir := startDir
	for {
		candidate := filepath.Join(dir, ".beads", "beads.db")
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			modTime := info.ModTime()
			if latest, err := latestModTimeForDB(candidate); err == nil {
				modTime = latest
			}
			return candidate, modTime, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", time.Time{}, fmt.Errorf("no beads db found from %s", startDir)
}
