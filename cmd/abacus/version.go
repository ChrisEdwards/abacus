package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Version information - injected at build time via ldflags
var (
	Version   = "dev"
	Build     = "unknown"
	BuildTime = ""
)

// printVersion prints the version information
func printVersion() {
	fmt.Printf("abacus version %s", Version)

	if Build != "unknown" && Build != "" {
		fmt.Printf(" (build: %s)", Build)
	}

	if BuildTime != "" {
		fmt.Printf(" [%s]", BuildTime)
	}

	fmt.Println()

	// Add Go version and platform info
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Try to get build info for development builds
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" && len(setting.Value) > 7 {
					fmt.Printf("Commit: %s\n", setting.Value[:7])
					break
				}
			}
		}
	}
}
