package beads

import (
	"context"
	"fmt"
)

// Export is not implemented for CLI client - use SQLite client for reads.
// In production, all read operations go through SQLite directly.
func (c *cliClient) Export(_ context.Context) ([]FullIssue, error) {
	return nil, fmt.Errorf("Export not implemented: CLI client only supports write operations; use SQLite client for reads")
}
