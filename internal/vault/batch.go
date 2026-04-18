package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// BatchResult holds the outcome of a single secret operation in a batch.
type BatchResult struct {
	Path    string
	Success bool
	Error   error
}

// Batcher runs bulk read/write operations across a set of paths.
type Batcher struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewBatcher creates a Batcher. Returns an error if client or logger is nil.
func NewBatcher(client *Client, logger *audit.Logger, dryRun bool) (*Batcher, error) {
	if client == nil {
		return nil, fmt.Errorf("batcher: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("batcher: logger is required")
	}
	return &Batcher{client: client, logger: logger, dryRun: dryRun}, nil
}

// WriteAll writes data from the provided map of path→secret data.
// In dry-run mode no writes are performed.
func (b *Batcher) WriteAll(secrets map[string]map[string]interface{}) []BatchResult {
	results := make([]BatchResult, 0, len(secrets))
	for path, data := range secrets {
		if b.dryRun {
			b.logger.Log("batch_write", map[string]interface{}{"path": path, "dry_run": true})
			results = append(results, BatchResult{Path: path, Success: true})
			continue
		}
		if err := b.client.WriteSecret(path, data); err != nil {
			b.logger.Log("batch_write_error", map[string]interface{}{"path": path, "error": err.Error()})
			results = append(results, BatchResult{Path: path, Success: false, Error: err})
			continue
		}
		b.logger.Log("batch_write", map[string]interface{}{"path": path, "dry_run": false})
		results = append(results, BatchResult{Path: path, Success: true})
	}
	return results
}

// DeleteAll removes secrets at the given paths.
// In dry-run mode no deletes are performed.
func (b *Batcher) DeleteAll(paths []string) []BatchResult {
	results := make([]BatchResult, 0, len(paths))
	for _, path := range paths {
		if b.dryRun {
			b.logger.Log("batch_delete", map[string]interface{}{"path": path, "dry_run": true})
			results = append(results, BatchResult{Path: path, Success: true})
			continue
		}
		if err := b.client.DeleteSecret(path); err != nil {
			b.logger.Log("batch_delete_error", map[string]interface{}{"path": path, "error": err.Error()})
			results = append(results, BatchResult{Path: path, Success: false, Error: err})
			continue
		}
		b.logger.Log("batch_delete", map[string]interface{}{"path": path, "dry_run": false})
		results = append(results, BatchResult{Path: path, Success: true})
	}
	return results
}
