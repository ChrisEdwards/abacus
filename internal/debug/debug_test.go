package debug

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_Disabled(t *testing.T) {
	// Reset state
	resetForTest()

	err := Init(false)
	if err != nil {
		t.Fatalf("Init(false) failed: %v", err)
	}

	if Enabled() {
		t.Error("Enabled() should return false when initialized with false")
	}

	// Logging should be no-ops
	Log("test message")
	Logf("test %s", "formatted")
}

func TestInit_Enabled(t *testing.T) {
	// Reset state
	resetForTest()

	// Use temp directory for log file
	tmpDir := t.TempDir()
	origGetLogPath := getLogPath
	getLogPath = func() (string, error) {
		return filepath.Join(tmpDir, LogDirName, LogFileName), nil
	}
	t.Cleanup(func() {
		getLogPath = origGetLogPath
		Close()
		resetForTest()
	})

	err := Init(true)
	if err != nil {
		t.Fatalf("Init(true) failed: %v", err)
	}

	if !Enabled() {
		t.Error("Enabled() should return true when initialized with true")
	}

	// Write some log messages
	Log("test message")
	Logf("test %s %d", "formatted", 42)

	// Verify log file was created and contains expected content
	logPath := filepath.Join(tmpDir, LogDirName, LogFileName)
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "debug log started") {
		t.Error("Log file should contain startup message")
	}
	if !strings.Contains(contentStr, "test message") {
		t.Error("Log file should contain 'test message'")
	}
	if !strings.Contains(contentStr, "test formatted 42") {
		t.Error("Log file should contain 'test formatted 42'")
	}
}

func TestInit_TruncatesExistingLog(t *testing.T) {
	// Reset state
	resetForTest()

	// Use temp directory for log file
	tmpDir := t.TempDir()
	origGetLogPath := getLogPath
	getLogPath = func() (string, error) {
		return filepath.Join(tmpDir, LogDirName, LogFileName), nil
	}
	t.Cleanup(func() {
		getLogPath = origGetLogPath
		Close()
		resetForTest()
	})

	// Create a pre-existing log file with content
	logDir := filepath.Join(tmpDir, LogDirName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}
	logPath := filepath.Join(logDir, LogFileName)
	if err := os.WriteFile(logPath, []byte("old log content that should be truncated\n"), 0600); err != nil {
		t.Fatalf("Failed to write pre-existing log: %v", err)
	}

	// Initialize with debug enabled
	err := Init(true)
	if err != nil {
		t.Fatalf("Init(true) failed: %v", err)
	}

	// Verify old content was truncated
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if strings.Contains(string(content), "old log content") {
		t.Error("Log file should have been truncated, but old content still present")
	}
	if !strings.Contains(string(content), "debug log started") {
		t.Error("Log file should contain new startup message")
	}
}

func TestClose(t *testing.T) {
	// Reset state
	resetForTest()

	// Use temp directory for log file
	tmpDir := t.TempDir()
	origGetLogPath := getLogPath
	getLogPath = func() (string, error) {
		return filepath.Join(tmpDir, LogDirName, LogFileName), nil
	}
	t.Cleanup(func() {
		getLogPath = origGetLogPath
		resetForTest()
	})

	err := Init(true)
	if err != nil {
		t.Fatalf("Init(true) failed: %v", err)
	}

	// Close should not panic
	Close()

	// Multiple closes should be safe
	Close()
	Close()
}

func TestGetLogPath(t *testing.T) {
	path, err := GetLogPath()
	if err != nil {
		t.Fatalf("GetLogPath() failed: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join(LogDirName, LogFileName)) {
		t.Errorf("GetLogPath() = %q, want suffix %q", path, filepath.Join(LogDirName, LogFileName))
	}
}

func TestLog_WhenDisabled(t *testing.T) {
	// Reset state
	resetForTest()

	err := Init(false)
	if err != nil {
		t.Fatalf("Init(false) failed: %v", err)
	}

	// These should be no-ops and not panic
	Log("test")
	Log("test", 123, "more")
	Logf("test %s", "fmt")
	Logf("test %d %s", 123, "fmt")
}

// resetForTest resets the package state for testing.
func resetForTest() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	enabled = false
	logger = nil
}
