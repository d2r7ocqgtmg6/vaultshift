package vault

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Restorer restores secrets from a previously captured snapshot.
type Restorer struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
}

// NewRestorer creates a new Restorer.
func NewRestorer(client *Client, logger *audit.Logger, dryRun bool) (*Restorer, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Restorer{client: client, logger: logger, dryRun: dryRun}, nil
}

// RestoreResult holds the outcome of a restore operation.
type RestoreResult struct {
	Restored []string
	Skipped  []string
	Errors   []string
}

// Restore writes all secrets from the snapshot to the destination prefix.
func (r *Restorer) Restore(ctx context.Context, snap *Snapshot, destPrefix string) (*RestoreResult, error) {
	result := &RestoreResult{}

	for path, data := range snap.Secrets {
		destPath := destPrefix + "/" + path
		if r.dryRun {
			r.logger.Log("restore", map[string]any{"path": destPath, "dry_run": true})
			resresult.Skipped, destPath)
			continue
		}
		if err := r.client.WriteSecret(ctx, destPath, data); err != nil {
			errMsg := fmt.Sprintf("%s: %v", destPath, err)
			r.logger.Log("restore_error", map[string]any{"path": destPath, "error": err.Error()})
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		r.logger.Log("restore", map[string]any{"path": destPath})
		result.Restored = append(result.Restored, destPath)
	}

	return result, nil
}
