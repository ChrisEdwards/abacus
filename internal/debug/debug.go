// Package debug provides debug logging infrastructure for abacus.
// Logging is only enabled when --debug flag is passed at startup.
// Logs are written to ~/.abacus/debug.log, truncated on each launch.
package debug

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// LogFileName is the name of the debug log file.
	LogFileName = "debug.log"
	// LogDirName is the name of the directory containing the log file.
	LogDirName = ".abacus"
)

var (
	mu      sync.RWMutex
	enabled bool
	logger  *log.Logger
	logFile *os.File

	// getLogPath is a function variable to allow overriding in tests.
	getLogPath = defaultGetLogPath
)

// Init initializes the debug logging system.
// If enable is false, all logging operations become no-ops.
// If enable is true, the log file is created/truncated at ~/.abacus/debug.log.
func Init(enable bool) error {
	mu.Lock()
	defer mu.Unlock()

	enabled = enable
	if !enable {
		logger = log.New(io.Discard, "", 0)
		return nil
	}

	logPath, err := getLogPath()
	if err != nil {
		return fmt.Errorf("determine log path: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(logPath)
	//nolint:gosec // G301: User config directory needs standard permissions
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	// Open log file, truncating if it exists
	//nolint:gosec // G304: Log path is computed from user home, not user input
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	logFile = f

	logger = log.New(f, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	logger.Printf("=== Abacus debug log started at %s ===", time.Now().Format(time.RFC3339))

	return nil
}

// Close closes the debug log file if open.
// Safe to call even if logging is disabled.
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// Log writes a debug message if debug logging is enabled.
// Arguments are handled in the manner of fmt.Print.
func Log(v ...any) {
	mu.RLock()
	defer mu.RUnlock()

	if !enabled || logger == nil {
		return
	}
	logger.Print(v...)
}

// Logf writes a formatted debug message if debug logging is enabled.
// Arguments are handled in the manner of fmt.Printf.
func Logf(format string, v ...any) {
	mu.RLock()
	defer mu.RUnlock()

	if !enabled || logger == nil {
		return
	}
	logger.Printf(format, v...)
}

// Enabled returns whether debug logging is currently enabled.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// defaultGetLogPath returns the path to the debug log file.
func defaultGetLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine user home: %w", err)
	}
	return filepath.Join(home, LogDirName, LogFileName), nil
}

// GetLogPath returns the path to the debug log file.
// Exported for use by other packages that need to know where logs are.
func GetLogPath() (string, error) {
	return getLogPath()
}
