package vault

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// RollbackRecord captures a single secret write that can be undone.
type RollbackRecord struct {
	Path  string
	Mount string
}

// Rollbacker holds the state needed to undo a migration run.
type Rollbacker struct {
	client  *Client
	records []RollbackRecord
	logger  *audit.Logger
}

// NewRollbacker creates a Rollbacker backed by the given client.
func NewRollbacker(c *Client, l *audit.Logger) *Rollbacker {
	return &Rollbacker{client: c, logger: l}
}

// Record registers a path that was successfully written during migration.
func (r *Rollbacker) Record(mount, path string) {
	r.records = append(r.records, RollbackRecord{Mount: mount, Path: path})
}

// Len returns the number of recorded writes.
func (r *Rollbacker) Len() int { return len(r.records) }

// Rollback deletes every recorded secret in reverse order.
// It returns a slice of errors encountered; a non-nil slice does not stop
// processing — all paths are attempted.
func (r *Rollbacker) Rollback(ctx context.Context) []error {
	var errs []error
	for i := len(r.records) - 1; i >= 0; i-- {
		rec := r.records[i]
		fullPath := fmt.Sprintf("%s/data/%s", rec.Mount, rec.Path)
		_, err := r.client.raw.Logical().DeleteWithContext(ctx, fullPath)
		if err != nil {
			r.logger.Log("rollback_error", map[string]interface{}{
				"path":  fullPath,
				"error": err.Error(),
			})
			errs = append(errs, fmt.Errorf("rollback delete %s: %w", fullPath, err))
			continue
		}
		r.logger.Log("rollback_ok", map[string]interface{}{"path": fullPath})
	}
	return errs
}
