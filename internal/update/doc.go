// Package update provides version checking and auto-update functionality.
//
// This package handles:
//   - Querying GitHub API for latest release information
//   - Comparing semantic versions to detect available updates
//   - Detecting installation method (Homebrew vs direct binary)
//   - Downloading and installing updates atomically
//
// The package is designed to be isolated from UI concerns. It returns
// structured data (UpdateInfo, VersionInfo) that the UI can present
// however it wants.
//
// Example usage:
//
//	checker := update.NewChecker(update.DefaultRepoOwner, update.DefaultRepoName)
//	info, err := checker.Check(ctx, currentVersion)
//	if err != nil {
//	    // handle error
//	}
//	if info.UpdateAvailable {
//	    // prompt user or auto-update
//	}
package update
