package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func findBeadsDB() (string, time.Time, error) {
	if envPath := os.Getenv("BEADS_DB"); envPath != "" {
		info, err := os.Stat(envPath)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("BEADS_DB points to %s: %w", envPath, err)
		}
		if info.IsDir() {
			return "", time.Time{}, fmt.Errorf("BEADS_DB must point to a file, got directory: %s", envPath)
		}
		modTime := info.ModTime()
		if latest, err := latestModTimeForDB(envPath); err == nil {
			modTime = latest
		}
		return envPath, modTime, nil
	}

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
		return "", time.Time{}, fmt.Errorf("locate beads db: %w", err)
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
