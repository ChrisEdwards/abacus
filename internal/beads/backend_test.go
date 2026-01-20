package beads

import (
	"context"
	"errors"
	"testing"

	"abacus/internal/config"
)

func TestBackendConstants(t *testing.T) {
	if BackendBd != "bd" {
		t.Errorf("BackendBd = %q, want 'bd'", BackendBd)
	}
	if BackendBr != "br" {
		t.Errorf("BackendBr = %q, want 'br'", BackendBr)
	}
}

func TestMinBrVersion(t *testing.T) {
	if MinBrVersion == "" {
		t.Error("MinBrVersion should not be empty")
	}
}

func TestErrorVariables(t *testing.T) {
	if ErrNoBackendAvailable == nil {
		t.Error("ErrNoBackendAvailable should not be nil")
	}
	if ErrBackendAmbiguous == nil {
		t.Error("ErrBackendAmbiguous should not be nil")
	}
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name   string
		binary string
		want   bool
	}{
		{
			name:   "go binary should exist",
			binary: "go",
			want:   true,
		},
		{
			name:   "nonexistent binary should not exist",
			binary: "definitely-not-a-real-binary-xyz123",
			want:   false,
		},
		{
			name:   "empty string should not exist",
			binary: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commandExists(tt.binary)
			if got != tt.want {
				t.Errorf("commandExists(%q) = %v, want %v", tt.binary, got, tt.want)
			}
		})
	}
}

func TestIsInteractiveTTY(t *testing.T) {
	// In test environment, stdin is typically not a TTY
	// This test just ensures the function runs without panicking
	_ = isInteractiveTTY()
}

// saveAndRestoreHooks saves current hook values and returns a function to restore them.
// MUST be called and deferred at the start of each test that modifies hooks.
func saveAndRestoreHooks(t *testing.T) func() {
	t.Helper()

	origCommandExists := commandExistsFunc
	origIsInteractiveTTY := isInteractiveTTYFunc
	origCheckBackendVersion := checkBackendVersionFunc
	origConfigGetProjectString := configGetProjectStringFunc
	origConfigSaveBackend := configSaveBackendFunc
	origPromptUserForBackend := promptUserForBackendFunc
	origPromptSwitchBackend := promptSwitchBackendFunc

	return func() {
		commandExistsFunc = origCommandExists
		isInteractiveTTYFunc = origIsInteractiveTTY
		checkBackendVersionFunc = origCheckBackendVersion
		configGetProjectStringFunc = origConfigGetProjectString
		configSaveBackendFunc = origConfigSaveBackend
		promptUserForBackendFunc = origPromptUserForBackend
		promptSwitchBackendFunc = origPromptSwitchBackend
	}
}

// TestDetectBackend_OnlyBr tests detection when only br is on PATH.
func TestDetectBackend_OnlyBr(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: only br exists
	commandExistsFunc = func(name string) bool {
		return name == "br"
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: version check passes
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return nil
	}
	// Mock: save backend succeeds
	configSaveBackendFunc = func(_ string) error {
		return nil
	}

	ctx := context.Background()
	got, err := DetectBackend(ctx, "")
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil", err)
	}
	if got != BackendBr {
		t.Errorf("DetectBackend() = %q, want %q", got, BackendBr)
	}
}

// TestDetectBackend_OnlyBd tests detection when only bd is on PATH.
func TestDetectBackend_OnlyBd(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: only bd exists
	commandExistsFunc = func(name string) bool {
		return name == "bd"
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: version check passes
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return nil
	}
	// Mock: save backend succeeds
	configSaveBackendFunc = func(_ string) error {
		return nil
	}

	ctx := context.Background()
	got, err := DetectBackend(ctx, "")
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil", err)
	}
	if got != BackendBd {
		t.Errorf("DetectBackend() = %q, want %q", got, BackendBd)
	}
}

// TestDetectBackend_NeitherAvailable tests error when neither backend is on PATH.
func TestDetectBackend_NeitherAvailable(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: neither binary exists
	commandExistsFunc = func(_ string) bool {
		return false
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if !errors.Is(err, ErrNoBackendAvailable) {
		t.Errorf("DetectBackend() error = %v, want %v", err, ErrNoBackendAvailable)
	}
}

// TestDetectBackend_StoredPreference tests using stored preference when valid.
func TestDetectBackend_StoredPreference(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: br exists
	commandExistsFunc = func(name string) bool {
		return name == "br"
	}
	// Mock: stored preference is "br"
	configGetProjectStringFunc = func(key string) string {
		if key == config.KeyBeadsBackend {
			return "br"
		}
		return ""
	}
	// Mock: version check passes
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return nil
	}

	ctx := context.Background()
	got, err := DetectBackend(ctx, "")
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil", err)
	}
	if got != BackendBr {
		t.Errorf("DetectBackend() = %q, want %q", got, BackendBr)
	}
}

// TestDetectBackend_CLIFlagOverride tests --backend flag takes priority.
func TestDetectBackend_CLIFlagOverride(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: both binaries exist
	commandExistsFunc = func(_ string) bool {
		return true
	}
	// Mock: stored preference is "bd"
	configGetProjectStringFunc = func(key string) string {
		if key == config.KeyBeadsBackend {
			return "bd"
		}
		return ""
	}
	// Mock: version check passes
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return nil
	}
	// Mock: save should NOT be called for CLI flag override
	saveCalled := false
	configSaveBackendFunc = func(_ string) error {
		saveCalled = true
		return nil
	}

	ctx := context.Background()
	// Pass --backend br flag, which should override stored "bd"
	got, err := DetectBackend(ctx, "br")
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil", err)
	}
	if got != BackendBr {
		t.Errorf("DetectBackend() = %q, want %q", got, BackendBr)
	}
	if saveCalled {
		t.Error("CLI flag override should not save to config")
	}
}

// TestDetectBackend_CLIFlagInvalid tests invalid --backend flag value.
func TestDetectBackend_CLIFlagInvalid(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	ctx := context.Background()
	_, err := DetectBackend(ctx, "invalid")
	if err == nil {
		t.Error("DetectBackend() should error on invalid --backend value")
	}
	if !contains(err.Error(), "invalid --backend value") {
		t.Errorf("error message should mention invalid flag, got: %v", err)
	}
}

// TestDetectBackend_CLIFlagBinaryNotFound tests --backend with missing binary.
func TestDetectBackend_CLIFlagBinaryNotFound(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: br does not exist
	commandExistsFunc = func(name string) bool {
		return name != "br"
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "br")
	if err == nil {
		t.Error("DetectBackend() should error when --backend binary not found")
	}
	if !contains(err.Error(), "not found in PATH") {
		t.Errorf("error message should mention PATH, got: %v", err)
	}
}

// TestDetectBackend_BothBinaries_NonTTY tests both binaries present in non-interactive mode.
func TestDetectBackend_BothBinaries_NonTTY(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: both binaries exist
	commandExistsFunc = func(_ string) bool {
		return true
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: not a TTY (e.g., CI environment)
	isInteractiveTTYFunc = func() bool {
		return false
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if !errors.Is(err, ErrBackendAmbiguous) {
		t.Errorf("DetectBackend() error = %v, want %v", err, ErrBackendAmbiguous)
	}
}

// TestDetectBackend_BothBinaries_Interactive tests both binaries with user prompt.
func TestDetectBackend_BothBinaries_Interactive(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: both binaries exist
	commandExistsFunc = func(_ string) bool {
		return true
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: is a TTY
	isInteractiveTTYFunc = func() bool {
		return true
	}
	// Mock: user selects br
	promptUserForBackendFunc = func() string {
		return BackendBr
	}
	// Mock: version check passes
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return nil
	}
	// Mock: save succeeds
	var savedBackend string
	configSaveBackendFunc = func(backend string) error {
		savedBackend = backend
		return nil
	}

	ctx := context.Background()
	got, err := DetectBackend(ctx, "")
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil", err)
	}
	if got != BackendBr {
		t.Errorf("DetectBackend() = %q, want %q", got, BackendBr)
	}
	if savedBackend != BackendBr {
		t.Errorf("saved backend = %q, want %q", savedBackend, BackendBr)
	}
}

// TestDetectBackend_StalePreference_SwitchAccepted tests switching when stored binary is missing.
func TestDetectBackend_StalePreference_SwitchAccepted(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: only bd exists (br is missing)
	commandExistsFunc = func(name string) bool {
		return name == "bd"
	}
	// Mock: stored preference is "br" (stale - br not on PATH)
	configGetProjectStringFunc = func(key string) string {
		if key == config.KeyBeadsBackend {
			return "br"
		}
		return ""
	}
	// Mock: is a TTY
	isInteractiveTTYFunc = func() bool {
		return true
	}
	// Mock: user accepts switch to bd
	promptSwitchBackendFunc = func(_ string) bool {
		return true
	}
	// Mock: version check passes
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return nil
	}
	// Mock: save succeeds
	var savedBackend string
	configSaveBackendFunc = func(backend string) error {
		savedBackend = backend
		return nil
	}

	ctx := context.Background()
	got, err := DetectBackend(ctx, "")
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil", err)
	}
	if got != BackendBd {
		t.Errorf("DetectBackend() = %q, want %q", got, BackendBd)
	}
	if savedBackend != BackendBd {
		t.Errorf("saved backend = %q, want %q", savedBackend, BackendBd)
	}
}

// TestDetectBackend_StalePreference_SwitchDeclined tests declining switch when stored binary is missing.
func TestDetectBackend_StalePreference_SwitchDeclined(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: only bd exists (br is missing)
	commandExistsFunc = func(name string) bool {
		return name == "bd"
	}
	// Mock: stored preference is "br" (stale)
	configGetProjectStringFunc = func(key string) string {
		if key == config.KeyBeadsBackend {
			return "br"
		}
		return ""
	}
	// Mock: is a TTY
	isInteractiveTTYFunc = func() bool {
		return true
	}
	// Mock: user declines switch
	promptSwitchBackendFunc = func(_ string) bool {
		return false
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if err == nil {
		t.Error("DetectBackend() should error when user declines switch")
	}
	if !contains(err.Error(), "cannot continue") {
		t.Errorf("error message should mention cannot continue, got: %v", err)
	}
}

// TestDetectBackend_StalePreference_NonTTY tests stale preference in non-TTY mode.
func TestDetectBackend_StalePreference_NonTTY(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: only bd exists (br is missing)
	commandExistsFunc = func(name string) bool {
		return name == "bd"
	}
	// Mock: stored preference is "br" (stale)
	configGetProjectStringFunc = func(key string) string {
		if key == config.KeyBeadsBackend {
			return "br"
		}
		return ""
	}
	// Mock: not a TTY
	isInteractiveTTYFunc = func() bool {
		return false
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if err == nil {
		t.Error("DetectBackend() should error in non-TTY mode with stale preference")
	}
	if !contains(err.Error(), "use --backend") {
		t.Errorf("error message should mention --backend override, got: %v", err)
	}
}

// TestDetectBackend_StalePreference_NeitherAvailable tests when both binaries are missing.
func TestDetectBackend_StalePreference_NeitherAvailable(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: neither binary exists
	commandExistsFunc = func(_ string) bool {
		return false
	}
	// Mock: stored preference is "br"
	configGetProjectStringFunc = func(key string) string {
		if key == config.KeyBeadsBackend {
			return "br"
		}
		return ""
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if err == nil {
		t.Error("DetectBackend() should error when neither binary available")
	}
	if !contains(err.Error(), "neither bd nor br found") {
		t.Errorf("error message should mention neither found, got: %v", err)
	}
}

// TestDetectBackend_VersionFallback_SwitchAccepted tests switching when version check fails.
func TestDetectBackend_VersionFallback_SwitchAccepted(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: both binaries exist
	commandExistsFunc = func(_ string) bool {
		return true
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: is a TTY
	isInteractiveTTYFunc = func() bool {
		return true
	}
	// Mock: user selects br initially
	promptUserForBackendFunc = func() string {
		return BackendBr
	}
	// Mock: br version check fails, bd passes
	checkBackendVersionFunc = func(_ context.Context, backend string) error {
		if backend == "br" {
			return errors.New("version too old")
		}
		return nil
	}
	// Mock: user accepts switch to bd
	promptSwitchBackendFunc = func(_ string) bool {
		return true
	}
	// Mock: save succeeds
	var savedBackend string
	configSaveBackendFunc = func(backend string) error {
		savedBackend = backend
		return nil
	}

	ctx := context.Background()
	got, err := DetectBackend(ctx, "")
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil", err)
	}
	if got != BackendBd {
		t.Errorf("DetectBackend() = %q, want %q (after fallback)", got, BackendBd)
	}
	if savedBackend != BackendBd {
		t.Errorf("saved backend = %q, want %q", savedBackend, BackendBd)
	}
}

// TestDetectBackend_VersionFallback_SwitchDeclined tests declining switch when version fails.
func TestDetectBackend_VersionFallback_SwitchDeclined(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: both binaries exist
	commandExistsFunc = func(_ string) bool {
		return true
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: is a TTY
	isInteractiveTTYFunc = func() bool {
		return true
	}
	// Mock: user selects br
	promptUserForBackendFunc = func() string {
		return BackendBr
	}
	// Mock: br version check fails
	checkBackendVersionFunc = func(_ context.Context, backend string) error {
		if backend == "br" {
			return errors.New("version too old")
		}
		return nil
	}
	// Mock: user declines switch
	promptSwitchBackendFunc = func(_ string) bool {
		return false
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if err == nil {
		t.Error("DetectBackend() should error when user declines version fallback")
	}
	if !contains(err.Error(), "user declined switch") {
		t.Errorf("error message should mention user declined, got: %v", err)
	}
}

// TestDetectBackend_VersionFallback_NoAlternative tests version failure with no alternative.
func TestDetectBackend_VersionFallback_NoAlternative(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: only br exists
	commandExistsFunc = func(name string) bool {
		return name == "br"
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: br version check fails
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return errors.New("version too old")
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if err == nil {
		t.Error("DetectBackend() should error when version fails with no alternative")
	}
	if !contains(err.Error(), "no alternative backend available") {
		t.Errorf("error message should mention no alternative, got: %v", err)
	}
}

// TestDetectBackend_VersionFallback_NonTTY tests version failure in non-TTY mode.
func TestDetectBackend_VersionFallback_NonTTY(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: both binaries exist
	commandExistsFunc = func(_ string) bool {
		return true
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: not a TTY
	isInteractiveTTYFunc = func() bool {
		return false
	}
	// This should error at the "both exist, no TTY" check before version check
	// But if we had a stored preference, it would get to version check

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if !errors.Is(err, ErrBackendAmbiguous) {
		t.Errorf("DetectBackend() error = %v, want %v", err, ErrBackendAmbiguous)
	}
}

// TestDetectBackend_SaveError tests handling of config save errors.
func TestDetectBackend_SaveError(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: only br exists
	commandExistsFunc = func(name string) bool {
		return name == "br"
	}
	// Mock: no stored preference
	configGetProjectStringFunc = func(_ string) string {
		return ""
	}
	// Mock: version check passes
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return nil
	}
	// Mock: save fails (but detection should still succeed)
	configSaveBackendFunc = func(_ string) error {
		return errors.New("no .beads directory")
	}

	ctx := context.Background()
	got, err := DetectBackend(ctx, "")
	// Save errors are logged but don't fail detection
	if err != nil {
		t.Fatalf("DetectBackend() error = %v, want nil (save errors are non-fatal)", err)
	}
	if got != BackendBr {
		t.Errorf("DetectBackend() = %q, want %q", got, BackendBr)
	}
}

// TestDetectBackend_StoredPreference_VersionFails tests stored preference with version failure.
func TestDetectBackend_StoredPreference_VersionFails(t *testing.T) {
	restore := saveAndRestoreHooks(t)
	defer restore()

	// Mock: br exists
	commandExistsFunc = func(name string) bool {
		return name == "br"
	}
	// Mock: stored preference is "br"
	configGetProjectStringFunc = func(key string) string {
		if key == config.KeyBeadsBackend {
			return "br"
		}
		return ""
	}
	// Mock: version check fails
	checkBackendVersionFunc = func(_ context.Context, _ string) error {
		return errors.New("version too old")
	}

	ctx := context.Background()
	_, err := DetectBackend(ctx, "")
	if err == nil {
		t.Error("DetectBackend() should error when stored preference version check fails")
	}
	if !contains(err.Error(), "version check failed") {
		t.Errorf("error message should mention version check failed, got: %v", err)
	}
}

// Helper function to check string contains substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestNewClientForBackend_Bd tests creating a bd backend client.
func TestNewClientForBackend_Bd(t *testing.T) {
	client, err := NewClientForBackend(BackendBd, "/tmp/test.db")
	if err != nil {
		t.Fatalf("NewClientForBackend(%q, path) error = %v, want nil", BackendBd, err)
	}
	if client == nil {
		t.Fatal("NewClientForBackend returned nil client")
	}
	// Verify it's the correct type (bdSQLiteClient)
	if _, ok := client.(*bdSQLiteClient); !ok {
		t.Errorf("expected *bdSQLiteClient, got %T", client)
	}
}

// TestNewClientForBackend_Br tests creating a br backend client.
func TestNewClientForBackend_Br(t *testing.T) {
	client, err := NewClientForBackend(BackendBr, "/tmp/test.db")
	if err != nil {
		t.Fatalf("NewClientForBackend(%q, path) error = %v, want nil", BackendBr, err)
	}
	if client == nil {
		t.Fatal("NewClientForBackend returned nil client")
	}
	// Verify it's the correct type (brSQLiteClient)
	if _, ok := client.(*brSQLiteClient); !ok {
		t.Errorf("expected *brSQLiteClient, got %T", client)
	}
}

// TestNewClientForBackend_UnknownBackend tests error handling for unknown backend.
func TestNewClientForBackend_UnknownBackend(t *testing.T) {
	_, err := NewClientForBackend("unknown", "/tmp/test.db")
	if err == nil {
		t.Error("NewClientForBackend(unknown, path) should return error")
	}
	if !contains(err.Error(), "unknown backend") {
		t.Errorf("error should mention 'unknown backend', got: %v", err)
	}
}

// TestNewClientForBackend_EmptyDbPath tests error handling for empty db path.
func TestNewClientForBackend_EmptyDbPath(t *testing.T) {
	_, err := NewClientForBackend(BackendBd, "")
	if err == nil {
		t.Error("NewClientForBackend with empty dbPath should return error")
	}
	if !contains(err.Error(), "dbPath is required") {
		t.Errorf("error should mention 'dbPath is required', got: %v", err)
	}
}

// TestNewClientForBackend_EmptyBackend tests error handling for empty backend.
func TestNewClientForBackend_EmptyBackend(t *testing.T) {
	_, err := NewClientForBackend("", "/tmp/test.db")
	if err == nil {
		t.Error("NewClientForBackend with empty backend should return error")
	}
	if !contains(err.Error(), "unknown backend") {
		t.Errorf("error should mention 'unknown backend', got: %v", err)
	}
}
