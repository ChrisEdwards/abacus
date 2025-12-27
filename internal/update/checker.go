package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// Default configuration values.
const (
	DefaultRepoOwner = "ChrisEdwards"
	DefaultRepoName  = "abacus"
	DefaultTimeout   = 5 * time.Second
)

// Error variables for specific error conditions.
var (
	ErrNetworkFailure = fmt.Errorf("network request failed")
	ErrRateLimited    = fmt.Errorf("rate limited by GitHub API")
	ErrInvalidVersion = fmt.Errorf("invalid version format")
)

// ReleaseAsset represents a downloadable file attached to a release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
	Size               int64  `json:"size"`
}

// ReleaseInfo contains information about a GitHub release.
type ReleaseInfo struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	Body        string         `json:"body"`
	HTMLURL     string         `json:"html_url"`
	PublishedAt time.Time      `json:"published_at"`
	Prerelease  bool           `json:"prerelease"`
	Draft       bool           `json:"draft"`
	Assets      []ReleaseAsset `json:"assets"`
}

// UpdateInfo contains the result of a version check.
type UpdateInfo struct {
	CurrentVersion  Version
	LatestVersion   Version
	UpdateAvailable bool
	ReleaseURL      string
	DownloadURL     string // Direct binary download URL for current platform
	ReleaseNotes    string
	PublishedAt     time.Time
	IsPrerelease    bool
	InstallMethod   InstallMethod
	UpdateCommand   string
	CheckedAt       time.Time
}

// InstallMethod indicates how the application was installed.
type InstallMethod int

const (
	// InstallUnknown indicates the installation method could not be determined.
	InstallUnknown InstallMethod = iota
	// InstallHomebrew indicates installation via Homebrew.
	InstallHomebrew
	// InstallDirect indicates a direct binary download.
	InstallDirect
)

// String returns the string representation of an InstallMethod.
func (m InstallMethod) String() string {
	switch m {
	case InstallHomebrew:
		return "homebrew"
	case InstallDirect:
		return "direct"
	default:
		return "unknown"
	}
}

// Checker handles version checking against GitHub releases.
type Checker struct {
	owner      string
	repo       string
	httpClient *http.Client
}

// CheckerOption configures a Checker.
type CheckerOption func(*Checker)

// WithHTTPClient sets a custom HTTP client for the checker.
func WithHTTPClient(client *http.Client) CheckerOption {
	return func(c *Checker) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) CheckerOption {
	return func(c *Checker) {
		c.httpClient.Timeout = timeout
	}
}

// NewChecker creates a new version checker for the specified repository.
func NewChecker(owner, repo string, opts ...CheckerOption) *Checker {
	c := &Checker{
		owner: owner,
		repo:  repo,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Check queries GitHub for the latest release and compares it to the current version.
// Returns nil without error for development builds or if the version cannot be parsed.
func (c *Checker) Check(ctx context.Context, currentVersion string) (*UpdateInfo, error) {
	// Skip check for development builds
	if currentVersion == "" || currentVersion == "dev" || currentVersion == "development" {
		return nil, nil
	}

	current, err := ParseVersion(currentVersion)
	if err != nil {
		// Silently skip check if version is unparseable (likely a dev build)
		return nil, nil
	}

	release, err := c.fetchLatestRelease(ctx)
	if err != nil {
		return nil, err
	}

	latest, err := ParseVersion(release.TagName)
	if err != nil {
		return nil, fmt.Errorf("parse latest version: %w", err)
	}

	info := &UpdateInfo{
		CurrentVersion:  current,
		LatestVersion:   latest,
		UpdateAvailable: current.LessThan(latest),
		ReleaseURL:      release.HTMLURL,
		DownloadURL:     findDownloadURL(release.Assets),
		ReleaseNotes:    release.Body,
		PublishedAt:     release.PublishedAt,
		IsPrerelease:    release.Prerelease,
		InstallMethod:   DetectInstallMethod(),
		CheckedAt:       time.Now(),
	}
	info.UpdateCommand = c.suggestUpdateCommand(info.InstallMethod)

	return info, nil
}

// fetchLatestRelease fetches the latest release from GitHub API.
func (c *Checker) fetchLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.owner, c.repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "abacus-update-checker")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkFailure, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrNetworkFailure, resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &release, nil
}

// suggestUpdateCommand returns the recommended update command based on install method.
func (c *Checker) suggestUpdateCommand(method InstallMethod) string {
	switch method {
	case InstallHomebrew:
		return "brew upgrade abacus"
	case InstallDirect:
		return "abacus --update"
	default:
		return "abacus --update"
	}
}

// findDownloadURL finds the download URL for the current platform from release assets.
func findDownloadURL(assets []ReleaseAsset) string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Build patterns to match against asset names
	patterns := buildAssetPatterns(os, arch)

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		for _, pattern := range patterns {
			if strings.Contains(name, pattern) {
				return asset.BrowserDownloadURL
			}
		}
	}

	return ""
}

// buildAssetPatterns returns patterns to match for the given OS/arch.
func buildAssetPatterns(os, arch string) []string {
	// Normalize arch names
	archPatterns := []string{arch}
	switch arch {
	case "amd64":
		archPatterns = append(archPatterns, "x86_64", "x64")
	case "arm64":
		archPatterns = append(archPatterns, "aarch64")
	}

	// Normalize OS names
	osPatterns := []string{os}
	switch os {
	case "darwin":
		osPatterns = append(osPatterns, "macos", "osx")
	}

	// Build all combinations
	var patterns []string
	for _, o := range osPatterns {
		for _, a := range archPatterns {
			patterns = append(patterns, o+"_"+a, o+"-"+a, a+"_"+o, a+"-"+o)
		}
	}

	return patterns
}
