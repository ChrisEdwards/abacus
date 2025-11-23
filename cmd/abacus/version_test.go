package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		build         string
		buildTime     string
		expectContain []string
	}{
		{
			name:          "dev build",
			version:       "dev",
			build:         "unknown",
			buildTime:     "",
			expectContain: []string{"abacus version dev", "Go version:", "OS/Arch:"},
		},
		{
			name:          "release build with commit",
			version:       "0.1.0",
			build:         "abc1234",
			buildTime:     "2025-11-22_12:00:00",
			expectContain: []string{"abacus version 0.1.0", "(build: abc1234)", "[2025-11-22_12:00:00]", "Go version:", "OS/Arch:"},
		},
		{
			name:          "release build without buildtime",
			version:       "1.0.0",
			build:         "def5678",
			buildTime:     "",
			expectContain: []string{"abacus version 1.0.0", "(build: def5678)", "Go version:", "OS/Arch:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersion := Version
			origBuild := Build
			origBuildTime := BuildTime
			defer func() {
				Version = origVersion
				Build = origBuild
				BuildTime = origBuildTime
			}()

			// Set test values
			Version = tt.version
			Build = tt.build
			BuildTime = tt.buildTime

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printVersion()

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check that all expected strings are present
			for _, expected := range tt.expectContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestVersionVariablesDefaults(t *testing.T) {
	// Test that default values are reasonable
	if Version != "dev" && Version == "" {
		t.Error("Version should have a default value")
	}

	if Build != "unknown" && Build == "" {
		t.Error("Build should have a default value")
	}
}
