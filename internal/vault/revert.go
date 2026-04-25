package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Reverter restores secrets at given paths to a previous snapshot state.
type Reverter struct {
	src    *Client
	logger *audit.Logger
	dryRun bool
}

// RevertResult holds the outcome of a single revert operation.
type RevertResult struct {
	Path    string
	Reverted bool
	Skipped  bool
	Err     error
}

// NewReverter constructs a Reverter.
func NewReverter(src *Client, logger *audit.Logger, dryRun bool) (*Reverter, error) {
	if src == nil {
		return nil, fmt.Errorf("revert: source client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("revert: logger is required")
	}
	return &Reverter{src: src, logger: logger, dryRun: dryRun}, nil
}

// Revert restores each path in the snapshot to the source Vault.
func (r *Reverter) Revert(snapshot map[string]map[string]interface{}) []RevertResult {
	results := make([]RevertResult, 0, len(snapshot))
	for path, data := range snapshot {
		res := RevertResult{Path: path}
		if r.dryRun {
			res.Skipped = true
			r.logger.Log("revert", map[string]interface{}{
				"path":    path,
				"dry_run": true,
			})
			results = append(results, res)
			continue
		}
		if err := r.src.WriteSecret(path, data); err != nil {
			res.Err = fmt.Errorf("revert %s: %w", path, err)
			r.logger.Log("revert_error", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
		} else {
			res.Reverted = true
			r.logger.Log("reverted", map[string]interface{}{"path": path})
		}
		results = append(results, res)
	}
	return results
}
