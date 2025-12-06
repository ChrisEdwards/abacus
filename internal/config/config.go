package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

const (
	KeyAutoRefreshSeconds = "auto-refresh-seconds"
	KeyRefreshInterval    = "refresh-interval" // Deprecated: use KeyAutoRefreshSeconds.
	KeyAutoRefresh        = "auto-refresh"     // Deprecated: use KeyAutoRefreshSeconds.
	KeyNoAutoRefresh      = "no-auto-refresh"  // Deprecated: use KeyAutoRefreshSeconds.
	KeySkipVersionCheck   = "skip-version-check"

	KeyDatabasePath = "database.path"
	KeyOutputFormat = "output.format"
	KeyOutputJSON   = "output.json"
)

const (
	defaultAutoRefreshSeconds = 3
	envPrefix                 = "AB"
)

type initSettings struct {
	workingDir        string
	projectConfigPath string
	userConfigPath    string
}

// Option configures Initialize behaviour. Useful for tests to override paths.
type Option func(*initSettings)

// WithWorkingDir overrides the directory used for project config discovery.
func WithWorkingDir(dir string) Option {
	return func(cfg *initSettings) {
		cfg.workingDir = dir
	}
}

// WithProjectConfig explicitly sets the project config path instead of discovery.
func WithProjectConfig(path string) Option {
	return func(cfg *initSettings) {
		cfg.projectConfigPath = path
	}
}

// WithUserConfig overrides the default user config path.
func WithUserConfig(path string) Option {
	return func(cfg *initSettings) {
		cfg.userConfigPath = path
	}
}

var (
	configOnce sync.Once
	configMu   sync.RWMutex
	configInst *viper.Viper
	initErr    error
)

// Initialize loads configuration using the precedence:
// defaults < user config < project config < environment variables < overrides.
func Initialize(opts ...Option) error {
	configOnce.Do(func() {
		settings := initSettings{}
		for _, opt := range opts {
			opt(&settings)
		}
		initErr = configure(&settings)
	})
	return initErr
}

// ApplyOverrides injects values typically coming from CLI flags.
func ApplyOverrides(overrides map[string]any) error {
	if len(overrides) == 0 {
		return nil
	}
	if err := Initialize(); err != nil {
		return err
	}
	configMu.Lock()
	defer configMu.Unlock()
	if configInst == nil {
		return fmt.Errorf("configuration not initialized")
	}
	for k, v := range overrides {
		configInst.Set(k, v)
	}
	return nil
}

// GetString fetches a string configuration value, initializing on demand.
func GetString(key string) string {
	v, err := getViper()
	if err != nil {
		return ""
	}
	return v.GetString(key)
}

// GetBool fetches a bool configuration value, initializing on demand.
func GetBool(key string) bool {
	v, err := getViper()
	if err != nil {
		return false
	}
	return v.GetBool(key)
}

// GetInt fetches an integer configuration value, initializing on demand.
func GetInt(key string) int {
	v, err := getViper()
	if err != nil {
		return 0
	}
	return v.GetInt(key)
}

// GetDuration fetches a duration configuration value, initializing on demand.
func GetDuration(key string) time.Duration {
	v, err := getViper()
	if err != nil {
		return 0
	}
	return v.GetDuration(key)
}

// Set updates a configuration key at runtime, initializing on demand.
func Set(key string, value any) error {
	if err := Initialize(); err != nil {
		return err
	}
	configMu.Lock()
	defer configMu.Unlock()
	if configInst == nil {
		return fmt.Errorf("configuration not initialized")
	}
	configInst.Set(key, value)
	return nil
}

func configure(settings *initSettings) error {
	workingDir := strings.TrimSpace(settings.workingDir)
	if workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine working directory: %w", err)
		}
		workingDir = wd
	}

	userConfigPath := strings.TrimSpace(settings.userConfigPath)
	if userConfigPath == "" {
		path, err := defaultUserConfigPath()
		if err != nil {
			return err
		}
		userConfigPath = path
	}

	projectConfigPath := strings.TrimSpace(settings.projectConfigPath)
	if projectConfigPath == "" {
		path, err := findProjectConfig(workingDir)
		if err != nil {
			return err
		}
		projectConfigPath = path
	}

	v := viper.New()
	v.SetConfigType("yaml")
	setDefaults(v)
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := mergeConfigFile(v, userConfigPath); err != nil {
		return fmt.Errorf("load user config: %w", err)
	}
	if err := mergeConfigFile(v, projectConfigPath); err != nil {
		return fmt.Errorf("load project config: %w", err)
	}
	applyLegacyAutoRefreshConfig(v)

	configMu.Lock()
	configInst = v
	configMu.Unlock()
	return nil
}

func mergeConfigFile(v *viper.Viper, path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("config path %s is a directory", path)
	}
	//nolint:gosec // G304: Config loader intentionally reads user and project config files
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := v.MergeConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func defaultUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine user home: %w", err)
	}
	return filepath.Join(home, ".abacus", "config.yaml"), nil
}

func findProjectConfig(startDir string) (string, error) {
	if strings.TrimSpace(startDir) == "" {
		return "", nil
	}
	dir := startDir
	for {
		candidate := filepath.Join(dir, ".abacus", "config.yaml")
		info, err := os.Stat(candidate)
		if err == nil {
			if info.IsDir() {
				return "", fmt.Errorf("config path %s is a directory", candidate)
			}
			return candidate, nil
		}
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("stat %s: %w", candidate, err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault(KeyDatabasePath, "")
	v.SetDefault(KeySkipVersionCheck, false)
	v.SetDefault(KeyOutputFormat, "rich")
	v.SetDefault(KeyOutputJSON, false)
	v.SetDefault(KeyAutoRefreshSeconds, defaultAutoRefreshSeconds)
}

func getViper() (*viper.Viper, error) {
	if err := Initialize(); err != nil {
		return nil, err
	}
	configMu.RLock()
	defer configMu.RUnlock()
	if configInst == nil {
		return nil, fmt.Errorf("configuration not initialized")
	}
	return configInst, nil
}

// reset clears package state for tests.
//
//nolint:unused // Used in config_test.go
func reset() {
	configMu.Lock()
	defer configMu.Unlock()
	configInst = nil
	initErr = nil
	configOnce = sync.Once{}
}

func applyLegacyAutoRefreshConfig(v *viper.Viper) {
	if v == nil {
		return
	}
	if hasExplicitAutoRefreshSeconds(v) {
		return
	}
	if v.IsSet(KeyNoAutoRefresh) && v.GetBool(KeyNoAutoRefresh) {
		v.Set(KeyAutoRefreshSeconds, 0)
		return
	}
	if v.IsSet(KeyAutoRefresh) && !v.GetBool(KeyAutoRefresh) {
		v.Set(KeyAutoRefreshSeconds, 0)
		return
	}
	if v.IsSet(KeyRefreshInterval) {
		seconds := durationToSeconds(v.GetDuration(KeyRefreshInterval))
		v.Set(KeyAutoRefreshSeconds, seconds)
	}
}

func hasExplicitAutoRefreshSeconds(v *viper.Viper) bool {
	if v.InConfig(KeyAutoRefreshSeconds) {
		return true
	}
	envKey := autoRefreshSecondsEnvKey()
	if _, ok := os.LookupEnv(envKey); ok {
		return true
	}
	return false
}

func autoRefreshSecondsEnvKey() string {
	replacer := strings.NewReplacer(".", "_", "-", "_")
	return strings.ToUpper(envPrefix) + "_" + strings.ToUpper(replacer.Replace(KeyAutoRefreshSeconds))
}

func durationToSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	seconds := int(d / time.Second)
	if d%time.Second != 0 {
		seconds++
	}
	if seconds <= 0 {
		seconds = 1
	}
	return seconds
}
