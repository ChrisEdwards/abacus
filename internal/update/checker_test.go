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

func TestCheckerCheckInvalidCurrentVersion(t *testing.T) {
	c := NewChecker("owner", "repo")
	ctx := context.Background()

	_, err := c.Check(ctx, "invalid")
	if err == nil {
		t.Error("Check() should error on invalid current version")
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
