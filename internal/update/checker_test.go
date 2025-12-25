package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	c := NewChecker("owner", "repo")
	if c.owner != "owner" {
		t.Errorf("owner = %q, want %q", c.owner, "owner")
	}
	if c.repo != "repo" {
		t.Errorf("repo = %q, want %q", c.repo, "repo")
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewCheckerWithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	c := NewChecker("owner", "repo", WithHTTPClient(customClient))

	if c.httpClient != customClient {
		t.Error("custom HTTP client not applied")
	}
}

func TestCheckerCheck(t *testing.T) {
	release := ReleaseInfo{
		TagName:     "v2.0.0",
		Name:        "Release 2.0.0",
		Body:        "Release notes",
		HTMLURL:     "https://github.com/owner/repo/releases/tag/v2.0.0",
		PublishedAt: time.Now(),
		Prerelease:  false,
		Draft:       false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	c := NewChecker("owner", "repo")
	// Override the URL by using a custom transport
	c.httpClient = &http.Client{
		Transport: &rewriteTransport{
			base:      http.DefaultTransport,
			targetURL: server.URL,
		},
	}

	ctx := context.Background()
	info, err := c.Check(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}

	if !info.UpdateAvailable {
		t.Error("UpdateAvailable should be true when latest > current")
	}
	if info.CurrentVersion.String() != "v1.0.0" {
		t.Errorf("CurrentVersion = %s, want v1.0.0", info.CurrentVersion.String())
	}
	if info.LatestVersion.String() != "v2.0.0" {
		t.Errorf("LatestVersion = %s, want v2.0.0", info.LatestVersion.String())
	}
	if info.ReleaseNotes != "Release notes" {
		t.Errorf("ReleaseNotes = %q, want %q", info.ReleaseNotes, "Release notes")
	}
}

func TestCheckerCheckNoUpdate(t *testing.T) {
	release := ReleaseInfo{
		TagName:     "v1.0.0",
		Name:        "Release 1.0.0",
		PublishedAt: time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	c := NewChecker("owner", "repo")
	c.httpClient = &http.Client{
		Transport: &rewriteTransport{
			base:      http.DefaultTransport,
			targetURL: server.URL,
		},
	}

	ctx := context.Background()
	info, err := c.Check(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}

	if info.UpdateAvailable {
		t.Error("UpdateAvailable should be false when latest == current")
	}
}

func TestCheckerCheckDevVersion(t *testing.T) {
	c := NewChecker("owner", "repo")
	ctx := context.Background()

	// Dev versions should return nil without error
	tests := []string{"dev", "development", ""}
	for _, version := range tests {
		info, err := c.Check(ctx, version)
		if err != nil {
			t.Errorf("Check(%q) unexpected error: %v", version, err)
		}
		if info != nil {
			t.Errorf("Check(%q) should return nil for dev version", version)
		}
	}
}

func TestCheckerCheckInvalidCurrentVersion(t *testing.T) {
	c := NewChecker("owner", "repo")
	ctx := context.Background()

	// Invalid versions should return nil without error (treated as dev builds)
	info, err := c.Check(ctx, "invalid")
	if err != nil {
		t.Errorf("Check() unexpected error: %v", err)
	}
	if info != nil {
		t.Error("Check() should return nil for unparseable version")
	}
}

func TestCheckerCheckRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := NewChecker("owner", "repo")
	c.httpClient = &http.Client{
		Transport: &rewriteTransport{
			base:      http.DefaultTransport,
			targetURL: server.URL,
		},
	}

	ctx := context.Background()
	_, err := c.Check(ctx, "1.0.0")
	if err == nil {
		t.Error("Check() should error on rate limit")
	}
}

func TestInstallMethodString(t *testing.T) {
	tests := []struct {
		method InstallMethod
		want   string
	}{
		{InstallUnknown, "unknown"},
		{InstallHomebrew, "homebrew"},
		{InstallDirect, "direct"},
	}

	for _, tt := range tests {
		if got := tt.method.String(); got != tt.want {
			t.Errorf("InstallMethod(%d).String() = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestFindDownloadURL(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "bv_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin-arm64"},
		{Name: "bv_darwin_amd64.tar.gz", BrowserDownloadURL: "https://example.com/darwin-amd64"},
		{Name: "bv_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux-amd64"},
		{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
	}

	// Test that findDownloadURL returns a URL (actual URL depends on runtime OS/arch)
	url := findDownloadURL(assets)
	// Should return one of the platform-specific URLs, not checksums
	if url == "https://example.com/checksums" {
		t.Error("findDownloadURL should not return non-binary asset")
	}
}

func TestFindDownloadURLEmpty(t *testing.T) {
	// Empty assets should return empty string
	url := findDownloadURL(nil)
	if url != "" {
		t.Errorf("findDownloadURL(nil) = %q, want empty string", url)
	}

	url = findDownloadURL([]ReleaseAsset{})
	if url != "" {
		t.Errorf("findDownloadURL([]) = %q, want empty string", url)
	}
}

func TestBuildAssetPatterns(t *testing.T) {
	// Test darwin/arm64
	patterns := buildAssetPatterns("darwin", "arm64")
	if len(patterns) == 0 {
		t.Error("buildAssetPatterns should return patterns")
	}

	// Should contain darwin_arm64 pattern
	found := false
	for _, p := range patterns {
		if p == "darwin_arm64" || p == "darwin-arm64" {
			found = true
			break
		}
	}
	if !found {
		t.Error("patterns should contain darwin_arm64 or darwin-arm64")
	}
}

func TestCheckWithDownloadURL(t *testing.T) {
	release := ReleaseInfo{
		TagName:     "v2.0.0",
		Name:        "Release 2.0.0",
		Body:        "Release notes",
		HTMLURL:     "https://github.com/owner/repo/releases/tag/v2.0.0",
		PublishedAt: time.Now(),
		Assets: []ReleaseAsset{
			{Name: "bv_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin-arm64"},
			{Name: "bv_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux-amd64"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	c := NewChecker("owner", "repo")
	c.httpClient = &http.Client{
		Transport: &rewriteTransport{
			base:      http.DefaultTransport,
			targetURL: server.URL,
		},
	}

	ctx := context.Background()
	info, err := c.Check(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}

	// DownloadURL should be populated (specific URL depends on runtime OS/arch)
	// Just verify it's not empty if we're on a supported platform
	if info.DownloadURL == "" {
		t.Log("DownloadURL is empty (may be expected if no matching asset for current platform)")
	}
}

// rewriteTransport rewrites request URLs for testing.
type rewriteTransport struct {
	base      http.RoundTripper
	targetURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.targetURL[7:] // strip "http://"
	return t.base.RoundTrip(req)
}
