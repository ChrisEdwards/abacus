package update

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewUpdater(t *testing.T) {
	u := NewUpdater("owner", "repo")
	if u.owner != "owner" {
		t.Errorf("owner = %q, want %q", u.owner, "owner")
	}
	if u.repo != "repo" {
		t.Errorf("repo = %q, want %q", u.repo, "repo")
	}
	if u.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestGetAssetName(t *testing.T) {
	name, err := getAssetName()
	if err != nil {
		t.Fatalf("getAssetName() error: %v", err)
	}

	// Verify it contains the expected OS
	switch runtime.GOOS {
	case "darwin":
		if name != "bv-darwin-arm64" && name != "bv-darwin-amd64" {
			t.Errorf("unexpected asset name for darwin: %s", name)
		}
	case "linux":
		if name != "bv-linux-arm64" && name != "bv-linux-amd64" {
			t.Errorf("unexpected asset name for linux: %s", name)
		}
	case "windows":
		if name != "bv-windows-amd64.exe" {
			t.Errorf("unexpected asset name for windows: %s", name)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	// Create a temp file with known content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	// Calculate expected checksum
	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	// Should succeed with correct checksum
	if err := VerifyChecksum(testFile, expected); err != nil {
		t.Errorf("VerifyChecksum() with correct checksum: %v", err)
	}

	// Should fail with incorrect checksum
	if err := VerifyChecksum(testFile, "wrong"); err == nil {
		t.Error("VerifyChecksum() should fail with wrong checksum")
	}

	// Should fail with non-existent file
	if err := VerifyChecksum(filepath.Join(tmpDir, "nonexistent"), expected); err == nil {
		t.Error("VerifyChecksum() should fail with non-existent file")
	}
}

func TestCheckWritePermission(t *testing.T) {
	// Should succeed in temp directory
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test")
	if err := checkWritePermission(testPath); err != nil {
		t.Errorf("checkWritePermission() in temp dir: %v", err)
	}
}

func TestDetectInstallMethod(t *testing.T) {
	// Just verify it doesn't panic and returns a valid value
	method := DetectInstallMethod()
	if method < InstallUnknown || method > InstallDirect {
		t.Errorf("DetectInstallMethod() returned invalid value: %d", method)
	}
}

func TestRollbackNoBackup(t *testing.T) {
	u := NewUpdater("owner", "repo")
	err := u.Rollback()
	// Should fail because there's no backup
	if err == nil {
		t.Error("Rollback() should fail when no backup exists")
	}
}
