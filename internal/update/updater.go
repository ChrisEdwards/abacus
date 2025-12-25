package update

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Error variables for updater-specific errors.
var (
	ErrPermissionDenied  = fmt.Errorf("permission denied")
	ErrChecksumMismatch  = fmt.Errorf("checksum verification failed")
	ErrUnsupportedOS     = fmt.Errorf("unsupported operating system")
	ErrDownloadFailed    = fmt.Errorf("download failed")
	ErrExtractionFailed  = fmt.Errorf("extraction failed")
	ErrWindowsNoAutoUpdate = fmt.Errorf("auto-update not supported on Windows; please download manually")
)

// Updater handles downloading and installing updates.
type Updater struct {
	owner      string
	repo       string
	httpClient *http.Client
}

// UpdaterOption configures an Updater.
type UpdaterOption func(*Updater)

// WithUpdaterHTTPClient sets a custom HTTP client for the updater.
func WithUpdaterHTTPClient(client *http.Client) UpdaterOption {
	return func(u *Updater) {
		u.httpClient = client
	}
}

// NewUpdater creates a new updater for the specified repository.
func NewUpdater(owner, repo string, opts ...UpdaterOption) *Updater {
	u := &Updater{
		owner: owner,
		repo:  repo,
		httpClient: &http.Client{
			Timeout: 0, // No timeout for downloads
		},
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// Update downloads and installs the specified version.
// It performs an atomic replacement of the current binary.
// Returns ErrWindowsNoAutoUpdate on Windows (manual update required).
func (u *Updater) Update(ctx context.Context, version string) error {
	// Windows cannot replace a running executable
	if runtime.GOOS == "windows" {
		return ErrWindowsNoAutoUpdate
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	// Check write permission
	if err := checkWritePermission(execPath); err != nil {
		return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
	}

	// Download and extract new binary
	tempBinary, cleanup, err := u.downloadAndExtract(ctx, version)
	if err != nil {
		return err
	}
	defer cleanup()

	// Backup current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("backup current binary: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(tempBinary, execPath); err != nil {
		// Attempt to restore backup
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("install new binary: %w", err)
	}

	// Make executable
	//nolint:gosec // G302: Binary needs to be executable
	if err := os.Chmod(execPath, 0755); err != nil {
		// Attempt to restore backup
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("set executable permission: %w", err)
	}

	return nil
}

// Rollback restores the previous version from backup.
func (u *Updater) Rollback() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	backupPath := execPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}

	if err := os.Rename(backupPath, execPath); err != nil {
		return fmt.Errorf("restore backup: %w", err)
	}

	return nil
}

// downloadAndExtract downloads and extracts the release tarball.
// Returns the path to the extracted binary and a cleanup function.
func (u *Updater) downloadAndExtract(ctx context.Context, version string) (string, func(), error) {
	assetName, err := getTarballName()
	if err != nil {
		return "", nil, err
	}

	// Normalize version tag
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	url := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/%s",
		u.owner, u.repo, version, assetName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "beads-viewer-updater")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("%w: status %d", ErrDownloadFailed, resp.StatusCode)
	}

	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "bv-update-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }

	// Extract tarball
	binaryPath, err := extractTarball(resp.Body, tempDir)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	return binaryPath, cleanup, nil
}

// extractTarball extracts a .tar.gz archive and returns the path to the binary.
func extractTarball(r io.Reader, destDir string) (string, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return "", fmt.Errorf("create gzip reader: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)
	var binaryPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Look for the binary (executable file, not a directory)
		name := filepath.Base(header.Name)
		if name == "bv" || name == "abacus" || strings.HasPrefix(name, "bv-") {
			destPath := filepath.Join(destDir, name)
			//nolint:gosec // G304: extracting to temp directory we control
			outFile, err := os.Create(destPath)
			if err != nil {
				return "", fmt.Errorf("create file: %w", err)
			}

			//nolint:gosec // G110: decompression bomb unlikely for known release assets
			if _, err := io.Copy(outFile, tr); err != nil {
				_ = outFile.Close()
				return "", fmt.Errorf("extract file: %w", err)
			}
			_ = outFile.Close()

			//nolint:gosec // G302: binary needs to be executable
			if err := os.Chmod(destPath, 0755); err != nil {
				return "", fmt.Errorf("chmod: %w", err)
			}

			binaryPath = destPath
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

// getTarballName returns the expected tarball name for the current OS/arch.
func getTarballName() (string, error) {
	goos := runtime.GOOS
	arch := runtime.GOARCH

	switch goos {
	case "darwin":
		if arch == "arm64" {
			return "bv_darwin_arm64.tar.gz", nil
		}
		return "bv_darwin_amd64.tar.gz", nil
	case "linux":
		if arch == "arm64" {
			return "bv_linux_arm64.tar.gz", nil
		}
		return "bv_linux_amd64.tar.gz", nil
	case "windows":
		return "", ErrWindowsNoAutoUpdate
	default:
		return "", fmt.Errorf("%w: %s/%s", ErrUnsupportedOS, goos, arch)
	}
}

// getAssetName returns the expected asset name for the current OS/arch.
func getAssetName() (string, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "darwin":
		if arch == "arm64" {
			return "bv-darwin-arm64", nil
		}
		return "bv-darwin-amd64", nil
	case "linux":
		if arch == "arm64" {
			return "bv-linux-arm64", nil
		}
		return "bv-linux-amd64", nil
	case "windows":
		return "bv-windows-amd64.exe", nil
	default:
		return "", fmt.Errorf("%w: %s/%s", ErrUnsupportedOS, os, arch)
	}
}

// checkWritePermission verifies the current process can write to the path.
func checkWritePermission(path string) error {
	dir := filepath.Dir(path)
	testFile := filepath.Join(dir, ".bv-update-test")

	//nolint:gosec // G304: Path is constructed from known binary directory
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	_ = f.Close()
	_ = os.Remove(testFile)
	return nil
}

// DetectInstallMethod determines how the application was installed.
func DetectInstallMethod() InstallMethod {
	execPath, err := os.Executable()
	if err != nil {
		return InstallUnknown
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return InstallUnknown
	}

	// Check for Homebrew installation
	if strings.Contains(execPath, "Cellar") || strings.Contains(execPath, "homebrew") {
		return InstallHomebrew
	}

	// Check if brew command recognizes it
	if isHomebrewInstalled() {
		return InstallHomebrew
	}

	return InstallDirect
}

// isHomebrewInstalled checks if the app is installed via Homebrew.
func isHomebrewInstalled() bool {
	cmd := exec.Command("brew", "list", "bv")
	err := cmd.Run()
	return err == nil
}

// VerifyChecksum verifies a file against an expected SHA256 checksum.
func VerifyChecksum(path, expected string) error {
	//nolint:gosec // G304: Path comes from caller; this is intentional for checksum verification
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hash file: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, expected, actual)
	}

	return nil
}

// ParseChecksumFile parses a checksums.txt file and returns a map of filename to checksum.
// Format: "sha256hash  filename" (two spaces between hash and filename)
func ParseChecksumFile(r io.Reader) (map[string]string, error) {
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on double space (standard format) or single space
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			parts = strings.SplitN(line, " ", 2)
		}
		if len(parts) != 2 {
			continue
		}

		hash := strings.TrimSpace(parts[0])
		filename := strings.TrimSpace(parts[1])
		// Remove any leading ./ or */
		filename = filepath.Base(filename)

		if hash != "" && filename != "" {
			checksums[filename] = hash
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}

	return checksums, nil
}
