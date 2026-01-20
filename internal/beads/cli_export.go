// Package beads provides client implementations for beads issue tracking.
//
// DEPRECATED: This file is kept for backward compatibility only.
// The Export method is not implemented for CLI clients - use SQLite client for reads.
package beads

// This file exists for backward compatibility.
// CLI clients (bdCLIClient, brCLIClient) implement Writer only.
// The Export method is provided by SQLite clients or the cliClientWrapper.
