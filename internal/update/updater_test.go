package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
		if name != "abacus-darwin-arm64" && name != "abacus-darwin-amd64" {
			t.Errorf("unexpected asset name for darwin: %s", name)
		}
	case "linux":
		if name != "abacus-linux-arm64" && name != "abacus-linux-amd64" {
			t.Errorf("unexpected asset name for linux: %s", name)
		}
	case "windows":
		if name != "abacus-windows-amd64.exe" {
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

func TestGetTarballName(t *testing.T) {
	version := "0.6.1"
	name, err := getTarballName(version)

	// On Windows, should return error
	if runtime.GOOS == "windows" {
		if err == nil {
			t.Error("getTarballName() should error on Windows")
		}
		return
	}

	if err != nil {
		t.Fatalf("getTarballName() error: %v", err)
	}

	// Verify it ends with .tar.gz
	if !strings.HasSuffix(name, ".tar.gz") {
		t.Errorf("tarball name should end with .tar.gz, got: %s", name)
	}

	// Verify it contains the version
	if !strings.Contains(name, version) {
		t.Errorf("tarball name should contain version %q, got: %s", version, name)
	}

	// Verify it contains the expected OS
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(name, "darwin") {
			t.Errorf("tarball name should contain 'darwin', got: %s", name)
		}
	case "linux":
		if !strings.Contains(name, "linux") {
			t.Errorf("tarball name should contain 'linux', got: %s", name)
		}
	}
}

func TestGetTarballName_StripsVPrefix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	// Version with v prefix should work the same
	name1, _ := getTarballName("0.6.1")
	name2, _ := getTarballName("v0.6.1")

	if name1 != name2 {
		t.Errorf("getTarballName should strip v prefix: %q != %q", name1, name2)
	}

	// Verify version is in the name without the v
	if !strings.Contains(name1, "0.6.1") {
		t.Errorf("tarball name should contain '0.6.1', got: %s", name1)
	}
	if strings.Contains(name1, "v0.6.1") {
		t.Errorf("tarball name should not contain 'v0.6.1', got: %s", name1)
	}
}

func TestExtractTarball(t *testing.T) {
	// Create a test tarball in memory
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add a binary file
	binaryContent := []byte("#!/bin/sh\necho hello")
	hdr := &tar.Header{
		Name: "abacus_1.0.0_darwin_arm64/abacus",
		Mode: 0755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("write content: %v", err)
	}

	_ = tw.Close()
	_ = gw.Close()

	// Extract to temp directory
	tmpDir := t.TempDir()
	binaryPath, err := extractTarball(&buf, tmpDir)
	if err != nil {
		t.Fatalf("extractTarball() error: %v", err)
	}

	// Verify binary was extracted
	if binaryPath == "" {
		t.Error("extractTarball() returned empty path")
	}

	// Verify file exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Error("extracted binary does not exist")
	}

	// Verify content
	content, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if !bytes.Equal(content, binaryContent) {
		t.Error("extracted content does not match")
	}
}

func TestExtractTarballNoBinary(t *testing.T) {
	// Create a tarball without a binary
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add only a text file
	content := []byte("just a text file")
	hdr := &tar.Header{
		Name: "readme.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	_ = tw.WriteHeader(hdr)
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gw.Close()

	tmpDir := t.TempDir()
	_, err := extractTarball(&buf, tmpDir)
	if err == nil {
		t.Error("extractTarball() should error when no binary found")
	}
}

func TestParseChecksumFile(t *testing.T) {
	input := `abc123def456  bv_darwin_arm64.tar.gz
789xyz  bv_linux_amd64.tar.gz
# comment line
invalid line without space

deadbeef  ./path/to/bv_darwin_amd64.tar.gz
`

	checksums, err := ParseChecksumFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseChecksumFile() error: %v", err)
	}

	// Verify expected entries
	tests := []struct {
		filename string
		checksum string
	}{
		{"bv_darwin_arm64.tar.gz", "abc123def456"},
		{"bv_linux_amd64.tar.gz", "789xyz"},
		{"bv_darwin_amd64.tar.gz", "deadbeef"},
	}

	for _, tt := range tests {
		got, ok := checksums[tt.filename]
		if !ok {
			t.Errorf("missing checksum for %s", tt.filename)
			continue
		}
		if got != tt.checksum {
			t.Errorf("checksum[%s] = %s, want %s", tt.filename, got, tt.checksum)
		}
	}
}

func TestParseChecksumFileEmpty(t *testing.T) {
	checksums, err := ParseChecksumFile(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ParseChecksumFile() error: %v", err)
	}
	if len(checksums) != 0 {
		t.Errorf("expected empty map, got %d entries", len(checksums))
	}
}

func TestHasBackup(t *testing.T) {
	u := NewUpdater("owner", "repo")

	// There should be no backup for the current executable
	// (unless a previous failed test left one, which is unlikely)
	// This is a basic sanity check that the function doesn't panic
	_ = u.HasBackup()
}

func TestCleanupBackupNoBackup(t *testing.T) {
	u := NewUpdater("owner", "repo")

	// Cleaning up when no backup exists should succeed (no-op)
	err := u.CleanupBackup()
	if err != nil {
		t.Errorf("CleanupBackup() with no backup should succeed, got: %v", err)
	}
}
